<script lang="ts">
  import { signup, signin, getRooms } from './api';
  import { currentUser, authToken, rooms } from './stores';
  import { connectSSE } from './sse';
  import { decodeJWT } from './jwt';
  import ThemeToggle from './ThemeToggle.svelte';
  import type { Room } from './types';

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let message = $state('');
  let loading = $state(false);
  let mode: 'signin' | 'signup' = $state('signin');

  async function handleSubmit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!username.trim() || !password) return;

    loading = true;
    error = '';
    message = '';

    if (mode === 'signup') {
      const result = await signup(username.trim(), password);
      if (result && !('_error' in result)) {
        message = 'Account created! Sign in below.';
        mode = 'signin';
      } else {
        error = result && '_error' in result ? result._error : 'Sign up failed.';
      }
    } else {
      const result = await signin(username.trim(), password);
      if (result && !('_error' in result) && result.data) {
        const token = result.data;
        localStorage.setItem('token', token);
        authToken.set(token);

        const claims = decodeJWT(token);
        const name = claims?.username;
        if (name) {
          currentUser.set({ username: name });
          connectSSE(token, name);

          getRooms().then((roomResult) => {
            if (Array.isArray(roomResult)) {
              rooms.set(roomResult as Room[]);
              localStorage.setItem('rooms', JSON.stringify(roomResult));
            }
          });
        }
      } else {
        error = result && '_error' in result ? result._error : 'Invalid username or password.';
      }
    }

    loading = false;
  }
</script>

<div class="login-container">
  <div class="login-card">
    <h1>Open Disc</h1>
    <p>{mode === 'signin' ? 'Sign in to continue' : 'Create an account'}</p>

    <form onsubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Username"
        bind:value={username}
        disabled={loading}
      />
      <input
        type="password"
        placeholder="Password"
        bind:value={password}
        disabled={loading}
      />
      <button type="submit" disabled={loading || !username.trim() || !password}>
        {loading ? (mode === 'signin' ? 'Signing in...' : 'Signing up...') : (mode === 'signin' ? 'Sign In' : 'Sign Up')}
      </button>
    </form>

    {#if message}
      <p class="success">{message}</p>
    {/if}

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <p class="toggle-link">
      {#if mode === 'signin'}
        Don't have an account? <button class="link-btn" onclick={() => { mode = 'signup'; error = ''; message = ''; }}>Sign Up</button>
      {:else}
        Already have an account? <button class="link-btn" onclick={() => { mode = 'signin'; error = ''; message = ''; }}>Sign In</button>
      {/if}
    </p>

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

  .success {
    color: var(--green, #859900);
    margin-top: 0.75rem;
    margin-bottom: 0;
    font-size: 0.85rem;
  }

  .toggle-link {
    margin-top: 1rem;
    margin-bottom: 0;
    font-size: 0.85rem;
  }

  .link-btn {
    background: none;
    border: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 0.85rem;
    padding: 0;
    text-decoration: underline;
  }

  .theme-row {
    margin-top: 1.5rem;
  }
</style>
