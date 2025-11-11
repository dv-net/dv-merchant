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
	"github.com/dv-net/dv-merchant/pkg/logger"
)

func NewClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner, opts ...SubClientOption) *Client {
	c := &Client{
		accessKey:  opt.AccessKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		passphrase: opt.PassPhrase,
		httpClient: http.DefaultClient,
		signer:     signer,
		store:      store,
	}
	for _, o := range opts {
		o(c)
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
	log        logger.Logger
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
	startTime := time.Now()
	baseURL := o.baseURL.String() + path
	var (
		req  *http.Request
		err  error
		j    []byte
		body string
	)

	if o.log != nil {
		o.log.Debugln("[EXCHANGE-API]: Preparing request",
			"exchange", "bitget",
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
		req, err = o.signer.SignRequest(ctx, req)
		if err != nil {
			return err
		}
	}

	if o.log != nil {
		o.log.Debugln("[EXCHANGE-API]: Sending request",
			"exchange", "bitget",
			"method", method,
			"url", o.baseURL.String()+path,
			"headers", sanitizeHeaders(req.Header),
			"body", sanitizeBody(body),
		)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		if o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: Request failed",
				"exchange", "bitget",
				"method", method,
				"endpoint", path,
				"error", err.Error(),
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
		}
		return err
	}
	defer resp.Body.Close()

	bb := new(bytes.Buffer)
	_, err = io.Copy(bb, resp.Body)
	if err != nil {
		return err
	}

	duration := time.Since(startTime)

	errRes := &bitget.ResponseError{}
	if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
		return err
	}

	if errRes.Code != bitgetresponses.ResponseCodeOK && errRes.Msg != "" {
		if o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: API error response",
				"exchange", "bitget",
				"method", method,
				"endpoint", path,
				"status_code", resp.StatusCode,
				"error", errorFromResponse(errRes).Error(),
				"duration_ms", duration.Milliseconds(),
			)
		}
		return errorFromResponse(errRes)
	}

	if o.log != nil {
		o.log.Debugln("[EXCHANGE-API]: Request completed",
			"exchange", "bitget",
			"method", method,
			"endpoint", path,
			"status_code", resp.StatusCode,
			"duration_ms", duration.Milliseconds(),
		)
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

// sanitizeHeaders removes sensitive headers for logging
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for k, v := range headers {
		if strings.Contains(strings.ToLower(k), "key") ||
			strings.Contains(strings.ToLower(k), "sign") ||
			strings.Contains(strings.ToLower(k), "passphrase") {
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

func errorFromResponse(o *bitget.ResponseError) error {
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
