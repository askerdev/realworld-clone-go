package simplejwt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Middleware struct {
	validator *Validator
}

func NewMiddleware(validator *Validator) *Middleware {
	return &Middleware{
		validator: validator,
	}
}

func (m *Middleware) HandleHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.getHeaderToken(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"statusCode": http.StatusUnauthorized,
				"message":    err.Error(),
			})
			return
		}

		ctx := ContextWithToken(r.Context(), token)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) getHeaderToken(header http.Header) (*jwt.Token, error) {
	auth := header.Get("Authorization")
	if len(auth) < 8 || !strings.HasPrefix(auth, "Token ") {
		return nil, fmt.Errorf("invalid header")
	}

	tokenString := strings.TrimPrefix(auth, "Token ")

	token, err := m.validator.Validate(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return token, nil
}
