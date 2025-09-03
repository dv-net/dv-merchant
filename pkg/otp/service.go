package otp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"
)

type Config struct {
	TTL time.Duration
}

type CodeGenerator func() int

type Service struct {
	config      *Config
	codeGen     CodeGenerator
	store       key_value.IKeyValue
	hashKeyFunc func(code int, identifier, purpose string) string
}

func New(cfg *Config, codeGen CodeGenerator, store key_value.IKeyValue) *Service {
	return &Service{
		config:      cfg,
		codeGen:     codeGen,
		store:       store,
		hashKeyFunc: defaultHashKeyFunc,
	}
}

func (s *Service) InitCode(ctx context.Context, identifier, purpose string) (int, error) {
	code := s.codeGen()
	key := s.hashKeyFunc(code, identifier, purpose)

	if err := s.store.Set(ctx, key, purpose, s.config.TTL); err != nil {
		return 0, fmt.Errorf("failed to store OTP: %w", err)
	}

	return code, nil
}

func (s *Service) VerifyCode(ctx context.Context, code int, identifier, purpose string) error {
	key := s.hashKeyFunc(code, identifier, purpose)
	storedPurpose, err := s.store.Get(ctx, key)
	if err != nil || storedPurpose == nil || storedPurpose.String() != purpose {
		return fmt.Errorf("invalid OTP code")
	}

	return s.store.Delete(ctx, key)
}

func defaultHashKeyFunc(code int, identifier, purpose string) string {
	keyString := fmt.Sprintf("%s:%s:%d", identifier, purpose, code)
	h := sha256.New()
	h.Write([]byte(keyString))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Service) SetHashKeyFunc(f func(code int, identifier, purpose string) string) {
	s.hashKeyFunc = f
}
