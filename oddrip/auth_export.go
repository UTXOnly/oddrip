package oddrip

import (
	"crypto/rsa"
	"net/http"

	"github.com/UTXOnly/oddrip/oddrip/internal/auth"
)

type AuthProvider interface {
	Apply(req *http.Request) error
}

func NewKalshiSigner(keyID string, privateKey *rsa.PrivateKey) AuthProvider {
	return auth.NewKalshiRSAPSSSigner(keyID, privateKey)
}

func ParsePrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	return auth.ParsePrivateKeyFromPEM(pemBytes)
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
