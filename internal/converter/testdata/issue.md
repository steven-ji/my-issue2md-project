---
title: "Bug: Connection timeout on large payloads"
url: https://github.com/owner/repo/issues/42
author: alice
created_at: 2025-01-15T10:30:00Z
state: open
labels:
  - bug
  - priority:high
type: issue
---

# Bug: Connection timeout on large payloads

**Author:** alice | **Created:** 2025-01-15 | **State:** Open | **Labels:** bug, priority:high

---

When sending payloads larger than 10MB, the connection consistently times out after 30 seconds.

## Comments

### Comment by bob — 2025-01-15T11:00:00Z

I can reproduce this. It seems related to the buffer size configuration.

### Comment by alice — 2025-01-15T11:15:00Z

> I can reproduce this. It seems related to the buffer size configuration.

Yes, increasing the buffer size to 20MB resolves it temporarily but is not a proper fix.
