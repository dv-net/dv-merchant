package gateio

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	GateTimestamp = "Timestamp"
	GateKey       = "KEY"
	GateSign      = "SIGN"
)

type ISigner interface {
	AddHeaders(req *http.Request) *http.Request
	SignRequest(ctx context.Context, req *http.Request) (*http.Request, error)
	GetTimestamp() string
}

type Signer struct {
	secretKey []byte
	accessKey string
}

func NewSigner(accessKey, secretKey string) ISigner {
	return &Signer{
		secretKey: []byte(secretKey),
		accessKey: accessKey,
	}
}

func (s *Signer) AddHeaders(req *http.Request) *http.Request {
	timestamp := s.GetTimestamp()
	req.Header.Add(GateTimestamp, timestamp)
	req.Header.Add(GateKey, s.accessKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	return req
}

func (s *Signer) SignRequest(_ context.Context, req *http.Request) (*http.Request, error) {
	timestamp := req.Header.Get(GateTimestamp)

	signature, err := s.sign(req, timestamp)
	if err != nil {
		return nil, err
	}
	req.Header.Add(GateSign, signature)

	return req, nil
}

// Request Method + "\n" + Request URL + "\n" + Query String + "\n" + HexEncode(SHA512(Request Payload)) + "\n" + Timestamp
func (s *Signer) sign(req *http.Request, timestamp string) (string, error) {
	var signatureString strings.Builder
	signatureString.WriteString(strings.ToUpper(req.Method))
	signatureString.WriteString("\n")
	signatureString.WriteString(req.URL.Path)
	signatureString.WriteString("\n")
	signatureString.WriteString(req.URL.RawQuery)
	signatureString.WriteString("\n")

	bodyHash, err := s.hashRequestBody(req)
	if err != nil {
		return "", err
	}
	signatureString.WriteString(bodyHash)
	signatureString.WriteString("\n")
	signatureString.WriteString(timestamp)

	// Generate HMAC-SHA512 signature
	h := hmac.New(sha512.New, s.secretKey)
	h.Write([]byte(signatureString.String()))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, nil
}

func (s *Signer) hashRequestBody(req *http.Request) (string, error) {
	// If no body, use SHA512 hash of empty string
	if req.Body == nil || req.ContentLength == 0 {
		h := sha512.New()
		h.Write([]byte(""))
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	// Read body
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore body for the actual request
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Hash the body
	h := sha512.New()
	h.Write(bodyBytes)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *Signer) GetTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
