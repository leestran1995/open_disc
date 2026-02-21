---
# open_disc-mfbx
title: Fix message fetch returning inconsistent results
status: done
type: bug
priority: high
tags:
    - backend
    - frontend
created_at: 2026-02-21T00:00:00Z
updated_at: 2026-02-21T00:00:00Z
parent: open_disc-533i
---

`GetMessagesByTimestamp` query in `postgresql/message.go` had `LIMIT 10` without `ORDER BY`, so PostgreSQL returned rows in arbitrary order. This caused messages to disappear on page refresh — some messages were excluded unpredictably by the limit.

Fix: Added `ORDER BY m.timestamp DESC` to consistently return the 10 most recent messages. Frontend `MessageList.svelte` reverses the result with `.reverse()` so messages display oldest-first.
