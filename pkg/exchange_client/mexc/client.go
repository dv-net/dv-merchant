package mexc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IMexcClient interface {
	Account() IMexcAccount
	Wallet() IMexcWallet
	Market() IMexcMarket
	Spot() IMexcSpot
}

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*BaseClient, error) {
	return &BaseClient{
		accountClient: NewAccountClient(opt, store, opts...),
		walletClient:  NewWalletClient(opt, store, opts...),
		marketClient:  NewMarketClient(opt, store, opts...),
		spotClient:    NewSpotClient(opt, store, opts...),
	}, nil
}

type BaseClient struct {
	accountClient IMexcAccount
	walletClient  IMexcWallet
	marketClient  IMexcMarket
	spotClient    IMexcSpot
}

func (o *BaseClient) Account() IMexcAccount { return o.accountClient }
func (o *BaseClient) Wallet() IMexcWallet   { return o.walletClient }
func (o *BaseClient) Market() IMexcMarket   { return o.marketClient }
func (o *BaseClient) Spot() IMexcSpot       { return o.spotClient }

type ClientOption func(c *Client)

func WithLogger(log logger.Logger) ClientOption {
	return func(c *Client) {
		c.log = log
	}
}

type ClientOptions struct {
	APIKey    string
	SecretKey string
	BaseURL   *url.URL
}

func NewClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:     opt.APIKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		httpClient: http.DefaultClient,
		signer:     NewSigner(opt.APIKey, opt.SecretKey),
		store:      store,
	}
	for _, o := range opts {
		o(c)
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
	signer     ISigner
	log        logger.Logger
}

func (o *Client) Do(ctx context.Context, method, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	if l, exists := o.limiters[endpoint]; exists {
		for {
			r, err := l.Get(ctx, utils.HashLimiterKey(endpoint, o.apiKey, o.secretKey))
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

func (o *Client) DoPlain(ctx context.Context, method, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	startTime := time.Now()

	if o.log != nil {
		o.log.Debugln(
			"[EXCHANGE-API]: Preparing request",
			"exchange", "mexc",
			"method", method,
			"endpoint", endpoint,
			"private", private,
		)
	}

	req, err := http.NewRequestWithContext(ctx, method, o.baseURL.String()+endpoint, http.NoBody)
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

	req = o.signer.SignRequest(ctx, req, private)
	req.Header.Set("Content-Type", "application/json")

	if o.log != nil {
		o.log.Debugln(
			"[EXCHANGE-API]: Sending request",
			"exchange", "mexc",
			"method", method,
			"url", o.baseURL.String()+endpoint,
			"query", sanitizeBody(req.URL.RawQuery),
			"headers", sanitizeHeaders(req.Header),
		)
	}

	res, err := o.httpClient.Do(req)
	if err != nil {
		if o.log != nil {
			o.log.Errorln(
				"[EXCHANGE-API]: Request failed",
				"exchange", "mexc",
				"method", method,
				"endpoint", endpoint,
				"error", err.Error(),
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
		}
		return err
	}
	defer res.Body.Close()

	bb := new(bytes.Buffer)
	if _, err = io.Copy(bb, res.Body); err != nil {
		return err
	}

	duration := time.Since(startTime)

	if res.StatusCode >= 400 {
		errRes := ErrorResponse{}
		if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil || errRes.Code == 0 {
			statusErr := errorFromStatus(res.StatusCode, bb.String())
			if o.log != nil {
				o.log.Errorln(
					"[EXCHANGE-API]: API error response",
					"exchange", "mexc",
					"method", method,
					"endpoint", endpoint,
					"status_code", res.StatusCode,
					"error", statusErr.Error(),
					"duration_ms", duration.Milliseconds(),
				)
			}
			return statusErr
		}
		if o.log != nil {
			o.log.Errorln(
				"[EXCHANGE-API]: API error response",
				"exchange", "mexc",
				"method", method,
				"endpoint", endpoint,
				"status_code", res.StatusCode,
				"error", errorFromResponse(&errRes).Error(),
				"duration_ms", duration.Milliseconds(),
			)
		}
		return errorFromResponse(&errRes)
	}

	if o.log != nil {
		o.log.Debugln(
			"[EXCHANGE-API]: Request completed",
			"exchange", "mexc",
			"method", method,
			"endpoint", endpoint,
			"status_code", res.StatusCode,
			"duration_ms", duration.Milliseconds(),
		)
	}

	return json.Unmarshal(bb.Bytes(), dest)
}
