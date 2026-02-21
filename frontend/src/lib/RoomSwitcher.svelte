<script>
  import { rooms, activeRoomId } from './stores.js';

  let { open, onclose } = $props();
  let query = $state('');
  let selectedIndex = $state(0);
  let inputEl = $state(null);

  let filtered = $derived(
    $rooms.filter((r) =>
      r.name.toLowerCase().includes(query.toLowerCase())
    )
  );

  $effect(() => {
    if (open && inputEl) {
      query = '';
      selectedIndex = 0;
      inputEl.focus();
    }
  });

  function select(room) {
    activeRoomId.set(room.id);
    onclose();
  }

  function onkeydown(e) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, filtered.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (filtered[selectedIndex]) {
        select(filtered[selectedIndex]);
      }
    } else if (e.key === 'Escape') {
      e.preventDefault();
      onclose();
    }
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="switcher-backdrop" onclick={onclose} onkeydown={onkeydown}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="switcher-modal" onclick={(e) => e.stopPropagation()}>
      <input
        bind:this={inputEl}
        bind:value={query}
        placeholder="Where would you like to go?"
        oninput={() => (selectedIndex = 0)}
      />
      <div class="switcher-results">
        {#each filtered as room, i}
          <button
            class:selected={i === selectedIndex}
            class:active={room.id === $activeRoomId}
            onclick={() => select(room)}
          >
            <span class="hash">#</span> {room.name}
          </button>
        {/each}
        {#if filtered.length === 0}
          <div class="no-results">No rooms found</div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .switcher-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 20vh;
    z-index: 100;
  }

  .switcher-modal {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 8px;
    width: 100%;
    max-width: 480px;
    max-height: 400px;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.3);
  }

  input {
    padding: 12px 16px;
    border: none;
    border-bottom: 1px solid var(--border);
    background: transparent;
    color: var(--text-primary);
    font-size: 1rem;
    outline: none;
    border-radius: 8px 8px 0 0;
  }

  input::placeholder {
    color: var(--text-primary);
    opacity: 0.5;
  }

  .switcher-results {
    overflow-y: auto;
    padding: 4px;
  }

  button {
    display: block;
    width: 100%;
    padding: 8px 12px;
    border: none;
    background: transparent;
    color: var(--text-primary);
    text-align: left;
    border-radius: 4px;
    font-size: 0.95rem;
  }

  button:hover,
  button.selected {
    background: var(--accent);
    color: var(--bg-primary);
  }

  button.active .hash {
    color: var(--accent);
  }

  button.selected .hash,
  button:hover .hash {
    color: inherit;
  }

  .hash {
    opacity: 0.6;
    margin-right: 4px;
  }

  .no-results {
    padding: 12px 16px;
    opacity: 0.5;
    text-align: center;
  }
</style>
