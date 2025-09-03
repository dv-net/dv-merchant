package admin_gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	admin_errors "github.com/dv-net/dv-merchant/pkg/admin_gateway/errors"
	admin_requests "github.com/dv-net/dv-merchant/pkg/admin_gateway/requests"
	admin_responses "github.com/dv-net/dv-merchant/pkg/admin_gateway/responses"

	"github.com/google/uuid"
)

const (
	MethodRequestOwnerAuth   = "/init-registration"
	MethodRequestOwnerInitTg = "/telegram/link-generation"
	MethodRequestUnlinkTg    = "/telegram/unlink-account"
	MethodRequestUnlinkCode  = "/telegram/unlink/code"
	MethodGetOwnerBalance    = "/owner-data"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type IOwner interface {
	GetAuthCode(ctx context.Context, req admin_requests.InitAuthRequest) (*admin_responses.InitAuthResponse, error)
	GetOwnerData(ctx context.Context, token string) (*admin_responses.OwnerDataResponse, error)
	InitOwnerTg(ctx context.Context, ownerToken string) (*admin_responses.InitOwnerTgResponse, error)
	ConfirmUnlinkTg(ctx context.Context, code, ownerToken string) error
	UnlinkOwnerTg(ctx context.Context, ownerID uuid.UUID, ownerToken string) error
}

var _ IOwner = (*Service)(nil)

func (s *Service) UnlinkOwnerTg(ctx context.Context, ownerID uuid.UUID, ownerToken string) error {
	req := map[string]string{"owner_id": ownerID.String()}
	encodedReq, err := json.Marshal(req)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodRequestUnlinkCode, http.MethodPost, encodedReq, map[string]string{
		"Authorization": "Bearer " + ownerToken,
	})
	if err != nil {
		var reqErr *admin_errors.RequestFailedError
		if errors.As(err, &reqErr) && reqErr.StatusCode == http.StatusUnauthorized {
			err = ErrUnauthenticated
		}

		return err
	}

	return nil
}

func (s *Service) ConfirmUnlinkTg(ctx context.Context, code, ownerToken string) error {
	req := map[string]string{"code": code}
	encodedReq, err := json.Marshal(req)
	if err != nil {
		return admin_errors.ErrPrepareRequest
	}

	_, err = s.sendRequest(ctx, MethodRequestUnlinkTg, http.MethodPost, encodedReq, map[string]string{
		"Authorization": "Bearer " + ownerToken,
	})
	if err != nil {
		var reqErr *admin_errors.RequestFailedError
		if errors.As(err, &reqErr) && reqErr.StatusCode == http.StatusUnauthorized {
			err = ErrUnauthenticated
		}

		return err
	}

	return nil
}

func (s *Service) GetAuthCode(ctx context.Context, req admin_requests.InitAuthRequest) (*admin_responses.InitAuthResponse, error) {
	encodedReq, err := json.Marshal(req)
	if err != nil {
		return nil, admin_errors.ErrPrepareRequest
	}

	resp, err := s.sendRequest(ctx, MethodRequestOwnerAuth, http.MethodPost, encodedReq, nil)
	if err != nil {
		return nil, err
	}

	preparedResp := &admin_responses.Result[admin_responses.InitAuthResponse]{}
	if err := json.Unmarshal(resp, preparedResp); err != nil {
		return nil, err
	}

	return &preparedResp.Data, nil
}

func (s *Service) GetOwnerData(ctx context.Context, token string) (*admin_responses.OwnerDataResponse, error) {
	resp, err := s.sendRequest(ctx, MethodGetOwnerBalance, http.MethodGet, nil, map[string]string{
		"Authorization": "Bearer " + token,
	})
	if err != nil {
		var reqErr *admin_errors.RequestFailedError
		if errors.As(err, &reqErr) && reqErr.StatusCode == http.StatusUnauthorized {
			err = ErrUnauthenticated
		}

		return nil, err
	}

	preparedResp := &admin_responses.Result[admin_responses.OwnerDataResponse]{}
	if err := json.Unmarshal(resp, preparedResp); err != nil {
		return nil, err
	}

	return &preparedResp.Data, nil
}

func (s *Service) InitOwnerTg(ctx context.Context, ownerToken string) (*admin_responses.InitOwnerTgResponse, error) {
	resp, err := s.sendRequest(ctx, MethodRequestOwnerInitTg, http.MethodPost, nil, map[string]string{
		"Authorization": "Bearer " + ownerToken,
	})
	if err != nil {
		return nil, err
	}

	preparedResp := &admin_responses.Result[admin_responses.InitOwnerTgResponse]{}
	if err := json.Unmarshal(resp, preparedResp); err != nil {
		return nil, err
	}

	return &preparedResp.Data, nil
}
