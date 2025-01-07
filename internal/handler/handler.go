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
	m := http.NewServeMux()

	// chiR := chi.NewRouter()

	// chiR.Get("/health", h.healthCheck)
	m.HandleFunc("GET /health", h.healthCheck)
	// chiR.Post("/api/users", h.register)
	m.HandleFunc("POST /api/users", h.register)
	// chiR.Post("/api/users/login", h.login)
	m.HandleFunc("POST /api/users/login", h.login)
	// chiR.Get("/api/profiles/{username}", h.profile)
	m.HandleFunc("GET /api/profiles/{username}", h.profile)
	// chiR.Get("/api/articles", h.listArticle)
	m.HandleFunc("GET /api/articles", h.listArticle)
	// chiR.Get("/api/articles/{slug}", h.articleBySlug)
	m.HandleFunc("GET /api/articles/{slug}", h.articleBySlug)
	// r.Get("/api/articles/feed", h.feedArticles)
	m.HandleFunc(
		"GET /api/articles/feed",
		// this is done like that because "most specific" rule not working with subrouting
		h.jwtMiddleware.HandleHTTP(http.HandlerFunc(h.feedArticles)).ServeHTTP,
	)
	// chiR.Get("/api/tags", h.listTags)
	m.HandleFunc("GET /api/tags", h.listTags)
	// chiR.Get("/api/articles/{slug}/comments", h.listComments)
	m.HandleFunc("GET /api/articles/{slug}/comments", h.listComments)

	auth := http.NewServeMux()
	// r.Get("/api/user", h.user)
	auth.HandleFunc("GET /api/user", h.user)
	// r.Put("/api/user", h.updateUser)
	auth.HandleFunc("PUT /api/user", h.updateUser)
	// r.Post("/api/profiles/{username}/follow", h.follow)
	auth.HandleFunc("POST /api/profiles/{username}/follow", h.follow)
	// r.Delete("/api/profiles/{username}/follow", h.unfollow)
	auth.HandleFunc("DELETE /api/profiles/{username}/follow", h.unfollow)
	// r.Post("/api/articles", h.createArticle)
	auth.HandleFunc("POST /api/articles", h.createArticle)
	// r.Put("/api/articles/{slug}", h.updateArticle)
	auth.HandleFunc("PUT /api/articles/{slug}", h.updateArticle)
	// r.Delete("/api/articles/{slug}", h.deleteArticle)
	auth.HandleFunc("DELETE /api/articles/{slug}", h.deleteArticle)
	// r.Post("/api/articles/{slug}/favorite", h.favoriteArticle)
	auth.HandleFunc("POST /api/articles/{slug}/favorite", h.favoriteArticle)
	// r.Delete("/api/articles/{slug}/favorite", h.unfavoriteArticle)
	auth.HandleFunc("DELETE /api/articles/{slug}/favorite", h.unfavoriteArticle)
	// r.Post("/api/articles/{slug}/comments", h.createComment)
	auth.HandleFunc("POST /api/articles/{slug}/comments", h.createComment)
	// r.Delete("/api/articles/{slug}/comments/{id}", h.deleteComment)
	auth.HandleFunc("DELETE /api/articles/{slug}/comments/{id}", h.deleteComment)

	m.Handle("/", h.jwtMiddleware.HandleHTTP(auth))

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
