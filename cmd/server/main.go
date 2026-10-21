package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/armin-meh82/live-chat/cmd/config"
	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/database"
	"github.com/armin-meh82/live-chat/internal/handler"
	"github.com/armin-meh82/live-chat/internal/ratelimit"
	"github.com/armin-meh82/live-chat/internal/repository"
	"github.com/armin-meh82/live-chat/internal/service"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found; using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config loading failed: ", err)
	}

	db, err := database.OpenPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("postgres connection failed: ", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("get postgres connection failed: ", err)
	}
	defer sqlDB.Close()

	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("invalid REDIS_URL: ", err)
	}

	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	startupCtx, startupCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer startupCancel()

	if err := redisClient.Ping(startupCtx).Err(); err != nil {
		log.Fatal("redis connection failed: ", err)
	}

	hub := chat.NewHub()

	messageRepo := repository.NewGormMessageRepository(db)
	chatService := service.NewChatService(hub, messageRepo)
	roomService := service.NewRoomService(hub)

	limiter := ratelimit.NewRedisLimiter(
		redisClient,
		5,
		20*time.Second,
	)

	wsHandler := handler.NewWebSocketHandler(
		hub,
		chatService,
		limiter,
	)
	httpHandler := handler.NewHTTPHandler(hub, roomService)

	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux, wsHandler)

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("server listening on http://localhost:%s", cfg.ServerPort)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}

	case <-ctx.Done():
		log.Println("shutting down server")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			_ = server.Close()
		}
	}

	log.Println("server stopped")
}
