package withdraw

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-processing/pkg/avalidator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
)

var (
	ErrWithdrawalInvalidAddress      = errors.New("invalid destination address for the blockchain")
	ErrWithdrawalInvalidAmount       = errors.New("amount must be greater than zero")
	ErrWithdrawalTwoFactorRequired   = errors.New("two-factor authentication must be enabled and confirmed for withdrawals to external addresses")
	ErrWithdrawalTwoFactorInvalid    = errors.New("invalid two-factor authentication code")
	ErrWithdrawalInsufficientBalance = errors.New("not enough funds on the wallet address for the withdrawal and its network fee")
	ErrWithdrawalAddressIsTaken      = errors.New("the wallet address has another transfer in progress")
)

// ExternalWithdrawalDTO is a withdrawal from a hot wallet address to any external address
type ExternalWithdrawalDTO struct {
	WalletAddressID uuid.UUID
	AddressTo       string
	Amount          decimal.Decimal
	TOTP            string
}

// WithdrawToExternal sends funds from the user's hot wallet address to an external address.
//
// The owner's TOTP is verified by processing (purpose withdrawal), the address must not belong
// to the owner's wallets: transfers to cold and processing wallets use the regular withdrawals.
// The merchant transfer is created when processing accepted the transfer or its result is unknown
// (then the transfer keeps the message and is updated by the status webhooks).
func (s *service) WithdrawToExternal(ctx context.Context, user *models.User, dto ExternalWithdrawalDTO) (*models.Transfer, error) {
	if !user.ProcessingOwnerID.Valid {
		return nil, ErrProcessingUninitialized
	}

	if !dto.Amount.IsPositive() {
		return nil, ErrWithdrawalInvalidAmount
	}

	walletAddress, err := s.storage.WalletAddresses().GetById(ctx, dto.WalletAddressID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWalletIsNotOwnedByUser
	}
	if err != nil {
		return nil, fmt.Errorf("fetch wallet address: %w", err)
	}

	if walletAddress.UserID != user.ID {
		return nil, ErrWalletIsNotOwnedByUser
	}

	curr, err := s.currencyService.GetCurrencyByID(ctx, walletAddress.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("fetch currency: %w", err)
	}
	if curr.Blockchain == nil {
		return nil, ErrFiatCurrencyIsNotSupported
	}

	if !avalidator.ValidateAddressByBlockchain(dto.AddressTo, walletAddress.Blockchain.String()) {
		return nil, ErrWithdrawalInvalidAddress
	}

	rate, err := s.currencyRate(ctx, user.RateSource.String(), curr)
	if err != nil {
		return nil, err
	}

	params := processing.FundsWithdrawalParams{
		OwnerID:         user.ProcessingOwnerID.UUID,
		RequestID:       uuid.New(),
		Blockchain:      walletAddress.Blockchain,
		FromAddress:     []string{walletAddress.Address},
		ToAddress:       []string{dto.AddressTo},
		ContractAddress: curr.ContractAddress.String,
		// processing does not allow the whole amount for withdrawals to external addresses
		WholeAmount: false,
		Amount:      dto.Amount.String(),
		Purpose:     processing.TransferPurposeWithdrawal,
		TOTP:        &dto.TOTP,
	}

	s.applyTronTransferKind(ctx, user, &params)

	var message *string
	if _, err := s.processing.FundsWithdrawal(ctx, params); err != nil {
		unknown, err := s.resolveExternalWithdrawalError(ctx, params.RequestID, err)
		if err != nil {
			return nil, err
		}
		if unknown != "" {
			message = &unknown
		}
	}

	transfer, err := s.storage.Transfers().Create(ctx, repo_transfers.CreateParams{
		ID:            params.RequestID,
		UserID:        user.ID,
		Kind:          models.TransferKindFromAddress,
		CurrencyID:    curr.ID,
		Status:        models.TransferStatusNew,
		Stage:         models.ResolveTransferStageByStatus(models.TransferStatusNew),
		FromAddresses: params.FromAddress,
		ToAddresses:   params.ToAddress,
		Amount:        dto.Amount,
		AmountUsd:     rate.Mul(dto.Amount),
		Blockchain:    walletAddress.Blockchain,
		Message:       message,
	})
	if err != nil {
		s.logger.Errorw("external withdrawal is created in processing, but the transfer is not saved",
			"transfer_id", params.RequestID,
			"error", err,
		)
		return nil, fmt.Errorf("transfer creation: %w", err)
	}

	return transfer, nil
}

