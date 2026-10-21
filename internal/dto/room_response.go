package dto

type RoomResponseDTO struct {
	RoomID         string `json:"roomId"`
	ConnectedUsers int    `json:"connectedUsers"`
	Active         bool   `json:"active"`
}