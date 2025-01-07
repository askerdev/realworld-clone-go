package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/askerdev/realworld-clone-go/internal/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/guregu/null/v5"
)

func (h *handler) listComments(w http.ResponseWriter, r *http.Request) {
	slugString := chi.URLParam(r, "slug")
	slug := null.NewString(slugString, len(slugString) > 0)

	if !slug.Valid {
		NotFoundError(w)
		return
	}

	var userID *uint64
	token, err := h.getHeaderToken(r.Header)
	if err == nil {
		user, err := h.userFromToken(token)
		if err == nil {
			userID = &user.ID
		}
	}

	comments, err := h.storage.SelectComments(r.Context(), &postgres.SelectCommentsParams{
		UserID:      userID,
		ArticleSlug: slug.String,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			JSON(w, map[string]any{
				"comments": []any{},
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
		"comments": comments,
	})
}

type CreateCommentRequestComment struct {
	Body string `json:"body"`
}

type CreateCommentRequest struct {
	Comment CreateCommentRequestComment `json:"comment"`
}

func (h *handler) createComment(w http.ResponseWriter, r *http.Request) {
	slugString := chi.URLParam(r, "slug")
	slug := null.NewString(slugString, len(slugString) > 0)

	if !slug.Valid {
		NotFoundError(w)
		return
	}

	var body CreateCommentRequest
	if err := ParseBody(r.Body, &body); err != nil {
		InvalidJSON(w)
		return
	}

	u := h.MustContextUser(r.Context())

	comment, err := h.storage.InsertComment(
		r.Context(),
		&postgres.InsertCommentParams{
			ArticleSlug: slug.String,
			UserID:      u.ID,
			Body:        body.Comment.Body,
		},
	)
	if err != nil {
		slog.Error(err.Error())
		AlreayExistsError(w)
		return
	}

	JSON(w, map[string]any{
		"comment": comment,
	})
}

func (h *handler) deleteComment(w http.ResponseWriter, r *http.Request) {
	slugString := chi.URLParam(r, "slug")
	slug := null.NewString(slugString, len(slugString) > 0)

	var commentID null.Int
	commentIDString := chi.URLParam(r, "id")
	if len(commentIDString) > 0 {
		commentIDInt, err := strconv.Atoi(commentIDString)
		fmt.Println(err)
		if err == nil && commentIDInt >= 0 {
			commentID.Int64 = int64(commentIDInt)
			commentID.Valid = true
		}
	}

	if !slug.Valid || !commentID.Valid {
		NotFoundError(w)
		return
	}

	u := h.MustContextUser(r.Context())

	err := h.storage.DeleteComment(
		r.Context(),
		&postgres.DeleteCommentParams{
			CommentID:   uint64(commentID.Int64),
			ArticleSlug: slug.String,
			UserID:      u.ID,
		},
	)
	if err != nil {
		slog.Error(err.Error())
		InternalServerError(w)
		return
	}
}
