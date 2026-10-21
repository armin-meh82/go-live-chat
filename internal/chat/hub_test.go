package chat

import "testing"

func newTestClient(hub *Hub, roomID string) *Client {
	client := NewClient(nil, roomID, "test-user")
	hub.Register(client)
	return client
}

func TestBroadcastOnlyToSameRoom(t *testing.T) {
	hub := NewHub()

	clientA := newTestClient(hub, "room-1")
	clientB := newTestClient(hub, "room-1")
	clientC := newTestClient(hub, "room-2")

	delivered := hub.Broadcast("room-1", []byte(`{"message":"Hello"}`))

	if delivered != 2 {
		t.Fatalf("expected 2 deliveries, got %d", delivered)
	}

	for _, client := range []*Client{clientA, clientB} {
		select {
		case message := <-client.Send:
			if string(message) != `{"message":"Hello"}` {
				t.Errorf("unexpected message: %s", message)
			}
		default:
			t.Error("expected client to receive message")
		}
	}

	select {
	case <-clientC.Send:
		t.Fatal("client in room-2 received room-1 message")
	default:
	}

	hub.Unregister(clientA)
	hub.Unregister(clientB)
	hub.Unregister(clientC)
}

func TestUnregisterRemovesClient(t *testing.T) {
	hub := NewHub()
	client := newTestClient(hub, "room-1")

	hub.Unregister(client)

	count, active := hub.RoomInfo("room-1")

	if active || count != 0 {
		t.Fatalf("expected empty room, got count=%d active=%v", count, active)
	}
}
