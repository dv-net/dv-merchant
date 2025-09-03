package clients

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
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	bitgetresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/responses"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
)

func NewClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner) *Client {
	c := &Client{
		accessKey:  opt.AccessKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		passphrase: opt.PassPhrase,
		httpClient: http.DefaultClient,
		signer:     signer,
		store:      store,
	}
	return c
}

type Client struct {
	accessKey  string
	secretKey  string
	passphrase string
	baseURL    *url.URL
	httpClient *http.Client
	store      limiter.Store
	limiters   map[string]*limiter.Limiter
	signer     bitget.ISigner
}

func (o *Client) Do(ctx context.Context, method, endpoint string, private bool, dest interface{}, params ...map[string]string) error {
	if l, exists := o.limiters[endpoint]; exists {
		for {
			r, err := l.Get(ctx, utils.HashLimiterKey(endpoint, o.accessKey, o.secretKey, o.passphrase))
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

	errRes := &bitget.ErrorResponse{}
	if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
		return err
	}

	if errRes.Code != bitgetresponses.ResponseCodeOK && errRes.Msg != "" {
		return errorFromResponse(errRes)
	}

	if err = json.Unmarshal(bb.Bytes(), &dest); err != nil {
		return err
	}
	return nil
}

func S2M(i any) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(i)
	_ = json.Unmarshal(j, &m)

	return m
}

func errorFromResponse(o *bitget.ErrorResponse) error {
	if o.Code == bitgetresponses.ResponseCodeOK {
		return nil
	}
	if o.Code == bitgetresponses.ResponseCodeTemporaryDisabled {
		return exchangeclient.ErrSoftLockByUserSecurityAction
	}
	if o.Code == bitgetresponses.ResponseCodeIncorrectPermission {
		return exchangeclient.ErrIncorrectAPIPermissions
	}
	if o.Code > bitgetresponses.ResponseCodeEmptyAccessKey && o.Code < bitgetresponses.ResponseCodeInvalidPassphrase {
		return exchangeclient.ErrInvalidAPICredentials
	}
	if o.Code == bitgetresponses.ResponseCodeInvalidIP {
		return exchangeclient.ErrInvalidIPAddress
	}
	return fmt.Errorf("bitget error: %s", o.Msg)
}
