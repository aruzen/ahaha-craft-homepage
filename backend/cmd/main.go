package main

import (
	"context"
	"errors"
	"fmt"
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
	hueCfg, err := loadHueSaveConfig()
	if err != nil {
		logger.Fatalf("hue save config error: %v", err)
	}
	hueSaveService, err := service.NewHueSaveService(hueRepo, logger, hueCfg)
	if err != nil {
		logger.Fatalf("hue save service init error: %v", err)
	}
	hueGetService := service.NewHueGetService(hueRepo, sessionRepo, userRepo, logger)

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

func loadHueSaveConfig() (service.HueSaveConfig, error) {
	endpoint := strings.TrimSpace(os.Getenv("HUE_API_ENDPOINT"))
	if endpoint == "" {
		return service.HueSaveConfig{}, fmt.Errorf("HUE_API_ENDPOINT is required")
	}
	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		return service.HueSaveConfig{}, fmt.Errorf("OPENAI_API_KEY is required")
	}
	return service.HueSaveConfig{
		Endpoint: endpoint,
		APIKey:   apiKey,
		SystemPrompt: `
あなたは心理テスト「Hue Are You」の結果生成AIです。
各ワードに対して選択された色から心理的特徴を分析し、最終的なrgb値(0〜255)と2〜4文程度の日本語メッセージを返してください。
分析には、選ばれた色に意識を向けるより「普通の人ならこう選ぶところをこの人はこの色を選んだので、こういう人なのだろう」という推察もしてください。
メッセージは分析の結果を伝えるのではなくふんわりした内容で、いいサービスだったと思ってもらえる分にしましょう。
`,
	}, nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := map[string]bool{
			"http://localhost:3000":   true,
			"http://ahaha-craft.org":  true,
			"https://ahaha-craft.org": true,
		}

		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")

			w.Header().Set("Access-Control-Allow-Methods",
				"GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers",
				"Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204
			return
		}

		next.ServeHTTP(w, r)
	})
}
