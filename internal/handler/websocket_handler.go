package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"

	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/dto"
	"github.com/armin-meh82/live-chat/internal/service"
)

const (
	maxMessageSize = 8192
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	writeWait      = 10 * time.Second
)

type WebSocketHandler struct {
	hub         *chat.Hub
	chatService *service.ChatService
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(
	hub *chat.Hub,
	chatService *service.ChatService,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}

				parsed, err := url.Parse(origin)
				return err == nil && parsed.Host == r.Host
			},
		},
	}
}

func (h *WebSocketHandler) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	roomID := r.PathValue("roomID")

	if !service.ValidRoomID(roomID) {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: "invalid room ID",
		})
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	userID, err := generateUserID()
	if err != nil {
		_ = conn.Close()
		return
	}

	client := chat.NewClient(conn, roomID, userID)
	h.hub.Register(client)

	go h.writePump(client)
	h.readPump(client)
}

func (h *WebSocketHandler) readPump(client *chat.Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)

	if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	client.Conn.SetPongHandler(func(string) error {
		return client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	limiter := rate.NewLimiter(rate.Limit(5), 10)

	for {
		messageType, data, err := client.Conn.ReadMessage()
		if err != nil {
			return
		}

		if messageType != websocket.TextMessage {
			if !sendError(client, "only text messages are supported") {
				return
			}
			continue
		}

		if !limiter.Allow() {
			if !sendError(client, "rate limit exceeded") {
				return
			}
			continue
		}

		var incoming dto.IncomingMessageDTO

		if err := json.Unmarshal(data, &incoming); err != nil {
			if !sendError(client, "invalid JSON message") {
				return
			}
			continue
		}

		if err := h.chatService.ProcessMessage(
			client.RoomID,
			client.UserID,
			incoming,
		); err != nil {
			if !sendError(client, "invalid message") {
				return
			}
		}
	}
}

func (h *WebSocketHandler) writePump(client *chat.Client) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}

			if !ok {
				_ = client.Conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				)
				return
			}

			if err := client.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {
				return
			}

		case <-ticker.C:
			if err := client.Conn.WriteControl(
				websocket.PingMessage,
				nil,
				time.Now().Add(writeWait),
			); err != nil {
				return
			}
		}
	}
}

func sendError(client *chat.Client, message string) bool {
	data, err := json.Marshal(dto.ErrorResponseDTO{
		Error: message,
	})
	if err != nil {
		return false
	}

	select {
	case client.Send <- data:
		return true
	default:
		return false
	}
}

func generateUserID() (string, error) {
	data := make([]byte, 16)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}
