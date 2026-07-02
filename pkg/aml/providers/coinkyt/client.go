package coinkyt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/shopspring/decimal"
)

type Client struct {
	cl      *http.Client
	baseURL *url.URL
	log     logger.Logger
}

const (
	MethodCheckTransaction = "/openapi/v1/transaction"
)

func NewCoinKyt(u *url.URL, l logger.Logger) *Client {
	return &Client{
		cl: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: u,
		log:     l,
	}
}

var _ aml.Client = (*Client)(nil)

func (c *Client) Name() string {
	return "coinkyt"
}

func (c *Client) InitCheckTransaction(ctx context.Context, dto aml.InitCheckDTO, auth aml.RequestAuthorizer) (*aml.CheckResponse, error) {
	values := url.Values{
		"blockchain":  {dto.TokenData.Blockchain},
		"token":       {nativeToEmpty(dto.TokenData.ContractAddress)},
		"transaction": {dto.TxID},
	}

	return c.doRequest(ctx, http.MethodGet, MethodCheckTransaction, values, auth, aml.CheckStatusNew)
}

// nativeToEmpty converts "native" contract address to empty string as required by CoinKyt API.
func nativeToEmpty(contractAddress string) string {
	if contractAddress == "native" {
		return ""
	}
	return contractAddress
}

func (c *Client) FetchCheckStatus(_ context.Context, _ string, _ aml.RequestAuthorizer) (*aml.CheckResponse, error) {
	// CoinKyt returns the result synchronously; polling is not needed.
	return nil, fmt.Errorf("coinkyt: FetchCheckStatus not supported, result is returned synchronously by InitCheckTransaction")
}

func (c *Client) TestRequestWithAuth(ctx context.Context, auth aml.RequestAuthorizer) error {
	u := *c.baseURL
	u.Path = MethodCheckTransaction

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if err := auth.Authorize(ctx, req); err != nil {
		return fmt.Errorf("authorizing request: %w", err)
	}

	resp, err := c.cl.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.log.Warnw("error closing response body", "error", cerr)
		}
	}()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	// 401/403 means bad credentials; anything else (e.g. 400 missing params) means auth passed
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: MethodCheckTransaction,
		}
	}

	return nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, values url.Values, auth aml.RequestAuthorizer, _ aml.CheckStatus) (*aml.CheckResponse, error) {
	u := *c.baseURL
	u.Path = endpoint
	u.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if auth != nil {
		if err := auth.Authorize(ctx, req); err != nil {
			return nil, fmt.Errorf("authorizing request: %w", err)
		}
	}

	reqURL := req.URL.String()

	resp, err := c.cl.Do(req)
	if err != nil {
		c.log.Errorw("sending CoinKyt HTTP request error", "error", err)
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.log.Warnw("error closing response body", "error", cerr)
		}
	}()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorw("reading CoinKyt response error", "error", err)
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.log.Debugw("CoinKyt non-200 response", "status", resp.StatusCode, "body", string(respBodyBytes), "url", reqURL)
		return nil, &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: endpoint,
		}
	}

	var response TransactionResponse
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		c.log.Errorw("parsing CoinKyt response error", "error", err)
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	reqJSON, _ := json.Marshal(values)

	riskLevel := response.ToAMLRiskLevel()
	return &aml.CheckResponse{
		ExternalID: response.ID,
		Score:      response.RiskScore.Mul(decimal.NewFromInt(100)),
		RiskLevel:  &riskLevel,
		Status:     aml.CheckStatusSuccess,
		HTTPStatus: resp.StatusCode,
		Request:    reqJSON,
		Response:   respBodyBytes,
	}, nil
}
