package entity

import "github.com/guregu/null/v5"

type User struct {
	ID       uint64      `json:"id"              db:"id"`
	Email    string      `json:"email"           db:"email"`
	Username string      `json:"username"        db:"username"`
	Bio      string      `json:"bio"             db:"bio"`
	Image    null.String `json:"image"           db:"image"`
	Password string      `json:"-"               db:"password"`
	Token    *string     `json:"token,omitempty"`
}
