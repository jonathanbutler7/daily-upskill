# Resilience

How Weave handles failures gracefully.

## Resilience Patterns Diagram

---

## Questions to Ask

- [ ] Do we use circuit breakers? Where?
- [ ] What's our retry strategy?
- [ ] How do we handle downstream service failures?
- [ ] What happens if the database goes down?
- [ ] How do we prevent cascading failures?
- [ ] What fallback mechanisms do we have?

---

## Key Concepts (System Design Interview)

### Circuit Breaker Pattern
```
States:
┌─────────┐    failures > threshold    ┌──────┐
│ CLOSED  │ ─────────────────────────► │ OPEN │
└─────────┘                            └──────┘
     ▲                                      │
     │                                      │ timeout
     │         ┌───────────┐                │
     └──────── │ HALF-OPEN │ ◄──────────────┘
   success     └───────────┘
```

- **Closed**: Normal operation, requests pass through
- **Open**: Fail fast, don't call downstream
- **Half-Open**: Test if downstream recovered

### Retry Strategies
| Strategy | Description | Use Case |
|----------|-------------|----------|
| Immediate | Retry right away | Transient network issues |
| Fixed delay | Wait fixed time | Rate limiting |
| Exponential backoff | Increasing delays | Overloaded services |
| Jitter | Random delay added | Prevent thundering herd |
| **Weave uses**: | [Fill in] | |

### Bulkhead Pattern
- Isolate failures to prevent cascade
- Separate thread pools/connections per dependency
- **Weave's implementation**: [Fill in]

### Timeout Strategies
| Type | Description | Typical Value |
|------|-------------|---------------|
| Connection timeout | Time to establish connection | 1-5 seconds |
| Read timeout | Time to receive response | 5-30 seconds |
| Total timeout | End-to-end request time | 30-60 seconds |
| **Weave's defaults**: | [Fill in] | |

### Fallback Strategies
- Return cached data
- Return default value
- Degrade gracefully
- Queue for later processing
- **Weave's fallbacks**: [Fill in]

---

## Weave Implementation

### Circuit Breaker Configuration
```go
// Example circuit breaker setup
// [Fill in with Weave's actual pattern]
cb := circuitbreaker.New(
    circuitbreaker.WithThreshold(5),
    circuitbreaker.WithTimeout(30 * time.Second),
)
```

### Retry Configuration
| Service/Dependency | Max Retries | Backoff | Timeout |
|-------------------|-------------|---------|---------|
| [Dependency 1] | [Count] | [Strategy] | [Duration] |
| [Dependency 2] | [Count] | [Strategy] | [Duration] |

### Timeout Configuration
| Operation | Timeout | Rationale |
|-----------|---------|-----------|
| Database query | [Duration] | [Why] |
| External API | [Duration] | [Why] |
| Inter-service | [Duration] | [Why] |

### Failure Scenarios
| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Database down | [Impact] | [What happens] |
| Cache down | [Impact] | [What happens] |
| External API down | [Impact] | [What happens] |
| Service A down | [Impact] | [What happens] |

### Graceful Degradation
| Feature | Degraded Behavior | Trigger |
|---------|-------------------|---------|
| [Feature 1] | [What happens] | [When] |
| [Feature 2] | [What happens] | [When] |

---

## Resources

- Resilience Documentation: [Link]
- Circuit Breaker Library: [Link]
- Failure Mode Analysis: [Link]

---

## Notes

*Add your notes here as you learn:*
