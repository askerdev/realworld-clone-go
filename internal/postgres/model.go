package postgres

import (
	"time"

	"github.com/guregu/null/v5"
)

type ArticleRow struct {
	ID             uint64    `db:"id"`
	Slug           string    `db:"slug"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	Body           string    `db:"body"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      null.Time `db:"updated_at"`
	AuthordID      uint64    `db:"author_id"`
	FavoritesCount uint64    `db:"favorites_count"`
}

type ArticleRowWithTagAndUser struct {
	ID             uint64      `db:"id"`
	Slug           string      `db:"slug"`
	Title          string      `db:"title"`
	Description    string      `db:"description"`
	Body           string      `db:"body"`
	CreatedAt      time.Time   `db:"created_at"`
	UpdatedAt      null.Time   `db:"updated_at"`
	AuthordID      uint64      `db:"author_id"`
	FavoritesCount uint64      `db:"favorites_count"`
	Tag            null.String `db:"article_tag"`
	SubscriberID   *uint64     `db:"subscriber_id"`
	UserID         uint64      `db:"user_id"`
	UserUsername   string      `db:"user_username"`
	UserImage      null.String `db:"user_image"`
	UserBio        string      `db:"user_bio"`
}
