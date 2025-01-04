package vo

import "errors"

type Username string

func NewUsername(value string) (Username, error) {
	if len(value) < 3 || len(value) > 64 {
		return "", errors.New("username length is invalid")
	}

	return Username(value), nil
}
