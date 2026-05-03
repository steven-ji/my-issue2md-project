---
title: "Fix: Increase buffer size for large payloads"
url: https://github.com/owner/repo/pull/43
author: bob
created_at: 2025-01-16T09:00:00Z
state: merged
labels:
  - fix
type: pull_request
---

# Fix: Increase buffer size for large payloads

**Author:** bob | **Created:** 2025-01-16 | **State:** Merged | **Labels:** fix

---

This PR increases the default buffer size from 10MB to 50MB to handle large payloads without timeout.

## Comments

### Comment by alice — 2025-01-16T09:30:00Z

LGTM, tested with 40MB payload.

### Review Comment by charlie — 2025-01-16T10:00:00Z

Should we also add a configurable max payload size? This feels like a hardcoded value.
