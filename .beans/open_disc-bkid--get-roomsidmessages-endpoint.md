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

Implemented differently than planned: `MessageService.GetMessagesByTimestamp` was added in `postgresql/message.go` and message history is delivered via SSE `historical_messages` events on connect (last 10 per room) rather than a REST endpoint. The frontend SSE handler loads these into the `messagesByRoom` store automatically. A REST endpoint for on-demand history loading could still be added later.
