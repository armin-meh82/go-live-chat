package chat

import "github.com/gorilla/websocket"

type Client struct {
	Conn   *websocket.Conn
	RoomID string
	UserID string
	Send   chan []byte
}

func NewClient(
	conn *websocket.Conn,
	roomID string,
	userID string,
) *Client {
	return &Client{
		Conn:   conn,
		RoomID: roomID,
		UserID: userID,
		Send:   make(chan []byte, 32),
	}
}
