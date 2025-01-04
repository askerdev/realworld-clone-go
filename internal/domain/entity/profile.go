package entity

import "github.com/guregu/null/v5"

type Profile struct {
	ID        uint64      `json:"-"         db:"id"`
	Username  string      `json:"username"  db:"username"`
	Bio       string      `json:"bio"       db:"bio"`
	Image     null.String `json:"image"     db:"image"`
	Following bool        `json:"following" db:"following"`
}
