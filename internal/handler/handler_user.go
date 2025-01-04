package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/askerdev/realworld-clone-go/internal/domain/vo"
	"github.com/askerdev/realworld-clone-go/internal/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/guregu/null/v5"
)

type RegisterRequestUser struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	User RegisterRequestUser `json:"user"`
}

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	var body RegisterRequest
	if err := ParseBody(r.Body, &body); err != nil {
		InvalidJSON(w)
		return
	}

	var errs FieldErrMap
	email, err := vo.NewEmail(body.User.Email)
	errs.AppendErr("email", err)

	username, err := vo.NewUsername(body.User.Username)
	errs.AppendErr("username", err)

	password, err := vo.NewPassword(body.User.Password)
	errs.AppendErr("password", err)

	if !errs.Empty() {
		ValidationError(w, errs)
		return
	}

	passHash, err := password.Hash()
	if err != nil {
		slog.Error("password hashing error", slog.String("msg", err.Error()))
		InternalServerError(w)
		return
	}

	u, err := h.storage.InsertUser(r.Context(), string(email), string(username), passHash)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUniqueConstraint):
			{
				ValidationError(w, FieldErrMap{
					"email": {"email or username already exists"},
				})
			}
			break
		default:
			{
				slog.Error("inserting user", slog.String("msg", err.Error()))
				InternalServerError(w)
			}
			break
		}
		return
	}

	JSON(w, map[string]any{
		"user": u,
	})
}

type LoginRequestUser struct {
	Email string `json:"email"`
	Pass  string `json:"password"`
}

type LoginRequest struct {
	User LoginRequestUser `json:"user"`
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := ParseBody(r.Body, &body); err != nil {
		InvalidJSON(w)
		return
	}

	var errs FieldErrMap
	email, err := vo.NewEmail(body.User.Email)
	errs.AppendErr("email", err)

	password, err := vo.NewPassword(body.User.Pass)
	errs.AppendErr("password", err)

	if !errs.Empty() {
		ValidationError(w, errs)
		return
	}

	u, err := h.storage.SelectUserByEmail(r.Context(), string(email))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			ValidationError(w, FieldErrMap{
				"email": {"email does not exists"},
			})
			break
		default:
			InternalServerError(w)
			break
		}
		return
	}

	if !password.Compare(u.Password) {
		ValidationError(w, FieldErrMap{
			"password": {"invalid password"},
		})
		return
	}

	token, err := h.issuer.Token(u)
	if err != nil {
		InternalServerError(w)
		return
	}

	u.Token = &token

	JSON(w, map[string]any{
		"user": u,
	})
}

func (h *handler) user(w http.ResponseWriter, r *http.Request) {
	JSON(w, map[string]any{
		"user": h.MustContextUser(r.Context()),
	})
}

type UpdateUserRequestUser struct {
	Email    null.String `json:"email"`
	Username null.String `json:"username"`
	Password null.String `json:"password"`
	Image    null.String `json:"image"`
	Bio      null.String `json:"bio"`
}

type UpdateUserRequest struct {
	User UpdateUserRequestUser `json:"user"`
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var body UpdateUserRequest
	if err := ParseBody(r.Body, &body); err != nil {
		InvalidJSON(w)
		return
	}

	if body.User.Password.Valid {
		var err error
		body.User.Password.String, err = vo.Password(body.User.Password.String).Hash()
		if err != nil {
			InternalServerError(w)
			return
		}
	}

	u := h.MustContextUser(r.Context())
	updatedUser, err := h.storage.UpdateUser(r.Context(), &postgres.UpdateUserParams{
		ID:       u.ID,
		Email:    body.User.Email,
		Username: body.User.Username,
		Password: body.User.Password,
		Image:    body.User.Image,
		Bio:      body.User.Bio,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error(err.Error())
		InternalServerError(w)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		slog.Error(err.Error())
		updatedUser = u
	}

	JSON(w, map[string]any{
		"user": updatedUser,
	})
}

func (h *handler) profile(w http.ResponseWriter, r *http.Request) {
	var id *uint64
	token, err := h.getHeaderToken(r.Header)
	if err == nil {
		user, err := h.userFromToken(token)
		if err == nil {
			id = &user.ID
		}
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		NotFoundError(w)
		return
	}

	profile, err := h.storage.SelectProfileByUsername(r.Context(), username, id)
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

	JSON(w, map[string]any{
		"profile": profile,
	})
}

func (h *handler) follow(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		NotFoundError(w)
		return
	}

	u := h.MustContextUser(r.Context())

	if username == u.Username {
		InternalServerError(w)
		return
	}

	profile, err := h.storage.FollowProfile(r.Context(), u.ID, username)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrSubscriptionAlreadyExists):
			NewError(err.Error(), http.StatusBadRequest).Write(w)
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

	JSON(w, map[string]any{
		"profile": profile,
	})
}

func (h *handler) unfollow(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		NotFoundError(w)
		return
	}

	u := h.MustContextUser(r.Context())

	if username == u.Username {
		InternalServerError(w)
		return
	}

	profile, err := h.storage.UnfollowProfile(r.Context(), u.ID, username)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrSubscriptionAlreadyExists):
			NewError(err.Error(), http.StatusBadRequest).Write(w)
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

	JSON(w, map[string]any{
		"profile": profile,
	})
}
