package bitok

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HMACAuthorizer implements RequestAuthorizer with HMAC-SHA256 signature
type HMACAuthorizer struct {
	secret    string
	accessKey string
}

// NewHMACAuthorizer creates a new HMACAuthorizer
func NewHMACAuthorizer(secret, accessKey string) *HMACAuthorizer {
	return &HMACAuthorizer{
		secret:    secret,
		accessKey: accessKey,
	}
}

// Authorize adds HMAC signature and headers to the request
func (a *HMACAuthorizer) Authorize(_ context.Context, req *http.Request) error {
	timestamp := time.Now().UTC().UnixMilli()

	// Get request body
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("read request body: %w", err)
		}
		// Restore body for subsequent reads
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	// Compact JSON body if valid
	if len(body) > 0 {
		var err error
		body, err = prepareBody(body)

		if err != nil {
			return fmt.Errorf("prepare body: %w", err)
		}
	}

	// Get endpoint with query parameters
	endpoint := req.URL.Path
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	if req.URL.RawQuery != "" {
		endpoint += "?" + req.URL.RawQuery
	}

	// Generate signature
	var builder strings.Builder
	builder.Grow(len(req.Method) + len(endpoint) + 32 + len(body) + 3)
	builder.WriteString(req.Method)
	builder.WriteByte('\n')
	builder.WriteString(endpoint)
	builder.WriteByte('\n')
	builder.WriteString(fmt.Sprintf("%d", timestamp))
	if len(body) > 0 {
		builder.WriteByte('\n')
		builder.Write(body)
	}

	h := hmac.New(sha256.New, []byte(a.secret))
	if _, err := h.Write([]byte(builder.String())); err != nil {
		return fmt.Errorf("compute HMAC: %w", err)
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Set headers
	req.Header["API-TIMESTAMP"] = []string{fmt.Sprintf("%d", timestamp)}
	req.Header["API-SIGNATURE"] = []string{signature}
	req.Header["API-KEY-ID"] = []string{a.accessKey}

	return nil
}

func prepareBody(body []byte) ([]byte, error) {
	if json.Valid(body) {
		var tmp any
		err := json.Unmarshal(body, &tmp)
		if err != nil {
			return nil, fmt.Errorf("parse JSON body for compact serialization: %w", err)
		}

		body, err = json.Marshal(tmp)
		if err != nil {
			return nil, fmt.Errorf("compact JSON serialization: %w", err)
		}
	}

	return body, nil
}
