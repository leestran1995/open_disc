---
# open_disc-lv66
title: Port unification (nginx/Caddy reverse proxy)
status: todo
type: task
priority: low
tags:
    - infra
created_at: 2026-02-19T03:21:39Z
updated_at: 2026-02-19T03:21:39Z
parent: open_disc-533i
---

TODO: Put nginx or Caddy in front of the two Go servers (8080 REST, 8081 SSE) to serve everything on one origin. Not blocking for dev (Vite proxy handles it), needed for production Docker deployment.
