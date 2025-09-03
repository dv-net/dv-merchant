package okx

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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ulule/limiter/v3"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	okxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/responses"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
)

type IOKXClient interface {
	Account() IOKXAccount
	Market() IOKXMarket
	Public() IOKXPublic
	Funding() IOKXFunding
	Trade() IOKXTrade
	SubAccount() IOKXSubAccount
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

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *BaseClient {
	c := &BaseClient{
		account:    NewAccount(opt, store, opts...),
		market:     NewMarket(opt, store, opts...),
		public:     NewPublicData(opt, store, opts...),
		funding:    NewFunding(opt, store, opts...),
		trade:      NewTrade(opt, store, opts...),
		subAccount: NewSubAccount(opt, store, opts...),
	}

	return c
}

type BaseClient struct {
	account    IOKXAccount
	market     IOKXMarket
	public     IOKXPublic
	funding    IOKXFunding
	trade      IOKXTrade
	subAccount IOKXSubAccount
}

func (o *BaseClient) Account() IOKXAccount       { return o.account }
func (o *BaseClient) Market() IOKXMarket         { return o.market }
func (o *BaseClient) Public() IOKXPublic         { return o.public }
func (o *BaseClient) Funding() IOKXFunding       { return o.funding }
func (o *BaseClient) Trade() IOKXTrade           { return o.trade }
func (o *BaseClient) SubAccount() IOKXSubAccount { return o.subAccount }

func NewClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:     opt.APIKey,
		secretKey:  opt.SecretKey,
		passPhrase: opt.Passphrase,
		baseURL:    opt.BaseURL,
		httpClient: http.DefaultClient,
		store:      store,
	}

	for _, option := range opts {
		option(c)
	}

	return c
}

type ClientOptions struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	BaseURL    *url.URL
}

type Client struct {
	apiKey     string
	secretKey  string
	passPhrase string
	baseURL    *url.URL
	httpClient *http.Client
	store      limiter.Store
	limiters   map[string]*limiter.Limiter
}

func (o *Client) Do(ctx context.Context, method string, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	if l, exists := o.limiters[endpoint]; exists {
		for {
			r, err := l.Get(ctx, utils.HashLimiterKey(endpoint, o.apiKey, o.secretKey, o.passPhrase))
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
				path += "?" + req.URL.RawQuery
			}
		}
	case http.MethodPost:
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
		req.Header.Add("Content-Type", "application/json")
	default:
		return fmt.Errorf("unsupported method %s", method)
	}
	if private {
		timestamp, sign := o.sign(method, path, body)
		req.Header.Add("OK-ACCESS-KEY", o.apiKey)
		req.Header.Add("OK-ACCESS-PASSPHRASE", o.passPhrase)
		req.Header.Add("OK-ACCESS-SIGN", sign)
		req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	}
	res, err := o.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bb := new(bytes.Buffer)
	_, err = io.Copy(bb, res.Body)
	if err != nil {
		return err
	}
	errRes := okxresponses.Basic{}
	if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
		return err
	}
	if err := errorFromResponse(&errRes); err != nil {
		return err
	}
	err = json.Unmarshal(bb.Bytes(), &dest)
	return err
}

func (o *Client) sign(method, path, body string) (string, string) {
	format := "2006-01-02T15:04:05.999Z07:00"
	t := time.Now().UTC().Format(format)
	ts := fmt.Sprint(t)
	s := ts + method + path + body
	p := []byte(s)
	h := hmac.New(sha256.New, []byte(o.secretKey))
	h.Write(p)
	return ts, base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func S2M(i interface{}) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
}

func errorFromResponse(errRes *okxresponses.Basic) error {
	isCredentialError := func(code int) bool {
		return (code >= okxresponses.ErrorCodeInvalidAccessKey &&
			code <= okxresponses.ErrorCodeAPIKeyNotExists && code != okxresponses.ErrorCodeIPWhitelist) || code == okxresponses.ErrorCodeAPIKeyNotExists
	}

	isWhitelistError := func(code int) bool {
		return code == okxresponses.ErrorCodeIPWhitelist
	}

	if strCode, ok := errRes.Code.(string); ok && strCode != "0" { //nolint:nestif
		if iCode, err := strconv.Atoi(strCode); err == nil {
			if isCredentialError(iCode) {
				return exchangeclient.ErrInvalidAPICredentials
			}
			if isWhitelistError(iCode) {
				return exchangeclient.ErrInvalidIPAddress
			}
		}
		return fmt.Errorf("okx error: %s (%s)", errRes.Msg, errRes.Code)
	}

	if intCode, ok := errRes.Code.(int); ok && intCode != 0 {
		if isCredentialError(intCode) {
			return exchangeclient.ErrInvalidAPICredentials
		}
		if isWhitelistError(intCode) {
			return exchangeclient.ErrInvalidIPAddress
		}
		return fmt.Errorf("okx error: %s (%s)", errRes.Msg, errRes.Code)
	}

	return nil
}
