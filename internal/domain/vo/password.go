package vo

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Password string

func NewPassword(value string) (Password, error) {
	if len(value) < 8 || len(value) > 255 {
		return "", errors.New("password length is invalid")
	}

	return Password(value), nil
}

func (p Password) Hash() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), 13)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p Password) Compare(hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
	return err == nil
}
