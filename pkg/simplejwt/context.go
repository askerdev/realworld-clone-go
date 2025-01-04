package simplejwt

import (
	"context"
	"errors"

	"github.com/askerdev/realworld-clone-go/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
)

type middlewareContextKey string

const tokenContextKey middlewareContextKey = "token"
const userContextKey middlewareContextKey = "user"

func ContextWithToken(ctx context.Context, token *jwt.Token) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

func ContextWithUser(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func ContextToken(ctx context.Context) (*jwt.Token, error) {
	val := ctx.Value(tokenContextKey)
	if val == nil {
		return nil, errors.New("no token in context")
	}

	t, ok := val.(*jwt.Token)
	if !ok {
		return nil, errors.New("unexpected token type in context")
	}

	return t, nil
}

func MustContextToken(ctx context.Context) *jwt.Token {
	t, err := ContextToken(ctx)
	if err != nil {
		panic(err)
	}

	return t
}
