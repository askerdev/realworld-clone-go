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
	UpdatedAt      time.Time `db:"updated_at"`
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
	UpdatedAt      time.Time   `db:"updated_at"`
	AuthordID      uint64      `db:"author_id"`
	FavoritesCount uint64      `db:"favorites_count"`
	Tag            null.String `db:"article_tag"`
	SubscriberID   *uint64     `db:"subscriber_id"`
	UserID         uint64      `db:"user_id"`
	UserUsername   string      `db:"user_username"`
	UserImage      null.String `db:"user_image"`
	UserBio        string      `db:"user_bio"`
	ArticlesCount  uint        `db:"articles_count"`
	FavoritedByID  *uint64     `db:"favorited_by_id"`
}

type CommentRow struct {
	ID           uint64      `db:"id"`
	Body         string      `db:"body"`
	AuthorID     uint64      `db:"author_id"`
	ArticleID    uint64      `db:"article_id"`
	UserUsername string      `db:"user_username"`
	UserImage    null.String `db:"user_image"`
	UserBio      string      `db:"user_bio"`
	SubscriberID *uint64     `db:"subscriber_id"`
	CreatedAt    time.Time   `db:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at"`
}
