package aml_bot

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestMD5Authorizer_Authorize tests MD5 token authorization
func TestMD5Authorizer_Authorize(t *testing.T) {
	const accessID = "123"
	const accessKey = "test-key"
	var uid = "tx123"

	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		uid      *string
		wantBody string
		wantErr  bool
	}{
		{
			name:     "POST check with form data",
			method:   http.MethodPost,
			path:     "/check/",
			body:     "hash=tx123&asset=BTC&locale=en&flow=advanced",
			uid:      nil,
			wantBody: "accessId=123&asset=BTC&flow=advanced&hash=tx123&locale=en&token=5313e9afc3fd9c690466e00f98e19f75",
			wantErr:  false,
		},
		{
			name:     "POST recheck with uid",
			method:   http.MethodPost,
			path:     "/recheck/",
			body:     "uid=tx123&locale=en",
			uid:      &uid,
			wantBody: "accessId=123&locale=en&token=b40e20561487cce62636bad5f0a16d38&uid=tx123",
			wantErr:  false,
		},
		{
			name:     "POST with empty body",
			method:   http.MethodPost,
			path:     "/check/",
			body:     "",
			uid:      nil,
			wantBody: "accessId=123&token=5313e9afc3fd9c690466e00f98e19f75",
			wantErr:  false,
		},
		{
			name:     "POST with invalid form data",
			method:   http.MethodPost,
			path:     "/check/",
			body:     "=invalid=%%%",
			uid:      nil,
			wantBody: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the request
			req, err := http.NewRequestWithContext(context.Background(), tt.method, "http://example.com"+tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create the authorizer
			auth := NewMD5Authorizer(accessID, accessKey, func() Option {
				if tt.uid != nil {
					return WithUID(*tt.uid)
				}
				return func(a *MD5Authorizer) {}
			}())

			// Perform authorization
			err = auth.Authorize(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// Verify the request body
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				t.Errorf("Failed to read request body: %v", err)
			}
			gotBody := string(bodyBytes)
			if gotBody != tt.wantBody {
				t.Errorf("Request body = %q, wanted %q", gotBody, tt.wantBody)
			}

			// Verify Content-Length
			if req.ContentLength != int64(len(gotBody)) {
				t.Errorf("ContentLength = %d, wanted %d", req.ContentLength, len(gotBody))
			}
		})
	}
}
