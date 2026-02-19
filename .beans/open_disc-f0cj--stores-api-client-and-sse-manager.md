---
# open_disc-f0cj
title: Stores, API client, and SSE manager
status: done
type: task
priority: normal
tags:
    - frontend
created_at: 2026-02-19T03:21:16Z
updated_at: 2026-02-19T03:21:47Z
parent: open_disc-533i
blocked_by:
    - open_disc-prg2
---

Create lib/stores.js (currentUser, rooms, activeRoomId, messagesByRoom). Create lib/api.js (fetch wrapper for all REST endpoints). Create lib/sse.js (EventSource connection manager with JSON parsing and dedup).
