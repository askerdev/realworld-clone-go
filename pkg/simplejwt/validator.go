package simplejwt

import (
	"crypto"
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	key   crypto.PublicKey
	cache Cache
}

func NewValidator(publicKeyPath string, cache Cache) (*Validator, error) {
	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key file: %w", err)
	}

	key, err := jwt.ParseEdPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse as ed public key: %w", err)
	}

	return &Validator{
		key:   key,
		cache: cache,
	}, nil
}

func (v *Validator) Validate(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, fmt.Errorf("unexpected singing method: %v", t.Header["alg"])
			}
			return v.key, nil
		})
	if err != nil {
		return nil, fmt.Errorf("unable to parse token string: %w", err)
	}

	iAt, err := token.Claims.GetIssuedAt()
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	sub, ok := claims["sub"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	id, ok := sub["id"].(float64)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	prevIAt, ok := v.cache.Get(uint64(id))
	if !ok {
		return nil, errors.New("invalid iat")
	}

	if iAt.Unix() < prevIAt.Unix() {
		return nil, errors.New("invalid iat")
	}

	return token, nil
}
