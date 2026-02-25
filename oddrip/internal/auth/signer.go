package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strconv"
)

func NewKalshiRSAPSSSigner(keyID string, privateKey *rsa.PrivateKey) *KalshiSigner {
	return &KalshiSigner{
		KeyID: keyID,
		SignRequest: func(method, path string, timestamp int64) (string, error) {
			msg := strconv.FormatInt(timestamp, 10) + method + path
			h := sha256.Sum256([]byte(msg))
			sig, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, h[:], &rsa.PSSOptions{
				SaltLength: rsa.PSSSaltLengthEqualsHash,
			})
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString(sig), nil
		},
	}
}

func ParsePrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	}
	k, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return k, nil
}
