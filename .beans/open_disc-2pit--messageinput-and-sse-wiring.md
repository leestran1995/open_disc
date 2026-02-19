---
# open_disc-2pit
title: MessageInput and SSE wiring
status: done
type: task
priority: normal
tags:
    - frontend
created_at: 2026-02-19T03:21:31Z
updated_at: 2026-02-19T03:21:47Z
parent: open_disc-533i
blocked_by:
    - open_disc-a7lk
    - open_disc-vrgz
---

MessageInput.svelte: text input, calls POST /messages on submit. Connect SSE on login (connectSSE in App.svelte after user set). Messages appear via SSE stream, no optimistic insert.
