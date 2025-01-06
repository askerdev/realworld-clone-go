package postgres

import "errors"

var (
	ErrUniqueConstraint          = errors.New("unique constraint")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrNoUpdateFields            = errors.New("no update fields")
	ErrNotFound                  = errors.New("resource not found")
)
