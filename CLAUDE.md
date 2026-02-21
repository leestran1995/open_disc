# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the backend
go run ./internal/main/main.go

# Build backend binary
go build -o open_disc ./internal/main

# Format/vet Go code
go fmt ./...
go vet ./...

# Run tests (when added)
go test ./...
go test ./internal/...  # single package

# Run the frontend (from project root)
cd frontend && bun install && bun run dev

# Build frontend for production
cd frontend && bun run build  # outputs to frontend/dist/
```

## Environment Setup

**1. PostgreSQL** — any modern version (15+). Create the database and schema:

```sql
CREATE DATABASE open_disc;
\c open_disc
CREATE SCHEMA open_discord;
CREATE TABLE open_discord.users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), nickname VARCHAR(255) NOT NULL, username VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NOT NULL);
CREATE TABLE open_discord.rooms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name VARCHAR(255) NOT NULL);
CREATE TABLE open_discord.messages (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), room_id UUID NOT NULL REFERENCES open_discord.rooms(id), message TEXT NOT NULL, username VARCHAR(255) NOT NULL, timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE open_discord.user_room_pivot (user_id UUID NOT NULL REFERENCES open_discord.users(id), room_id UUID NOT NULL REFERENCES open_discord.rooms(id), PRIMARY KEY (user_id, room_id));
```

**2. `local.env`** — create in project root (gitignored):

```
DATABASE_URL=postgres://youruser@localhost:5432/open_disc
JWT_SECRET=your-secret-key-here
```

**3. Frontend deps** — requires [bun](https://bun.sh):

```bash
cd frontend && bun install
```

## Running in Development

Start both servers (two terminals):

```bash
# Terminal 1: Backend (ports 8080 + 8081)
go run ./internal/main/main.go

# Terminal 2: Frontend (port 4000)
cd frontend && bun run dev
```

Open http://localhost:4000. The Vite dev server (hardcoded to port 4000 via `strictPort`) proxies `/api/*` to the Go REST server (8080) and `/sse/*` to the SSE server (8081), so no CORS config is needed during dev.

## Architecture

**Two HTTP servers on startup:**
- Port **8080** — Gin router for REST API
- Port **8081** — Standard Go HTTP server for SSE (Server-Sent Events)

**Backend layer structure:**

| Layer | Path | Responsibility |
|-------|------|----------------|
| Domain models | `internal/domain/` | `User`, `Message`, `Room`, interfaces |
| HTTP handlers | `internal/http/` | Request parsing, response writing, auth middleware |
| Auth | `internal/auth/` | JWT token generation/validation, password hashing (argon2id), signup logic |
| Business logic | `internal/logic/` | In-memory room/connection state |
| Data access | `internal/postgresql/` | DB queries via pgx |
| Services wiring | `internal/util/` | Dependency injection, service initialization |
| Entry point | `internal/main/` | Server startup |

**Frontend structure** (`frontend/src/`):

| File | Responsibility |
|------|----------------|
| `main.js` | Svelte 5 `mount()` entry point |
| `app.css` | Solarized light/dark CSS custom properties |
| `App.svelte` | Root layout, login gate, JWT-based session restore from localStorage |
| `lib/stores.js` | Shared writable stores: `currentUser`, `rooms`, `activeRoomId`, `messagesByRoom`, `authToken` |
| `lib/api.js` | Auth-aware REST client with JWT Bearer token. Signup, signin, room CRUD, message send/fetch via `/api/*` |
| `lib/sse.js` | Fetch-based SSE client with JWT auth — connects via `/sse/connect/:username`, unwraps `RoomEvent` envelope, handles `new_message`, `user_joined`, `user_left`. Discovers rooms via `user_joined` events. |
| `lib/theme.js` | Dark/light toggle with `localStorage` persistence |
| `lib/Login.svelte` | Username/password auth form with signup + signin modes |
| `lib/Sidebar.svelte` | Room list, create room, logout clears JWT token |
| `lib/RoomHeader.svelte` | Displays `# room-name` for active room |
| `lib/MessageList.svelte` | Renders messages from store, fetches messages via REST on room select, auto-scrolls |
| `lib/MessageInput.svelte` | Text input, sends on Enter (no user_id in payload) |
| `lib/Message.svelte` | Single message row: username, timestamp, text |
| `lib/ThemeToggle.svelte` | Dark/light mode button |

**Svelte 5 patterns used:** `$state`, `$derived`, `$effect`, `$props()` for component-local state. `writable` stores from `svelte/store` for cross-component shared state. `onMount` for initial data fetching.

**Real-time messaging flow:**

1. Client connects to `GET /connect/:username` with `Authorization: Bearer <token>` header (SSE endpoint on port 8081)
2. Backend validates JWT, extracts username, connects client to all rooms, sends `user_joined` events for each room
3. Frontend uses fetch-based SSE (not EventSource) to support Authorization header
4. Client discovers rooms via `user_joined` events, fetches messages per room via `GET /messages/:room_id`
5. When `POST /messages` is called, the message is saved to DB (username from JWT) and fanned out to all connected clients in that room as a `RoomEvent` with type `new_message`
6. All SSE events use the `RoomEvent` envelope: `{ "room_event_type": "...", "payload": <JSON> }`
7. Frontend `sse.js` unwraps the envelope, dispatches by event type, deduplicates by message `id`, and updates the `messagesByRoom` store
8. Disconnection removes the client from all room maps; client reconnects with exponential backoff

**In-memory state** (`logic/room_connections.go`): `rooms map[UUID]*Room` where each `Room` holds a map of connected `RoomClient`s, each with a `SendChannel chan RoomEvent` (buffered, capacity 50). New rooms are registered in this map at startup (from DB) and on create (via `RoomHandler`).

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/ping` | No | Health check |
| POST | `/signup` | No | Create account `{username, password}` |
| POST | `/signin` | No | Sign in `{username, password}`, returns JWT |
| POST | `/rooms` | Yes | Create room `{name}` |
| GET | `/rooms/:id` | Yes | Get room |
| POST | `/rooms/:id/join` | Yes | Join room `{user_id}` |
| POST | `/messages` | Yes | Post message `{room_id, message}` (username from JWT) |
| GET | `/messages/:room_id` | Yes | Get messages, optional `?timestamp=` for pagination |
| GET | `/connect/:username` | Yes | SSE stream (port 8081) — sends `user_joined`, `user_left`, `new_message` events |

## Remaining Work (beans tickets)

See `.beans/` for full ticket details.

| Ticket | Description | Notes |
|--------|-------------|-------|
| `open_disc-pokb` | CORS middleware | Not needed during dev (Vite proxy handles it). Needed for production when frontend is served separately. |
| `open_disc-lv66` | Port unification | nginx/Caddy reverse proxy to serve REST + SSE on one origin. Low priority. |
