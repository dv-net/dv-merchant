package coinkyt

import (
	"context"
	"net/http"
)

type APIKeyAuthorizer struct {
	apiKey string
}

func NewAPIKeyAuthorizer(apiKey string) *APIKeyAuthorizer {
	return &APIKeyAuthorizer{apiKey: apiKey}
}

func (a *APIKeyAuthorizer) Authorize(_ context.Context, req *http.Request) error {
	req.Header.Set("X-API-Key", a.apiKey)
	return nil
}
