import { messagesByRoom } from './stores.js';

let eventSource = null;
const seenIds = new Set();

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

function handleHistoricalMessages(messages) {
  if (!Array.isArray(messages)) return;

  messagesByRoom.update((current) => {
    const updated = { ...current };
    for (const msg of messages) {
      if (!msg.room_id) continue;
      if (msg.id) {
        if (seenIds.has(msg.id)) continue;
        seenIds.add(msg.id);
      }
      const roomMessages = updated[msg.room_id] || [];
      updated[msg.room_id] = [...roomMessages, msg];
    }
    return updated;
  });
}

function handleEvent(event) {
  let roomEvent;
  try {
    roomEvent = JSON.parse(event.data);
  } catch {
    return;
  }

  const { room_event_type, payload } = roomEvent;

  switch (room_event_type) {
    case 'new_message':
      handleNewMessage(typeof payload === 'string' ? JSON.parse(payload) : payload);
      break;
    case 'historical_messages':
      handleHistoricalMessages(typeof payload === 'string' ? JSON.parse(payload) : payload);
      break;
    case 'user_joined':
    case 'user_left':
      // Could surface these as UI notifications later
      break;
  }
}

export function connectSSE(userId) {
  disconnectSSE();
  eventSource = new EventSource(`/sse/connect/${userId}`);
  eventSource.onmessage = handleEvent;
}

export function disconnectSSE() {
  if (eventSource) {
    eventSource.close();
    eventSource = null;
  }
  seenIds.clear();
}
