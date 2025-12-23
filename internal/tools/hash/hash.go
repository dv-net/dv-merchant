package hash //nolint:revive

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func SHA256(plainTextToken string) string {
	hash := sha256.New()
	hash.Write([]byte(plainTextToken))
	hashedBytes := hash.Sum(nil)
	return hex.EncodeToString(hashedBytes)
}

func SHA256Signature(data []byte, secretKey string) string {
	sign := sha256.New()
	sign.Write(append(data, []byte(secretKey)...))
	return hex.EncodeToString(sign.Sum(nil))
}

func SHA256ConnectionHash(exchangeSlug string, input ...string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("input cannot be empty")
	}
	if exchangeSlug == "" {
		return "", fmt.Errorf("exchangeSlug cannot be empty")
	}
	hasher := sha256.New()
	secretString := strings.Join(input, "")
	hasher.Write([]byte(secretString))
	hashedBytes := hasher.Sum(nil)
	return fmt.Sprintf("%s_%s", exchangeSlug, hex.EncodeToString(hashedBytes)), nil
}
