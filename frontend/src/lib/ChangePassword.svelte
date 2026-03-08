<script lang="ts">
  import { changePassword } from './api';
  import { checkPasswordStrength, isPasswordValid } from './password';
  import PasswordStrength from './PasswordStrength.svelte';

  interface Props {
    onclose: () => void;
  }

  let { onclose }: Props = $props();

  let oldPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');
  let error = $state('');
  let success = $state('');
  let loading = $state(false);

  let newPasswordValid = $derived(isPasswordValid(checkPasswordStrength(newPassword)));
  let passwordsMatch = $derived(newPassword === confirmPassword);
  let canSubmit = $derived(
    oldPassword.length > 0 && newPasswordValid && passwordsMatch && !loading
  );

  async function handleSubmit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!canSubmit) return;

    loading = true;
    error = '';
    success = '';

    const result = await changePassword(oldPassword, newPassword);
    if (result && !('_error' in result)) {
      success = 'Password changed successfully.';
      oldPassword = '';
      newPassword = '';
      confirmPassword = '';
    } else {
      error = result && '_error' in result ? result._error : 'Password change failed.';
    }

    loading = false;
  }

  function handleBackdropClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) onclose();
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div class="overlay" onclick={handleBackdropClick}>
  <div class="modal">
    <div class="modal-header">
      <h2>Change Password</h2>
      <button class="close-btn" onclick={onclose}>&times;</button>
    </div>

    <form onsubmit={handleSubmit}>
      <input
        type="password"
        placeholder="Current password"
        bind:value={oldPassword}
        disabled={loading}
      />
      <input
        type="password"
        placeholder="New password"
        bind:value={newPassword}
        disabled={loading}
      />
      <PasswordStrength password={newPassword} />
      {#if confirmPassword && !passwordsMatch}
        <p class="mismatch">Passwords do not match.</p>
      {/if}
      <input
        type="password"
        placeholder="Confirm new password"
        bind:value={confirmPassword}
        disabled={loading}
      />
      <button type="submit" disabled={!canSubmit}>
        {loading ? 'Changing...' : 'Change Password'}
      </button>
    </form>

    {#if success}
      <p class="success">{success}</p>
    {/if}
    {#if error}
      <p class="error">{error}</p>
    {/if}
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 2rem;
    width: 100%;
    max-width: 360px;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.25rem;
  }

  h2 {
    color: var(--text-heading);
    font-size: 1.2rem;
    margin: 0;
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-primary);
    font-size: 1.4rem;
    cursor: pointer;
    padding: 0;
    line-height: 1;
    opacity: 0.7;
  }

  .close-btn:hover {
    opacity: 1;
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
    margin: 0.75rem 0 0;
    font-size: 0.85rem;
  }

  .success {
    color: var(--green);
    margin: 0.75rem 0 0;
    font-size: 0.85rem;
  }

  .mismatch {
    color: var(--red);
    font-size: 0.8rem;
    margin: -0.25rem 0 0;
  }
</style>
