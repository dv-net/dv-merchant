package aml_bot

import (
	"bytes"
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

const (
	endpointCheck   = "/"
	endpointRecheck = "/recheck/"
	endpointHistory = "/history/"
)

// Client is an AMLBot AML service client.
type Client struct {
	baseURL *url.URL
	log     logger.Logger
	client  *http.Client
}

// NewAMBot creates a new AMLBot client.
func NewAMBot(u *url.URL, l logger.Logger) *Client {
	return &Client{
		baseURL: u,
		log:     l,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

var _ aml.Client = (*Client)(nil)

// Name returns the name of the AML provider.
func (c *Client) Name() string {
	return "amlbot"
}

// InitCheckTransaction initiates a transaction check with AMLBot.
func (c *Client) InitCheckTransaction(ctx context.Context, dto aml.InitCheckDTO, auth aml.RequestAuthorizer) (*aml.CheckResponse, error) {
	// Prepare form data
	values := url.Values{
		"hash":      {dto.TxID},
		"asset":     {dto.TokenData.Blockchain},
		"direction": {DirectionFromAML(dto.Direction).String()},
		"address":   {dto.OutputAddress},
		"locale":    {"en"},
		"flow":      {"advanced"},
	}

	return c.doRequest(ctx, http.MethodPost, endpointCheck, values, auth, aml.CheckStatusNew)
}

// FetchCheckStatus fetches the status of a previous check.
func (c *Client) FetchCheckStatus(ctx context.Context, checkID string, auth aml.RequestAuthorizer) (*aml.CheckResponse, error) {
	// Prepare form data
	values := url.Values{
		"uid":    {checkID},
		"locale": {"en"},
	}

	return c.doRequest(ctx, http.MethodPost, endpointRecheck, values, auth, aml.CheckStatusNew)
}

// TestRequestWithAuth sends request with required auth for creds test
func (c *Client) TestRequestWithAuth(ctx context.Context, auth aml.RequestAuthorizer) error {
	// Prepare empty form data
	values := url.Values{}
	values.Set("page", "1")

	// Perform request
	req, _, err := c.prepareRequest(ctx, http.MethodPost, endpointHistory, values, auth)
	if err != nil {
		return fmt.Errorf("preparing request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("sending AML HTTP request error", err)
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.log.Warn("error closing response body", "error", cerr)
		}
	}()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("parsing AML response error", err)
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: req.URL.Path,
		}
	}

	var response ErrorResponse
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		c.log.Error("parsing AML response error", err)
		return fmt.Errorf("decoding response: %w", err)
	}

	if !response.Result {
		return &aml.RequestFailedError{
			StatusCode: http.StatusOK,
			Body:       respBodyBytes,
			RequestURL: req.URL.Path,
		}
	}

	return nil
}

// doRequest performs HTTP request
func (c *Client) doRequest(ctx context.Context, method, endpoint string, values url.Values, auth aml.RequestAuthorizer, defaultStatus aml.CheckStatus) (*aml.CheckResponse, error) {
	req, reqBodyBytes, err := c.prepareRequest(ctx, method, endpoint, values, auth)
	if err != nil {
		return nil, fmt.Errorf("preparing request: %w", err)
	}

	jsonRequest, err := QueryToJSON(string(reqBodyBytes))
	if err != nil {
		c.log.Error("error converting request to JSON", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("sending AML HTTP request error", err)
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.log.Warn("error closing response body", "error", cerr)
		}
	}()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("parsing AML response error", err)
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: req.URL.Path,
		}
	}

	var response Response
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		c.log.Error("parsing AML response error", err)
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if !response.Result {
		c.log.Debug("AML API error", "url", req.URL.String(), "description", response.Description)
		return nil, &aml.RequestFailedError{
			StatusCode: http.StatusOK,
			Body:       respBodyBytes,
			RequestURL: req.URL.Path,
		}
	}

	// Check if response data exists
	if response.Data == nil {
		c.log.Debug("AML response has no data", "url", req.URL.String())
		return nil, &aml.RequestFailedError{
			StatusCode: http.StatusOK,
			Body:       respBodyBytes,
			RequestURL: req.URL.Path,
		}
	}

	status := CheckStatus(response.Data.Status).ToAMLStatus()
	if status == "" {
		status = defaultStatus
	}

	riskScore, riskLevel := prepareRiskData(response.Data.RiskScore, status)

	return &aml.CheckResponse{
		ExternalID: response.Data.UID,
		Score:      riskScore,
		RiskLevel:  &riskLevel,
		Status:     status,
		HTTPStatus: resp.StatusCode,
		Request:    []byte(jsonRequest),
		Response:   respBodyBytes,
	}, nil
}

func QueryToJSON(query string) (string, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("parsing AML request query: %w", err)
	}

	data := make(map[string]interface{})
	for key, val := range values {
		if len(val) > 0 {
			data[key] = val[0]
		}
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("serializing JSON query: %w", err)
	}

	return string(jsonBytes), nil
}

// prepareRequest prepares an HTTP request with form data and authorization.
func (c *Client) prepareRequest(ctx context.Context, httpMethod, endpoint string, values url.Values, auth aml.RequestAuthorizer) (*http.Request, []byte, error) {
	// Encode form data
	bodyBytes := []byte(values.Encode())

	// Create HTTP request
	u := *c.baseURL
	u.Path = endpoint
	req, err := http.NewRequestWithContext(ctx, httpMethod, u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	// Authorize request
	if auth != nil {
		if err := auth.Authorize(ctx, req); err != nil {
			return nil, nil, fmt.Errorf("authorizing request: %w", err)
		}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, bodyBytes, nil
}

// prepareRiskData converts risk score to 100-percentage system and determines generic risk level
func prepareRiskData(score decimal.Decimal, currentStatus aml.CheckStatus) (decimal.Decimal, aml.CheckRiskLevel) {
	// Convert score to 100-percentage system
	percentageScore := score.Mul(decimal.NewFromInt(100))

	// For pending checks with zeroed risk - is undefined
	if currentStatus == aml.CheckStatusNew && score.IsZero() {
		return percentageScore, aml.CheckRiskLevelUndefined
	}

	switch {
	case score.Equal(decimal.Zero):
		return percentageScore, aml.CheckRiskLevelNone
	case score.LessThanOrEqual(decimal.NewFromFloat(0.20)):
		return percentageScore, aml.CheckRiskLevelLow
	case score.LessThanOrEqual(decimal.NewFromFloat(0.79)):
		return percentageScore, aml.CheckRiskLevelMedium
	default:
		return percentageScore, aml.CheckRiskLevelSevere
	}
}
