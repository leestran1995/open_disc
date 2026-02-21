import { nameToEmoji } from 'gemoji';

const emojiNames = Object.keys(nameToEmoji);

export function replaceEmoji(text) {
  return text.replace(/:([a-zA-Z0-9_+-]+):/g, (match, name) => {
    return nameToEmoji[name] || match;
  });
}

export function searchEmoji(query, limit = 8) {
  if (!query) return [];
  const lower = query.toLowerCase();
  const results = [];
  for (const name of emojiNames) {
    if (name.startsWith(lower)) {
      results.push({ name, emoji: nameToEmoji[name] });
      if (results.length >= limit) break;
    }
  }
  return results;
}
