package interceptors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/tools/hash"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-processing/api/processing/client/v1/clientv1connect"
)

var errEmptyClientCredentials = fmt.Errorf("empty client credentials")

type SignInterceptor struct {
	settingService setting.ISettingService
}

func NewSignInterceptor(
	settingService setting.ISettingService,
) *SignInterceptor {
	return &SignInterceptor{
		settingService: settingService,
	}
}

// WrapUnary wraps the unary function
func (i *SignInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		// check sign key
		if err := i.signRequest(ctx, req); err != nil {
			return nil, connect.NewError(connect.CodeCanceled, err)
		}

		return next(ctx, req)
	}
}

// WrapStreamingClient wraps the streaming server function
func (*SignInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

// WrapStreamingHandler wraps the streaming server function
func (i *SignInterceptor) WrapStreamingHandler(_ connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(
		_ context.Context,
		_ connect.StreamingHandlerConn,
	) error {
		return fmt.Errorf("streaming is not supported")
	}
}

// signRequest signs request by client_secret
func (i *SignInterceptor) signRequest(ctx context.Context, req connect.AnyRequest) error {
	// skip signing for create client
	if req.Spec().Procedure == clientv1connect.ClientServiceCreateProcedure {
		return nil
	}

	rootSettings, err := i.settingService.GetRootSettings(ctx)
	if err != nil {
		return fmt.Errorf("get root setting: %w", err)
	}

	var clientID, clientSecret string
	for _, stng := range rootSettings {
		if stng.Name == setting.ProcessingClientKey {
			clientSecret = stng.Value
		}
		if stng.Name == setting.ProcessingClientID {
			clientID = stng.Value
		}
	}

	// check client credentials
	if clientID == "" || clientSecret == "" {
		return errEmptyClientCredentials
	}

	// marshal payload
	payload, err := json.Marshal(req.Any())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	sign := hash.SHA256Signature(payload, clientSecret)

	// set sign key
	req.Header().Set("X-Sign", sign)
	req.Header().Set("X-Client-ID", clientID)

	return nil
}
