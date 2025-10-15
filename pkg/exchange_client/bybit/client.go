package bybit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ulule/limiter/v3"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *BaseClient {
	c := &BaseClient{
		account: NewAccount(opt, store, opts...),
		market:  NewMarket(opt, store, opts...),
		trade:   NewTrade(opt, store, opts...),
	}
	return c
}

type BaseClient struct {
	account IBybitAccount
	market  IBybitMarket
	trade   ITradeClient
}

func (o *BaseClient) Account() IBybitAccount { return o.account }
func (o *BaseClient) Market() IBybitMarket   { return o.market }
func (o *BaseClient) Trade() ITradeClient    { return o.trade }

func NewClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:     opt.KeyAPI,
		secretKey:  opt.KeySecret,
		baseURL:    opt.BaseURL,
		httpClient: http.DefaultClient,
		store:      store,
		limiters:   make(map[string]*limiter.Limiter),
	}

	for _, option := range opts {
		option(c)
	}

	return c
}

type Client struct {
	apiKey     string
	secretKey  string
	baseURL    *url.URL
	httpClient *http.Client
	store      limiter.Store
	limiters   map[string]*limiter.Limiter
	log        logger.Logger
}

func (o *Client) Do(ctx context.Context, method string, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	if l, exists := o.limiters[endpoint]; exists {
		for {
			r, err := l.Get(ctx, utils.HashLimiterKey(endpoint, o.apiKey, o.secretKey, ""))
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

func (o *Client) DoPlain(ctx context.Context, method, path string, private bool, dest interface{}, params ...map[string]string) error { //nolint:gocyclo
	startTime := time.Now()
	baseURL := o.baseURL.String() + path
	var (
		req  *http.Request
		err  error
		j    []byte
		body string
	)

	if o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Preparing request",
			"exchange", "bybit",
			"method", method,
			"endpoint", path,
			"private", private,
		)
	}

	switch method {
	case http.MethodGet:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
		if err != nil {
			return err
		}

		if len(params) > 0 {
			q := req.URL.Query()
			for k, v := range params[0] {
				q.Add(k, strings.ReplaceAll(v, "\"", ""))
			}
			req.URL.RawQuery = q.Encode()
			if len(params[0]) > 0 {
				path += "?" + req.URL.RawQuery
			}
		}
	case http.MethodPost:
		if len(params) > 0 {
			j, err = json.Marshal(params[0])
			if err != nil {
				return err
			}
			body = string(j)
		} else {
			body = ""
		}
		req, err = http.NewRequestWithContext(ctx, method, baseURL, bytes.NewBuffer(j))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
	default:
		return fmt.Errorf("unsupported method %s", method)
	}

	if private {
		timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
		recvWindow := "5000"

		req.Header.Set(SignTypeKey, "2")
		req.Header.Set(APIRequestKey, o.apiKey)
		req.Header.Set(TimestampKey, timestamp)
		req.Header.Set(RecvWindowKey, recvWindow)

		var signatureBase []byte
		if method == "POST" {
			signatureBase = []byte(timestamp + o.apiKey + recvWindow + body)
		} else {
			queryString := ""
			if req.URL.RawQuery != "" {
				queryString = req.URL.RawQuery
			}
			signatureBase = []byte(timestamp + o.apiKey + recvWindow + queryString)
		}

		hmac256 := hmac.New(sha256.New, []byte(o.secretKey))
		hmac256.Write(signatureBase)
		signature := hex.EncodeToString(hmac256.Sum(nil))
		req.Header.Set(SignatureKey, signature)
	}

	if o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Sending request",
			"exchange", "bybit",
			"method", method,
			"url", o.baseURL.String()+path,
			"headers", sanitizeHeaders(req.Header),
			"body", sanitizeBody(body),
		)
	}

	res, err := o.httpClient.Do(req)
	if err != nil {
		if o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: Request failed",
				"exchange", "bybit",
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

	if res.StatusCode >= http.StatusBadRequest {
		if o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: API error response",
				"exchange", "bybit",
				"method", method,
				"endpoint", path,
				"status_code", res.StatusCode,
				"error", exchangeclient.ErrInvalidAPICredentials.Error(),
				"duration_ms", duration.Milliseconds(),
			)
		}
		return exchangeclient.ErrInvalidAPICredentials
	}

	var apiErr APIError
	err = json.Unmarshal(bb.Bytes(), &apiErr)
	if err != nil {
		return fmt.Errorf("error parsing response: %w, response: %s", err, bb.String())
	}

	if apiErr.Code != 0 {
		if err := errorFromResponse(&apiErr); err != nil {
			if o.log != nil {
				o.log.Errorln("[EXCHANGE-API]: API error response",
					"exchange", "bybit",
					"method", method,
					"endpoint", path,
					"status_code", res.StatusCode,
					"error", err.Error(),
					"duration_ms", duration.Milliseconds(),
				)
			}
			return err
		}
	}

	if o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Request completed",
			"exchange", "bybit",
			"method", method,
			"endpoint", path,
			"status_code", res.StatusCode,
			"duration_ms", duration.Milliseconds(),
		)
	}

	if err = json.Unmarshal(bb.Bytes(), &dest); err != nil {
		return fmt.Errorf("error parsing error response: %w, response: %s", err, bb.String())
	}

	return err
}

type ClientOption func(*Client)

func WithBaseURL(baseURL *url.URL) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithLogger(log logger.Logger) ClientOption {
	return func(c *Client) {
		c.log = log
	}
}

func FormatTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func GetCurrentTime() int64 {
	now := time.Now()
	unixNano := now.UnixNano()
	timeStamp := unixNano / int64(time.Millisecond)
	return timeStamp
}

type ClientOptions struct {
	KeyAPI    string
	KeySecret string
	BaseURL   *url.URL
}

func S2M(i interface{}) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
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

func errorFromResponse(errRes *APIError) error {
	isCredentialError := func(code int64) bool {
		return (code >= ErrAPIKeyInvalid && code <= ErrAPISecretInvalid)
	}

	isWhitelistError := func(code int64) bool {
		return code == ErrIPWhitelist
	}

	if errRes.Code != 0 {
		if isCredentialError(errRes.Code) {
			return exchangeclient.ErrInvalidAPICredentials
		}
		if isWhitelistError(errRes.Code) {
			return exchangeclient.ErrInvalidIPAddress
		}
		return fmt.Errorf("bybit error: %s (%d)", errRes.Message, errRes.Code)
	}

	return nil
}
