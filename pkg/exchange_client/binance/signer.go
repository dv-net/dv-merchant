package binance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"
)

const defaultRecvWindow = 5000

type SecurityLevel int

const (
	SecurityLevelNone SecurityLevel = iota
	SecurityLevelAPIKey
	SecurityLevelSigned
)

type Option func(req *http.Request)

type ISigner interface {
	SignRequest(ctx context.Context, req *http.Request, securityLevel SecurityLevel) *http.Request
}

func NewSigner(apiKey string, secretKey string) ISigner {
	return &Signer{
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

type Signer struct {
	apiKey    string
	secretKey string
}

func (o *Signer) SignRequest(_ context.Context, req *http.Request, securityLevel SecurityLevel) *http.Request {
	if securityLevel == SecurityLevelNone {
		return req
	}
	if securityLevel == SecurityLevelAPIKey || securityLevel == SecurityLevelSigned {
		req.Header.Set("X-MBX-APIKEY", o.apiKey)
	}
	if securityLevel == SecurityLevelSigned {
		q := req.URL.Query()
		q.Add("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		q.Add("recvWindow", strconv.FormatInt(defaultRecvWindow, 10))
		req.URL.RawQuery = q.Encode()

		signature := generateSignature(o.secretKey, req.URL.RawQuery)
		q.Add("signature", signature)
		req.URL.RawQuery = q.Encode()
	}

	return req
}

func generateSignature(secretKey, data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
