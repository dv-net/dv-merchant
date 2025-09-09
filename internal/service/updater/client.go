package updater

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/service/updater/requests"
	"github.com/dv-net/dv-merchant/internal/service/updater/responses"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/goccy/go-json"
)

const (
	getNewVersionURL     = "/api/v1/version/dv-merchant"
	updateProcessingURL  = "/api/v1/update"
	getUpdaterVersionURL = "/api/v1/version"
	pingURL              = "/ping"
)

type IUpdateClient interface {
	CheckNewVersion(ctx context.Context) (*responses.GetNewVersionResponse, error)
	Update(ctx context.Context) error
	Ping(ctx context.Context) error
	GetUpdaterVersion(ctx context.Context) (*responses.GetUpdaterVersionResponse, error)
}

type Option func(c *Client)

type Client struct {
	logger     logger.Logger
	baseURL    string
	httpClient *http.Client
}

func NewClient(log logger.Logger, conf *config.Config, opts ...Option) (*Client, error) {
	svc := &Client{
		httpClient: http.DefaultClient,
		logger:     log,
	}

	svc.baseURL = conf.Updater.BaseURL

	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

func (u *Client) CheckNewVersion(ctx context.Context) (*responses.GetNewVersionResponse, error) {
	response := &responses.GetNewVersionResponse{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.baseURL+getNewVersionURL, http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	responseBytes := new(bytes.Buffer)
	if _, err := responseBytes.ReadFrom(res.Body); err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned unexpected status code: %d", res.StatusCode)
	}
	if err := json.Unmarshal(responseBytes.Bytes(), response); err != nil {
		return nil, err
	}
	return response, nil
}

func (u *Client) Update(ctx context.Context) error {
	response := &responses.UpdateResponse{}
	request := &requests.UpdateRequest{
		Name: "dv-merchant",
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.baseURL+updateProcessingURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	res, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	responseBytes := new(bytes.Buffer)
	if _, err := responseBytes.ReadFrom(res.Body); err != nil {
		return nil
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned unexpected status code: %d", res.StatusCode)
	}
	if err := json.Unmarshal(responseBytes.Bytes(), response); err != nil {
		return err
	}
	return nil
}

func (u *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	url, err := url.JoinPath(u.baseURL, pingURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}
	req, err := http.NewRequestWithContext(pingCtx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}

	res, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (u *Client) GetUpdaterVersion(ctx context.Context) (*responses.GetUpdaterVersionResponse, error) {
	response := &responses.GetUpdaterVersionResponse{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.baseURL+"/api/v1/version", http.NoBody)
	if err != nil {
		return nil, err
	}
	res, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	responseBytes := new(bytes.Buffer)
	if _, err := responseBytes.ReadFrom(res.Body); err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned unexpected status code: %d", res.StatusCode)
	}
	if err := json.Unmarshal(responseBytes.Bytes(), response); err != nil {
		return nil, err
	}
	return response, nil
}
