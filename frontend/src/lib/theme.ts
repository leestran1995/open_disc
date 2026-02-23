import { writable } from 'svelte/store';

type Theme = 'light' | 'dark';

const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : null;
const systemPreference: Theme = typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
const initial: Theme = stored === 'dark' || stored === 'light' ? stored : systemPreference;
export const theme = writable<Theme>(initial);

theme.subscribe((value) => {
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', value === 'dark');
  }
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('theme', value);
  }
});

export function toggleTheme(): void {
  theme.update((current) => (current === 'dark' ? 'light' : 'dark'));
}
