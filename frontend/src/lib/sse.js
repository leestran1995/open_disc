import { messagesByRoom, rooms } from './stores.js';
import { getRooms } from './api.js';

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

async function handleRoomCreated(_roomName) {
  const allRooms = await getRooms();
  if (Array.isArray(allRooms)) {
    rooms.set(allRooms);
    localStorage.setItem('rooms', JSON.stringify(allRooms));
  }
}

function handleEvent(eventType, rawData) {
  let parsed;
  try {
    parsed = JSON.parse(rawData);
  } catch {
    parsed = rawData;
  }

  switch (eventType) {
    case 'new_message':
      handleNewMessage(parsed);
      break;
    case 'user_joined':
      // Server-scoped event (username string), not room-scoped
      break;
    case 'user_left':
      // Server-scoped event (username string), not room-scoped
      break;
    case 'room_created':
      handleRoomCreated(parsed);
      break;
  }
}

function processChunk(buffer) {
  const events = buffer.split('\n\n');
  const remainder = events.pop();

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

  fetch(`/sse/connect`, {
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
