package entity

import (
	"time"

	"github.com/guregu/null/v5"
)

type Article struct {
	ID             uint64    `json:"id"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount uint64    `json:"favoritesCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      null.Time `json:"updatedAt"`
	Author         *Profile  `json:"author"`
}
