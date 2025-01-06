package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/askerdev/realworld-clone-go/internal/postgres"
	"github.com/go-chi/chi/v5"
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
		AlreayExistsError(w)
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

func (h *handler) articleBySlug(w http.ResponseWriter, r *http.Request) {
	slugString := chi.URLParam(r, "slug")
	slug := null.NewString(slugString, len(slugString) > 0)

	if !slug.Valid {
		NotFoundError(w)
		return
	}

	article, err := h.storage.SelectArticles(r.Context(), &postgres.SelectArticlesParams{
		Slug:  slug,
		Limit: null.IntFrom(1),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			NotFoundError(w)
			break
		default:
			InternalServerError(w)
			break
		}
		return
	}
	if len(article) == 0 {
		NotFoundError(w)
		return
	}

	JSON(w, map[string]any{
		"article": article[0],
	})
}

type UpdateArticleRequestArticle struct {
	Title       null.String `json:"title"`
	Description null.String `json:"description"`
	Body        null.String `json:"body"`
}

type UpdateArticleRequest struct {
	Article UpdateArticleRequestArticle `json:"article"`
}

func (h *handler) updateArticle(w http.ResponseWriter, r *http.Request) {
	slugString := chi.URLParam(r, "slug")
	slugField := null.NewString(slugString, len(slugString) > 0)

	var body UpdateArticleRequest
	if err := ParseBody(r.Body, &body); err != nil {
		InvalidJSON(w)
		return
	}

	if !slugField.Valid {
		NotFoundError(w)
		return
	}

	var newSlug null.String
	if body.Article.Title.Valid {
		newSlug = null.StringFrom(slug.Make(body.Article.Title.String))
	}

	err := h.storage.UpdateArticle(r.Context(), &postgres.UpdateArticleParams{
		OriginalSlug: slugField.String,
		Slug:         newSlug,
		Title:        body.Article.Title,
		Description:  body.Article.Description,
		Body:         body.Article.Body,
	})

	if err != nil {
		InternalServerError(w)
		return
	}

	if newSlug.Valid {
		slugField = newSlug
	}

	article, err := h.storage.SelectArticles(r.Context(), &postgres.SelectArticlesParams{
		Slug:  slugField,
		Limit: null.IntFrom(1),
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			NotFoundError(w)
			break
		default:
			InternalServerError(w)
			break
		}
		return
	}
	if len(article) == 0 {
		NotFoundError(w)
		return
	}

	JSON(w, map[string]any{
		"article": article[0],
	})
}

func (h *handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	slugString := chi.URLParam(r, "slug")
	slug := null.NewString(slugString, len(slugString) > 0)

	if !slug.Valid {
		NotFoundError(w)
		return
	}

	u := h.MustContextUser(r.Context())
	err := h.storage.RemoveArticle(r.Context(), slug.String, u.ID)
	if err != nil {
		slog.Error(err.Error())
		switch {
		case errors.Is(err, postgres.ErrNotFound):
			NotFoundError(w)
			break
		case errors.Is(err, sql.ErrNoRows):
			NotFoundError(w)
			break
		default:
			InternalServerError(w)
			break
		}
		return
	}
}
