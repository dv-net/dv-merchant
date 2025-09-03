package interceptors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/metrics"

	"connectrpc.com/connect"
)

const httpSchemePattern = "http"

const (
	defaultServiceName = "unknown"
	defaultMethodName  = "unknown"
)

type ProcessingMetricInterceptor struct {
	m *metrics.PrometheusMetrics
}

func NewProcessingMetric(metrics *metrics.PrometheusMetrics) *ProcessingMetricInterceptor {
	return &ProcessingMetricInterceptor{m: metrics}
}

// WrapUnary wraps the unary function
func (i *ProcessingMetricInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		now := time.Now()
		defer func() {
			i.setRequestDuration(req.Spec().Procedure, time.Since(now))
		}()

		res, err := next(ctx, req)
		var code connect.Code
		if err != nil {
			code = connect.CodeOf(err)
		}

		if i.isHTTPByScheme(req.Spec().Schema) {
			i.m.ProcessingHTTPResponseCode(code.String())
		} else {
			i.m.ProcessingRPCResponseCode(code.String())
		}

		return res, err
	}
}

// WrapStreamingClient wraps the streaming server function
func (*ProcessingMetricInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		return next(ctx, spec)
	}
}

// WrapStreamingHandler wraps the streaming server function
func (i *ProcessingMetricInterceptor) WrapStreamingHandler(_ connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(
		_ context.Context,
		_ connect.StreamingHandlerConn,
	) error {
		return fmt.Errorf("streaming is not supported")
	}
}

func (i *ProcessingMetricInterceptor) isHTTPByScheme(scheme any) bool {
	converted, ok := scheme.(string)
	if !ok {
		return false
	}

	return strings.Contains(converted, httpSchemePattern)
}

func (i *ProcessingMetricInterceptor) setRequestDuration(rpcMethod string, duration time.Duration) {
	service := defaultServiceName
	method := defaultMethodName

	rpcMethod = strings.TrimPrefix(rpcMethod, "/")
	if i := strings.Index(rpcMethod, "/"); i >= 0 { //nolint:gocritic
		service, method = rpcMethod[:i], rpcMethod[i+1:]
	}

	i.m.ProcessingRequestDuration(service, method, duration.Seconds())
}
