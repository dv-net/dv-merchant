package turnstile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type VerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"` //nolint:tagliatelle
}

type VerifyResponse struct {
	Success bool `json:"success"`
}

type Verifier interface {
	Verify(ctx context.Context, ipAddress, token string) error
}

type Service struct {
	cl        *http.Client
	URL       *url.URL
	secretKey string
	enabled   bool
}

func New(baseURL, secret string, enabled bool) (*Service, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Service{cl: http.DefaultClient, URL: u, secretKey: secret, enabled: enabled}, nil
}

func (s *Service) Verify(ctx context.Context, ipAddress, token string) error {
	if !s.enabled {
		return nil
	}

	verifyReq := VerifyRequest{
		Secret:   s.secretKey,
		Response: token,
		RemoteIP: ipAddress,
	}
	body, err := json.Marshal(verifyReq)
	if err != nil {
		return errors.New("turnstile: failed to marshal verify request")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.URL.JoinPath("/turnstile/v0/siteverify").String(),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("turnstile: failed to create verify request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.cl.Do(req)
	if err != nil {
		return errors.New("turnstile: failed to verify token")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("turnstile: failed to read response body")
	}

	var outcome VerifyResponse
	if err := json.Unmarshal(respBody, &outcome); err != nil {
		return errors.New("turnstile: failed to unmarshal verify response")
	}

	if !outcome.Success {
		return errors.New("turnstile: failed to verify token")
	}

	return nil
}
