package bingx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/logger"
)

const defaultTimeout = 30 * time.Second

type IBingxClient interface {
	Account() IBingxAccount
	Market() IBingxMarket
	Spot() IBingxSpot
	Wallet() IBingxWallet
}

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*BaseClient, error) {
	return &BaseClient{
		accountClient: NewAccountClient(opt, store, opts...),
		marketClient:  NewMarketClient(opt, store, opts...),
		spotClient:    NewSpotClient(opt, store, opts...),
		walletClient:  NewWalletClient(opt, store, opts...),
	}, nil
}

type BaseClient struct {
	accountClient IBingxAccount
	marketClient  IBingxMarket
	spotClient    IBingxSpot
	walletClient  IBingxWallet
}

func (o *BaseClient) Account() IBingxAccount { return o.accountClient }
func (o *BaseClient) Market() IBingxMarket   { return o.marketClient }
func (o *BaseClient) Spot() IBingxSpot       { return o.spotClient }
func (o *BaseClient) Wallet() IBingxWallet   { return o.walletClient }

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

func NewClient(opt *ClientOptions, _ limiter.Store, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:     opt.APIKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		signer:     NewSigner(opt.APIKey, opt.SecretKey),
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
	intervals  map[string]time.Duration
	signer     ISigner
	log        logger.Logger
}

const (
	pacerMaxKeys = 256
	pacerKeyTTL  = time.Hour
)

var (
	pacerMu   sync.Mutex
	pacerSlot = make(map[string]time.Time)
)

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (o *Client) Do(ctx context.Context, method, endpoint string, private bool, dest any, params ...map[string]string) error {
	if err := o.waitTurn(ctx, endpoint); err != nil {
		return err
	}

	return o.DoPlain(ctx, method, endpoint, private, dest, params...)
}

func (o *Client) waitTurn(ctx context.Context, endpoint string) error {
	interval, exists := o.intervals[endpoint]
	if !exists || interval <= 0 {
		return nil
	}

	slot := reserveSlot(o.apiKey+"|"+endpoint, interval)

	wait := time.Until(slot)
	if wait <= 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func reserveSlot(key string, interval time.Duration) time.Time {
	pacerMu.Lock()
	defer pacerMu.Unlock()

	now := time.Now()
	if len(pacerSlot) > pacerMaxKeys {
		for k, v := range pacerSlot {
			if v.Before(now.Add(-pacerKeyTTL)) {
				delete(pacerSlot, k)
			}
		}
	}

	slot := pacerSlot[key].Add(interval)
	if slot.Before(now) {
		slot = now
	}
	pacerSlot[key] = slot

	return slot
}

func (o *Client) DoPlain(ctx context.Context, method, endpoint string, private bool, dest any, params ...map[string]string) error {
	startTime := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, o.baseURL.String()+endpoint, http.NoBody)
	if err != nil {
		return err
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params[0] {
			q.Set(k, v)
		}
		req.URL.RawQuery = plainQuery(q)
	}

	req = o.signer.SignRequest(ctx, req, private)
	req.Header.Set("Content-Type", "application/json")

	if o.log != nil {
		o.log.Debugln(
			"[EXCHANGE-API]: Sending request",
			"exchange", "bingx",
			"method", method,
			"endpoint", endpoint,
			"query", sanitizeBody(req.URL.RawQuery),
			"headers", sanitizeHeaders(req.Header),
		)
	}

	res, err := o.httpClient.Do(req)
	if err != nil {
		if o.log != nil {
			o.log.Errorln(
				"[EXCHANGE-API]: Request failed",
				"exchange", "bingx",
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
		return fmt.Errorf("bingx error: status %d, body: %s", res.StatusCode, sanitizeBody(bb.String()))
	}

	body := bytes.TrimSpace(bb.Bytes())
	if len(body) == 0 {
		return nil
	}

	if body[0] == '[' {
		if dest == nil {
			return nil
		}
		return json.Unmarshal(body, dest)
	}

	env := envelope{}
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("bingx: decode envelope: %w, body: %s", err, sanitizeBody(bb.String()))
	}

	if env.Code != successCode {
		apiErr := errorFromResponse(&ErrorResponse{Code: env.Code, Msg: env.Msg})
		if o.log != nil {
			o.log.Errorln(
				"[EXCHANGE-API]: API error response",
				"exchange", "bingx",
				"method", method,
				"endpoint", endpoint,
				"error", apiErr.Error(),
				"duration_ms", duration.Milliseconds(),
			)
		}
		return apiErr
	}

	if o.log != nil {
		o.log.Debugln(
			"[EXCHANGE-API]: Request completed",
			"exchange", "bingx",
			"method", method,
			"endpoint", endpoint,
			"duration_ms", duration.Milliseconds(),
		)
	}

	if dest == nil || len(env.Data) == 0 {
		return nil
	}

	return json.Unmarshal(env.Data, dest)
}
