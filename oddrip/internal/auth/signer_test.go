package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var testKey = sync.OnceValues(func() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
})

func TestKalshiRSAPSSSignerApply(t *testing.T) {
	key, err := testKey()
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/trade-api/v2/markets?limit=5", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewKalshiRSAPSSSigner("key-id", key).Apply(req); err != nil {
		t.Fatal(err)
	}

	if got := req.Header.Get("KALSHI-ACCESS-KEY"); got != "key-id" {
		t.Errorf("KALSHI-ACCESS-KEY = %q, want %q", got, "key-id")
	}
	tsHeader := req.Header.Get("KALSHI-ACCESS-TIMESTAMP")
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		t.Fatalf("KALSHI-ACCESS-TIMESTAMP = %q: %v", tsHeader, err)
	}
	if drift := time.Now().UnixMilli() - ts; drift < -1000 || drift > 60_000 {
		t.Errorf("timestamp %d is not recent milliseconds (drift %dms)", ts, drift)
	}
	sig, err := base64.StdEncoding.DecodeString(req.Header.Get("KALSHI-ACCESS-SIGNATURE"))
	if err != nil {
		t.Fatalf("KALSHI-ACCESS-SIGNATURE not base64: %v", err)
	}

	opts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}
	if req.URL.Path != "/trade-api/v2/markets" {
		t.Fatalf("req.URL.Path = %q", req.URL.Path)
	}
	h := sha256.Sum256([]byte(tsHeader + http.MethodGet + req.URL.Path))
	if err := rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, h[:], sig, opts); err != nil {
		t.Fatalf("signature over timestamp+method+path did not verify: %v", err)
	}
	h = sha256.Sum256([]byte(tsHeader + http.MethodGet + req.URL.RequestURI()))
	if err := rsa.VerifyPSS(&key.PublicKey, crypto.SHA256, h[:], sig, opts); err == nil {
		t.Fatal("signature verifies over path with query string; query must be excluded")
	}
}

func TestKalshiSignerApplyError(t *testing.T) {
	errSign := errors.New("sign failed")
	s := &KalshiSigner{KeyID: "k", SignRequest: func(string, string, int64) (string, error) {
		return "", errSign
	}}
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(req); !errors.Is(err, errSign) {
		t.Fatalf("err = %v, want %v", err, errSign)
	}
	if req.Header.Get("KALSHI-ACCESS-KEY") != "" {
		t.Error("headers set despite signing error")
	}
}

func TestParsePrivateKeyFromPEM(t *testing.T) {
	key, err := testKey()
	if err != nil {
		t.Fatal(err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, typ string
		der       []byte
	}{
		{"pkcs8", "PRIVATE KEY", pkcs8},
		{"pkcs1", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePrivateKeyFromPEM(pem.EncodeToMemory(&pem.Block{Type: tc.typ, Bytes: tc.der}))
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(key) {
				t.Fatal("parsed key differs from original")
			}
		})
	}

	t.Run("garbage", func(t *testing.T) {
		if _, err := ParsePrivateKeyFromPEM([]byte("not a pem")); err == nil {
			t.Fatal("expected error for non-PEM input")
		}
		junk := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("garbage")})
		if _, err := ParsePrivateKeyFromPEM(junk); err == nil {
			t.Fatal("expected error for PEM with garbage body")
		}
	})

	t.Run("ecdsa", func(t *testing.T) {
		ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		der, err := x509.MarshalPKCS8PrivateKey(ec)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ParsePrivateKeyFromPEM(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
		if err == nil || !strings.Contains(err.Error(), "not an RSA private key") {
			t.Fatalf("err = %v, want \"not an RSA private key\"", err)
		}
	})
}
