package vo

import (
	"errors"
	"net/mail"
)

type Email string

func NewEmail(value string) (Email, error) {
	_, err := mail.ParseAddress(value)
	if err != nil {
		return "", errors.New("invalid email")
	}

	return Email(value), nil
}
