import { writable } from 'svelte/store';

const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : null;
const systemPreference = typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
export const theme = writable(stored || systemPreference);

theme.subscribe((value) => {
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', value === 'dark');
  }
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('theme', value);
  }
});

export function toggleTheme() {
  theme.update((current) => (current === 'dark' ? 'light' : 'dark'));
}
