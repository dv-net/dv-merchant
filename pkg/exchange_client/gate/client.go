package gateio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ulule/limiter/v3"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IGateClient interface {
	Wallet() IGateWallet
}

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*BaseClient, error) {
	// Create a client with logging enabled by default
	client := NewClient(opt, store, opts...)
	client.logEnabled = true

	c := &BaseClient{
		opts:          opt,
		accountClient: NewAccountClient(opt, store, opts...),
		walletClient:  NewWalletClient(opt, store, opts...),
		spotClient:    NewSpotClient(opt, store, opts...),
	}
	return c, nil
}

type BaseClient struct {
	opts          *ClientOptions
	accountClient IGateAccount
	walletClient  IGateWallet
	spotClient    IGateSpot
}

func (o *BaseClient) Account() IGateAccount { return o.accountClient }
func (o *BaseClient) Wallet() IGateWallet   { return o.walletClient }
func (o *BaseClient) Spot() IGateSpot       { return o.spotClient }

type ClientOption func(c *Client)

func WithLogger(log logger.Logger) ClientOption {
	return func(c *Client) {
		c.log = log
	}
}

func NewClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *Client {
	c := &Client{
		accessKey:  opt.AccessKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		httpClient: http.DefaultClient,
		signer:     NewSigner(opt.AccessKey, opt.SecretKey),
		store:      store,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type ClientOptions struct {
	AccessKey string
	SecretKey string
	BaseURL   *url.URL
}

type Client struct {
	accessKey  string
	secretKey  string
	baseURL    *url.URL
	httpClient *http.Client
	store      limiter.Store
	limiters   map[string]*limiter.Limiter
	signer     ISigner
	log        logger.Logger
	logEnabled bool
}

func (o *Client) Do(ctx context.Context, method, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	if l, exists := o.limiters[endpoint]; exists {
		for {
			r, err := l.Get(ctx, utils.HashLimiterKey(endpoint, o.accessKey, o.secretKey))
			if err != nil {
				return err
			}
			if !r.Reached {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Until(time.Unix(r.Reset, 0).Add(time.Second))):
			}
		}
	}
	return o.DoPlain(ctx, method, endpoint, private, dest, params...)
}

func (o *Client) DoPlain(ctx context.Context, method, path string, private bool, dest interface{}, params ...map[string]string) error {
	startTime := time.Now()
	baseURL := o.baseURL.String() + path
	var (
		req  *http.Request
		err  error
		j    []byte
		body string
	)

	if o.logEnabled && o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Preparing request",
			"exchange", "gateio",
			"method", method,
			"endpoint", path,
			"private", private,
		)
	}

	switch method {
	case http.MethodGet, http.MethodDelete:
		req, err = http.NewRequestWithContext(ctx, method, baseURL, nil)
		if err != nil {
			return err
		}

		if len(params) > 0 {
			q := req.URL.Query()
			for k, v := range params[0] {
				q.Add(k, strings.ReplaceAll(v, "\"", ""))
			}
			req.URL.RawQuery = q.Encode()
		}
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		j, err = json.Marshal(params[0])
		if err != nil {
			return err
		}
		body = string(j)
		if body == "{}" {
			body = ""
		}
		req, err = http.NewRequestWithContext(ctx, method, baseURL, bytes.NewBuffer(j))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported method %s", method)
	}

	req = o.signer.AddHeaders(req)

	if private {
		req, err = o.signer.SignRequest(ctx, req)
		if err != nil {
			return err
		}
	}

	if o.logEnabled && o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Sending request",
			"exchange", "gateio",
			"method", method,
			"url", o.baseURL.String()+path,
			"headers", sanitizeHeaders(req.Header),
			"body", sanitizeBody(body),
		)
	}

	res, err := o.httpClient.Do(req)
	if err != nil {
		if o.logEnabled && o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: Request failed",
				"exchange", "gateio",
				"method", method,
				"endpoint", path,
				"error", err.Error(),
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
		}
		return err
	}
	defer res.Body.Close()

	bb := new(bytes.Buffer)
	_, err = io.Copy(bb, res.Body)
	if err != nil {
		return err
	}

	duration := time.Since(startTime)

	// Check for errors
	if res.StatusCode >= 400 {
		errRes := ErrorResponse{}
		if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
			return fmt.Errorf("gate.io error: status %d, body: %s", res.StatusCode, bb.String())
		}
		if o.logEnabled && o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: API error response",
				"exchange", "gateio",
				"method", method,
				"endpoint", path,
				"status_code", res.StatusCode,
				"error", errorFromResponse(&errRes).Error(),
				"duration_ms", duration.Milliseconds(),
			)
		}
		return errorFromResponse(&errRes)
	}

	if o.logEnabled && o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Request completed",
			"exchange", "gateio",
			"method", method,
			"endpoint", path,
			"status_code", res.StatusCode,
			"duration_ms", duration.Milliseconds(),
		)
	}

	err = json.Unmarshal(bb.Bytes(), &dest)
	return err
}

// sanitizeHeaders removes sensitive headers for logging
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for k, v := range headers {
		if strings.Contains(strings.ToLower(k), "key") ||
			strings.Contains(strings.ToLower(k), "sign") ||
			strings.Contains(strings.ToLower(k), "signature") {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = strings.Join(v, ",")
		}
	}
	return sanitized
}

// sanitizeBody truncates long bodies and masks sensitive data for logging
func sanitizeBody(body string) string {
	if len(body) == 0 {
		return "(empty)"
	}
	if len(body) > 500 {
		return body[:500] + "... (truncated)"
	}
	return body
}

// ErrorResponse represents gate.io API error response
type ErrorResponse struct {
	Label   string `json:"label"`   // Error code
	Message string `json:"message"` // Error message
}

func errorFromResponse(errRes *ErrorResponse) error {
	switch errRes.Label {
	case "INVALID_KEY", "INVALID_SIGNATURE", "KEY_NOT_FOUND":
		return exchangeclient.ErrInvalidAPICredentials
	case "IP_FORBIDDEN", "IP_NOT_IN_WHITELIST":
		return exchangeclient.ErrInvalidIPAddress
	case "FORBIDDEN":
		// Could be IP or permission issue, check message
		if strings.Contains(strings.ToLower(errRes.Message), "ip") {
			return exchangeclient.ErrInvalidIPAddress
		}
		return wrapGateError(errRes.Label, errRes.Message)
	default:
		return wrapGateError(errRes.Label, errRes.Message)
	}
}

// wrapGateError wraps Gate.io errors with centralized errors when applicable
func wrapGateError(label, message string) error {
	msgLower := strings.ToLower(message)

	// Check for balance-related errors
	if strings.Contains(msgLower, "balance") {
		return fmt.Errorf("gate.io error: %s - %s: %w", label, message, exchangeclient.ErrWithdrawalBalanceLocked)
	}

	return fmt.Errorf("gate.io error: %s - %s", label, message)
}

func S2M(i interface{}) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
}
