<script>
  import { createRoom, joinRoom, getRoomMessages } from './api.js';
  import { rooms, activeRoomId, currentUser, messagesByRoom } from './stores.js';
  import ThemeToggle from './ThemeToggle.svelte';
  import { connectSSE, disconnectSSE } from './sse.js';

  let newRoomName = $state('');
  let creating = $state(false);

  async function handleCreateRoom(e) {
    e.preventDefault();
    const user = $currentUser;
    if (!newRoomName.trim() || !user || creating) return;

    creating = true;
    const room = await createRoom(newRoomName.trim());
    if (room) {
      await joinRoom(room.id, user.user_id);
      rooms.update((list) => [...list, room]);
      activeRoomId.set(room.id);
      // Reconnect SSE so the backend registers this new room subscription
      connectSSE(user.user_id);
      newRoomName = '';
    }
    creating = false;
  }

  async function selectRoom(roomId) {
    activeRoomId.set(roomId);
    // Load message history if we don't have it yet
    const current = $messagesByRoom;
    if (!current[roomId]) {
      const msgs = await getRoomMessages(roomId);
      if (msgs) {
        messagesByRoom.update((m) => ({ ...m, [roomId]: msgs }));
      }
    }
  }

  function logout() {
    disconnectSSE();
    localStorage.removeItem('user_id');
    currentUser.set(null);
    rooms.set([]);
    activeRoomId.set(null);
    messagesByRoom.set({});
  }
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <h1>Open Disc</h1>
    <ThemeToggle />
  </div>

  <div class="room-list">
    {#each $rooms as room (room.id)}
      <button
        class="room-item"
        class:active={$activeRoomId === room.id}
        onclick={() => selectRoom(room.id)}
      >
        # {room.name}
      </button>
    {/each}
  </div>

  <form class="create-room" onsubmit={handleCreateRoom}>
    <input
      type="text"
      placeholder="New room name"
      bind:value={newRoomName}
      disabled={creating}
    />
    <button type="submit" disabled={creating || !newRoomName.trim()}>+</button>
  </form>

  <div class="sidebar-footer">
    <span class="username">{$currentUser?.nickname}</span>
    <button class="logout" onclick={logout}>Log out</button>
  </div>
</aside>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    height: 100%;
    overflow: hidden;
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--border);
  }

  .sidebar-header h1 {
    color: var(--text-heading);
    font-size: 1.1rem;
    font-weight: 700;
  }

  .room-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem 0;
  }

  .room-item {
    display: block;
    width: 100%;
    text-align: left;
    padding: 0.4rem 1rem;
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: 0.9rem;
    border-radius: 0;
  }

  .room-item:hover {
    background: var(--bg-primary);
  }

  .room-item.active {
    background: var(--bg-primary);
    color: var(--accent);
    font-weight: 600;
  }

  .create-room {
    display: flex;
    gap: 0.4rem;
    padding: 0.5rem 0.75rem;
    border-top: 1px solid var(--border);
  }

  .create-room input {
    flex: 1;
    padding: 0.4em 0.6em;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 0.85rem;
    outline: none;
  }

  .create-room input:focus {
    border-color: var(--accent);
  }

  .create-room button {
    padding: 0.4em 0.7em;
    background: var(--accent);
    color: #fdf6e3;
    border: none;
    border-radius: 4px;
    font-weight: 700;
    font-size: 0.9rem;
  }

  .create-room button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .sidebar-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5rem 1rem;
    border-top: 1px solid var(--border);
    font-size: 0.85rem;
  }

  .username {
    color: var(--text-heading);
    font-weight: 600;
  }

  .logout {
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: 0.8rem;
    opacity: 0.7;
    padding: 0.2em 0.4em;
  }

  .logout:hover {
    opacity: 1;
    color: var(--red);
  }
</style>
