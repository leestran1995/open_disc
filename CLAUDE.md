# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the backend
go run ./main/main.go

# Build backend binary
go build -o open_disc ./main

# Format/vet Go code
go fmt ./...
go vet ./...

# Run tests (when added)
go test ./...
go test ./postgresql/...  # single package

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
CREATE TABLE open_discord.users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), nickname VARCHAR(255) NOT NULL);
CREATE TABLE open_discord.rooms (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name VARCHAR(255) NOT NULL);
CREATE TABLE open_discord.messages (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), room_id UUID NOT NULL REFERENCES open_discord.rooms(id), message TEXT NOT NULL, user_id UUID NOT NULL REFERENCES open_discord.users(id), timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE open_discord.user_room_pivot (user_id UUID NOT NULL REFERENCES open_discord.users(id), room_id UUID NOT NULL REFERENCES open_discord.rooms(id), PRIMARY KEY (user_id, room_id));
```

**2. `local.env`** — create in project root (gitignored):

```
DATABASE_URL=postgres://youruser@localhost:5432/open_disc
```

**3. Frontend deps** — requires [bun](https://bun.sh):

```bash
cd frontend && bun install
```

## Running in Development

Start both servers (two terminals):

```bash
# Terminal 1: Backend (ports 8080 + 8081)
go run ./main/main.go

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
| Domain models | `*.go` (root) | `User`, `Message`, `Room`, interfaces |
| HTTP handlers | `http/` | Request parsing, response writing |
| Business logic | `logic/` | In-memory room/connection state |
| Data access | `postgresql/` | DB queries via pgx |

**Frontend structure** (`frontend/src/`):

| File | Responsibility |
|------|----------------|
| `main.js` | Svelte 5 `mount()` entry point |
| `app.css` | Solarized light/dark CSS custom properties |
| `App.svelte` | Root layout, login gate, session restore from localStorage |
| `lib/stores.js` | Shared writable stores: `currentUser`, `rooms`, `activeRoomId`, `messagesByRoom` |
| `lib/api.js` | REST client — all fetch calls go through `/api/*` (Vite proxies to 8080). Uses `room_id` in message payloads. |
| `lib/sse.js` | `EventSource` manager — connects via `/sse/connect/:userId`, unwraps `RoomEvent` envelope, handles `new_message`, `historical_messages`, `user_joined`, `user_left` |
| `lib/theme.js` | Dark/light toggle with `localStorage` persistence |
| `lib/Login.svelte` | Nickname form, creates user via API, connects SSE |
| `lib/Sidebar.svelte` | Room list, create room + join, logout, reconnects SSE after joining rooms |
| `lib/RoomHeader.svelte` | Displays `# room-name` for active room |
| `lib/MessageList.svelte` | Renders messages from store, auto-scrolls to bottom |
| `lib/MessageInput.svelte` | Text input, sends on Enter, retains focus after send |
| `lib/Message.svelte` | Single message row: nickname/ID, timestamp, text |
| `lib/ThemeToggle.svelte` | Dark/light mode button |

**Svelte 5 patterns used:** `$state`, `$derived`, `$effect`, `$props()` for component-local state. `writable` stores from `svelte/store` for cross-component shared state. `onMount` for initial data fetching.

**Real-time messaging flow:**

1. Client connects to `GET /connect/:userId` (SSE endpoint on port 8081)
2. `http/sse_handler.go` looks up the user's rooms, sends `historical_messages` events (last 10 per room), then registers the client's send channel with each room
3. When `POST /messages` is called, the message is saved to DB and fanned out to all connected clients in that room via their Go channels as a `RoomEvent` with type `new_message`
4. All SSE events use the `RoomEvent` envelope: `{ "room_event_type": "...", "payload": <JSON> }`
5. Frontend `sse.js` unwraps the envelope, dispatches by event type, deduplicates by message `id`, and updates the `messagesByRoom` store
6. Disconnection (context cancel) removes the client from all room maps
7. `user_joined` / `user_left` events are broadcast to room members on connect/disconnect

**In-memory state** (`logic/room_connections.go`): `rooms map[UUID]*Room` where each `Room` holds a map of connected `RoomClient`s, each with a `SendChannel chan RoomEvent` (buffered, capacity 50). New rooms are registered in this map at startup (from DB) and on create (via `RoomHandler`).

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ping` | Health check |
| POST | `/users` | Create user `{nickname}` |
| GET | `/users/:id` | Get user |
| POST | `/rooms` | Create room `{name}` |
| GET | `/rooms/:id` | Get room |
| POST | `/rooms/:id/join` | Join room `{user_id}` |
| GET | `/users/:id/rooms` | Get rooms for user |
| POST | `/messages` | Post message `{room_id, message, user_id}` |
| GET | `/connect/:userId` | SSE stream — sends `historical_messages` on connect, then `new_message`, `user_joined`, `user_left` events |

## Remaining Work (beans tickets)

See `.beans/` for full ticket details.

| Ticket | Description | Notes |
|--------|-------------|-------|
| `open_disc-pokb` | CORS middleware | Not needed during dev (Vite proxy handles it). Needed for production when frontend is served separately. |
| `open_disc-lv66` | Port unification | nginx/Caddy reverse proxy to serve REST + SSE on one origin. Low priority. |
