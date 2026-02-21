import { get } from 'svelte/store';
import { messagesByRoom, rooms } from './stores.js';
import { getRoom } from './api.js';

let abortController = null;
const seenIds = new Set();
let reconnectTimeout = null;
let reconnectDelay = 1000;

function handleNewMessage(msg) {
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

async function handleUserJoined(payload) {
  if (!payload || !payload.room_id) return;

  const currentRooms = get(rooms);
  if (currentRooms.some((r) => r.id === payload.room_id)) return;

  const room = await getRoom(payload.room_id);
  if (!room) return;

  rooms.update((current) => {
    if (current.some((r) => r.id === room.id)) return current;
    const updated = [...current, room];
    localStorage.setItem('rooms', JSON.stringify(updated));
    return updated;
  });
}

function handleEvent(roomEvent) {
  const { room_event_type, payload } = roomEvent;
  const parsed = typeof payload === 'string' ? JSON.parse(payload) : payload;

  switch (room_event_type) {
    case 'new_message':
      handleNewMessage(parsed);
      break;
    case 'user_joined':
      handleUserJoined(parsed);
      break;
    case 'user_left':
      break;
  }
}

function processChunk(buffer) {
  const events = buffer.split('\n\n');
  const remainder = events.pop();

  for (const event of events) {
    const lines = event.split('\n');
    let data = '';
    for (const line of lines) {
      if (line.startsWith('data: ')) {
        data += line.slice(6);
      }
    }
    if (!data) continue;
    try {
      const roomEvent = JSON.parse(data);
      handleEvent(roomEvent);
    } catch {
      // skip malformed events
    }
  }

  return remainder;
}

async function readStream(reader) {
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      buffer = processChunk(buffer);
    }
  } catch (err) {
    if (err.name === 'AbortError') return;
    throw err;
  }
}

let _currentToken = null;
let _currentUsername = null;

export function connectSSE(token, username) {
  disconnectSSE();

  _currentToken = token;
  _currentUsername = username;
  reconnectDelay = 1000;

  startConnection(token, username);
}

function startConnection(token, username) {
  abortController = new AbortController();

  fetch(`/sse/connect/${username}`, {
    headers: { Authorization: `Bearer ${token}` },
    signal: abortController.signal,
  })
    .then((response) => {
      if (!response.ok) throw new Error(`SSE connect failed: ${response.status}`);
      reconnectDelay = 1000;
      return readStream(response.body.getReader());
    })
    .then(() => {
      scheduleReconnect();
    })
    .catch((err) => {
      if (err.name === 'AbortError') return;
      scheduleReconnect();
    });
}

function scheduleReconnect() {
  if (!_currentToken || !_currentUsername) return;
  reconnectTimeout = setTimeout(() => {
    startConnection(_currentToken, _currentUsername);
    reconnectDelay = Math.min(reconnectDelay * 2, 30000);
  }, reconnectDelay);
}

export function disconnectSSE() {
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
