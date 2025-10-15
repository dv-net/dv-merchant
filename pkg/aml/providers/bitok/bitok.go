package bitok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

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
	MethodManualCheckTransfer = "/v1/manual-checks/check-transfer/"
	MethodManualCheck         = "/v1/manual-checks/"
	MethodSupportedCoins      = "/v1/basics/networks/"
)

func NewBitOK(u *url.URL, l logger.Logger) *Client {
	return &Client{
		cl:      http.DefaultClient,
		baseURL: u,
		log:     l,
	}
}

type ScoreRequestDTO struct {
	TxHash       string
	Blockchain   string
	CurrencyCode string
}

type APICredentialsDTO struct {
	Secret    string
	AccessKey string
}

// TestRequestWithAuth sends test request with required auth for credentials test
func (b *Client) TestRequestWithAuth(ctx context.Context, auth aml.RequestAuthorizer) error {
	// Prepare the GET request (no body for GET)
	httpReq, _, err := b.prepareRequest(ctx, http.MethodGet, MethodSupportedCoins, nil, auth)
	if err != nil {
		return fmt.Errorf("prepare request: %w", err)
	}

	// Send the request
	resp, err := b.cl.Do(httpReq)
	if err != nil {
		b.log.Errorw("sending BitOK HTTP request error", "error", err)
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			b.log.Warnw("failed to close response body", "error", cerr)
		}
	}()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		b.log.Errorw("parsing BitOK response error", "error", err)
		return fmt.Errorf("read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		b.log.Debugw(
			"failed BitOK networks request",
			"http_code", resp.StatusCode,
			"body", string(respBodyBytes),
		)
		return &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: httpReq.URL.Path,
		}
	}

	// Decode response
	var response map[string]interface{}
	if err := json.Unmarshal(respBodyBytes, &response); err != nil {
		b.log.Errorw("parsing BitOK response error", "error", err)
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (b *Client) InitCheckTransaction(ctx context.Context, dto aml.InitCheckDTO, auth aml.RequestAuthorizer) (*aml.CheckResponse, error) {
	reqBody := &CheckTransferRequest{
		Network:       dto.TokenData.Blockchain,
		TokenID:       dto.TokenData.ContractAddress,
		TxHash:        dto.TxID,
		OutputAddress: dto.OutputAddress,
		Direction:     DirectionFromAML(dto.Direction).String(),
	}

	preparedReq, reqBodyBytes, err := b.prepareRequest(ctx, http.MethodPost, MethodManualCheckTransfer, reqBody, auth)
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}

	// Send request
	resp, err := b.cl.Do(preparedReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			b.log.Warnw("failed to close response body", "error", cerr)
		}
	}()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		b.log.Debugw(
			"failed bitok init check request",
			"http_code", resp.StatusCode,
			"body", string(respBodyBytes),
			"request", string(reqBodyBytes),
		)

		return nil, &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: preparedReq.URL.Path,
		}
	}

	// Decode response
	var apiResp CheckTransferResponse
	if err = json.Unmarshal(respBodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	riskLevel := aml.CheckRiskLevel(apiResp.RiskLevel)
	return &aml.CheckResponse{
		ExternalID: apiResp.ID,
		Score:      apiResp.RiskScore.Mul(decimal.NewFromInt(100)), // Calculate percentage
		RiskLevel:  &riskLevel,
		Status:     aml.CheckStatusNew,
		HTTPStatus: resp.StatusCode,
		Request:    reqBodyBytes,
		Response:   respBodyBytes,
	}, nil
}

// FetchCheckStatus retrieves the status of a check by its external ID.
func (b *Client) FetchCheckStatus(ctx context.Context, checkID string, auth aml.RequestAuthorizer) (*aml.CheckResponse, error) {
	if checkID == "" {
		return nil, fmt.Errorf("checkID cannot be empty")
	}

	// Prepare the request URL
	u := *b.baseURL
	u.Path = path.Join(MethodManualCheck, checkID)

	// Prepare the GET request (no body for GET)
	httpReq, _, err := b.prepareRequest(ctx, http.MethodGet, u.Path, nil, auth)
	if err != nil {
		return nil, fmt.Errorf("prepare request: %w", err)
	}

	// Send the request
	resp, err := b.cl.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			b.log.Warnw("failed to close response body", "error", cerr)
		}
	}()

	// Read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, &aml.RequestFailedError{
			StatusCode: resp.StatusCode,
			Body:       respBodyBytes,
			RequestURL: httpReq.URL.Path,
		}
	}

	// Decode response
	var apiResp CheckTransferResponse
	if err = json.Unmarshal(respBodyBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	riskLevel := aml.CheckRiskLevel(apiResp.RiskLevel)
	return &aml.CheckResponse{
		ExternalID: apiResp.ID,
		Score:      apiResp.RiskScore.Mul(decimal.NewFromInt(100)), // Calculate percentage,
		RiskLevel:  &riskLevel,
		Status:     CheckStatus(apiResp.CheckStatus).ToAMLStatus(),
		HTTPStatus: resp.StatusCode,
		Response:   respBodyBytes,
	}, nil
}

// prepareRequest prepares an HTTP request with signature
func (b *Client) prepareRequest(ctx context.Context, httpMethod, apiMethod string, req json.Marshaler, auth aml.RequestAuthorizer) (*http.Request, json.RawMessage, error) {
	// Marshal request body
	var bodyBytes []byte
	if req != nil {
		var err error
		bodyBytes, err = req.MarshalJSON()
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request: %w", err)
		}

		bodyBytes, err = prepareBody(bodyBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare request: %w", err)
		}
	}

	// Construct the full URL using ResolveReference
	relativeURL, err := url.Parse(apiMethod)
	if err != nil {
		return nil, nil, fmt.Errorf("parse api method path: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, httpMethod, b.baseURL.ResolveReference(relativeURL).String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("create HTTP request: %w", err)
	}

	if auth != nil {
		if err = auth.Authorize(ctx, httpReq); err != nil {
			return nil, nil, fmt.Errorf("authorize request: %w", err)
		}
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, bodyBytes, nil
}
