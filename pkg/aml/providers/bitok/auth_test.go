package bitok

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHMACAuthorizer_Authorize tests the Authorize method of HMACAuthorizer
func TestHMACAuthorizer_Authorize(t *testing.T) {
	const secret = "test-secret"
	const accessKey = "test-access-key"

	tests := []struct {
		name       string
		method     string
		path       string
		query      string
		body       string
		wantBody   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "POST with JSON body",
			method:   http.MethodPost,
			path:     "/v1/manual-checks/check-transfer/",
			query:    "",
			body:     `{"network":"ethereum","token_id":"0x123","tx_hash":"0x456","direction":"in","risk_model":"origin_of_funds"}`,
			wantBody: `{"network":"ethereum","token_id":"0x123","tx_hash":"0x456","direction":"in","risk_model":"origin_of_funds"}`,
			wantErr:  false,
		},
		{
			name:     "GET without body",
			method:   http.MethodGet,
			path:     "/v1/manual-checks/3fa85f64-5717-4562-b3fc-2c963f66afa6/",
			query:    "",
			body:     "",
			wantBody: "",
			wantErr:  false,
		},
		{
			name:     "GET with query parameters",
			method:   http.MethodGet,
			path:     "/v1/manual-checks/3fa85f64-5717-4562-b3fc-2c963f66afa6/",
			query:    "param1=value1",
			body:     "",
			wantBody: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			urlStr := "http://example.com" + tt.path
			if tt.query != "" {
				urlStr += "?" + tt.query
			}
			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}
			req, err := http.NewRequestWithContext(context.Background(), tt.method, urlStr, bodyReader)
			require.NoError(t, err)

			// Create authorizer
			auth := NewHMACAuthorizer(secret, accessKey)

			// Call Authorize
			err = auth.Authorize(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.wantErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Authorize() error = %v, wantErrMsg %q", err, tt.wantErrMsg)
				}
			}

			if tt.wantErr {
				return
			}

			// Read body for verification
			var bodyBytes []byte
			if req.Body != nil {
				bodyBytes, err = io.ReadAll(req.Body)
				if err != nil {
					t.Errorf("Failed to read body: %v", err)
				}
			}

			// Check body preservation
			if tt.wantBody != "" {
				gotBody := string(bodyBytes)
				if gotBody != tt.wantBody {
					t.Errorf("Body = %q, want %q", gotBody, tt.wantBody)
				}
			}

			// Check headers
			if len(req.Header["API-KEY-ID"]) != 1 || req.Header["API-KEY-ID"][0] != accessKey {
				t.Errorf("API-KEY-ID = %v, want %q", req.Header["API-KEY-ID"], accessKey)
			}
			if len(req.Header["API-TIMESTAMP"]) != 1 {
				t.Errorf("API-TIMESTAMP = %v, want single value", req.Header["API-TIMESTAMP"])
			}
			if len(req.Header["API-SIGNATURE"]) != 1 {
				t.Errorf("API-SIGNATURE = %v, want single value", req.Header["API-SIGNATURE"])
			}
		})
	}
}
