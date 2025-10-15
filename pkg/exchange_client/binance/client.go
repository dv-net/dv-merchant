package binance

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

	"github.com/go-playground/validator/v10"
	"github.com/google/go-querystring/query"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	binanceresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/responses"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IBinanceClient interface {
	MarketData() IMarketClient
	Wallet() IWalletClient
	Spot() ISpotClient
}

func NewBaseClient(opt *ClientOptions) (*BaseClient, error) {
	v := validator.New()
	if err := v.Struct(opt); err != nil {
		return nil, err
	}

	c := &BaseClient{}

	marketDataClient, err := NewMarketData(opt)
	if err != nil {
		return nil, err
	}
	walletClient, err := NewWallet(opt)
	if err != nil {
		return nil, err
	}
	spotClient, err := NewSpotClient(opt)
	if err != nil {
		return nil, err
	}

	c.market = marketDataClient
	c.wallet = walletClient
	c.spot = spotClient

	return c, nil
}

type BaseClient struct {
	market IMarketClient
	wallet IWalletClient
	spot   ISpotClient
}

func (o *BaseClient) MarketData() IMarketClient { return o.market }
func (o *BaseClient) Wallet() IWalletClient     { return o.wallet }
func (o *BaseClient) Spot() ISpotClient         { return o.spot }

func NewClient(opt *ClientOptions) (*Client, error) {
	v := validator.New()
	if err := v.Struct(opt); err != nil {
		return nil, err
	}

	client := &Client{
		validator:  v,
		apiKey:     opt.APIKey,
		secretKey:  opt.SecretKey,
		baseURL:    opt.BaseURL,
		httpClient: http.DefaultClient,
		signer:     NewSigner(opt.APIKey, opt.SecretKey),
	}

	return client, nil
}

type Client struct {
	apiKey     string
	secretKey  string
	baseURL    *url.URL
	httpClient *http.Client
	validator  *validator.Validate
	signer     ISigner
	log        logger.Logger
	logEnabled bool
}

type ClientOptions struct {
	APIKey       string   `validate:"required_if=PublicClient false"`
	SecretKey    string   `validate:"required_if=PublicClient false"`
	BaseURL      *url.URL `validate:"required_if=PublicClient false"`
	PublicClient bool
}

func (o *Client) Do(ctx context.Context, req *http.Request, level SecurityLevel, dest interface{}) error {
	startTime := time.Now()

	if o.logEnabled && o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Preparing request",
			"exchange", "binance",
			"method", req.Method,
			"endpoint", req.URL.Path,
			"security_level", level,
		)
	}

	req = o.signer.SignRequest(ctx, req, level)

	if o.logEnabled && o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Sending request",
			"exchange", "binance",
			"method", req.Method,
			"url", req.URL.String(),
			"headers", sanitizeHeaders(req.Header),
		)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		if o.logEnabled && o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: Request failed",
				"exchange", "binance",
				"method", req.Method,
				"endpoint", req.URL.Path,
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

	if resp.StatusCode != http.StatusOK {
		errRes := binanceresponses.ErrorResponse{}
		if err = json.Unmarshal(bb.Bytes(), &errRes); err != nil {
			return err
		}
		if o.logEnabled && o.log != nil {
			o.log.Errorln("[EXCHANGE-API]: API error response",
				"exchange", "binance",
				"method", req.Method,
				"endpoint", req.URL.Path,
				"status_code", resp.StatusCode,
				"error", errorFromResponse(&errRes).Error(),
				"duration_ms", duration.Milliseconds(),
			)
		}
		return errorFromResponse(&errRes)
	}

	// StatusTeapot is a custom status code that indicated IP being banned for Retry-After duration.
	// StatusTooManyRequests indicates that API forces us to stop and come back after Retry-After duration.
	if resp.StatusCode == http.StatusTeapot || resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return fmt.Errorf("IP is banned for %s seconds", retryAfter)
		// TODO: Implement retry logic when it will eventually be needed.
	}

	if o.logEnabled && o.log != nil {
		o.log.Infoln("[EXCHANGE-API]: Request completed",
			"exchange", "binance",
			"method", req.Method,
			"endpoint", req.URL.Path,
			"status_code", resp.StatusCode,
			"duration_ms", duration.Milliseconds(),
		)
	}

	if err = json.Unmarshal(bb.Bytes(), dest); err != nil {
		return err
	}
	return nil
}

func (o *Client) assembleRequest(dto interface{}, req *http.Request) (*http.Request, error) {
	switch req.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
		if err := o.validator.Struct(dto); err != nil {
			return nil, err
		}
		qVals, err := query.Values(dto)
		if err != nil {
			return nil, err
		}
		req.URL.RawQuery = qVals.Encode()
		return req, nil
	default:
		return nil, fmt.Errorf("unsupported method %s", req.Method)
	}
}

// sanitizeHeaders removes sensitive headers for logging
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for k, v := range headers {
		if strings.Contains(strings.ToLower(k), "key") ||
			strings.Contains(strings.ToLower(k), "sign") ||
			strings.Contains(strings.ToLower(k), "signature") {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = strings.Join(v, ",")
		}
	}
	return sanitized
}

func errorFromResponse(o *binanceresponses.ErrorResponse) error {
	if o.Code == binanceresponses.ResponseCodeInvalidSecretKey || o.Code == binanceresponses.ResponseCodeInvalidAPIKey {
		return exchangeclient.ErrInvalidAPICredentials
	}
	if o.Code == binanceresponses.ResponseCodeInvalidPermissionOnAPI {
		return exchangeclient.ErrIncorrectAPIPermissions
	}
	if o.Code == binanceresponses.ResponseCodeInvalidPermissionsOrIP {
		return exchangeclient.ErrInvalidIPAddress
	}
	return fmt.Errorf("binance error: %s (%d)", o.Msg, o.Code)
}
