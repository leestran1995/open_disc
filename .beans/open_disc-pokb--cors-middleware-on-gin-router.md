---
# open_disc-pokb
title: CORS middleware on Gin router
status: todo
type: task
priority: normal
tags:
    - backend
    - lee
created_at: 2026-02-19T03:21:04Z
updated_at: 2026-02-19T03:21:04Z
parent: open_disc-533i
---

Add gin-contrib/cors middleware to the Gin router before route binding. Allow localhost:4000 during dev (Vite dev server port). Not blocking for dev since Vite proxies `/api/*` and `/sse/*` to the Go servers, but needed for production when frontend is served separately.
