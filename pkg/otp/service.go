package otp

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"
)

const DefaultMaxAttempts = 5

var (
	ErrInvalidCode     = errors.New("invalid OTP code")
	ErrTooManyAttempts = errors.New("too many OTP attempts, request a new code later")
)

type Config struct {
	TTL         time.Duration
	MaxAttempts int64
}

type CodeGenerator func() int

type StringCodeGenerator func() (string, error)
type Service struct {
	config  *Config
	codeGen CodeGenerator
	store   key_value.IKeyValue
}

func New(cfg *Config, codeGen CodeGenerator, store key_value.IKeyValue) *Service {
	return &Service{
		config:  cfg,
		codeGen: codeGen,
		store:   store,
	}
}

func (s *Service) InitCode(ctx context.Context, identifier, purpose string) (int, error) {
	code := s.codeGen()
	if err := s.storeCode(ctx, strconv.Itoa(code), identifier, purpose); err != nil {
		return 0, err
	}

	return code, nil
}

func (s *Service) VerifyCode(ctx context.Context, code int, identifier, purpose string) error {
	return s.verify(ctx, strconv.Itoa(code), identifier, purpose)
}

func (s *Service) InitStringCode(ctx context.Context, identifier, purpose string, gen StringCodeGenerator) (string, error) {
	code, err := gen()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.storeCode(ctx, code, identifier, purpose); err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) VerifyStringCode(ctx context.Context, code, identifier, purpose string) error {
	return s.verify(ctx, code, identifier, purpose)
}

func (s *Service) storeCode(ctx context.Context, code, identifier, purpose string) error {
	if err := s.store.Set(ctx, codeKey(identifier, purpose), hashCode(code, identifier, purpose), s.config.TTL); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	return nil
}

func (s *Service) verify(ctx context.Context, code, identifier, purpose string) error {
	cKey := codeKey(identifier, purpose)
	aKey := attemptsKey(identifier, purpose)

	if err := s.store.IncrementCounterWithLimit(ctx, aKey, s.maxAttempts(), s.config.TTL); err != nil {
		if errors.Is(err, key_value.ErrCounterLimitExceeded) {
			_ = s.store.Delete(ctx, cKey)
			return ErrTooManyAttempts
		}
		return fmt.Errorf("count OTP attempt: %w", err)
	}

	stored, err := s.store.Get(ctx, cKey)
	if err != nil || len(stored) == 0 {
		return ErrInvalidCode
	}

	expected := hashCode(code, identifier, purpose)
	if subtle.ConstantTimeCompare(stored.Bytes(), []byte(expected)) != 1 {
		return ErrInvalidCode
	}

	if err := s.store.Delete(ctx, cKey); err != nil {
		return fmt.Errorf("delete used OTP: %w", err)
	}
	_ = s.store.Delete(ctx, aKey)

	return nil
}

func (s *Service) maxAttempts() int64 {
	if s.config.MaxAttempts > 0 {
		return s.config.MaxAttempts
	}
	return DefaultMaxAttempts
}

func codeKey(identifier, purpose string) string {
	return "otp:code:" + sha256Hex(identifier+":"+purpose)
}

func attemptsKey(identifier, purpose string) string {
	return "otp:attempts:" + sha256Hex(identifier+":"+purpose)
}

func hashCode(code, identifier, purpose string) string {
	return sha256Hex(identifier + ":" + purpose + ":" + code)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
