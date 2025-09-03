package admin_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"
	admin_responses "github.com/dv-net/dv-merchant/pkg/admin_gateway/responses"
)

const (
	MethodHeartBeat = "/system/heartbeat"
	MethodCheckMyIP = "/check-my-ip"
)

type ISystem interface {
	HeartBeat(ctx context.Context, analyticsData *admin_requests.AnalyticsData) error
	CheckIP(ctx context.Context) (string, error)
}

var _ ISystem = (*Service)(nil)

func (s *Service) HeartBeat(ctx context.Context, analyticsData *admin_requests.AnalyticsData) error {
	data := admin_requests.HeartBeatRequest{
		TS: time.Now(),
	}
	if analyticsData != nil {
		data.AnalyticsData = analyticsData
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal heart beat request: %w", err)
	}

	_, err = s.sendRequest(ctx, MethodHeartBeat, http.MethodPost, buf, nil)
	return err
}

func (s *Service) CheckIP(ctx context.Context) (string, error) {
	res, err := s.sendPublicRequest(ctx, MethodCheckMyIP, http.MethodGet, nil, nil)
	if err != nil {
		return "", err
	}

	preparedResp := &admin_responses.Result[admin_responses.CheckMyIPResponse]{}
	if err := json.Unmarshal(res, preparedResp); err != nil {
		return "", err
	}

	return preparedResp.Data.IP, nil
}
