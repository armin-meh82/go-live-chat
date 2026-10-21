package chat

import "sync"

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.RoomID] == nil {
		h.rooms[client.RoomID] = make(map[*Client]struct{})
	}

	h.rooms[client.RoomID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[client.RoomID]
	if !exists {
		return
	}

	if _, exists := room[client]; !exists {
		return
	}

	delete(room, client)
	close(client.Send)

	if len(room) == 0 {
		delete(h.rooms, client.RoomID)
	}
}

func (h *Hub) Broadcast(roomID string, message []byte) int {
	h.mu.RLock()

	room := h.rooms[roomID]
	var slowClients []*Client
	delivered := 0

	for client := range room {
		select {
		case client.Send <- message:
			delivered++
		default:
			slowClients = append(slowClients, client)
		}
	}

	h.mu.RUnlock()

	for _, client := range slowClients {
		h.Unregister(client)
		_ = client.Conn.Close()
	}

	return delivered
}

func (h *Hub) RoomInfo(roomID string) (int, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return 0, false
	}

	return len(room), true
}
