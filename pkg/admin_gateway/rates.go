package admin_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	admin_responses "github.com/dv-net/dv-merchant/pkg/admin_gateway/responses"
)

const (
	MethodFetchRateBySource = "/rate/%s"
)

type IRates interface {
	FetchRateBySource(ctx context.Context, source string) ([]admin_responses.RatesResponse, error)
	FetchAllRates(ctx context.Context) ([]admin_responses.RatesResponse, error)
}

var _ INotification = (*Service)(nil)

func (s *Service) FetchRateBySource(ctx context.Context, source string) ([]admin_responses.RatesResponse, error) {
	resp, err := s.sendPublicRequest(ctx, fmt.Sprintf(MethodFetchRateBySource, source), http.MethodGet, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	preparedResp := &admin_responses.Result[[]admin_responses.RatesResponse]{}
	if err = json.Unmarshal(resp, preparedResp); err != nil {
		return nil, err
	}

	return preparedResp.Data, nil
}

func (s *Service) FetchAllRates(ctx context.Context) ([]admin_responses.RatesResponse, error) {
	resp, err := s.sendPublicRequest(ctx, fmt.Sprintf(MethodFetchRateBySource, ""), http.MethodGet, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	preparedResp := &admin_responses.Result[[]admin_responses.RatesResponse]{}
	if err = json.Unmarshal(resp, preparedResp); err != nil {
		return nil, err
	}

	return preparedResp.Data, nil
}
