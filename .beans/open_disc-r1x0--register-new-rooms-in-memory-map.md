---
# open_disc-r1x0
title: Register new rooms in in-memory map on create
status: done
type: bug
priority: high
tags:
    - backend
created_at: 2026-02-19T04:00:00Z
updated_at: 2026-02-19T04:00:00Z
parent: open_disc-533i
---

RoomHandler.HandleCreateRoom saved rooms to DB but never added them to the in-memory rooms map used by MessageService.Send. This caused a nil pointer panic when sending messages to newly created rooms. Fixed by giving RoomHandler access to the rooms map and registering on create.
