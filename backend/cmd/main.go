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
		Handler:           newHTTPHandler(pool, logger),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          logger,
	}
}

func newHTTPHandler(pool *pgxpool.Pool, logger *log.Logger) http.Handler {
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewLoginSessionRepository(pool)
	hueRepo := repository.NewHueRepository(pool)

	signInService := service.NewSignInService(userRepo, sessionRepo, logger)
	loginService := service.NewLoginService(userRepo, sessionRepo, logger)
	hueSaveService := service.NewHueSaveService(hueRepo, logger)
	hueGetService := service.NewHueGetService(hueRepo, sessionRepo, logger)

	mux := http.NewServeMux()
	mux.Handle("/api/sign-in", withCORS(handler.NewSignInHandler(signInService)))
	mux.Handle("/api/login", withCORS(handler.NewLoginHandler(loginService)))
	mux.Handle("/api/hue-are-you/save-result", withCORS(handler.NewHueSaveHandler(hueSaveService)))
	mux.Handle("/api/hue-are-you/get-data", withCORS(handler.NewHueGetHandler(hueGetService)))

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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
