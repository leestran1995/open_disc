// ---------------------------------------------------------------------
// Domain types
//
// Field names use snake_case to match Go JSON tags exactly, so
// JSON.parse(response) produces objects that satisfy these interfaces
// with zero transformation.
//
// Go type mappings:
//   uuid.UUID  → string   (JSON serializes UUIDs as strings)
//   time.Time  → string   (JSON serializes as RFC 3339, e.g. "2024-01-15T09:30:00Z")
//   int        → number   (JSON number)
//   bool       → boolean
// ---------------------------------------------------------------------

/** Go: domain.Room (server.go) */
export interface Room {
  id: string;
  name: string;
  sort_order: number;
  starred: boolean;
}

/** Go: domain.Message (message.go) */
export interface Message {
  id: string;
  room_id: string;
  message: string;
  timestamp: string;
  username: string;
}

/**
 * Go: domain.User (user.go)
 *
 * Not currently consumed by the frontend — auth only extracts `username`
 * from the JWT, and there's no user-list endpoint yet. Defined here to
 * document the Go contract for when we add presence or profiles.
 */
export interface User {
  user_id: string;
  nickname: string;
  username: string;
  is_online: boolean;
}

/**
 * Go: auth.Claims (token.go) — embeds jwt.RegisteredClaims
 *
 * Decoded client-side without verification (backend validates on every
 * request). Fields beyond `username` + `exp` are optional because we
 * only need them for session restore, not security.
 */
export interface JWTClaims {
  id: string;
  username: string;
  exp: number;
  iat: number;
  nbf?: number;
  iss?: string;
  sub?: string;
}

// ---------------------------------------------------------------------
// API result handling
//
// Every api.ts function returns ApiResult<T>, which is a three-state
// discriminated union:
//
//   T         → success: the parsed response body
//   ApiError  → server returned an error (4xx/5xx)
//   null      → network failure (fetch threw)
//
// Callers narrow with:  if (result && !('_error' in result)) { ... }
//
// The discriminant field is `_error` (not `error`) to avoid collisions
// with response payloads — e.g. gin.H{"error": "..."} from the server
// is rewritten to { _error: "..." } by api.ts so it can't be confused
// with a success response that happens to have an `error` field.
// ---------------------------------------------------------------------

export interface ApiError {
  _error: string;
}

export type ApiResult<T> = T | ApiError | null;

/** Type guard alternative to `'_error' in result`. */
export function isApiError(result: unknown): result is ApiError {
  return result != null && typeof result === 'object' && '_error' in result;
}

// ---------------------------------------------------------------------
// SSE event types
//
// Mirrors Go ServerEventType constants (message.go). Defined for
// documentation; sse.ts uses plain `string` for the event type
// parameter since the wire data is unvalidated.
// ---------------------------------------------------------------------

export type ServerEventType =
  | 'new_message'
  | 'user_joined'
  | 'user_left'
  | 'room_created'
  | 'room_deleted';

/**
 * Go: ServerEvent (message.go)
 *
 * Wraps all persisted server events. Events sent via SSE with
 * server_event_order > 0 arrive in this envelope; legacy events
 * (user_joined/user_left) arrive as bare payloads with order 0.
 */
export interface ServerEvent {
  server_event_type: ServerEventType;
  server_event_id: string;
  server_event_order: number;
  server_event_time: string;
  payload: unknown;
}

// ---------------------------------------------------------------------
// Frontend-only types
// ---------------------------------------------------------------------

export interface EmojiSuggestion {
  name: string;
  emoji: string;
}

/** Store shape: room ID → ordered message array. */
export type MessagesByRoom = Record<string, Message[]>;

/** GET /events → gin.H{"server_events": [...]} */
export interface ServerEventsResponse {
  server_events: ServerEvent[];
}

// ---------------------------------------------------------------------
// API response envelopes
//
// These match the exact gin.H{} shapes returned by each Go handler.
// Endpoints that return a struct directly (getRooms → Room[],
// createRoom → Room) don't need a wrapper type.
// ---------------------------------------------------------------------

/** POST /signin → gin.H{"data": mintedToken} */
export interface SigninResponse {
  data: string;
}

/** POST /signup → gin.H{"data": "ok"} */
export interface SignupResponse {
  data: string;
}

/** GET /messages/:room_id → gin.H{"messages": result} */
export interface MessagesResponse {
  messages: Message[];
}

/** POST /messages → gin.H{"message": r} */
export interface MessageCreateResponse {
  message: Message;
}
