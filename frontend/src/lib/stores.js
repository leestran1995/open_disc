import { writable } from 'svelte/store';

export const currentUser = writable(null);
export const rooms = writable([]);
export const activeRoomId = writable(null);
export const messagesByRoom = writable({});
