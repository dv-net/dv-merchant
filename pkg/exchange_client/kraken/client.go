package kraken

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

const (
	exchangeName       = "kraken"
	defaultHTTPTimeout = 30 * time.Second
)

// Private endpoints share one rate counter per API key on Kraken's side
// (decays ~0.33 points per second for Starter tier), so we throttle all of them together.
var privateRate = limiter.Rate{Limit: 10, Period: 30 * time.Second}

// Public endpoints are limited by Kraken per IP (about 1 request per second),
// so one counter is shared by all clients in the process.
var (
	publicRate     = limiter.Rate{Limit: 1, Period: time.Second}
	publicLimitKey = utils.HashLimiterKey(exchangeName, "public")
)

type IKrakenClient interface {
	Account() IKrakenAccount
	Market() IKrakenMarket
	Spot() IKrakenSpot
	Wallet() IKrakenWallet
}

// NewBaseClient builds all sub-clients on top of a single HTTP client,
// so that they share one signer and nonce sequence.
func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*BaseClient, error) {
	c, err := NewClient(opt, store, opts...)
	if err != nil {
		return nil, err
	}

	return &BaseClient{
		accountClient: NewAccountClient(c),
		marketClient:  NewMarketClient(c),
		spotClient:    NewSpotClient(c),
		walletClient:  NewWalletClient(c),
	}, nil
}

type BaseClient struct {
	accountClient IKrakenAccount
	marketClient  IKrakenMarket
	spotClient    IKrakenSpot
	walletClient  IKrakenWallet
}

var _ IKrakenClient = (*BaseClient)(nil)

func (o *BaseClient) Account() IKrakenAccount { return o.accountClient }
func (o *BaseClient) Market() IKrakenMarket   { return o.marketClient }
func (o *BaseClient) Spot() IKrakenSpot       { return o.spotClient }
func (o *BaseClient) Wallet() IKrakenWallet   { return o.walletClient }

type ClientOption func(c *Client)

func WithLogger(log logger.Logger) ClientOption {
	return func(c *Client) {
		c.log = log
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

type ClientOptions struct {
	APIKey string
	// SecretKey is the base64 encoded "Private Key" issued by Kraken.
	SecretKey string
	BaseURL   *url.URL
}

type Client struct {
	apiKey        string
	baseURL       *url.URL
	httpClient    *http.Client
	privateLimit  *limiter.Limiter
	publicLimit   *limiter.Limiter
	signer        ISigner
	nonce         *nonceSequence
	log           logger.Logger
	privateLimKey string
}

func NewClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*Client, error) {
	if opt == nil || opt.BaseURL == nil {
		return nil, fmt.Errorf("kraken: base url is required")
	}

	signer, err := NewSigner(opt.SecretKey)
	if err != nil {
		return nil, err
	}

	c := &Client{
		apiKey:        opt.APIKey,
		baseURL:       opt.BaseURL,
		httpClient:    &http.Client{Timeout: defaultHTTPTimeout},
		signer:        signer,
		nonce:         nonceFor(opt.APIKey),
		privateLimKey: utils.HashLimiterKey(exchangeName, "private", opt.APIKey),
	}
	if store != nil {
		c.privateLimit = limiter.New(store, privateRate)
		c.publicLimit = limiter.New(store, publicRate)
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// envelope is the common wrapper of every Kraken REST response.
type envelope[T any] struct {
	Error  []string `json:"error"`
	Result T        `json:"result"`
}

// DoPublic performs GET /0/public/{method}.
func DoPublic[T any](ctx context.Context, c *Client, endpoint string, params url.Values) (T, error) {
	var zero T

	if err := waitLimit(ctx, c.publicLimit, publicLimitKey); err != nil {
		return zero, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.String()+endpoint, http.NoBody)
	if err != nil {
		return zero, err
	}
	if len(params) > 0 {
		req.URL.RawQuery = params.Encode()
	}

	return send[T](c, req, endpoint)
}

// DoPrivate performs a signed POST /0/private/{method}.
func DoPrivate[T any](ctx context.Context, c *Client, endpoint string, params url.Values) (T, error) {
	var zero T

	if err := waitLimit(ctx, c.privateLimit, c.privateLimKey); err != nil {
		return zero, err
	}

	// Kraken rejects nonces that arrive out of order, so private calls
	// for one API key are sent strictly one after another.
	if err := c.nonce.lock(ctx); err != nil {
		return zero, err
	}
	defer c.nonce.unlock()

	body := url.Values{}
	maps.Copy(body, params)
	nonce := c.nonce.next()
	body.Set("nonce", nonce)
	encoded := body.Encode()

	signature, err := c.signer.Sign(endpoint, nonce, encoded)
	if err != nil {
		return zero, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.String()+endpoint, strings.NewReader(encoded))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.Header.Set("API-Key", c.apiKey)
	req.Header.Set("API-Sign", signature)

	return send[T](c, req, endpoint)
}

func waitLimit(ctx context.Context, l *limiter.Limiter, key string) error {
	if l == nil {
		return nil
	}
	for {
		r, err := l.Get(ctx, key)
		if err != nil {
			return err
		}
		if !r.Reached {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Until(time.Unix(r.Reset, 0).Add(time.Second))):
		}
	}
}

func send[T any](c *Client, req *http.Request, endpoint string) (T, error) {
	var zero T
	startTime := time.Now()

	if c.log != nil {
		c.log.Debugln(
			"[EXCHANGE-API]: Sending request",
			"exchange", exchangeName,
			"method", req.Method,
			"endpoint", endpoint,
			"query", sanitizeBody(req.URL.RawQuery),
			"headers", sanitizeHeaders(req.Header),
		)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logError(req.Method, endpoint, 0, err, startTime)
		return zero, err
	}
	defer res.Body.Close()

	bb := new(bytes.Buffer)
	if _, err = io.Copy(bb, res.Body); err != nil {
		return zero, err
	}

	env := envelope[T]{}
	if err = json.Unmarshal(bb.Bytes(), &env); err != nil {
		// Non-JSON bodies come from proxies/CDN (HTML error pages, 429, 5xx).
		if res.StatusCode >= http.StatusBadRequest {
			statusErr := errorFromStatus(res.StatusCode, bb.String())
			c.logError(req.Method, endpoint, res.StatusCode, statusErr, startTime)
			return zero, statusErr
		}
		return zero, fmt.Errorf("kraken: decode response of %s: %w", endpoint, err)
	}

	if len(env.Error) > 0 {
		apiErr := errorFromResponse(env.Error)
		c.logError(req.Method, endpoint, res.StatusCode, apiErr, startTime)
		return zero, apiErr
	}

	if res.StatusCode >= http.StatusBadRequest {
		statusErr := errorFromStatus(res.StatusCode, bb.String())
		c.logError(req.Method, endpoint, res.StatusCode, statusErr, startTime)
		return zero, statusErr
	}

	if c.log != nil {
		c.log.Debugln(
			"[EXCHANGE-API]: Request completed",
			"exchange", exchangeName,
			"method", req.Method,
			"endpoint", endpoint,
			"status_code", res.StatusCode,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
	}

	return env.Result, nil
}

func (c *Client) logError(method, endpoint string, status int, err error, startTime time.Time) {
	if c.log == nil {
		return
	}
	c.log.Errorln(
		"[EXCHANGE-API]: Request failed",
		"exchange", exchangeName,
		"method", method,
		"endpoint", endpoint,
		"status_code", status,
		"error", err.Error(),
		"duration_ms", time.Since(startTime).Milliseconds(),
	)
}