// applyTronTransferKind sets the user's tron transfer kind, processing requires it for tron withdrawals
func (s *service) applyTronTransferKind(ctx context.Context, user *models.User, params *processing.FundsWithdrawalParams) {
	if !params.Blockchain.KindWithdrawalRequired() {
		return
	}

	tronTransferType, err := s.settings.GetModelSetting(ctx, setting.TransferType, setting.IModelSetting(user))
	if err != nil || tronTransferType == nil {
		params.Kind = new(string(setting.TransferByBurnTRX))
		return
	}

	params.Kind = new(tronTransferType.Value)
}

// resolveExternalWithdrawalError returns the error converted for the admin when processing certainly
// did not create the transfer. When processing did not answer, the transfer is looked up by the request id:
// if it exists or the lookup fails too, nil error is returned and the transfer is saved (with the returned
// message when the result is unknown), so status webhooks find it and the withdrawal is not repeated by the admin.
func (s *service) resolveExternalWithdrawalError(ctx context.Context, requestID uuid.UUID, err error) (string, error) {
	if code, ok := processing.ErrorRPCCode(err); ok {
		switch code {
		case processing.RPCCodeTwoFactorRequired:
			return "", fmt.Errorf("%w: %w", ErrWithdrawalTwoFactorRequired, err)
		case processing.RPCCodeTwoFactorInvalid:
			return "", fmt.Errorf("%w: %w", ErrWithdrawalTwoFactorInvalid, err)
		case processing.RPCCodeAddressEmptyBalance, CodeStatusNotEnoughBalance:
			return "", fmt.Errorf("%w: %w", ErrWithdrawalInsufficientBalance, err)
		case CodeStatusAddressIsTaken:
			return "", fmt.Errorf("%w: %w", ErrWithdrawalAddressIsTaken, err)
		case CodeStatusBlockchainIsDisabled:
			return "", fmt.Errorf("%w: %w", ErrProcessingExplorerUnavailable, err)
		default:
			return "", fmt.Errorf("create withdrawal in processing: %w", err)
		}
	}

	if !isAmbiguousProcessingError(err) {
		return "", fmt.Errorf("create withdrawal in processing: %w", err)
	}

	_, lookupErr := s.processing.GetTransferByRequestID(ctx, requestID)
	switch {
	case lookupErr == nil:
		s.logger.Warnw("processing returned an error, but the withdrawal transfer exists",
			"transfer_id", requestID,
			"error", err,
		)
		return "", nil
	case errors.Is(lookupErr, processing.ErrTransferNotFound):
		if connect.CodeOf(err) == connect.CodeUnavailable {
			return "", ErrProcessingUnavailable
		}
		return "", fmt.Errorf("create withdrawal in processing: %w", err)
	default:
		s.logger.Errorw("external withdrawal result in processing is unknown",
			"transfer_id", requestID,
			"error", err,
			"lookup_error", lookupErr,
		)
		return "withdrawal result in processing is unknown: " + err.Error(), nil
	}
}

// isAmbiguousProcessingError reports whether processing may have created the transfer despite the error
func isAmbiguousProcessingError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return false
	}

	switch connectErr.Code() {
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded, connect.CodeCanceled, connect.CodeUnknown, connect.CodeAborted:
		return true
	default:
		return false
	}
}
