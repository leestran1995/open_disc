---
# open_disc-a7lk
title: MessageList, Message, and history loading
status: done
type: task
priority: normal
tags:
    - frontend
created_at: 2026-02-19T03:21:28Z
updated_at: 2026-02-19T03:21:47Z
parent: open_disc-533i
blocked_by:
    - open_disc-gk6k
    - open_disc-bkid
---

MessageList.svelte: fetches GET /rooms/:id/messages when activeRoomId changes, auto-scrolls to bottom. Message.svelte: displays text, timestamp, sender nickname (cached user lookups).
