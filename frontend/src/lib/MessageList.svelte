<script>
  import Message from './Message.svelte';
  import { messagesByRoom, activeRoomId } from './stores.js';

  let container;

  let messages = $derived($messagesByRoom[$activeRoomId] || []);

  $effect(() => {
    // Re-run when messages change
    messages;
    if (container) {
      requestAnimationFrame(() => {
        container.scrollTop = container.scrollHeight;
      });
    }
  });
</script>

<div class="message-list" bind:this={container}>
  {#if !$activeRoomId}
    <div class="empty">Select a room to start chatting</div>
  {:else if messages.length === 0}
    <div class="empty">No messages yet. Say something!</div>
  {:else}
    {#each messages as msg (msg.id || messages.indexOf(msg))}
      <Message message={msg} />
    {/each}
  {/if}
</div>

<style>
  .message-list {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem 0;
  }

  .empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-primary);
    opacity: 0.5;
    font-size: 0.9rem;
  }
</style>
