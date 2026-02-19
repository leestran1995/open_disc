---
# open_disc-gk6k
title: Sidebar and room list
status: done
type: task
priority: normal
tags:
    - frontend
created_at: 2026-02-19T03:21:19Z
updated_at: 2026-02-19T03:21:47Z
parent: open_disc-533i
blocked_by:
    - open_disc-13pj
    - open_disc-w0nf
---

Sidebar.svelte: fetches user rooms on mount, lists them, clicking sets activeRoomId. Create Room UI (input + button) calls POST /rooms then POST /rooms/:id/join then refreshes list.
