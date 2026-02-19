<script>
  import { rooms, activeRoomId } from './stores.js';

  let roomName = $derived(() => {
    const id = $activeRoomId;
    if (!id) return null;
    const room = $rooms.find((r) => r.id === id);
    return room ? room.name : null;
  });
</script>

<div class="room-header">
  {#if roomName()}
    <h2># {roomName()}</h2>
  {:else}
    <h2 class="empty">Select a room</h2>
  {/if}
</div>

<style>
  .room-header {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg-primary);
  }

  h2 {
    color: var(--text-heading);
    font-size: 1.1rem;
    font-weight: 600;
  }

  .empty {
    opacity: 0.5;
  }
</style>
