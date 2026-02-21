<script>
  import { sendMessage } from './api.js';
  import { currentUser, activeRoomId } from './stores.js';

  let text = $state('');
  let inputEl;

  async function handleSend() {
    const roomId = $activeRoomId;
    if (!text.trim() || !roomId || !$currentUser) return;

    const msg = text.trim();
    text = '';
    await sendMessage(roomId, msg);
    requestAnimationFrame(() => inputEl?.focus());
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }
</script>

{#if $activeRoomId}
  <div class="message-input">
    <input
      type="text"
      placeholder="Send a message..."
      bind:this={inputEl}
      bind:value={text}
      onkeydown={handleKeydown}
    />
    <button onclick={handleSend} disabled={!text.trim()}>
      Send
    </button>
  </div>
{/if}

<style>
  .message-input {
    display: flex;
    gap: 0.5rem;
    padding: 0.75rem;
    border-top: 1px solid var(--border);
    background: var(--bg-primary);
  }

  input {
    flex: 1;
    padding: 0.5em 0.75em;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--bg-secondary);
    color: var(--text-primary);
    outline: none;
  }

  input:focus {
    border-color: var(--accent);
  }

  button {
    padding: 0.5em 1em;
    background: var(--accent);
    color: #fdf6e3;
    border: none;
    border-radius: 4px;
    font-weight: 600;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
