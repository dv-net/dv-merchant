package mexc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultRecvWindow = 5000

type ISigner interface {
	SignRequest(ctx context.Context, req *http.Request, private bool) *http.Request
}

type Signer struct {
	apiKey    string
	secretKey string
}

func NewSigner(apiKey, secretKey string) ISigner {
	return &Signer{
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

func (s *Signer) SignRequest(_ context.Context, req *http.Request, private bool) *http.Request {
	if !private {
		return req
	}

	req.Header.Set("X-MEXC-APIKEY", s.apiKey)

	q := req.URL.Query()
	q.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	q.Set("recvWindow", strconv.Itoa(defaultRecvWindow))
	message := strings.ReplaceAll(q.Encode(), "+", "%20")

	signature := generateSignature(s.secretKey, message)
	req.URL.RawQuery = message + "&signature=" + signature

	return req
}

func generateSignature(secretKey, data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
