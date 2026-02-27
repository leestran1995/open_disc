import { get } from 'svelte/store';
import { authToken } from './stores';
import type { ApiResult, SigninResponse, SignupResponse, MessagesResponse, MessageCreateResponse, Room, ServerEventsResponse } from './types';

const BASE = import.meta.env.VITE_API_BASE || '/api';

/**
 * Generic HTTP client. Attaches JWT from the auth store and returns a
 * three-state result:
 *
 *   T        — success (parsed JSON body)
 *   ApiError — server returned 4xx/5xx (error message in `_error`)
 *   null     — network failure (fetch threw)
 *
 * The `_error` field is a client-side discriminant — the server sends
 * `{ "error": "..." }` which we rewrite to `{ _error: "..." }` so
 * callers can narrow with `'_error' in result`.
 */
async function request<T>(path: string, options: RequestInit = {}): Promise<ApiResult<T>> {
  try {
    const token = get(authToken);
    // RequestInit.headers is a wide union type (HeadersInit); we only
    // pass plain objects, so the cast is safe for our usage.
    const headers: Record<string, string> = { 'Content-Type': 'application/json', ...options.headers as Record<string, string> };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const res = await fetch(`${BASE}${path}`, {
      headers,
      ...options,
    });
    if (!res.ok) {
      try {
        const body = await res.json();
        return { _error: body.error || `Request failed (${res.status})` };
      } catch {
        return { _error: `Request failed (${res.status})` };
      }
    }
    return await res.json() as T;
  } catch {
    return null;
  }
}

export function signup(username: string, password: string): Promise<ApiResult<SignupResponse>> {
  return request<SignupResponse>('/signup', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function signin(username: string, password: string): Promise<ApiResult<SigninResponse>> {
  return request<SigninResponse>('/signin', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

/** POST /rooms — Go returns the Room struct directly, not wrapped in gin.H. */
export function createRoom(name: string): Promise<ApiResult<Room>> {
  return request<Room>('/rooms', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function getRoom(id: string): Promise<ApiResult<Room>> {
  return request<Room>(`/rooms/${id}`);
}

/** GET /messages/:room_id — Go wraps in gin.H{"messages": [...]}.  */
export function getMessages(roomId: string, timestamp?: string): Promise<ApiResult<MessagesResponse>> {
  const query = timestamp ? `?timestamp=${encodeURIComponent(timestamp)}` : '';
  return request<MessagesResponse>(`/messages/${roomId}${query}`);
}

export function sendMessage(roomId: string, message: string): Promise<ApiResult<MessageCreateResponse>> {
  return request<MessageCreateResponse>('/messages', {
    method: 'POST',
    body: JSON.stringify({ room_id: roomId, message }),
  });
}

/** GET /rooms — Go returns Room[] directly, not wrapped in gin.H. */
export function getRooms(): Promise<ApiResult<Room[]>> {
  return request<Room[]>('/rooms');
}

/**
 * PUT /rooms/order — Go returns c.JSON(200, nil), so the success body
 * is JSON null, which maps to the `null` branch of ApiResult.
 * Callers only check for `_error` to decide whether to rollback.
 */
export function updateRoomOrder(roomIds: string[]): Promise<ApiResult<null>> {
  return request<null>('/rooms/order', {
    method: 'PUT',
    body: JSON.stringify({ room_ids: roomIds }),
  });
}

/** PUT /rooms/:roomId/star — star a room. Backend returns empty 200. */
export function starRoom(roomId: string): Promise<ApiResult<null>> {
  return request<null>(`/rooms/${roomId}/star`, { method: 'PUT' });
}

/** DELETE /rooms/:roomId/star — unstar a room. Backend returns empty 200. */
export function unstarRoom(roomId: string): Promise<ApiResult<null>> {
  return request<null>(`/rooms/${roomId}/star`, { method: 'DELETE' });
}

/** GET /events — fetch server events by order range for gap-fill on reconnect. */
export function getServerEvents(orderStart: number, orderEnd?: number): Promise<ApiResult<ServerEventsResponse>> {
  let query = `?event_order_start=${orderStart}`;
  if (orderEnd !== undefined) query += `&event_order_end=${orderEnd}`;
  return request<ServerEventsResponse>(`/events${query}`);
}
