package postgres

import "errors"

var (
	ErrUniqueConstraint          = errors.New("unique constraint")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
)