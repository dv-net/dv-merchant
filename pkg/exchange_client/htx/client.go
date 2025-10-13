package htx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"

	"github.com/ulule/limiter/v3"
)

type IHTXClient interface {
	Account() IHTXAccount
	Common() IHTXCommon
	Market() IHTXMarket
	Order() IHTXOrder
	User() IHTXUser
	Wallet() IHTXWallet
}

type ClientOption func(c *Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithCustomBaseURL(baseURL *url.URL) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*BaseClient, error) {
	c := &BaseClient{
		opts:          opt,
		accountClient: NewAccountClient(opt, store, opts...),
		marketClient:  NewMarketClient(opt, store, opts...),
		orderClient:   NewOrderClient(opt, store, opts...),
		commonClient:  NewCommonClient(opt, store, opts...),
		userClient:    NewUserClient(opt, store, opts...),
		walletClient:  NewWalletClient(opt, store, opts...),
	}
	return c, nil
}

type BaseClient struct {
	opts          *ClientOptions
	accountClient IHTXAccount
	marketClient  IHTXMarket
	orderClient   IHTXOrder
	commonClient  IHTXCommon
	userClient    IHTXUser
	walletClient  IHTXWallet
}

func (o *BaseClient) Account() IHTXAccount { return o.accountClient }
func (o *BaseClient) Market() IHTXMarket   { return o.marketClient }
func (o *BaseClient) Order() IHTXOrder     { return o.orderClient }
func (o *BaseClient) Common() IHTXCommon   { return o.commonClient }
func (o *BaseClient) User() IHTXUser       { return o.userClient }
func (o *BaseClient) Wallet() IHTXWallet   { return o.walletClient }
func (o *BaseClient) AccessKey() string    { return o.opts.AccessKey }

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
}

func (o *Client) Do(ctx context.Context, method, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	for {
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
		err := o.DoPlain(ctx, method, endpoint, private, dest, params...)
		if err != nil && errors.Is(err, htxmodels.ErrHtxRateLimitExceeded) {
			// Server-side rate limit hit, wait and retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(3 * time.Second):
				continue
			}
		}
		return err
	}
}

func (o *Client) DoPlain(ctx context.Context, method, path string, private bool, dest interface{}, params ...map[string]string) error {
	baseURL := o.baseURL.String() + path
	var (
		req  *http.Request
		err  error
		j    []byte
		body string
	)
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
				path += "?" + req.URL.RawQuery //nolint:ineffassign
			}
		}
	case http.MethodPost:
		j, err = json.Marshal(params[0])
		if err != nil {
			return err
		}
		body = string(j)
		if body == "{}" {
			body = "" //nolint:ineffassign
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
		req, err = o.signer.SignRequest(ctx, req)
		if err != nil {
			return err
		}
	}
	resp, err := o.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bb := new(bytes.Buffer)
	_, err = io.Copy(bb, resp.Body)
	if err != nil {
		return err
	}
	errRes := ResponseV1Error{}
	if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
		return err
	}
	if errRes.Status == "" && errRes.StatusCode != http.StatusOK {
		errRes := ResponseV2Error{}
		if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
			return err
		}
		return errorFromResponse(errRes, "v2")
	}
	if errRes.Status != htxresponses.StatusOK {
		if errRes.StatusCode != http.StatusOK || errRes.ErrMsg != "" {
			return errorFromResponse(errRes, "v1")
		}
	}

	if err = json.Unmarshal(bb.Bytes(), &dest); err != nil {
		return err
	}
	return nil
}

func (o *Client) AccessKey() string { return o.accessKey }
func (o *Client) SecretKey() string { return o.secretKey }

func S2M(i interface{}) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
}

func errorFromResponse(err any, version string) error {
	if version == "v1" { //nolint:nestif
		if errRes, ok := err.(ResponseV1Error); ok {
			if strings.Contains(errRes.ErrMsg, "Incorrect IP address") {
				return exchangeclient.ErrInvalidIPAddress
			}
			if strings.Contains(errRes.ErrMsg, "Verification failure") {
				return exchangeclient.ErrInvalidAPICredentials
			}
			if errRes.ErrCode == "rate-too-many-requests" {
				return fmt.Errorf("htx error: %w", htxmodels.ErrHtxRateLimitExceeded)
			}
			return wrapHtxError(fmt.Sprintf("%v", errRes.ErrCode), fmt.Sprintf("%v", errRes.ErrMsg))
		}
	}
	if version == "v2" {
		if errRes, ok := err.(ResponseV2Error); ok {
			if errRes.Code == 12005 {
				return exchangeclient.ErrInvalidIPAddress
			}
			return wrapHtxV2Error(errRes.Code, errRes.Message)
		}
	}
	return errors.New("unknown htx error")
}

// wrapHtxError wraps HTX v1 errors with centralized errors when applicable
func wrapHtxError(code, msg string) error {
	msgLower := strings.ToLower(msg)

	// Check for withdrawal confirmation limit or unsafe deposit errors
	if strings.Contains(msgLower, "withdrawal confirmation limit") ||
		strings.Contains(code, "dw-withdraw-unsafe-deposit-only") {
		return fmt.Errorf("htx error: %s: %w", code, exchangeclient.ErrWithdrawalBalanceLocked)
	}

	return fmt.Errorf("htx error: %s", code)
}

// wrapHtxV2Error wraps HTX v2 errors with centralized errors when applicable
func wrapHtxV2Error(code int, msg string) error {
	return fmt.Errorf("htx error: %s, %d", msg, code)
}
