---
# open_disc-bkid
title: GET /rooms/:id/messages endpoint
status: done
type: task
priority: normal
tags:
    - backend
    - lee
created_at: 2026-02-19T03:21:07Z
updated_at: 2026-02-19T03:21:07Z
parent: open_disc-533i
---

Add GetByRoomID method to MessageService and expose as HTTP route. Returns last 50 messages for a room ordered by timestamp. Needed for message history on room switch.

Implemented as `GET /messages/:room_id` with optional `?timestamp=` cursor for pagination. `MessageService.GetMessagesByTimestamp` in `postgresql/message.go` returns last 10 messages ordered by `timestamp DESC`. Frontend fetches via `getMessages()` in `api.js` on room select, reverses to display oldest-first. SSE no longer delivers historical messages on connect — client fetches via REST instead.
