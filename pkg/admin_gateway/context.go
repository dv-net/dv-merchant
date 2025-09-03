package admin_gateway

import (
	"golang.org/x/net/context"
)

type ServiceContextKey string

const (
	ServiceContextKeyIdentity ServiceContextKey = "user"
)

type ServiceIdentity struct {
	ClientID  string
	SecretKey string
}

func PrepareServiceContext(ctx context.Context, adminSecret, clientID string) context.Context {
	return context.WithValue(ctx, ServiceContextKeyIdentity, &ServiceIdentity{
		ClientID:  clientID,
		SecretKey: adminSecret,
	})
}

func IdentityFromContext(ctx context.Context) *ServiceIdentity {
	if res, ok := ctx.Value(ServiceContextKeyIdentity).(*ServiceIdentity); ok {
		return res
	}

	return nil
}
