package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

type ISigner interface {
	Sign(path, nonce, body string) (string, error)
}

type Signer struct {
	secret []byte
	err    error
}

func NewSigner(secretKey string) (*Signer, error) {
	if secretKey == "" || secretKey == "-" {
		return &Signer{err: fmt.Errorf("kraken: secret key is not set")}, nil
	}
	secret, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		return nil, fmt.Errorf("kraken: secret key is not valid base64: %w", err)
	}
	return &Signer{secret: secret}, nil
}

// Sign computes base64(HMAC-SHA512(secret, path + SHA256(nonce + body))).
func (s *Signer) Sign(path, nonce, body string) (string, error) {
	if s.err != nil {
		return "", s.err
	}

	sha := sha256.Sum256([]byte(nonce + body))

	mac := hmac.New(sha512.New, s.secret)
	mac.Write([]byte(path))
	mac.Write(sha[:])

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
