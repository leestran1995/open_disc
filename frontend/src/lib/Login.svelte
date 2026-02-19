<script>
  import { createUser } from './api.js';
  import { currentUser } from './stores.js';
  import { connectSSE } from './sse.js';
  import ThemeToggle from './ThemeToggle.svelte';

  let nickname = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!nickname.trim()) return;

    loading = true;
    error = '';

    const user = await createUser(nickname.trim());
    if (user) {
      localStorage.setItem('user_id', user.user_id);
      currentUser.set(user);
      connectSSE(user.user_id);
    } else {
      error = 'Failed to create user. Is the server running?';
    }

    loading = false;
  }
</script>

<div class="login-container">
  <div class="login-card">
    <h1>Open Disc</h1>
    <p>Enter a nickname to get started</p>

    <form onsubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Nickname"
        bind:value={nickname}
        disabled={loading}
      />
      <button type="submit" disabled={loading || !nickname.trim()}>
        {loading ? 'Joining...' : 'Join'}
      </button>
    </form>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <div class="theme-row">
      <ThemeToggle />
    </div>
  </div>
</div>

<style>
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background: var(--bg-primary);
  }

  .login-card {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 2.5rem;
    width: 100%;
    max-width: 360px;
    text-align: center;
  }

  h1 {
    color: var(--text-heading);
    margin-bottom: 0.25rem;
    font-size: 1.8rem;
  }

  p {
    color: var(--text-primary);
    margin-bottom: 1.5rem;
    font-size: 0.9rem;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  input {
    padding: 0.6em 0.8em;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    outline: none;
  }

  input:focus {
    border-color: var(--accent);
  }

  button[type='submit'] {
    padding: 0.6em;
    background: var(--accent);
    color: #fdf6e3;
    border: none;
    border-radius: 4px;
    font-weight: 600;
  }

  button[type='submit']:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .error {
    color: var(--red);
    margin-top: 0.75rem;
    margin-bottom: 0;
    font-size: 0.85rem;
  }

  .theme-row {
    margin-top: 1.5rem;
  }
</style>
