import { get } from 'svelte/store';
import { authToken } from './stores.js';

const BASE = '/api';

async function request(path, options = {}) {
  try {
    const token = get(authToken);
    const headers = { 'Content-Type': 'application/json', ...options.headers };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const res = await fetch(`${BASE}${path}`, {
      headers,
      ...options,
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export function signup(username, password) {
  return request('/signup', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function signin(username, password) {
  return request('/signin', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
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

export function getMessages(roomId, timestamp) {
  const query = timestamp ? `?timestamp=${encodeURIComponent(timestamp)}` : '';
  return request(`/messages/${roomId}${query}`);
}

export function sendMessage(roomId, message) {
  return request('/messages', {
    method: 'POST',
    body: JSON.stringify({ room_id: roomId, message }),
  });
}
