const BASE = '/api';

async function request(path, options = {}) {
  try {
    const res = await fetch(`${BASE}${path}`, {
      headers: { 'Content-Type': 'application/json', ...options.headers },
      ...options,
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export function createUser(nickname) {
  return request('/users', {
    method: 'POST',
    body: JSON.stringify({ nickname }),
  });
}

export function getUser(id) {
  return request(`/users/${id}`);
}

export function createRoom(name) {
  return request('/rooms', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export function getRoom(id) {
  return request(`/rooms/${id}`);
}

export function getUserRooms(userId) {
  return request(`/users/${userId}/rooms`);
}

export function joinRoom(roomId, userId) {
  return request(`/rooms/${roomId}/join`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId }),
  });
}

export function sendMessage(roomId, message, userId) {
  return request('/messages', {
    method: 'POST',
    body: JSON.stringify({ room_id: roomId, message, user_id: userId }),
  });
}