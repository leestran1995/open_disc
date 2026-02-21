---
# open_disc-em0j
title: Emoji shortcode support
status: done
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-02-21T00:00:00Z
updated_at: 2026-02-21T00:00:00Z
parent: open_disc-533i
---

Discord/Slack-style `:shortcode:` to unicode emoji conversion (e.g. `:smile:` -> 😄). Frontend-only — messages stored as raw text in DB, shortcodes replaced at render time so historical messages benefit retroactively. Uses `gemoji` npm package (~1,900 shortcodes).

Three parts:
1. `lib/emoji.js` — `replaceEmoji(text)` replaces `:shortcode:` patterns with unicode. `searchEmoji(query, limit)` returns prefix-matched results for autocomplete.
2. `lib/Message.svelte` — `$derived` renders shortcodes in displayed messages.
3. `lib/MessageInput.svelte` — Autocomplete popup appears after typing `:` + 2 chars. Arrow keys navigate, Tab/Enter select, Escape dismisses. Inline replacement also triggers when user types the closing `:`.
