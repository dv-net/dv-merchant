package auth

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/tools/str"
	"github.com/google/uuid"
)

const (
	walletCodeLetters        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	walletCodeDigits         = "0123456789"
	walletCodeResendCooldown = 60 * time.Second
)

func generateWalletCode() (string, error) {
	letters, err := str.RandomStringFromCharset(walletCodeLetters, 3)
	if err != nil {
		return "", err
	}
	digits, err := str.RandomStringFromCharset(walletCodeDigits, 3)
	if err != nil {
		return "", err
	}
	return letters + digits, nil
}

func walletCodeResendCooldownKey(walletID uuid.UUID) string {
	return "wallet_refund_cooldown:" + walletID.String()
}

func walletOTPPurpose(walletID uuid.UUID) string {
	return "wallet_refund:" + walletID.String()
}
