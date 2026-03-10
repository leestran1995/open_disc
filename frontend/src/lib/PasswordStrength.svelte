<script lang="ts">
  import { checkPasswordStrength } from './password';

  interface Props {
    password: string;
  }

  let { password }: Props = $props();

  let result = $derived(checkPasswordStrength(password));

  const criteria = [
    { key: 'has_uppercase', label: 'Uppercase letter' },
    { key: 'has_lowercase', label: 'Lowercase letter' },
    { key: 'has_number', label: 'Number' },
    { key: 'has_special', label: 'Special character (!@#$%^&*()-+)' },
    { key: 'has_eight_chars', label: '8+ characters' },
  ] as const;
</script>

{#if password}
  <ul class="strength-list">
    {#each criteria as { key, label }}
      <li class:pass={result[key]}>
        <span class="icon">{result[key] ? '\u2713' : '\u2717'}</span>
        {label}
      </li>
    {/each}
  </ul>
{/if}

<style>
  .strength-list {
    list-style: none;
    padding: 0;
    margin: -0.25rem 0 0;
    text-align: left;
    font-size: 0.8rem;
  }

  li {
    color: var(--red);
    padding: 0.15em 0;
  }

  li.pass {
    color: var(--green);
  }

  .icon {
    display: inline-block;
    width: 1.2em;
    font-weight: 600;
  }
</style>
