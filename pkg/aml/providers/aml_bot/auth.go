package aml_bot

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MD5Authorizer implements RequestAuthorizer with MD5 token authentication.
type MD5Authorizer struct {
	accessID  string
	accessKey string
	uid       *string // Optional uid, used only for recheck requests
}

// Option configures MD5Authorizer.
type Option func(*MD5Authorizer)

// WithUID sets the uid for MD5Authorizer (used in recheck requests).
func WithUID(uid string) Option {
	return func(a *MD5Authorizer) {
		a.uid = &uid
	}
}

// NewMD5Authorizer creates a new MD5Authorizer.
func NewMD5Authorizer(accessID, accessKey string, opts ...Option) *MD5Authorizer {
	auth := &MD5Authorizer{
		accessID:  accessID,
		accessKey: accessKey,
	}
	for _, opt := range opts {
		opt(auth)
	}
	return auth
}

// Authorize adds the MD5 token and access ID to the request.
func (a *MD5Authorizer) Authorize(_ context.Context, req *http.Request) error {
	// Generate MD5 token: md5('uid:access_key:accessId') for recheck with uid, md5('access_key:accessId') otherwise
	hash := a.generateHash()
	token := hex.EncodeToString(hash[:])

	// Add form data
	values := req.URL.Query()
	if req.Method == http.MethodPost {
		// For POST requests, parse form data from body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		if len(body) > 0 {
			values, err = url.ParseQuery(string(body))
			if err != nil {
				return fmt.Errorf("parse form data: %w", err)
			}
		}
	}

	// Add accessId and token to form data
	values.Set("accessId", a.accessID)
	values.Set("token", token)

	// Update request body for POST
	if req.Method == http.MethodPost {
		req.Body = io.NopCloser(strings.NewReader(values.Encode()))
		req.ContentLength = int64(len(values.Encode()))
	} else {
		req.URL.RawQuery = values.Encode()
	}

	return nil
}

func (a *MD5Authorizer) generateHash() [16]byte {
	var builder strings.Builder

	size := len(a.accessKey) + len(a.accessID) + 1
	if a.uid != nil {
		size += len(*a.uid) + 1
	}
	builder.Grow(size)

	if a.uid != nil {
		builder.WriteString(*a.uid)
		builder.WriteByte(':')
	}
	builder.WriteString(a.accessKey)
	builder.WriteByte(':')
	builder.WriteString(a.accessID)

	return md5.Sum([]byte(builder.String())) //nolint:gosec
}
