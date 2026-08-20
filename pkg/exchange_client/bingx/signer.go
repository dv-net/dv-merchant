package bingx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
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

	req.Header.Set("X-BX-APIKEY", s.apiKey)

	params := req.URL.Query()
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", strconv.Itoa(defaultRecvWindow))

	message := plainQuery(params)
	req.URL.RawQuery = message + "&signature=" + generateSignature(s.secretKey, message)

	return req
}

func plainQuery(params url.Values) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+params.Get(key))
	}

	return strings.Join(pairs, "&")
}

func generateSignature(secretKey, data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
