# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the application
go run ./main/main.go

# Build binary
go build -o open_disc ./main

# Format code
go fmt ./...

# Vet code
go vet ./...

# Run tests (when added)
go test ./...
go test ./postgresql/...  # single package
```

## Environment Setup

Create a `local.env` file in the project root (gitignored):

```
DATABASE_URL=postgres://user:password@localhost:5432/database
```

The app uses the `open_discord` PostgreSQL schema. Required tables:

```sql
CREATE TABLE open_discord.users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), nickname VARCHAR(255) NOT NULL);
CREATE TABLE open_discord.rooms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name VARCHAR(255) NOT NULL);
CREATE TABLE open_discord.messages (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), server_id UUID NOT NULL REFERENCES open_discord.rooms(id), message TEXT NOT NULL, user_id UUID NOT NULL REFERENCES open_discord.users(id), timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE open_discord.user_room_pivot (user_id UUID NOT NULL REFERENCES open_discord.users(id), room_id UUID NOT NULL REFERENCES open_discord.rooms(id), PRIMARY KEY (user_id, room_id));
```

## Architecture

**Two HTTP servers on startup:**
- Port **8080** — Gin router for REST API
- Port **8081** — Standard Go HTTP server for SSE (Server-Sent Events)

**Layer structure:**

| Layer | Path | Responsibility |
|-------|------|----------------|
| Domain models | `*.go` (root) | `User`, `Message`, interfaces |
| HTTP handlers | `http/` | Request parsing, response writing |
| Business logic | `logic/` | In-memory room/connection state |
| Data access | `postgresql/` | DB queries via pgx |

**Real-time messaging flow:**

1. Client connects to `GET /connect/:userId` (SSE endpoint on port 8081)
2. `logic/room_connections.go` registers the client's send channel with all rooms the user has joined
3. When `POST /messages` is called, the message is saved to DB and fanned out to all connected clients in that room via their Go channels
4. Disconnection (context cancel) removes the client from all room maps

**In-memory state** (`logic/room_connections.go`): `rooms map[UUID]*Room` where each `Room` holds a slice of connected `Client`s, each with a `SendChannel chan Message`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ping` | Health check |
| POST | `/users` | Create user `{nickname}` |
| GET | `/users/:id` | Get user |
| POST | `/rooms` | Create room `{name}` |
| GET | `/rooms/:id` | Get room |
| POST | `/rooms/:id/join` | Join room `{user_id}` |
| POST | `/messages` | Post message `{server_id, message, user_id}` |
| GET | `/connect/:userId` | SSE stream for real-time messages |
