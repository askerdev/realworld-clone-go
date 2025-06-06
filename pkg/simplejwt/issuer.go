package simplejwt

import (
	"crypto"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	key   crypto.PrivateKey
	cache Cache
}

func NewIssuer(privateKeyPath string, cache Cache) (*Issuer, error) {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		panic(fmt.Errorf("unable to read private key file: %w", err))
	}

	key, err := jwt.ParseEdPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse as ed private key: %w", err)
	}

	return &Issuer{
		key:   key,
		cache: cache,
	}, nil
}

func (i *Issuer) Token(authID uint64, data any) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.MapClaims{
		"aud": "api",
		"nbf": now.Unix(),
		"iat": now.Unix(),
		"exp": now.Add(15 * time.Minute).Unix(),
		"iss": "http://localhost:8080",
		"sub": data,
	})

	tokenString, err := token.SignedString(i.key)
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}

	err = i.cache.Set(authID, now)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
