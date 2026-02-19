import { messagesByRoom } from './stores.js';

let eventSource = null;
const seenIds = new Set();

function handleMessage(event) {
  let msg;
  try {
    msg = JSON.parse(event.data);
  } catch {
    // Current backend sends "Event <text>" — show as raw fallback
    msg = {
      id: null,
      server_id: null,
      message: event.data,
      user_id: 'system',
      timestamp: new Date().toISOString(),
    };
  }

  // Deduplicate by id when available
  if (msg.id) {
    if (seenIds.has(msg.id)) return;
    seenIds.add(msg.id);
  }

  if (!msg.server_id) return;

  messagesByRoom.update((current) => {
    const roomMessages = current[msg.server_id] || [];
    return { ...current, [msg.server_id]: [...roomMessages, msg] };
  });
}

export function connectSSE(userId) {
  disconnectSSE();
  eventSource = new EventSource(`/sse/connect/${userId}`);
  eventSource.onmessage = handleMessage;
}

export function disconnectSSE() {
  if (eventSource) {
    eventSource.close();
    eventSource = null;
  }
  seenIds.clear();
}
