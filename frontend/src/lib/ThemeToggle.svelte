<script lang="ts">
  import { theme, toggleTheme } from './theme';
</script>

<button class="theme-toggle" onclick={toggleTheme}>
  {$theme === 'dark' ? 'Light' : 'Dark'} mode
</button>

<style>
  .theme-toggle {
    background: var(--bg-secondary);
    color: var(--text-primary);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 0.4em 0.8em;
    font-size: 0.8rem;
  }

  .theme-toggle:hover {
    border-color: var(--accent);
  }
</style>
