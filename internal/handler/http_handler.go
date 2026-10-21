package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/armin-meh82/live-chat/internal/chat"
	"github.com/armin-meh82/live-chat/internal/dto"
	"github.com/armin-meh82/live-chat/internal/service"
)

type HTTPHandler struct {
	roomService *service.RoomService
}

func NewHTTPHandler(
	hub *chat.Hub,
	roomService *service.RoomService,
) *HTTPHandler {
	return &HTTPHandler{
		roomService: roomService,
	}
}

func (h *HTTPHandler) RegisterRoutes(
	mux *http.ServeMux,
	wsHandler *WebSocketHandler,
) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/v1/rooms/{roomID}", h.GetRoom)
	mux.Handle("GET /ws/rooms/{roomID}", wsHandler)
}

func (h *HTTPHandler) Health(
	w http.ResponseWriter,
	r *http.Request,
) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *HTTPHandler) GetRoom(
	w http.ResponseWriter,
	r *http.Request,
) {
	roomID := r.PathValue("roomID")

	room, err := h.roomService.GetRoom(roomID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRoomID) {
			writeJSON(w, http.StatusBadRequest, dto.ErrorResponseDTO{
				Error: "invalid room ID",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
