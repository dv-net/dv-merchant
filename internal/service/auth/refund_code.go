package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/pkg/key_value"

	"github.com/google/uuid"
)

const (
	refundCodeResendCooldown = 60 * time.Second
	// refundCodeMaxAttempts limits the code checks of one blocked transaction within refundCodeAttemptsWindow
	refundCodeMaxAttempts    = 5
	refundCodeAttemptsWindow = 30 * time.Minute
)

func refundCodeCooldownKey(blockedTransactionID uuid.UUID) string {
	return "refund_api_code_cooldown:" + blockedTransactionID.String()
}

func refundCodeAttemptsKey(blockedTransactionID uuid.UUID) string {
	return "refund_api_code_attempts:" + blockedTransactionID.String()
}

// refundCodePurpose binds the code to the refund of the blocked transaction to the destination address:
// a code sent for one address does not confirm a refund to another one
func refundCodePurpose(walletID, blockedTransactionID uuid.UUID, destinationAddress string) string {
	return "refund_api:" + walletID.String() + ":" + blockedTransactionID.String() + ":" + destinationAddress
}

// SendRefundConfirmationCode sends a one-time code to the wallet email. The code confirms a refund
// of the blocked transaction to the destination address requested by the store through the external api.
func (s *Service) SendRefundConfirmationCode(ctx context.Context, walletID, blockedTransactionID uuid.UUID, destinationAddress string) error {
	w, err := s.walletService.GetWallet(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWalletNotFound, err)
	}
	if !w.Email.Valid || w.Email.String == "" {
		return ErrWalletEmailMissing
	}

	if err := s.storage.KeyValue().IncrementCounterWithLimit(ctx, refundCodeCooldownKey(blockedTransactionID), 1,
		refundCodeResendCooldown); err != nil {
		if errors.Is(err, key_value.ErrCounterLimitExceeded) {
			return ErrRefundCodeCooldown
		}
		return fmt.Errorf("check refund code cooldown: %w", err)
	}

	code, err := s.otpService.InitStringCode(ctx, "", refundCodePurpose(walletID, blockedTransactionID, destinationAddress), generateWalletCode)
	if err != nil {
		return fmt.Errorf("init refund code: %w", err)
	}

	storeID := w.StoreID
	go s.notifyService.SendSystemEmail(context.WithoutCancel(ctx), models.NotificationTypeRefundVerificationCode, w.Email.String,
		&notify.RefundVerificationCodeData{
			Language: w.Locale,
			Code:     code,
		}, &models.NotificationArgs{StoreID: &storeID})

	return nil
}

// VerifyRefundConfirmationCode checks the code sent by SendRefundConfirmationCode and returns the wallet email.
// The code is single use, the number of checks per blocked transaction is limited.
func (s *Service) VerifyRefundConfirmationCode(ctx context.Context, walletID, blockedTransactionID uuid.UUID, destinationAddress, code string) (string, error) {
	w, err := s.walletService.GetWallet(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrWalletNotFound, err)
	}
	if !w.Email.Valid || w.Email.String == "" {
		return "", ErrWalletEmailMissing
	}

	if err := s.storage.KeyValue().IncrementCounterWithLimit(ctx, refundCodeAttemptsKey(blockedTransactionID), refundCodeMaxAttempts,
		refundCodeAttemptsWindow); err != nil {
		if errors.Is(err, key_value.ErrCounterLimitExceeded) {
			return "", ErrRefundCodeAttemptsLimit
		}
		return "", fmt.Errorf("check refund code attempts: %w", err)
	}

	purpose := refundCodePurpose(walletID, blockedTransactionID, destinationAddress)
	if err := s.otpService.VerifyStringCode(ctx, strings.ToUpper(strings.TrimSpace(code)), "", purpose); err != nil {
		return "", ErrInvalidOTPCode
	}

	return w.Email.String, nil
}
