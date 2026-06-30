# Logging

How Weave collects, stores, and queries logs.

## Logging Architecture Diagram

---

## Questions to Ask

- [ ] Where do logs go? How do I query them?
- [ ] What's our log retention policy?
- [ ] What should I log vs what's too verbose?
- [ ] How do we correlate logs across services?
- [ ] What log levels do we use?
- [ ] How do I add structured logging to a service?
- [ ] What's the cost of logging? Any limits?

---

## Key Concepts (System Design Interview)

### Log Levels
| Level | Description | When to Use |
|-------|-------------|-------------|
| DEBUG | Detailed diagnostic info | Development, troubleshooting |
| INFO | General operational events | Normal operations |
| WARN | Potential issues | Degraded but functional |
| ERROR | Failures that need attention | Request failures, exceptions |
| FATAL | System cannot continue | Startup failures, crashes |
| **Weave's conventions**: | [Fill in] | |

### Structured Logging
```json
{
  "timestamp": "2026-04-15T10:30:00Z",
  "level": "INFO",
  "service": "user-service",
  "trace_id": "abc123",
  "message": "User created",
  "user_id": "12345"
}
```
- Easier to parse and query
- Consistent format across services
- **Weave's format**: [Fill in]

### Log Aggregation Patterns
| Pattern | Description | Example |
|---------|-------------|---------|
| Sidecar | Agent per pod | Fluentd sidecar |
| DaemonSet | Agent per node | Fluentd DaemonSet |
| Direct | App ships logs | App → Elasticsearch |
| **Weave uses**: | [Fill in] | |

### Correlation
- **Trace ID**: Links logs across services for a request
- **Span ID**: Links logs within a service
- **Request ID**: User-facing identifier
- **Weave's correlation strategy**: [Fill in]

---

## Weave Implementation

### Logging Stack
| Component | Technology | Purpose |
|-----------|------------|---------|
| Collection | [Fluentd / Fluent Bit / etc.] | [Purpose] |
| Storage | [Elasticsearch / Loki / etc.] | [Purpose] |
| Query UI | [Kibana / Grafana / etc.] | [Purpose] |

### Log Retention
| Log Type | Retention | Archive |
|----------|-----------|---------|
| Application | [Duration] | [Yes/No] |
| Access | [Duration] | [Yes/No] |
| Audit | [Duration] | [Yes/No] |

### Querying Logs
```
# Example queries for [your log tool]
# [Fill in with actual query syntax]

# Find errors for a service
service:user-service AND level:ERROR

# Find logs for a specific request
trace_id:abc123

# Find slow requests
response_time:>1000
```

### Logging Best Practices at Weave
1. [Practice 1]
2. [Practice 2]
3. [Practice 3]

### Adding Logging to a Service
```go
// Example: How to log in Go at Weave
// [Fill in with actual pattern]
```

### Common Log Queries
| What You're Looking For | Query |
|------------------------|-------|
| Errors in last hour | [Query] |
| Logs for a user | [Query] |
| Slow requests | [Query] |

---

## Resources

- Logging Documentation: [Link]
- Log Query UI: [Link]
- Logging Guidelines: [Link]

---

## Notes

*Add your notes here as you learn:*
