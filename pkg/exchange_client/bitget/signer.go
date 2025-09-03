package bitget

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ISigner interface {
	SignRequest(ctx context.Context, req *http.Request) (*http.Request, error)
	GetTimestamp() string
}

type Signer struct {
	secretKey  []byte
	accesskey  string
	passphrase string
}

func NewSigner(accessKey, secretkey, passphrase string) ISigner {
	return &Signer{
		secretKey:  []byte(secretkey),
		accesskey:  accessKey,
		passphrase: passphrase,
	}
}

func (o *Signer) SignRequest(_ context.Context, req *http.Request) (*http.Request, error) {
	req.Header.Add(BitGetTimeStamp, o.GetTimestamp())
	req.Header.Add(BitGetAccessKey, o.accesskey)
	req.Header.Add(BitGetAccessPassphrase, o.passphrase)
	req.Header.Add(BitGetAccessSign, o.sign(req))
	req.Header.Add(BitGetLocale, "en-US")
	return req, nil
}

func (o *Signer) sign(req *http.Request) string {
	if req.ContentLength == 0 || req.Body == http.NoBody {
		payload := &strings.Builder{}
		payload.WriteString(req.Header.Get(BitGetTimeStamp))
		payload.WriteString(strings.ToUpper(req.Method))
		payload.WriteString(req.URL.Path)
		if req.URL.RawQuery != "" {
			payload.WriteString("?")
			payload.WriteString(o.sortQuery(req.URL.Query()))
		}

		hash := hmac.New(sha256.New, o.secretKey)
		hash.Write([]byte(payload.String()))
		return base64.StdEncoding.EncodeToString(hash.Sum(nil))
	}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return ""
	}
	req.Body = io.NopCloser(bytes.NewBuffer(buf))

	body := bytes.NewBuffer(buf)
	payload := &strings.Builder{}
	payload.WriteString(req.Header.Get(BitGetTimeStamp))
	payload.WriteString(strings.ToUpper(req.Method))
	payload.WriteString(req.URL.Path)
	if req.URL.RawQuery != "" {
		payload.WriteString("?")
		payload.WriteString(o.sortQuery(req.URL.Query()))
	}
	if body.Len() > 0 {
		payload.Write(body.Bytes())
	}

	hash := hmac.New(sha256.New, o.secretKey)
	hash.Write([]byte(payload.String()))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func (o *Signer) GetTimestamp() string {
	ts := time.Now().UnixMilli()
	return strconv.FormatInt(ts, 10)
}

func (o *Signer) sortQuery(query url.Values) string {
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var sb strings.Builder
	for i, key := range keys {
		values := query[key]
		for j, value := range values {
			if i == 0 && j == 0 {
				sb.WriteString(key + "=" + value)
			} else {
				sb.WriteString("&" + key + "=" + value)
			}
		}
	}

	return sb.String()
}
