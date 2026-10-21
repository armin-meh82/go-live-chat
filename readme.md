# Go Live Chat

A real-time multi-room chat service built with Go and WebSocket for live-streaming platforms.

## Features

- Real-time messaging with WebSocket
- Multiple isolated chat rooms
- Concurrent client management using goroutines and channels
- Ping/Pong heartbeat and disconnected-client handling
- Message validation and a 500-character limit
- Redis-based rate limiting: 5 messages per 20 seconds per connection
- PostgreSQL message persistence using GORM
- HTTP health check and room information endpoints

## Tech Stack

- Go
- Gorilla WebSocket
- PostgreSQL
- GORM
- Redis
- `net/http`
- `godotenv`

## Getting Started

### Requirements

- Go
- PostgreSQL
- Redis

### Installation

Clone the repository:

```bash
git clone https://github.com/armin-meh82/go-live-chat.git
cd go-live-chat
```

Create a `.env` file based on `.env.example` and configure your database, Redis, and server settings.

Install dependencies and run the server:

```bash
go mod download
go run ./cmd/server
```

The API examples below use port `3000`. Make sure `SERVER_PORT=3000` is configured in your `.env` file.

## HTTP API

### Health Check

`GET /health`

```http
http://localhost:3000/health
```

Checks whether the service is running.

### Room Information

`GET /api/v1/rooms/{roomID}`

Returns room information and the number of currently connected users. The count updates as users connect and disconnect.

| Room | Endpoint |
|---|---|
| Test room | `GET /api/v1/rooms/test-room` |
| Room A | `GET /api/v1/rooms/room-a` |
| Room B | `GET /api/v1/rooms/room-b` |

Example:

```http
GET http://localhost:3000/api/v1/rooms/test-room
```

## WebSocket API

Connect to a room using a WebSocket client.

| Room | WebSocket URL |
|---|---|
| Test room | `ws://localhost:3000/ws/rooms/test-room` |
| Room A | `ws://localhost:3000/ws/rooms/room-a` |
| Room B | `ws://localhost:3000/ws/rooms/room-b` |

### Send a Message

Send the following JSON after connecting:

```json
{
  "type": "chat.message",
  "message": "Hello everyone!"
}
```

The server broadcasts the message to clients connected to the same room. The outgoing message contains server-generated metadata, including message ID, room ID, user ID, and timestamp.

Messages are limited to 500 characters. Empty or invalid messages are rejected.

## Rate Limiting

Redis enforces a limit of 5 messages per 20-second window per connection. Excess messages are rejected to help protect the service from abuse.

## Database

PostgreSQL stores chat messages using GORM. Database connection settings are provided through environment variables.

## Testing

Run the test suite:

```bash
go test ./...
```

Check for data races:

```bash
go test -race ./...
```

## License

No license has been specified yet.
