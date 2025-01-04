package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/askerdev/realworld-clone-go/internal/domain/entity"
	"github.com/askerdev/realworld-clone-go/internal/postgres"
	"github.com/askerdev/realworld-clone-go/pkg/simplejwt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/guregu/null/v5"
	"github.com/jmoiron/sqlx"
)

type handler struct {
	storage       *postgres.Storage
	issuer        *simplejwt.Issuer
	validator     *simplejwt.Validator
	jwtMiddleware *simplejwt.Middleware
}

func New(db *sqlx.DB, issuer *simplejwt.Issuer, validator *simplejwt.Validator) *handler {
	return &handler{
		storage:       postgres.NewStorage(db),
		issuer:        issuer,
		validator:     validator,
		jwtMiddleware: simplejwt.NewMiddleware(validator),
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := chi.NewRouter()

	m.Get("/health", h.healthCheck)
	m.Post("/api/users", h.register)
	m.Post("/api/users/login", h.login)
	m.Get("/api/profiles/{username}", h.profile)
	m.Get("/api/articles", h.listArticle)

	m.With(h.jwtMiddleware.HandleHTTP).
		Group(func(r chi.Router) {
			r.Get("/api/user", h.user)
			r.Put("/api/user", h.updateUser)
			r.Post("/api/profiles/{username}/follow", h.follow)
			r.Delete("/api/profiles/{username}/follow", h.unfollow)
			r.Post("/api/articles", h.createArticle)
		})

	m.ServeHTTP(w, r)
}

func (h *handler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "ok!",
	})
}

func (h *handler) getHeaderToken(header http.Header) (*jwt.Token, error) {
	auth := header.Get("Authorization")
	if len(auth) < 8 || !strings.HasPrefix(auth, "Token ") {
		return nil, fmt.Errorf("invalid header")
	}

	tokenString := strings.TrimPrefix(auth, "Token ")

	token, err := h.validator.Validate(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil
}

func (h *handler) userFromToken(token *jwt.Token) (*entity.User, error) {

	claims := token.Claims.(jwt.MapClaims)
	sub, ok := claims["sub"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	id, ok := sub["id"].(float64)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	email, ok := sub["email"].(string)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	username, ok := sub["username"].(string)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	bio, ok := sub["bio"].(string)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	var image null.String
	if img, ok := sub["image"].(string); ok {
		image.Valid = true
		image.String = img
	}

	return &entity.User{
		ID:       uint64(id),
		Email:    email,
		Username: username,
		Bio:      bio,
		Image:    image,
	}, nil
}

func (h *handler) ContextUser(ctx context.Context) (*entity.User, error) {
	token, err := simplejwt.ContextToken(ctx)
	if err != nil {
		return nil, err
	}
	return h.userFromToken(token)
}

func (h *handler) MustContextUser(ctx context.Context) *entity.User {
	u, err := h.ContextUser(ctx)
	if err != nil {
		panic(err)
	}

	return u
}
