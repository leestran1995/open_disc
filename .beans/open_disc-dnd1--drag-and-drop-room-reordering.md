---
# open_disc-dnd1
title: Drag-and-drop room reordering in sidebar
status: done
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-02-21T00:00:00Z
updated_at: 2026-02-22T00:00:00Z
parent: open_disc-533i
---

Frontend support for the backend room ordering API (`sort_order` field, `GET /rooms` sorted, `PUT /rooms/order` bulk reorder).

Four files changed:
1. `lib/api.js` — Added `getRooms()` and `updateRoomOrder(roomIds)` functions.
2. `App.svelte` — Fetches `GET /rooms` on session restore to replace localStorage cache with canonical server order.
3. `lib/Login.svelte` — Fetches `GET /rooms` after signin for correct initial order.
4. `lib/Sidebar.svelte` — Native HTML5 drag-and-drop: draggable room items, 2px accent drop indicator between rooms, optimistic reorder with rollback on API failure. Dragged room fades to 0.3 opacity. No-op detection for drag-to-same-position.
