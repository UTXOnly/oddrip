package auth

import (
	"net/http"
	"strconv"
	"time"
)

type Provider interface {
	Apply(req *http.Request) error
}

type StaticHeaders struct {
	Headers http.Header
}

func (s *StaticHeaders) Apply(req *http.Request) error {
	for k, v := range s.Headers {
		req.Header[k] = v
	}
	return nil
}

type BearerToken string

func (b BearerToken) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+string(b))
	return nil
}

type KalshiSigner struct {
	KeyID      string
	SignRequest func(method, path string, timestamp int64) (signature string, err error)
}

func (k *KalshiSigner) Apply(req *http.Request) error {
	path := req.URL.Path
	ts := nowMilliseconds()
	sig, err := k.SignRequest(req.Method, path, ts)
	if err != nil {
		return err
	}
	req.Header.Set("KALSHI-ACCESS-KEY", k.KeyID)
	req.Header.Set("KALSHI-ACCESS-SIGNATURE", sig)
	req.Header.Set("KALSHI-ACCESS-TIMESTAMP", formatInt64(ts))
	return nil
}

func nowMilliseconds() int64     { return time.Now().UnixMilli() }
func formatInt64(n int64) string { return strconv.FormatInt(n, 10) }
