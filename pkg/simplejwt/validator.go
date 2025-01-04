package simplejwt

import (
	"crypto"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	key crypto.PublicKey
}

func NewValidator(publicKeyPath string) (*Validator, error) {
	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key file: %w", err)
	}

	key, err := jwt.ParseEdPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse as ed public key: %w", err)
	}

	return &Validator{
		key: key,
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

	return token, nil
}
