# rate-limiter-service Architecture

## System Diagram
The following Mermaid.js sequence diagram maps the core workflow and interactions:

```mermaid
sequenceDiagram
    client->>Gateway: Request
Gateway->>RateLimiter: Check limit
RateLimiter->>Redis: EVAL Lua script
Redis-->>RateLimiter: Token count
RateLimiter-->>Gateway: Allow/Deny
```

## Component Breakdown
- **Core Technology**: Go, Redis Lua
- **Design Paradigm**: Emphasizes high availability, fault tolerance, and security.

## Security & Scaling Considerations
- Strict boundary validations.
- Horizontal scalability achieved via stateless workers.
- Encrypted data at rest and in transit.
