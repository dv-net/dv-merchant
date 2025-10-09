package kucoin

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
	"strings"
	"time"

	"github.com/ulule/limiter/v3"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	kucoinresponse "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/responses"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/key_value"
)

type IKucoinClient interface {
	Account() IKucoinAccount
	Market() IKucoinMarket
	Public() IKucoinPublic
	Spot() IKucoinSpot
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

func NewBaseClient(opt *ClientOptions, store limiter.Store, cache key_value.IKeyValue, opts ...ClientOption) *BaseClient {
	c := &BaseClient{
		account: NewAccount(opt, store, opts...),
		market:  NewMarket(opt, store, opts...),
		public:  NewPublic(opt, store, cache, opts...),
		spot:    NewSpot(opt, store, opts...),
	}

	return c
}

type BaseClient struct {
	account IKucoinAccount
	market  IKucoinMarket
	public  IKucoinPublic
	spot    IKucoinSpot
}

func (o *BaseClient) Account() IKucoinAccount { return o.account }
func (o *BaseClient) Market() IKucoinMarket   { return o.market }
func (o *BaseClient) Public() IKucoinPublic   { return o.public }
func (o *BaseClient) Spot() IKucoinSpot       { return o.spot }

func NewClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:     opt.KeyAPI,
		secretKey:  opt.KeySecret,
		passPhrase: opt.KeyPassphrase,
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
	KeyAPI        string
	KeySecret     string
	KeyPassphrase string
	BaseURL       *url.URL
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

	req.Header.Set("KC-API-KEY-VERSION", "3")

	if private {
		timestamp, signature := o.sign(method, path, body)
		passphraseSig := o.signPassphrase()

		req.Header.Add("KC-API-KEY", o.apiKey)
		req.Header.Add("KC-API-SIGN", signature)
		req.Header.Add("KC-API-TIMESTAMP", timestamp)
		req.Header.Add("KC-API-PASSPHRASE", passphraseSig)
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

	errRes := kucoinresponse.Basic{}
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
	ts := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	s := ts + strings.ToUpper(method) + path + body
	h := hmac.New(sha256.New, []byte(o.secretKey))
	h.Write([]byte(s))
	return ts, base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (o *Client) signPassphrase() string {
	h := hmac.New(sha256.New, []byte(o.secretKey))
	h.Write([]byte(o.passPhrase))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func S2M(i interface{}) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
}

func errorFromResponse(errRes *kucoinresponse.Basic) error {
	isCredentialError := func(code int) bool {
		return (code >= kucoinresponse.ErrorCodeMissingAPICreds &&
			code <= kucoinresponse.ErrorCodeInvalidSignature && code != kucoinresponse.ErrorCodeIPWhitelist)
	}

	isWhitelistError := func(code int) bool {
		return code == kucoinresponse.ErrorCodeIPWhitelist
	}

	if errRes.Code != kucoinresponse.SuccessCodeOK {
		if isCredentialError(errRes.Code) {
			return exchangeclient.ErrInvalidAPICredentials
		}
		if isWhitelistError(errRes.Code) {
			return exchangeclient.ErrInvalidIPAddress
		}
		if errRes.Code == kucoinresponse.ErrorWithdrawalTooFast {
			return exchangeclient.ErrRateLimited
		}
		// Convert msg to string and wrap error
		msgStr := fmt.Sprintf("%v", errRes.Msg)
		return wrapKuCoinError(msgStr, errRes.Code)
	}

	return nil
}
