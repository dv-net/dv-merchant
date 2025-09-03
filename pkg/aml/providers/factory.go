package providers

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/aml_bot"
	"github.com/dv-net/dv-merchant/pkg/aml/providers/bitok"

	"github.com/puzpuzpuz/xsync/v3"
)

// AuthorizerCreator authorizer function-creator for concrete AML-provider
type AuthorizerCreator func(ctx context.Context, creds map[aml.AuthKeyType]string, externalID string) (aml.RequestAuthorizer, error)

// ProviderFactory keeps creation logic with concurrently-safe registry
type ProviderFactory interface {
	GetClient(slug aml.ProviderSlug) (aml.Client, error)
	CreateAuthorizer(ctx context.Context, slug aml.ProviderSlug, creds map[aml.AuthKeyType]string, externalID string) (aml.RequestAuthorizer, error)
	RegisterProvider(slug aml.ProviderSlug, client aml.Client, authCreator AuthorizerCreator)
	GetAllRegisteredProviders() []aml.ProviderSlug
}

type providerFactoryImpl struct {
	clients *xsync.MapOf[aml.ProviderSlug, aml.Client]
	auths   *xsync.MapOf[aml.ProviderSlug, AuthorizerCreator]
}

func NewFactory() ProviderFactory {
	return &providerFactoryImpl{
		clients: xsync.NewMapOf[aml.ProviderSlug, aml.Client](),
		auths:   xsync.NewMapOf[aml.ProviderSlug, AuthorizerCreator](),
	}
}

// RegisterProvider registers AML provider with its authorizer
func (f *providerFactoryImpl) RegisterProvider(slug aml.ProviderSlug, client aml.Client, authCreator AuthorizerCreator) {
	f.clients.Store(slug, client)
	f.auths.Store(slug, authCreator)
}

func (f *providerFactoryImpl) GetClient(slug aml.ProviderSlug) (aml.Client, error) {
	client, ok := f.clients.Load(slug)
	if !ok {
		return nil, fmt.Errorf("provider %s not registered", slug)
	}

	return client, nil
}

func (f *providerFactoryImpl) CreateAuthorizer(ctx context.Context, slug aml.ProviderSlug, creds map[aml.AuthKeyType]string, externalID string) (aml.RequestAuthorizer, error) {
	creator, ok := f.auths.Load(slug)
	if !ok {
		return nil, fmt.Errorf("authorizer for provider %s not registered", slug)
	}

	return creator(ctx, creds, externalID)
}

func (f *providerFactoryImpl) GetAllRegisteredProviders() []aml.ProviderSlug {
	result := make([]aml.ProviderSlug, 0)
	f.clients.Range(func(slug aml.ProviderSlug, _ aml.Client) bool {
		result = append(result, slug)
		return true
	})

	return result
}

// CreateBitOKAuthorizer define creator for BitOK authorizer factory
func CreateBitOKAuthorizer() AuthorizerCreator {
	return func(_ context.Context, creds map[aml.AuthKeyType]string, _ string) (aml.RequestAuthorizer, error) {
		accessKeyID, ok := creds[aml.KeyAccessKeyID]
		if !ok || accessKeyID == "" {
			return nil, fmt.Errorf("access key ID not provided")
		}
		secret, ok := creds[aml.KeySecret]
		if !ok || secret == "" {
			return nil, fmt.Errorf("secret not provided")
		}

		return bitok.NewHMACAuthorizer(secret, accessKeyID), nil
	}
}

// CreateAMLBotAuthorizer define creator for AMLBot authorizer factory
func CreateAMLBotAuthorizer() AuthorizerCreator {
	return func(_ context.Context, creds map[aml.AuthKeyType]string, externalID string) (aml.RequestAuthorizer, error) {
		accessID, ok := creds[aml.KeyAccessID]
		if !ok || accessID == "" {
			return nil, fmt.Errorf("access ID not provided")
		}
		accessKey, ok := creds[aml.KeyAccessKey]
		if !ok || accessKey == "" {
			return nil, fmt.Errorf("access key not provided")
		}

		var opts []aml_bot.Option
		if externalID != "" {
			opts = append(opts, aml_bot.WithUID(externalID))
		}

		return aml_bot.NewMD5Authorizer(accessID, accessKey, opts...), nil
	}
}
