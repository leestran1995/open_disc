import { nameToEmoji } from 'gemoji';
import type { EmojiSuggestion } from './types';

const emojiNames = Object.keys(nameToEmoji);

export function replaceEmoji(text: string): string {
  return text.replace(/:([a-zA-Z0-9_+-]+):/g, (match, name: string) => {
    return nameToEmoji[name] || match;
  });
}

export function searchEmoji(query: string, limit = 8): EmojiSuggestion[] {
  if (!query) return [];
  const lower = query.toLowerCase();
  const results: EmojiSuggestion[] = [];
  for (const name of emojiNames) {
    if (name.startsWith(lower)) {
      results.push({ name, emoji: nameToEmoji[name] });
      if (results.length >= limit) break;
    }
  }
  return results;
}
