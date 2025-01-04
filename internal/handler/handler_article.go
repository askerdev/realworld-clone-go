package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/askerdev/realworld-clone-go/internal/postgres"
	"github.com/gosimple/slug"
	"github.com/guregu/null/v5"
)

type CreateArticleRequestArticle struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

type CreateArticleRequest struct {
	Article CreateArticleRequestArticle `json:"article"`
}

func (h *handler) createArticle(w http.ResponseWriter, r *http.Request) {
	var body CreateArticleRequest
	if err := ParseBody(r.Body, &body); err != nil {
		InvalidJSON(w)
		return
	}

	u := h.MustContextUser(r.Context())
	slug := slug.Make(body.Article.Title)
	article, err := h.storage.CreateArticle(
		r.Context(),
		&postgres.CreateArticleParams{
			AuthorID:    u.ID,
			Slug:        slug,
			Title:       body.Article.Title,
			Description: body.Article.Description,
			Body:        body.Article.Body,
			TagList:     body.Article.TagList,
		},
	)
	if err != nil {
		slog.Error(err.Error())
		InternalServerError(w)
		return
	}

	JSON(w, map[string]any{
		"article": article,
	})
}

func (h *handler) listArticle(w http.ResponseWriter, r *http.Request) {
	var id *uint64
	token, err := h.getHeaderToken(r.Header)
	if err == nil {
		user, err := h.userFromToken(token)
		if err == nil {
			id = &user.ID
		}
	}

	author := r.URL.Query().Get("author")
	tag := r.URL.Query().Get("tag")
	favorited := r.URL.Query().Get("favorited")
	var limit null.Int
	var offset null.Int
	limitString := r.URL.Query().Get("limit")
	if len(limitString) > 0 {
		limitInt, err := strconv.Atoi(limitString)
		if err == nil && limitInt > 0 && limitInt <= 20 {
			limit.Int64 = int64(limitInt)
			limit.Valid = true
		}
	}
	offsetString := r.URL.Query().Get("offset")
	if len(offsetString) > 0 {
		offsetInt, err := strconv.Atoi(offsetString)
		if err == nil && offsetInt > 0 && offsetInt <= 20 {
			offset.Int64 = int64(offsetInt)
			offset.Valid = true
		}
	}

	articles, err := h.storage.SelectArticles(r.Context(), &postgres.SelectArticlesParams{
		UserID:              id,
		AuthorUsername:      null.NewString(author, len(author) > 0),
		Tag:                 null.NewString(tag, len(tag) > 0),
		FavoritedByUsername: null.NewString(favorited, len(favorited) > 0),
		Limit:               limit,
		Offset:              offset,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			JSON(w, map[string]any{
				"articles": []any{},
			})
			break
		default:
			slog.Error(err.Error())
			InternalServerError(w)
			break
		}
		return
	}

	JSON(w, map[string]any{
		"articles": articles,
	})
}
