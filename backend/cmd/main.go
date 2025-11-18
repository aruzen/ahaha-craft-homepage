package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/handler"
	infraDB "backend/internal/infra/db"
	"backend/internal/repository"
	"backend/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := infraDB.NewConnection(ctx)
	if err != nil {
		logger.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	server := newHTTPServer(pool, logger)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Printf("graceful shutdown error: %v", err)
		}
	}()

	logger.Printf("server listening on %s", server.Addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatalf("server stopped with error: %v", err)
	}

	logger.Println("server stopped")
}

func newHTTPServer(pool *pgxpool.Pool, logger *log.Logger) *http.Server {
	return &http.Server{
		Addr:              serverAddr(),
		Handler:           newHTTPHandler(pool),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          logger,
	}
}

func newHTTPHandler(pool *pgxpool.Pool) http.Handler {
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewLoginSessionRepository(pool)
	hueRepo := repository.NewHueRepository(pool)

	signInService := service.NewSignInService(userRepo, sessionRepo)
	loginService := service.NewLoginService(userRepo, sessionRepo)
	hueSaveService := service.NewHueSaveService(hueRepo)
	hueGetService := service.NewHueGetService(hueRepo, sessionRepo)

	mux := http.NewServeMux()
	mux.Handle("/api/sign-in", handler.NewSignInHandler(signInService))
	mux.Handle("/api/login", handler.NewLoginHandler(loginService))
	mux.Handle("/api/hue-are-you/save-result", handler.NewHueSaveHandler(hueSaveService))
	mux.Handle("/api/hue-are-you/get-data", handler.NewHueGetHandler(hueGetService))

	return mux
}

func serverAddr() string {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
