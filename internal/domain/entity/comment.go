package entity

import (
	"time"

	"github.com/guregu/null/v5"
)

type Comment struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt null.Time `json:"updatedAt"`
	Body      string    `json:"body"`
	Author    Profile   `json:"author"`
}
