import { writable, type Writable } from 'svelte/store';
import type { Room, MessagesByRoom } from './types';

export const authToken: Writable<string | null> = writable(localStorage.getItem('token'));

// Only holds { username } because auth extracts just the username from
// the JWT — the backend has no "get current user" endpoint that returns
// a full User object. This is intentionally not the User type from types.ts.
export const currentUser: Writable<{ username: string } | null> = writable(null);

export const rooms: Writable<Room[]> = writable([]);
export const activeRoomId: Writable<string | null> = writable(null);
export const messagesByRoom: Writable<MessagesByRoom> = writable({});
