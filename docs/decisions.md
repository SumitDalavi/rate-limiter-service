# Decisions

## ADR-001: Lua Scripting for Token Bucket
**Date:** 2026-08-29  
**Status:** Accepted

**Context:**  
We need to evaluate the token bucket algorithm. Fetching tokens, calculating time deltas, and writing back the new token count requires multiple network round trips and risks race conditions.

**Decision:**  
We will use a Redis Lua script to perform the entire read-evaluate-write operation atomically on the Redis server.

**Consequences:**  
- ✅ Completely eliminates race conditions.
- ✅ Reduces network overhead to a single `EVALSHA` call per request.
- ⚠️ Logic is tied to Redis and written in Lua rather than Go.
