import { messagesByRoom, rooms } from './stores';
import { getRooms } from './api';
import type { Message, Room } from './types';

// --- Module-level connection state ---

let abortController: AbortController | null = null;
const seenIds = new Set<string>();
// ReturnType<typeof setTimeout> avoids Node vs browser timer ID mismatch.
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
let reconnectDelay = 1000;

// --- SSE event handlers ---

function handleNewMessage(msg: Message): void {
  if (!msg || !msg.room_id) return;

  if (msg.id) {
    if (seenIds.has(msg.id)) return;
    seenIds.add(msg.id);
  }

  messagesByRoom.update((current) => {
    const roomMessages = current[msg.room_id] || [];
    return { ...current, [msg.room_id]: [...roomMessages, msg] };
  });
}

async function handleRoomCreated(_roomName: string): Promise<void> {
  const allRooms = await getRooms();
  if (Array.isArray(allRooms)) {
    rooms.set(allRooms as Room[]);
    localStorage.setItem('rooms', JSON.stringify(allRooms));
  }
}

/**
 * Dispatch a parsed SSE event. The `eventType` and `rawData` come from
 * parsing Gin's c.SSEvent() wire format, so they're unvalidated strings.
 * We use `as` casts after the switch because each event type has a known
 * payload shape from the Go backend, but TypeScript can't infer that
 * from JSON.parse().
 *
 * Note: the backend currently double-fans-out new_message and room_created
 * events (once as a bare payload, once wrapped in a ServerEvent envelope).
 * The envelope version naturally gets dropped — handleNewMessage bails on
 * missing top-level room_id, and handleRoomCreated is idempotent.
 */
function handleEvent(eventType: string, rawData: string): void {
  let parsed: unknown;
  try {
    parsed = JSON.parse(rawData);
  } catch {
    parsed = rawData;
  }

  switch (eventType) {
    case 'new_message':
      handleNewMessage(parsed as Message);
      break;
    case 'user_joined':
      break;
    case 'user_left':
      break;
    case 'room_created':
      handleRoomCreated(parsed as string);
      break;
  }
}

// --- SSE wire format parser ---
//
// Gin's c.SSEvent() produces the standard SSE format:
//
//   event: new_message
//   data: {"id":"...","room_id":"...","message":"hello",...}
//
// Events are separated by double newlines. The `data:` and `event:`
// prefixes may or may not have a space after the colon (per SSE spec),
// so we check charAt to determine the slice offset.

function processChunk(buffer: string): string {
  const events = buffer.split('\n\n');
  // split() always returns at least one element, so pop() is safe here.
  const remainder = events.pop()!;

  for (const event of events) {
    const lines = event.split('\n');
    let data = '';
    let eventType = '';
    for (const line of lines) {
      if (line.startsWith('data:')) {
        data += line.slice(line.charAt(5) === ' ' ? 6 : 5);
      } else if (line.startsWith('event:')) {
        eventType = line.slice(line.charAt(6) === ' ' ? 7 : 6);
      }
    }
    if (!data || !eventType) continue;
    handleEvent(eventType, data);
  }

  return remainder;
}

async function readStream(reader: ReadableStreamDefaultReader<Uint8Array>): Promise<void> {
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      buffer = processChunk(buffer);
    }
  } catch (err: unknown) {
    if (err instanceof Error && err.name === 'AbortError') return;
    throw err;
  }
}

// --- Public API: connect / disconnect ---
//
// Module-level _currentToken/_currentUsername track the active session
// so scheduleReconnect can re-establish the connection on drop.

let _currentToken: string | null = null;
let _currentUsername: string | null = null;

export function connectSSE(token: string, username: string): void {
  disconnectSSE();

  _currentToken = token;
  _currentUsername = username;
  reconnectDelay = 1000;

  startConnection(token, username);
}

function startConnection(token: string, username: string): void {
  abortController = new AbortController();

  const sseBase = import.meta.env.VITE_SSE_BASE || '/sse';
  fetch(`${sseBase}/connect`, {
    headers: { Authorization: `Bearer ${token}` },
    signal: abortController.signal,
  })
    .then((response) => {
      if (!response.ok) throw new Error(`SSE connect failed: ${response.status}`);
      reconnectDelay = 1000;
      // response.body is non-null for successful fetch responses.
      return readStream(response.body!.getReader());
    })
    .then(() => {
      scheduleReconnect();
    })
    .catch((err: unknown) => {
      if (err instanceof Error && err.name === 'AbortError') return;
      scheduleReconnect();
    });
}

function scheduleReconnect(): void {
  if (!_currentToken || !_currentUsername) return;
  reconnectTimeout = setTimeout(() => {
    // Non-null assertions are safe: the guard above ensures both are
    // set when we schedule, and disconnectSSE() clears the timeout
    // synchronously before nulling them.
    startConnection(_currentToken!, _currentUsername!);
    reconnectDelay = Math.min(reconnectDelay * 2, 30000);
  }, reconnectDelay);
}

export function disconnectSSE(): void {
  _currentToken = null;
  _currentUsername = null;
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }
  if (abortController) {
    abortController.abort();
    abortController = null;
  }
  seenIds.clear();
}
