---
# open_disc-sse1
title: SSE connect route no longer takes username path parameter
status: done
type: bug
priority: high
tags:
    - frontend
created_at: 2026-02-22T00:00:00Z
updated_at: 2026-02-22T00:00:00Z
---

Backend changed `/connect/:username` to `/connect` (username now comes from JWT), but frontend `sse.js` was still hitting `/sse/connect/${username}`. This caused the SSE connection to fail silently, so no new messages were received.

Fix: Updated `sse.js` to fetch `/sse/connect` without the username path parameter.
