package refund

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-processing/pkg/avalidator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type IRefundService interface { //nolint:interfacebloat
	CreateRefund(ctx context.Context, dto CreateRefundDTO) (*models.RefundRequest, error)
	EstimateRefund(ctx context.Context, dto EstimateRefundDTO) (*EstimateDTO, error)
	CreateRefundToSender(ctx context.Context, dto ToSenderDTO) (*models.RefundRequest, error)
	EstimateRefundToSender(ctx context.Context, dto ToSenderDTO) (*EstimateDTO, error)
	GetCabinet(ctx context.Context, walletID uuid.UUID) (map[string][]*CabinetItem, error)
	GetUnclaimed(ctx context.Context, walletID uuid.UUID) ([]*CabinetItem, error)
	GetPendingReviewByUser(ctx context.Context, userID uuid.UUID) ([]*RequestWithTxDTO, error)
	FindByUser(ctx context.Context, dto FindDTO) (*storecmn.FindResponseWithFullPagination[*repo_refund_requests.FindRow], error)
	RejectRefund(ctx context.Context, dto RejectRefundDTO) (*RequestWithTxDTO, error)
	ApproveRefund(ctx context.Context, dto ApproveRefundDTO) (*RequestWithTxDTO, error)
	CancelRefund(ctx context.Context, dto CancelRefundDTO) (*RequestWithTxDTO, error)
	ExecuteRefundNow(ctx context.Context, dto ExecuteRefundNowDTO) (*RequestWithTxDTO, error)
}

var (
	ErrRefundAlreadyRequested  = errors.New("a refund request already exists for this transaction")
	ErrRefundNotOwned          = errors.New("refund request does not belong to the user")
	ErrRefundNotPendingReview  = errors.New("refund request is not pending review")
	ErrRefundTwoFactorRequired = errors.New("two-factor authentication must be enabled and confirmed to approve refunds")
	ErrRefundTwoFactorInvalid  = errors.New("invalid two-factor authentication code")
	ErrRefundInsufficientFee   = errors.New("not enough native coin on the deposit address to pay the refund network fee")
	ErrRefundInsufficientFunds = errors.New("not enough funds on the deposit address to refund the deposit")
	ErrRefundSourceInvalid     = errors.New("the deposit cannot be refunded: it is not confirmed or does not match the refund")
	ErrRefundAlreadyExists     = errors.New("a refund for this deposit is already in progress")
	ErrRefundNotCancelable     = errors.New("refund cannot be canceled: it is not scheduled or is already being sent")
	ErrRefundNotScheduled      = errors.New("refund is not scheduled or is already being sent")
	ErrRefundBlockedTxNotFound = errors.New("blocked transaction not found")
	ErrRefundInvalidAddress    = errors.New("invalid destination address for the blockchain")
	ErrRefundAuthRequired      = errors.New("email verification is required to request a refund")
	ErrRefundNotSender         = errors.New("the address is not a sender of the deposit, request a refund to it with the email verification")
)

// refundExecutor sends refunds to processing
type refundExecutor interface {
	CheckRefund(ctx context.Context, ref *models.RefundRequest) (*processing.CheckRefundResult, error)
	CreateRefundTransfer(ctx context.Context, ref *models.RefundRequest, opts ...repos.Option) (*models.Transfer, error)
	SendRefund(ctx context.Context, ref *models.RefundRequest, transferID uuid.UUID, totp *string) (*processing.FundsWithdrawalResult, error)
}

// userSettings reads the store owner settings
type userSettings interface {
	GetModelSetting(ctx context.Context, name string, model setting.IModelSetting) (*models.Setting, error)
}

type Service struct {
	logger     logger.Logger
	storage    storage.IStorage
	executor   refundExecutor
	processing processing.IProcessingTransfer
	settings   userSettings
}

func NewService(
	log logger.Logger,
	st storage.IStorage,
	executor refundExecutor,
	processingTransfer processing.IProcessingTransfer,
	settings userSettings,
) *Service {
	return &Service{
		logger:     log,
		storage:    st,
		executor:   executor,
		processing: processingTransfer,
		settings:   settings,
	}
}

