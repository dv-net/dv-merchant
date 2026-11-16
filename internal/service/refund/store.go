package refund

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) EstimateRefundByStore(ctx context.Context, dto StoreRefundDTO) (*EstimateDTO, error) {
	checked, err := s.checkClaimByStore(ctx, dto)
	if err != nil {
		return nil, err
	}

	return s.buildEstimate(checked)
}

func (s *Service) SendRefundCodeByStore(ctx context.Context, dto StoreRefundDTO) error {
	checked, err := s.checkClaimByStore(ctx, dto)
	if err != nil {
		return err
	}

	if checked.result.ToSender {
		return ErrRefundCodeNotRequired
	}

	if err := checkFeeCovered(checked); err != nil {
		return err
	}

	if err := s.codes.SendRefundConfirmationCode(ctx, checked.btx.WalletID, checked.btx.ID, checked.destinationAddress); err != nil {
		return convertCodeError(err)
	}

	return nil
}

func (s *Service) CreateRefundByStore(ctx context.Context, dto StoreRefundDTO) (*RequestWithTxDTO, error) {
	checked, err := s.checkClaimByStore(ctx, dto)
	if err != nil {
		return nil, err
	}

	var ref *models.RefundRequest
	if checked.result.ToSender {
		ref, err = s.createRefund(ctx, checked, pgtype.Text{}, constants.RefundAuthMethodAPI)
	} else {
		ref, err = s.createConfirmedRefundByStore(ctx, checked, dto.Code)
	}
	if err != nil {
		return nil, err
	}

	return s.enrichRefundRequest(ctx, ref)
}

func (s *Service) GetRefundByStore(ctx context.Context, storeID, refundRequestID uuid.UUID) (*RequestWithTxDTO, error) {
	ref, err := s.storage.RefundRequests().GetById(ctx, refundRequestID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRefundNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch refund request: %w", err)
	}

	if ref.StoreID != storeID {
		return nil, fmt.Errorf("%w: store id mismatch", ErrRefundNotFound)
	}

	return s.enrichRefundRequest(ctx, ref)
}

func (s *Service) createConfirmedRefundByStore(ctx context.Context, checked *claimCheck, code string) (*models.RefundRequest, error) {
	if code == "" {
		return nil, ErrRefundCodeRequired
	}

	if err := checkFeeCovered(checked); err != nil {
		return nil, err
	}

	email, err := s.codes.VerifyRefundConfirmationCode(ctx, checked.btx.WalletID, checked.btx.ID, checked.destinationAddress, code)
	if err != nil {
		return nil, convertCodeError(err)
	}

	return s.createRefund(ctx, checked, pgtype.Text{String: email, Valid: true}, constants.RefundAuthMethodAPIEmail)
}

func (s *Service) checkClaimByStore(ctx context.Context, dto StoreRefundDTO) (*claimCheck, error) {
	btx, err := s.storage.BlockedTransactions().GetById(ctx, dto.BlockedTransactionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRefundBlockedTxNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transaction: %w", err)
	}

	if btx.StoreID != dto.StoreID {
		return nil, fmt.Errorf("%w: store id mismatch", ErrRefundBlockedTxNotFound)
	}

	return s.checkClaim(ctx, btx.WalletID, btx.ID, dto.DestinationAddress, false)
}

func checkFeeCovered(checked *claimCheck) error {
	if !checked.result.FeeCovered() {
		return fmt.Errorf("%w: fee shortage %s, activation required %t",
			ErrRefundInsufficientFee, checked.result.FeeShortage.Decimal, checked.result.ActivationRequired)
	}
	return nil
}

func convertCodeError(err error) error {
	switch {
	case errors.Is(err, auth.ErrInvalidOTPCode):
		return fmt.Errorf("%w: %w", ErrRefundCodeInvalid, err)
	case errors.Is(err, auth.ErrRefundCodeCooldown):
		return fmt.Errorf("%w: %w", ErrRefundCodeCooldown, err)
	case errors.Is(err, auth.ErrRefundCodeAttemptsLimit):
		return fmt.Errorf("%w: %w", ErrRefundCodeAttempts, err)
	case errors.Is(err, auth.ErrWalletEmailMissing):
		return fmt.Errorf("%w: %w", ErrRefundNoEmail, err)
	}
	return err
}
