<script>
  import { replaceEmoji } from './emoji.js';

  let { message } = $props();

  let displayText = $derived(() => replaceEmoji(message.message));

  let displayName = $derived(() => {
    if (message.username === 'system') return 'system';
    return message.username;
  });

  let time = $derived(() => {
    if (!message.timestamp) return '';
    const d = new Date(message.timestamp);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  });
</script>

<div class="message">
  <span class="user">{displayName()}</span>
  <span class="time">{time()}</span>
  <span class="text">{displayText()}</span>
</div>

<style>
  .message {
    padding: 0.3rem 0.75rem;
    display: flex;
    gap: 0.5rem;
    align-items: baseline;
    line-height: 1.4;
  }

  .message:hover {
    background: var(--bg-secondary);
  }

  .user {
    color: var(--accent);
    font-weight: 600;
    font-size: 0.85rem;
    flex-shrink: 0;
  }

  .time {
    color: var(--text-primary);
    opacity: 0.5;
    font-size: 0.75rem;
    flex-shrink: 0;
  }

  .text {
    color: var(--text-primary);
    font-size: 0.9rem;
    word-break: break-word;
  }
</style>