// RefundToSenderWithoutAuth reports whether the user (the store owner) allows the payer to request
// a refund back to the deposit sender without the email verification (setting refund_auth_mode).
// The verification is required when the setting is not set.
func (s *Service) RefundToSenderWithoutAuth(ctx context.Context, userID uuid.UUID) (bool, error) {
	res, err := s.settings.GetModelSetting(ctx, setting.RefundAuthMode, &models.User{ID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get refund auth mode setting: %w", err)
	}

	return res.Value == setting.RefundAuthModeCustomAddressOnly, nil
}

// checkClaim runs the checks of a refund claim without creating anything: the blocked transaction
// belongs to the wallet, it is not claimed yet, the destination address is valid and processing
// can send the refund (the deposit is valid and the network fee is covered).
//
// A claim without the email verification (toSender) is allowed only by the owner setting, and the
// payer's address must be one of the deposit senders. Processing determines the senders from the
// blockchain (every input of a btc like deposit) and confirms the address (ToSender), so such claim
// can never send the funds to an address which did not send the deposit.
func (s *Service) checkClaim(ctx context.Context, walletID, blockedTransactionID uuid.UUID, destinationAddress string, toSender bool) (*claimCheck, error) {
	btx, err := s.storage.BlockedTransactions().GetById(ctx, blockedTransactionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRefundBlockedTxNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transaction: %w", err)
	}

	// a transaction of another wallet is reported as not found, its existence is not disclosed
	if btx.WalletID != walletID {
		return nil, fmt.Errorf("%w: wallet id mismatch", ErrRefundBlockedTxNotFound)
	}

	// the user of the blocked transaction is the store owner
	if toSender {
		allowed, err := s.RefundToSenderWithoutAuth(ctx, btx.UserID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrRefundAuthRequired
		}
	}

	if _, err := s.storage.RefundRequests().GetByBlockedTransactionID(ctx, blockedTransactionID); err == nil {
		return nil, ErrRefundAlreadyRequested
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check existing refund request: %w", err)
	}

	tx, err := s.storage.Transactions().GetById(ctx, btx.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch transaction: %w", err)
	}

	if !avalidator.ValidateAddressByBlockchain(destinationAddress, tx.Blockchain.String()) {
		return nil, fmt.Errorf("%w: %s", ErrRefundInvalidAddress, tx.Blockchain.String())
	}

	// the refund must be sendable before the request is created: processing checks the deposit
	// and that the deposit address holds enough native coin to pay the network fee
	check, err := s.executor.CheckRefund(ctx, &models.RefundRequest{
		BlockedTransactionID: blockedTransactionID,
		WalletID:             walletID,
		StoreID:              btx.StoreID,
		DestinationAddress:   destinationAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("check refund: %w", convertProcessingError(err))
	}

	if toSender && !check.ToSender {
		return nil, fmt.Errorf("%w: %s", ErrRefundNotSender, destinationAddress)
	}

	return &claimCheck{btx: btx, tx: tx, result: check, destinationAddress: destinationAddress}, nil
}

// EstimateRefund returns the network fee and the amount the payer receives for a refund
// to the destination address. Nothing is created, the fee may change by the time of sending.
func (s *Service) EstimateRefund(ctx context.Context, dto EstimateRefundDTO) (*EstimateDTO, error) {
	checked, err := s.checkClaim(ctx, dto.WalletID, dto.BlockedTransactionID, dto.DestinationAddress, false)
	if err != nil {
		return nil, err
	}

	return s.buildEstimate(checked)
}

// EstimateRefundToSender estimates a refund to a deposit sender requested without the email verification
func (s *Service) EstimateRefundToSender(ctx context.Context, dto ToSenderDTO) (*EstimateDTO, error) {
	checked, err := s.checkClaim(ctx, dto.WalletID, dto.BlockedTransactionID, dto.DestinationAddress, true)
	if err != nil {
		return nil, err
	}

	return s.buildEstimate(checked)
}

func (s *Service) buildEstimate(checked *claimCheck) (*EstimateDTO, error) {
	feeCurrencyID, err := checked.tx.Blockchain.NativeCurrency()
	if err != nil {
		return nil, fmt.Errorf("resolve native currency: %w", err)
	}

	res := &EstimateDTO{
		BlockedTransactionID: checked.btx.ID,
		DestinationAddress:   checked.destinationAddress,
		Amount:               checked.tx.Amount,
		CurrencyID:           checked.tx.CurrencyID,
		FeeCurrencyID:        feeCurrencyID,
		FeeFromAmount:        checked.result.FeeFromAmount,
		FeeCovered:           checked.result.FeeCovered(),
		FeeBalance:           checked.result.FeeBalance,
		FeeShortage:          checked.result.FeeShortage,
		ActivationRequired:   checked.result.ActivationRequired,
		RequiresReview:       !checked.result.ToSender,
	}

	if checked.result.EstimatedFee != nil {
		fee, err := decimal.NewFromString(*checked.result.EstimatedFee)
		if err != nil {
			return nil, fmt.Errorf("parse estimated fee %q: %w", *checked.result.EstimatedFee, err)
		}
		res.EstimatedFee = decimal.NewNullDecimal(fee)
	}

	// the received amount is unknown when the fee is taken from the amount but not estimated (btc like)
	switch {
	case !res.FeeFromAmount:
		res.AmountToReceive = decimal.NewNullDecimal(res.Amount)
	case res.EstimatedFee.Valid:
		res.AmountToReceive = decimal.NewNullDecimal(decimal.Max(res.Amount.Sub(res.EstimatedFee.Decimal), decimal.Zero))
	}

	return res, nil
}

// CreateRefund creates a refund request of the payer verified by the wallet email
func (s *Service) CreateRefund(ctx context.Context, dto CreateRefundDTO) (*models.RefundRequest, error) {
	checked, err := s.checkClaim(ctx, dto.WalletID, dto.BlockedTransactionID, dto.DestinationAddress, false)
	if err != nil {
		return nil, err
	}

	return s.createRefund(ctx, checked, pgtype.Text{String: dto.Email, Valid: dto.Email != ""}, constants.RefundAuthMethodEmail)
}

// CreateRefundToSender creates a refund request to a deposit sender without the email
// verification, when the store owner allows it (setting refund_auth_mode)
func (s *Service) CreateRefundToSender(ctx context.Context, dto ToSenderDTO) (*models.RefundRequest, error) {
	checked, err := s.checkClaim(ctx, dto.WalletID, dto.BlockedTransactionID, dto.DestinationAddress, true)
	if err != nil {
		return nil, err
	}

	return s.createRefund(ctx, checked, pgtype.Text{}, constants.RefundAuthMethodNone)
}

func (s *Service) createRefund(ctx context.Context, checked *claimCheck, email pgtype.Text, authMethod string) (*models.RefundRequest, error) {
	btx, check := checked.btx, checked.result

	// processing reports a not covered fee of a token refund in the check result, not as an error:
	// such refund cannot be sent until the deposit address is topped up
	if !check.FeeCovered() {
		return nil, fmt.Errorf("%w: fee shortage %s, activation required %t",
			ErrRefundInsufficientFee, check.FeeShortage.Decimal, check.ActivationRequired)
	}

	ref, err := s.storage.RefundRequests().Create(ctx, repo_refund_requests.CreateParams{
		BlockedTransactionID: btx.ID,
		WalletID:             btx.WalletID,
		StoreID:              btx.StoreID,
		DestinationAddress:   checked.destinationAddress,
		Email:                email,
		Status:               constants.RefundStatusPendingReview,
		AuthMethod:           authMethod,
	})
	if err != nil {
		// a concurrent claim of the same blocked transaction (refund_requests_blocked_transaction_id_key)
		if _, ok := errors.AsType[*pgerror.UniqueConstraintError](pgerror.ParseError(err)); ok {
			return nil, ErrRefundAlreadyRequested
		}
		return nil, fmt.Errorf("create refund request: %w", err)
	}

	// a refund to a new address is sent only after the manual review.
	// Processing determines the deposit senders from the blockchain
	if !check.ToSender {
		return ref, nil
	}

	scheduled, err := s.sendRefund(ctx, ref.ID, nil, false)
	if err != nil {
		// the request stays pending review with the reason (the admin can approve it manually)
		// or in processing when the result is unknown (resolved by the reconciliation)
		s.logger.Errorw("failed to schedule refund to the deposit sender",
			"refund_request_id", ref.ID,
			"error", err,
		)

		current, getErr := s.storage.RefundRequests().GetById(ctx, ref.ID)
		if getErr != nil {
			s.logger.Errorw("failed to fetch refund request", "refund_request_id", ref.ID, "error", getErr)
			return ref, nil
		}

		return current, nil
	}

	return scheduled, nil
}

// sendRefund claims a pending review request and sends the refund to processing.
//
// Without TOTP processing accepts only refunds back to the deposit sender and delays them
// by the cooldown (status scheduled). With TOTP the refund is sent immediately (status processing).
//
// The call to processing is never made inside a database transaction:
//  1. the request is claimed (pending_review -> processing) and its refund transfer is created
//     and attached in one committed transaction, so status webhooks always find the request;
//  2. the refund is sent to processing with the transfer id as the request id;
//  3. the result is saved with a conditional update, a status webhook may have already moved the request.
//
// When processing certainly did not create the refund, the request goes back to pending review.
// When the result is unknown (network error), the request stays in processing and is resolved by
// ReconcileInFlight: sending it again could create a second refund.
func (s *Service) sendRefund(ctx context.Context, refundRequestID uuid.UUID, totp *string, reviewed bool) (*models.RefundRequest, error) {
	claimed, transfer, err := s.claimForSending(ctx, refundRequestID, reviewed)
	if err != nil {
		return nil, err
	}

	transferID := uuid.NullUUID{UUID: transfer.ID, Valid: true}

	result, sendErr := s.executor.SendRefund(ctx, claimed, transfer.ID, totp)
	if sendErr == nil {
		return s.markSent(ctx, claimed.ID, transferID, result.ExecuteAfter)
	}

	sendErr = convertProcessingError(sendErr)

	// processing did not answer: the transfer may be created, it is resolved by the reconciliation
	if isAmbiguousProcessingError(sendErr) {
		s.logger.Warnw("refund result in processing is unknown, the request is left for reconciliation",
			"refund_request_id", claimed.ID,
			"transfer_id", transfer.ID,
			"error", sendErr,
		)

		if setErr := s.storage.RefundRequests().SetLastError(ctx, repo_refund_requests.SetLastErrorParams{
			ID:        claimed.ID,
			LastError: pgtype.Text{String: TruncateRefundError("refund result in processing is unknown: " + sendErr.Error()), Valid: true},
		}); setErr != nil {
			s.logger.Errorw("failed to save refund request error", "refund_request_id", claimed.ID, "error", setErr)
		}

		return nil, fmt.Errorf("send refund: %w", sendErr)
	}

	// processing answered with an error: make sure the transfer was not created anyway
	info, lookupErr := s.processing.GetTransferByRequestID(ctx, transfer.ID)
	switch {
	case lookupErr == nil:
		s.logger.Warnw("processing returned an error, but the refund transfer exists",
			"refund_request_id", claimed.ID,
			"transfer_id", transfer.ID,
			"error", sendErr,
		)
		return s.markSent(ctx, claimed.ID, transferID, info.ExecuteAfter)
	case !errors.Is(lookupErr, processing.ErrTransferNotFound):
		s.logger.Warnw("failed to check the refund transfer in processing, the request is left for reconciliation",
			"refund_request_id", claimed.ID,
			"transfer_id", transfer.ID,
			"error", sendErr,
			"lookup_error", lookupErr,
		)
		return nil, fmt.Errorf("send refund: %w", sendErr)
	}

	if err := s.revertUnsent(ctx, claimed.ID, transfer.ID, sendErr.Error()); err != nil {
		s.logger.Errorw("failed to return refund request to pending review",
			"refund_request_id", claimed.ID,
			"transfer_id", transfer.ID,
			"error", err,
		)
	}

	return nil, fmt.Errorf("send refund: %w", sendErr)
}

// claimForSending moves the request pending_review -> processing and attaches a new refund transfer
func (s *Service) claimForSending(ctx context.Context, refundRequestID uuid.UUID, reviewed bool) (*models.RefundRequest, *models.Transfer, error) {
	var (
		claimed  *models.RefundRequest
		transfer *models.Transfer
	)

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		ref, err := s.storage.RefundRequests(repos.WithTx(tx)).GetById(ctx, refundRequestID)
		if err != nil {
			return fmt.Errorf("fetch refund request: %w", err)
		}

		if ref.Status != constants.RefundStatusPendingReview {
			return ErrRefundNotPendingReview
		}

		transfer, err = s.executor.CreateRefundTransfer(ctx, ref, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("create refund transfer: %w", err)
		}

		var reviewedAt pgtype.Timestamp
		if reviewed {
			reviewedAt = pgtypeutils.EncodeTime(time.Now())
		}

		// conditional update: the request may be approved or rejected at the same time
		claimed, err = s.storage.RefundRequests(repos.WithTx(tx)).ClaimForSending(ctx, repo_refund_requests.ClaimForSendingParams{
			ID:            refundRequestID,
			CurrentStatus: constants.RefundStatusPendingReview,
			NewStatus:     constants.RefundStatusProcessing,
			TransferID:    uuid.NullUUID{UUID: transfer.ID, Valid: true},
			ReviewedAt:    reviewedAt,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRefundNotPendingReview
		}
		if err != nil {
			return fmt.Errorf("claim refund request: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return claimed, transfer, nil
}

// markSent saves the result of a created refund: scheduled with the cooldown or processing.
// A status webhook may have already moved the request further, then the current state is returned.
func (s *Service) markSent(ctx context.Context, id uuid.UUID, transferID uuid.NullUUID, executeAfter *time.Time) (*models.RefundRequest, error) {
	status := constants.RefundStatusProcessing
	var executeAfterTS pgtype.Timestamp
	if executeAfter != nil {
		status = constants.RefundStatusScheduled
		executeAfterTS = pgtypeutils.EncodeTime(executeAfter.UTC())
	}

	updated, err := s.storage.RefundRequests().MarkSent(ctx, repo_refund_requests.MarkSentParams{
		ID:            id,
		TransferID:    transferID,
		CurrentStatus: constants.RefundStatusProcessing,
		NewStatus:     status,
		ExecuteAfter:  executeAfterTS,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return s.storage.RefundRequests().GetById(ctx, id)
	}
	if err != nil {
		return nil, fmt.Errorf("update sent refund request: %w", err)
	}

	return updated, nil
}

// revertUnsent returns the request to pending review when processing did not create the refund
// and fails its refund transfer
func (s *Service) revertUnsent(ctx context.Context, id, transferID uuid.UUID, reason string) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		rows, err := s.storage.RefundRequests(repos.WithTx(tx)).RevertUnsent(ctx, repo_refund_requests.RevertUnsentParams{
			ID:            id,
			TransferID:    uuid.NullUUID{UUID: transferID, Valid: true},
			CurrentStatus: constants.RefundStatusProcessing,
			NewStatus:     constants.RefundStatusPendingReview,
			LastError:     pgtype.Text{String: TruncateRefundError(reason), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("revert refund request: %w", err)
		}

		// the request was moved by a webhook or another reconciliation, the transfer is not touched
		if rows == 0 {
			return nil
		}

		message := TruncateRefundError(reason)
		if err := s.storage.Transfers(repos.WithTx(tx)).UpdateTransferStatus(ctx, repo_transfers.UpdateTransferStatusParams{
			ID:      transferID,
			Status:  models.TransferStatusFailed,
			Stage:   models.ResolveTransferStageByStatus(models.TransferStatusFailed),
			Message: &message,
		}); err != nil {
			return fmt.Errorf("fail refund transfer: %w", err)
		}

		return nil
	})
}

// isAmbiguousProcessingError reports whether processing may have handled the request despite the error:
// the request did not reach processing or its answer was lost
func isAmbiguousProcessingError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		// not an rpc error: the refund failed before the call to processing
		return false
	}

	switch connectErr.Code() {
	case connect.CodeUnavailable, connect.CodeDeadlineExceeded, connect.CodeCanceled, connect.CodeUnknown, connect.CodeAborted:
		return true
	default:
		return false
	}
}

// convertProcessingError maps processing rpc codes of refunds to errors shown to the admin
func convertProcessingError(err error) error {
	code, ok := processing.ErrorRPCCode(err)
	if !ok {
		return err
	}

	switch code {
	case processing.RPCCodeTwoFactorRequired:
		return fmt.Errorf("%w: %w", ErrRefundTwoFactorRequired, err)
	case processing.RPCCodeTwoFactorInvalid:
		return fmt.Errorf("%w: %w", ErrRefundTwoFactorInvalid, err)
	case processing.RPCCodeRefundInsufficientFee:
		return fmt.Errorf("%w: %w", ErrRefundInsufficientFee, err)
	case processing.RPCCodeAddressEmptyBalance:
		return fmt.Errorf("%w: %w", ErrRefundInsufficientFunds, err)
	case processing.RPCCodeRefundSourceInvalid:
		return fmt.Errorf("%w: %w", ErrRefundSourceInvalid, err)
	case processing.RPCCodeRefundAlreadyExists:
		return fmt.Errorf("%w: %w", ErrRefundAlreadyExists, err)
	default:
		return err
	}
}

func (s *Service) GetCabinet(ctx context.Context, walletID uuid.UUID) (map[string][]*CabinetItem, error) {
	blocked, err := s.storage.BlockedTransactions().GetAllWithTxByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transactions: %w", err)
	}

	refunds, err := s.storage.RefundRequests().GetAllByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("fetch refund requests: %w", err)
	}

	return buildCabinet(blocked, refunds), nil
}

func (s *Service) GetUnclaimed(ctx context.Context, walletID uuid.UUID) ([]*CabinetItem, error) {
	blocked, err := s.storage.BlockedTransactions().GetUnclaimedByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("fetch unclaimed blocked transactions: %w", err)
	}

	// the transactions of one wallet belong to one store owner, the setting is read once per owner
	withoutAuth := make(map[uuid.UUID]bool, 1)

	items := make([]*CabinetItem, 0, len(blocked))
	for _, b := range blocked {
		allowed, ok := withoutAuth[b.UserID]
		if !ok {
			// the list is still returned when the setting cannot be read, the email verification stays available
			if allowed, err = s.RefundToSenderWithoutAuth(ctx, b.UserID); err != nil {
				s.logger.Errorw("failed to get refund auth mode", "wallet_id", walletID, "error", err)
			}
			withoutAuth[b.UserID] = allowed
		}

		items = append(items, &CabinetItem{
			BlockedTransactionID: b.ID,
			TransactionID:        b.TransactionID,
			TxHash:               b.TxHash,
			Blockchain:           b.Blockchain,
			CurrencyID:           b.CurrencyID,
			CurrencyCode:         b.CurrencyCode,
			Amount:               b.Amount,
			AmountUsd:            b.AmountUsd,
			FromAddress:          b.FromAddress,
			ToAddress:            b.ToAddress,
			RiskLevel:            b.RiskLevel,
			Score:                b.Score,
			BlockReasons:         b.BlockReasons,
			CreatedAt:            b.CreatedAt,

			RefundToSenderWithoutAuth: allowed,
		})
	}
	return items, nil
}

