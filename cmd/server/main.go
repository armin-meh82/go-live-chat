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

	"github.com/joho/godotenv"

	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/handler"
	"github.com/armin-meh82/live-chat/internal/service"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config loading failed:", err)
	}

	hub := chat.NewHub()

	chatService := service.NewChatService(hub)
	roomService := service.NewRoomService(hub)

	wsHandler := handler.NewWebSocketHandler(hub, chatService)
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
