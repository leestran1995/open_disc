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

# Type-check frontend (svelte-check)
cd frontend && bun run check
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

Open http://localhost:4000. The Vite dev server (hardcoded to port 4000 via `strictPort`) proxies `/api/*` and `/sse/*` to the Go server (8080), so no CORS config is needed during dev.

## Architecture

**Single HTTP server on startup:**
- Port **8080** — Gin router for REST API + SSE (Server-Sent Events)

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
| `main.ts` | Svelte 5 `mount()` entry point |
| `app.css` | Solarized light/dark CSS custom properties |
| `App.svelte` | Root layout, login gate, JWT-based session restore from localStorage |
| `lib/types.ts` | Shared domain types (`Room`, `Message`, `User`, `JWTClaims`), API response/request types, `ApiResult<T>` discriminated union |
| `lib/stores.ts` | Typed writable stores: `currentUser`, `rooms`, `activeRoomId`, `messagesByRoom`, `authToken` |
| `lib/api.ts` | Auth-aware REST client with JWT Bearer token. Generic `request<T>()` returning `ApiResult<T>`. Signup, signin, room CRUD, message send/fetch via `/api/*` |
| `lib/sse.ts` | Fetch-based SSE client with JWT auth — connects via `/sse/connect` (username from JWT), parses Gin `c.SSEvent()` format (`event:` + `data:` lines), handles `new_message`, `user_joined`, `user_left`, `room_created`. |
| `lib/jwt.ts` | Shared JWT decode helper (base64 payload extraction, no verification), returns `JWTClaims \| null` |
| `lib/emoji.ts` | `:shortcode:` to unicode emoji replacement (`replaceEmoji`) + prefix search for autocomplete (`searchEmoji`). Uses `gemoji` package. |
| `lib/theme.ts` | Dark/light toggle with `localStorage` persistence, auto-detects system preference |
| `lib/Login.svelte` | Username/password auth form with signup + signin modes |
| `lib/Sidebar.svelte` | Room list, create room, logout clears JWT token |
| `lib/RoomHeader.svelte` | Displays `# room-name` for active room |
| `lib/MessageList.svelte` | Renders messages from store, fetches messages via REST on room select (reverses DESC results), auto-scrolls |
| `lib/MessageInput.svelte` | Text input with emoji autocomplete popup (`:` + 2 chars triggers suggestions, arrow keys/Tab/Enter to select). Inline shortcode replacement on type. Sends on Enter. |
| `lib/Message.svelte` | Single message row: username, timestamp, text. Renders `:shortcode:` emoji via `$derived`. |
| `lib/ThemeToggle.svelte` | Dark/light mode button |

**Svelte 5 patterns used:** `$state`, `$derived`, `$effect`, `$props()` for component-local state. `writable` stores from `svelte/store` for cross-component shared state. `onMount` for initial data fetching. **TypeScript:** strict mode enabled, `$props()` typed via `interface Props`, `ApiResult<T>` narrowed with `'_error' in result`.

**Real-time messaging flow:**

1. Client connects to `GET /connect` with `Authorization: Bearer <token>` header (SSE on main Gin router, port 8080, username extracted from JWT)
2. Backend validates JWT, extracts username, registers client, sends server-scoped `user_joined` event (username string) to all connected clients
3. Frontend uses fetch-based SSE (not EventSource) to support Authorization header
4. Client fetches rooms via `GET /rooms` REST call on connect, fetches messages per room via `GET /messages/:room_id`
5. When `POST /messages` is called, the message is saved to DB (username from JWT) and fanned out to all connected clients as SSE `new_message` event
6. SSE uses Gin's native `c.SSEvent()` format: `event: <type>\ndata: <payload>\n\n`. Struct payloads (messages) are JSON-encoded; string payloads (usernames, room names) are plain text
7. Frontend `sse.ts` parses `event:` and `data:` lines, dispatches by event type, deduplicates messages by `id`, and updates stores
8. `user_joined` and `user_left` are server-scoped events (payload is a username string), not room-scoped. Currently no-ops in the frontend
9. `room_created` event triggers a REST refetch of the room list to update the sidebar
10. Disconnection removes the client from the registry; client reconnects with exponential backoff

**In-memory state** (`internal/logic/room_connections.go`): `rooms map[UUID]*Room` where each `Room` holds a map of connected `RoomClient`s, each with a `SendChannel chan RoomEvent` (buffered, capacity 50). New rooms are registered in this map at startup (from DB) and on create (via `RoomHandler`). Client registry is keyed by username (one SSE connection per user).

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
| GET | `/connect` | Yes | SSE stream (port 8080, username from JWT) — sends `user_joined`, `user_left`, `new_message`, `room_created` events |

## Remaining Work (beans tickets)

See `.beans/` for full ticket details.

| Ticket | Description | Notes |
|--------|-------------|-------|
| `open_disc-pokb` | CORS middleware | Not needed during dev (Vite proxy handles it). Needed for production when frontend is served separately. |
| `open_disc-lv66` | Port unification | nginx/Caddy reverse proxy to serve REST + SSE on one origin. Low priority. |