// GetPendingReviewByUser returns refund requests which need the admin's attention:
// pending review (approve / reject), scheduled refunds to the deposit sender (cancel)
// and refunds failed after the broadcast (manual check of the transaction)
func (s *Service) GetPendingReviewByUser(ctx context.Context, userID uuid.UUID) ([]*RequestWithTxDTO, error) {
	var list []*repo_refund_requests.GetAllByUserIDAndStatusRow
	for _, status := range []string{constants.RefundStatusPendingReview, constants.RefundStatusScheduled, constants.RefundStatusFailed} {
		rows, err := s.storage.RefundRequests().GetAllByUserIDAndStatus(ctx, repo_refund_requests.GetAllByUserIDAndStatusParams{
			UserID: userID,
			Status: status,
		})
		if err != nil {
			return nil, fmt.Errorf("fetch %s refund requests: %w", status, err)
		}
		list = append(list, rows...)
	}

	// newest first, as returned by the query for a single status
	slices.SortFunc(list, func(a, b *repo_refund_requests.GetAllByUserIDAndStatusRow) int {
		return b.CreatedAt.Time.Compare(a.CreatedAt.Time)
	})

	items := make([]*RequestWithTxDTO, 0, len(list))
	for _, row := range list {
		items = append(items, &RequestWithTxDTO{
			RefundRequest: *refundRequestFromPendingRow(row),
			TransactionID: row.TransactionID,
			TxHash:        row.TxHash,
			Amount:        row.Amount,
			AmountUsd:     row.AmountUsd,
			CurrencyID:    row.CurrencyID,
			CurrencyCode:  row.CurrencyCode,
			Blockchain:    row.Blockchain,
			FromAddress:   row.FromAddress,
			ToAddress:     row.ToAddress,
			RiskLevel:     row.RiskLevel,
			Score:         row.Score,
			BlockReasons:  row.BlockReasons,
		})
	}
	return items, nil
}

