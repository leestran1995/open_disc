import { get } from 'svelte/store';
import { userIdUsernameMap } from './stores';
import { getAllUsers, getUserById } from './api';
import type { User } from './types';

/** Fetch all users and populate the user ID -> username map. */
export async function loadAllUsers(): Promise<void> {
  const result = await getAllUsers();
  if (Array.isArray(result)) {
    const map: Record<string, string> = {};
    for (const user of result as User[]) {
      map[user.user_id] = user.nickname || user.username;
    }
    userIdUsernameMap.set(map);
  }
}

/** Resolve a user_id to a display name. Returns the name or 'Unknown' if not in the map. */
export function resolveUsername(userId: string): string {
  const map = get(userIdUsernameMap);
  return map[userId] ?? 'Unknown';
}

/**
 * Ensure a user_id exists in the map. If not, fetches from the server
 * and updates the store. Returns the display name.
 */
export async function ensureUser(userId: string): Promise<string> {
  console.log("Ensuring user ID:", userId);
  const map = get(userIdUsernameMap);
  if (map[userId]) return map[userId];

  const result = await getUserById(userId);
  if (result && !('_error' in result)) {
    const user = result as User;
    const displayName = user.nickname || user.username;
    userIdUsernameMap.update((current) => ({ ...current, [userId]: displayName }));
    return displayName;
  }

  if (result && '_error' in result) {
    console.error(`Error fetching user ${userId}:`, result._error);
  }

  return 'Unknown';
}
