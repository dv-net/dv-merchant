package htx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

type SigningMethod string

func (o SigningMethod) String() string {
	return string(o)
}

const (
	SigningMethodEd25519    SigningMethod = "Ed25519"
	SigningMethodHmacSHA256 SigningMethod = "HmacSHA256"
)

const (
	SignatureVersion2 = "2"
)

type ISigner interface {
	SignRequest(ctx context.Context, req *http.Request) (*http.Request, error)
	SecretKey() []byte
}

func NewSigner(accessKey, secretKey string) ISigner {
	return &Signer{
		secretKey: []byte(secretKey),
		accessKey: accessKey,
	}
}

type Signer struct {
	secretKey []byte
	accessKey string
}

func (o *Signer) SecretKey() []byte { return o.secretKey }

func (o *Signer) SignRequest(_ context.Context, req *http.Request) (*http.Request, error) {
	query := req.URL.Query()
	query.Set("AccessKeyId", o.accessKey)
	query.Set("SignatureMethod", SigningMethodHmacSHA256.String())
	query.Set("SignatureVersion", SignatureVersion2)
	query.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05"))
	req.URL.RawQuery = query.Encode()

	signature := o.sign(req)
	query.Set("Signature", signature)
	req.URL.RawQuery = query.Encode()

	return req, nil
}

func (o *Signer) sign(req *http.Request) string {
	data := fmt.Sprintf("%s\n%s\n%s\n%s", req.Method, req.URL.Host, req.URL.Path, req.URL.RawQuery)
	h := hmac.New(sha256.New, o.secretKey)
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
