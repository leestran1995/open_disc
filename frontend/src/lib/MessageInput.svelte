<script>
  import { sendMessage } from './api.js';
  import { currentUser, activeRoomId } from './stores.js';
  import { replaceEmoji, searchEmoji } from './emoji.js';

  let text = $state('');
  let inputEl;

  $effect(() => {
    if ($activeRoomId) {
      requestAnimationFrame(() => inputEl?.focus());
    }
  });

  let suggestions = $state([]);
  let selectedIndex = $state(0);
  let colonStart = $state(-1);

  function getPartialShortcode() {
    const el = inputEl;
    if (!el) return null;
    const pos = el.selectionStart;
    const before = text.slice(0, pos);
    const match = before.match(/:([a-zA-Z0-9_+-]*)$/);
    if (!match) return null;
    return { query: match[1], start: pos - match[0].length };
  }

  function updateSuggestions() {
    const partial = getPartialShortcode();
    if (partial && partial.query.length >= 2) {
      colonStart = partial.start;
      suggestions = searchEmoji(partial.query);
      selectedIndex = 0;
    } else {
      suggestions = [];
      colonStart = -1;
    }
  }

  function applySuggestion(name) {
    const el = inputEl;
    if (!el || colonStart < 0) return;
    const pos = el.selectionStart;
    const before = text.slice(0, colonStart);
    const after = text.slice(pos);
    const replaced = replaceEmoji(`:${name}:`);
    text = before + replaced + after;
    suggestions = [];
    colonStart = -1;
    const newPos = before.length + replaced.length;
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(newPos, newPos);
    });
  }

  function handleInput() {
    const el = inputEl;
    if (!el) return;
    const pos = el.selectionStart;
    const replaced = replaceEmoji(text);
    if (replaced !== text) {
      const diff = text.length - replaced.length;
      text = replaced;
      requestAnimationFrame(() => el.setSelectionRange(pos - diff, pos - diff));
      suggestions = [];
      colonStart = -1;
    } else {
      updateSuggestions();
    }
  }

  async function handleSend() {
    const roomId = $activeRoomId;
    if (!text.trim() || !roomId || !$currentUser) return;

    const msg = text.trim();
    text = '';
    suggestions = [];
    colonStart = -1;
    await sendMessage(roomId, msg);
    requestAnimationFrame(() => inputEl?.focus());
  }

  function handleKeydown(e) {
    if (suggestions.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        selectedIndex = (selectedIndex + 1) % suggestions.length;
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        selectedIndex = (selectedIndex - 1 + suggestions.length) % suggestions.length;
        return;
      }
      if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault();
        applySuggestion(suggestions[selectedIndex].name);
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        suggestions = [];
        colonStart = -1;
        return;
      }
    }

    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }
</script>

{#if $activeRoomId}
  <div class="message-input">
    {#if suggestions.length > 0}
      <ul class="emoji-suggestions">
        {#each suggestions as item, i}
          <li>
            <button
              class="emoji-option"
              class:selected={i === selectedIndex}
              onmousedown={(e) => { e.preventDefault(); applySuggestion(item.name); }}
              onmouseenter={() => selectedIndex = i}
            >
              <span class="emoji-char">{item.emoji}</span>
              <span class="emoji-name">:{item.name}:</span>
            </button>
          </li>
        {/each}
      </ul>
    {/if}
    <input
      type="text"
      placeholder="Send a message..."
      bind:this={inputEl}
      bind:value={text}
      oninput={handleInput}
      onkeydown={handleKeydown}
    />
    <button onclick={handleSend} disabled={!text.trim()}>
      Send
    </button>
  </div>
{/if}

<style>
  .message-input {
    position: relative;
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

  .message-input > button {
    padding: 0.5em 1em;
    background: var(--accent);
    color: #fdf6e3;
    border: none;
    border-radius: 4px;
    font-weight: 600;
  }

  .message-input > button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .emoji-suggestions {
    position: absolute;
    bottom: 100%;
    left: 0.75rem;
    right: 0.75rem;
    margin: 0;
    padding: 0.25rem;
    list-style: none;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 6px;
    box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.15);
    max-height: 240px;
    overflow-y: auto;
  }

  .emoji-option {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    padding: 0.35rem 0.5rem;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: var(--text-primary);
    font-size: 0.9rem;
    cursor: pointer;
    text-align: left;
  }

  .emoji-option:hover,
  .emoji-option.selected {
    background: var(--accent);
    color: #fdf6e3;
  }

  .emoji-char {
    font-size: 1.2rem;
    width: 1.5em;
    text-align: center;
    flex-shrink: 0;
  }

  .emoji-name {
    opacity: 0.85;
  }
</style>