func (s *Service) RejectRefund(ctx context.Context, dto RejectRefundDTO) (*RequestWithTxDTO, error) {
	if err := s.checkRefundOwner(ctx, dto.RefundRequestID, dto.UserID); err != nil {
		return nil, err
	}

	var updated *models.RefundRequest
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// conditional update: the request may be approved at the same time
		claimed, err := s.storage.RefundRequests(repos.WithTx(tx)).UpdateStatusIfCurrent(ctx, repo_refund_requests.UpdateStatusIfCurrentParams{
			ID:            dto.RefundRequestID,
			CurrentStatus: constants.RefundStatusPendingReview,
			NewStatus:     constants.RefundStatusRejected,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRefundNotPendingReview
		}
		if err != nil {
			return fmt.Errorf("reject refund request: %w", err)
		}

		updated, err = s.storage.RefundRequests(repos.WithTx(tx)).Update(ctx, repo_refund_requests.UpdateParams{
			ID:           claimed.ID,
			Status:       constants.RefundStatusRejected,
			TransferID:   claimed.TransferID,
			ReviewedAt:   pgtypeutils.EncodeTime(time.Now()),
			ExecuteAfter: claimed.ExecuteAfter,
			LastError:    claimed.LastError,
		})
		if err != nil {
			return fmt.Errorf("reject refund request: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.enrichRefundRequest(ctx, updated)
}

// ApproveRefund sends a pending review refund to processing. The owner's TOTP is required:
// with it processing sends the refund to any external address immediately.
func (s *Service) ApproveRefund(ctx context.Context, dto ApproveRefundDTO) (*RequestWithTxDTO, error) {
	if err := s.checkRefundOwner(ctx, dto.RefundRequestID, dto.UserID); err != nil {
		return nil, err
	}

	if dto.TOTP == "" {
		return nil, ErrRefundTwoFactorInvalid
	}

	updated, err := s.sendRefund(ctx, dto.RefundRequestID, &dto.TOTP, true)
	if err != nil {
		return nil, err
	}

	return s.enrichRefundRequest(ctx, updated)
}

// CancelRefund cancels a scheduled refund to the deposit sender before processing sends it.
// The request goes back to pending review: the admin can approve the refund manually or reject it.
func (s *Service) CancelRefund(ctx context.Context, dto CancelRefundDTO) (*RequestWithTxDTO, error) {
	if err := s.checkRefundOwner(ctx, dto.RefundRequestID, dto.UserID); err != nil {
		return nil, err
	}

	ref, err := s.storage.RefundRequests().GetById(ctx, dto.RefundRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetch refund request: %w", err)
	}

	if ref.Status != constants.RefundStatusScheduled || !ref.TransferID.Valid {
		return nil, ErrRefundNotCancelable
	}

	user, err := s.storage.Users().GetByID(ctx, dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	if !user.ProcessingOwnerID.Valid {
		return nil, fmt.Errorf("processing owner is not initialized")
	}

	if _, err := s.processing.CancelTransfer(ctx, user.ProcessingOwnerID.UUID, ref.TransferID.UUID); err != nil {
		if code, ok := processing.ErrorRPCCode(err); ok && code == processing.RPCCodeTransferNotCancelable {
			return nil, fmt.Errorf("%w: %w", ErrRefundNotCancelable, err)
		}
		return nil, fmt.Errorf("cancel refund in processing: %w", err)
	}

	// processing also sends the canceled status webhook, the conditional update makes it a no-op
	if _, err := s.storage.RefundRequests().UpdateStatusByTransferID(ctx, repo_refund_requests.UpdateStatusByTransferIDParams{
		TransferID:      ref.TransferID,
		CurrentStatuses: []string{constants.RefundStatusScheduled},
		NewStatus:       constants.RefundStatusPendingReview,
		ResetTransfer:   true,
		LastError:       pgtype.Text{String: "scheduled refund was canceled by the admin", Valid: true},
	}); err != nil {
		return nil, fmt.Errorf("update canceled refund request: %w", err)
	}

	updated, err := s.storage.RefundRequests().GetById(ctx, dto.RefundRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetch refund request: %w", err)
	}

	return s.enrichRefundRequest(ctx, updated)
}

// ExecuteRefundNow sends a scheduled refund to the deposit sender immediately, without waiting
// for the cooldown. The owner's TOTP is required, processing verifies it.
func (s *Service) ExecuteRefundNow(ctx context.Context, dto ExecuteRefundNowDTO) (*RequestWithTxDTO, error) {
	if err := s.checkRefundOwner(ctx, dto.RefundRequestID, dto.UserID); err != nil {
		return nil, err
	}

	if dto.TOTP == "" {
		return nil, ErrRefundTwoFactorInvalid
	}

	ref, err := s.storage.RefundRequests().GetById(ctx, dto.RefundRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetch refund request: %w", err)
	}

	if ref.Status != constants.RefundStatusScheduled || !ref.TransferID.Valid {
		return nil, ErrRefundNotScheduled
	}

	user, err := s.storage.Users().GetByID(ctx, dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	if !user.ProcessingOwnerID.Valid {
		return nil, fmt.Errorf("processing owner is not initialized")
	}

	if _, err := s.processing.ExecuteTransferNow(ctx, user.ProcessingOwnerID.UUID, ref.TransferID.UUID, dto.TOTP); err != nil {
		if code, ok := processing.ErrorRPCCode(err); ok && code == processing.RPCCodeTransferNotCancelable {
			return nil, fmt.Errorf("%w: %w", ErrRefundNotScheduled, err)
		}
		return nil, fmt.Errorf("execute refund now: %w", convertProcessingError(err))
	}

	// processing picks the transfer up within a second and its status webhook also moves
	// the request to processing, so the request may already be updated
	var updated *models.RefundRequest
	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		current, err := s.storage.RefundRequests(repos.WithTx(tx)).UpdateStatusIfCurrent(ctx, repo_refund_requests.UpdateStatusIfCurrentParams{
			ID:            ref.ID,
			CurrentStatus: constants.RefundStatusScheduled,
			NewStatus:     constants.RefundStatusProcessing,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			updated, err = s.storage.RefundRequests(repos.WithTx(tx)).GetById(ctx, ref.ID)
			return err
		}
		if err != nil {
			return fmt.Errorf("update refund request: %w", err)
		}

		updated, err = s.storage.RefundRequests(repos.WithTx(tx)).Update(ctx, repo_refund_requests.UpdateParams{
			ID:         current.ID,
			Status:     current.Status,
			TransferID: current.TransferID,
			ReviewedAt: pgtypeutils.EncodeTime(time.Now()),
			// the refund is sent now, the cooldown is not actual anymore
			ExecuteAfter: pgtype.Timestamp{},
			LastError:    pgtype.Text{},
		})
		if err != nil {
			return fmt.Errorf("update refund request: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.enrichRefundRequest(ctx, updated)
}

// checkRefundOwner checks that the refund request belongs to a store of the user
func (s *Service) checkRefundOwner(ctx context.Context, refundRequestID, userID uuid.UUID) error {
	ref, err := s.storage.RefundRequests().GetById(ctx, refundRequestID)
	if err != nil {
		return fmt.Errorf("fetch refund request: %w", err)
	}

	store, err := s.storage.Stores().GetByID(ctx, ref.StoreID)
	if err != nil {
		return fmt.Errorf("fetch refund request store: %w", err)
	}

	if store.UserID != userID {
		return ErrRefundNotOwned
	}

	return nil
}

func (s *Service) enrichRefundRequest(ctx context.Context, ref *models.RefundRequest) (*RequestWithTxDTO, error) {
	btx, err := s.storage.BlockedTransactions().GetById(ctx, ref.BlockedTransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transaction %s: %w", ref.BlockedTransactionID, err)
	}

	tx, err := s.storage.Transactions().GetById(ctx, btx.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch transaction %s: %w", btx.TransactionID, err)
	}

	currency, err := s.storage.Currencies().GetByID(ctx, tx.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("fetch currency %s: %w", tx.CurrencyID, err)
	}

	return &RequestWithTxDTO{
		RefundRequest: *ref,
		TransactionID: tx.ID,
		TxHash:        tx.TxHash,
		Amount:        tx.Amount,
		AmountUsd:     tx.AmountUsd,
		CurrencyID:    tx.CurrencyID,
		CurrencyCode:  currency.Code,
		Blockchain:    tx.Blockchain,
		FromAddress:   tx.FromAddress,
		ToAddress:     tx.ToAddress,
		RiskLevel:     btx.RiskLevel,
		Score:         btx.Score,
		BlockReasons:  btx.BlockReasons,
	}, nil
}

// TruncateRefundError limits the error saved in the refund request
func TruncateRefundError(msg string) string {
	const maxLen = 1000
	if len(msg) > maxLen {
		return msg[:maxLen]
	}
	return msg
}

func refundRequestFromPendingRow(row *repo_refund_requests.GetAllByUserIDAndStatusRow) *models.RefundRequest {
	return &models.RefundRequest{
		ID:                   row.ID,
		BlockedTransactionID: row.BlockedTransactionID,
		WalletID:             row.WalletID,
		StoreID:              row.StoreID,
		TransferID:           row.TransferID,
		DestinationAddress:   row.DestinationAddress,
		Status:               row.Status,
		Email:                row.Email,
		AuthMethod:           row.AuthMethod,
		ReviewedAt:           row.ReviewedAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		ExecuteAfter:         row.ExecuteAfter,
		LastError:            row.LastError,
	}
}
