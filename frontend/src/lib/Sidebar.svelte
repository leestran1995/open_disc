<script lang="ts">
  import { createRoom, updateRoomOrder, starRoom, unstarRoom } from './api';
  import { authToken, rooms, activeRoomId, currentUser, messagesByRoom } from './stores';
  import { get } from 'svelte/store';
  import ThemeToggle from './ThemeToggle.svelte';
  import { connectSSE, disconnectSSE } from './sse';
  import type { Room } from './types';

  let newRoomName = $state('');
  let creating = $state(false);

  let draggedRoomId: string | null = $state(null);
  let dropTargetIndex: number | null = $state(null);

  function handleDragStart(e: DragEvent, room: Room): void {
    draggedRoomId = room.id;
    e.dataTransfer!.effectAllowed = 'move';
  }

  function handleDragOver(e: DragEvent, index: number): void {
    e.preventDefault();
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const midY = rect.top + rect.height / 2;
    dropTargetIndex = e.clientY < midY ? index : index + 1;
  }

  function handleDragLeave(e: DragEvent): void {
    if (!(e.currentTarget as HTMLElement).contains(e.relatedTarget as Node)) {
      dropTargetIndex = null;
    }
  }

  function handleDrop(e: DragEvent): void {
    e.preventDefault();
    if (draggedRoomId == null || dropTargetIndex == null) return;

    const currentRooms = get(rooms);
    const dragIndex = currentRooms.findIndex((r) => r.id === draggedRoomId);
    if (dragIndex === -1) return;

    // Dropping at same position or adjacent (no-op)
    if (dropTargetIndex === dragIndex || dropTargetIndex === dragIndex + 1) {
      draggedRoomId = null;
      dropTargetIndex = null;
      return;
    }

    const snapshot = [...currentRooms];
    const reordered = [...currentRooms];
    const [moved] = reordered.splice(dragIndex, 1);
    const insertAt = dropTargetIndex > dragIndex ? dropTargetIndex - 1 : dropTargetIndex;
    reordered.splice(insertAt, 0, moved);

    // Optimistic update
    rooms.set(reordered);
    localStorage.setItem('rooms', JSON.stringify(reordered));

    draggedRoomId = null;
    dropTargetIndex = null;

    updateRoomOrder(reordered.map((r) => r.id)).then((result) => {
      if (result && '_error' in result) {
        rooms.set(snapshot);
        localStorage.setItem('rooms', JSON.stringify(snapshot));
      }
    });
  }

  function handleDragEnd(): void {
    draggedRoomId = null;
    dropTargetIndex = null;
  }

  async function handleCreateRoom(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    const user = $currentUser;
    if (!newRoomName.trim() || !user || creating) return;

    creating = true;
    const room = await createRoom(newRoomName.trim());
    if (room && !('_error' in room)) {
      rooms.update((list) => {
        const updated = [...list, room];
        localStorage.setItem('rooms', JSON.stringify(updated));
        return updated;
      });
      activeRoomId.set(room.id);
      connectSSE(get(authToken)!, user.username);
      updateRoomOrder(get(rooms).map((r) => r.id));
      newRoomName = '';
    }
    creating = false;
  }

  function selectRoom(roomId: string): void {
    activeRoomId.set(roomId);
  }

  async function toggleStar(e: MouseEvent, room: Room): Promise<void> {
    e.stopPropagation();
    const wasStarred = room.starred;

    // Optimistic update
    rooms.update((list) =>
      list.map((r) => r.id === room.id ? { ...r, starred: !wasStarred } : r)
    );

    const result = wasStarred ? await unstarRoom(room.id) : await starRoom(room.id);
    if (result && '_error' in result) {
      // Rollback
      rooms.update((list) =>
        list.map((r) => r.id === room.id ? { ...r, starred: wasStarred } : r)
      );
    }
  }

  function logout(): void {
    disconnectSSE();
    localStorage.removeItem('token');
    localStorage.removeItem('rooms');
    localStorage.removeItem('activeRoomId');
    authToken.set(null);
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

  <div class="room-list" ondragleave={handleDragLeave}>
    {#each $rooms as room, index (room.id)}
      {#if dropTargetIndex === index && draggedRoomId !== room.id}
        <div class="drop-indicator"></div>
      {/if}
      <button
        class="room-item"
        class:active={$activeRoomId === room.id}
        class:dragging={draggedRoomId === room.id}
        draggable="true"
        ondragstart={(e) => handleDragStart(e, room)}
        ondragover={(e) => handleDragOver(e, index)}
        ondrop={handleDrop}
        ondragend={handleDragEnd}
        onclick={() => selectRoom(room.id)}
      >
        <span class="room-name"># {room.name}</span>
        <span
          class="star-btn"
          class:starred={room.starred}
          onclick={(e: MouseEvent) => toggleStar(e, room)}
          onkeydown={(e: KeyboardEvent) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleStar(e as unknown as MouseEvent, room); } }}
          role="button"
          tabindex="-1"
          title={room.starred ? 'Unstar room' : 'Star room'}
        >
          {room.starred ? '\u2605' : '\u2606'}
        </span>
      </button>
    {/each}
    {#if dropTargetIndex === $rooms.length}
      <div class="drop-indicator"></div>
    {/if}
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
    <span class="username">{$currentUser?.username}</span>
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
    display: flex;
    align-items: center;
    width: 100%;
    text-align: left;
    padding: 0.4rem 1rem;
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: 0.9rem;
    border-radius: 0;
    overflow: hidden;
    white-space: nowrap;
    cursor: grab;
  }

  .room-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .star-btn {
    flex-shrink: 0;
    background: none;
    border: none;
    color: var(--text-primary);
    opacity: 0.3;
    font-size: 0.85rem;
    padding: 0 0.2em;
    cursor: pointer;
    line-height: 1;
  }

  .room-item:hover .star-btn {
    opacity: 0.6;
  }

  .star-btn.starred {
    opacity: 1;
    color: var(--yellow);
  }

  .room-item.dragging {
    opacity: 0.3;
  }

  .drop-indicator {
    height: 2px;
    background: var(--accent);
    margin: 0 0.75rem;
    border-radius: 1px;
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
    min-width: 0;
    overflow: hidden;
  }

  .create-room input {
    flex: 1;
    min-width: 0;
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
    flex-shrink: 0;
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
