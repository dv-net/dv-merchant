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

	if res.StatusCode >= http.StatusBadRequest {
		return exchangeclient.ErrInvalidAPICredentials
	}

	var apiErr APIError
	err = json.Unmarshal(bb.Bytes(), &apiErr)
	if err != nil {
		return fmt.Errorf("error parsing response: %w, response: %s", err, bb.String())
	}

	if apiErr.Code != 0 {
		if err := errorFromResponse(&apiErr); err != nil {
			return err
		}
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
