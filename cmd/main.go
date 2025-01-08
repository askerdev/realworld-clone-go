package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/askerdev/realworld-clone-go/internal/handler"
	"github.com/askerdev/realworld-clone-go/internal/mem"
	"github.com/askerdev/realworld-clone-go/pkg/simplejwt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	privateKeyPath, publicKeyPath := os.Args[1], os.Args[2]

	jwtCache := mem.NewJWTCache()

	issuer, err := simplejwt.NewIssuer(privateKeyPath, jwtCache)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	validator, err := simplejwt.NewValidator(publicKeyPath, jwtCache)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	db, err := sqlx.Connect("pgx", os.Getenv("POSTGRES_URL"))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	err = db.Ping()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	h := handler.New(
		db,
		issuer,
		validator,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: h,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()

		slog.Warn("Shutting down server!")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error(err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
			return
		}
	}()

	slog.Info("Listening on :8080")

	wg.Wait()

	slog.Warn("Server stopped!")
}
