---
# open_disc-vrgz
title: SSE JSON event format
status: done
type: task
priority: normal
tags:
    - backend
created_at: 2026-02-19T03:21:02Z
updated_at: 2026-02-19T03:21:02Z
parent: open_disc-533i
---

Replace debug string SSE payload with JSON-serialized Message struct. Lee is handling this on develop branch.

Implemented in `sse_refactor` branch with a generic `RoomEvent` envelope (`message.go`):
```json
{ "room_event_type": "new_message|user_joined|user_left|historical_messages", "payload": <JSON> }
```
- `new_message`: payload is a single `Message`
- `historical_messages`: payload is an array of `Message` (sent on SSE connect, last 10 per room)
- `user_joined` / `user_left`: payload is `{ room_id, user_id }`

Frontend SSE handler (`lib/sse.js`) updated in `frontend-sse-refactor` branch to unwrap the envelope and dispatch by event type.
