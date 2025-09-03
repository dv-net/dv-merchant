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
)

type IGateClient interface {
	Wallet() IGateWallet
}

func NewBaseClient(opt *ClientOptions, store limiter.Store) (*BaseClient, error) {
	c := &BaseClient{
		opts:          opt,
		accountClient: NewAccountClient(opt, store),
		walletClient:  NewWalletClient(opt, store),
		spotClient:    NewSpotClient(opt, store),
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

func NewClient(opt *ClientOptions, store limiter.Store) *Client {
	c := &Client{
		accessKey:  opt.AccessKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		httpClient: http.DefaultClient,
		signer:     NewSigner(opt.AccessKey, opt.SecretKey),
		store:      store,
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
	baseURL := o.baseURL.String() + path
	var (
		req  *http.Request
		err  error
		j    []byte
		body string
	)
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
			body = "" //nolint:ineffassign
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

	// Check for errors
	if res.StatusCode >= 400 {
		errRes := ErrorResponse{}
		if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
			return fmt.Errorf("gate.io error: status %d, body: %s", res.StatusCode, bb.String())
		}
		return errorFromResponse(&errRes)
	}

	err = json.Unmarshal(bb.Bytes(), &dest)
	return err
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
		return fmt.Errorf("gate.io error: %s - %s", errRes.Label, errRes.Message)
	default:
		return fmt.Errorf("gate.io error: %s - %s", errRes.Label, errRes.Message)
	}
}

func S2M(i interface{}) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
}
