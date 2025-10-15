package admin_gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	admin_errors "github.com/dv-net/dv-merchant/pkg/admin_gateway/errors"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type Service struct {
	cl         *http.Client
	log        logger.Logger
	appVersion string
	baseURL    string
	logStatus  bool
}

func New(baseURL string, appVersion string, log logger.Logger, logStatus bool) *Service {
	return &Service{
		cl:         http.DefaultClient,
		log:        log,
		appVersion: appVersion,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		logStatus:  logStatus,
	}
}

func (s *Service) sendRequest(
	ctx context.Context,
	apiMethod string,
	httpMethod string,
	body []byte,
	customHeaders map[string]string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, s.baseURL+apiMethod, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}

	identity := IdentityFromContext(ctx)
	if identity == nil {
		return nil, fmt.Errorf("service is not identified")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Version", s.appVersion)
	req.Header.Set("X-Client-ID", identity.ClientID)
	req.Header.Set("X-Sign", getHash(string(body), identity.SecretKey))
	if len(customHeaders) > 0 {
		for k, v := range customHeaders {
			req.Header.Set(k, v)
		}
	}
	if s.logStatus {
		s.log.Infoln("[DV-API]: Sending request",
			"method", httpMethod,
			"url", s.baseURL+apiMethod,
			"headers", req.Header,
			"body", string(body),
		)
	}

	resp, err := s.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send dv-admin request: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			s.log.Errorw("failed to close response body", "error", err)
		}
	}()

	parsedBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if s.logStatus {
		s.log.Infoln("[DV-API]: Received response",
			"status", resp.StatusCode,
			"body", string(parsedBody),
		)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &admin_errors.RequestFailedError{
			RequestURL: s.baseURL + apiMethod,
			StatusCode: resp.StatusCode,
			Body:       parsedBody,
		}
	}

	return parsedBody, nil
}

func (s *Service) sendPublicRequest(
	ctx context.Context,
	apiMethod string,
	httpMethod string,
	body []byte,
	customHeaders map[string]string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, s.baseURL+apiMethod, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if len(customHeaders) > 0 {
		for k, v := range customHeaders {
			req.Header.Set(k, v)
		}
	}
	if s.logStatus {
		s.log.Infoln("[DV-API]: Sending public request ",
			"method", httpMethod,
			"url", s.baseURL+apiMethod,
			"headers", req.Header,
			"body", string(body),
		)
	}

	resp, err := s.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send dv-admin request: %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			s.log.Errorw("failed to close response body", "error", err)
		}
	}()

	parsedBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if s.logStatus {
		s.log.Infoln("[DV-API]: Received public response ",
			"status", resp.StatusCode,
			"body", string(parsedBody),
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &admin_errors.RequestFailedError{
			RequestURL: s.baseURL + apiMethod,
			StatusCode: resp.StatusCode,
			Body:       parsedBody,
		}
	}

	return parsedBody, nil
}

func SHA256Signature(data []byte, secretKey string) string {
	sign := sha256.New()
	sign.Write(append(data, []byte(secretKey)...))
	return hex.EncodeToString(sign.Sum(nil))
}

func getHash(rawJSONString string, secretKey string) string {
	var compactBody bytes.Buffer
	_ = json.Compact(&compactBody, []byte(rawJSONString))

	return SHA256Signature(compactBody.Bytes(), secretKey)
}
