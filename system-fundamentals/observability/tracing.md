# Distributed Tracing

How Weave traces requests across services.

## Tracing Architecture Diagram

---

## Questions to Ask

- [ ] Do we use distributed tracing? What tool?
- [ ] How do I trace a request across services?
- [ ] What's the sampling rate for traces?
- [ ] How do we debug performance issues?
- [ ] How do I add tracing to a new service?
- [ ] What's the overhead of tracing?

---

## Key Concepts (System Design Interview)

### Tracing Terminology
| Term | Description |
|------|-------------|
| Trace | End-to-end journey of a request |
| Span | Single operation within a trace |
| Trace ID | Unique identifier for the entire trace |
| Span ID | Unique identifier for a single span |
| Parent Span | The span that called this span |
| Baggage | Context propagated across services |

### Trace Anatomy
```
Trace ID: abc123
├── Span: API Gateway (10ms)
│   ├── Span: Auth Check (2ms)
│   └── Span: Service A (8ms)
│       ├── Span: Database Query (3ms)
│       └── Span: Service B (4ms)
│           └── Span: Cache Lookup (1ms)
```

### Sampling Strategies
| Strategy | Description | Use Case |
|----------|-------------|----------|
| Head-based | Decide at start | Simple, predictable |
| Tail-based | Decide at end | Capture interesting traces |
| Rate-based | Fixed percentage | Cost control |
| Adaptive | Dynamic rate | Balance cost/coverage |
| **Weave uses**: | [Fill in] | |

### Context Propagation
- **W3C Trace Context**: Standard header format
- **B3**: Zipkin's format
- **How Weave propagates context**: [Fill in]

---

## Weave Implementation

### Tracing Stack
| Component | Technology | Purpose |
|-----------|------------|---------|
| Instrumentation | [OpenTelemetry / etc.] | [Purpose] |
| Collection | [Jaeger / Zipkin / etc.] | [Purpose] |
| Storage | [Elasticsearch / etc.] | [Purpose] |
| UI | [Jaeger UI / etc.] | [Purpose] |

### Sampling Configuration
- **Default rate**: [Fill in]
- **How to adjust**: [Fill in]
- **Cost considerations**: [Fill in]

### Instrumenting a Service
```go
// Example: How to add tracing in Go at Weave
// [Fill in with actual pattern]

// Creating a span
span := tracer.StartSpan("operation-name")
defer span.Finish()

// Adding tags
span.SetTag("user_id", userID)

// Propagating context
ctx = opentracing.ContextWithSpan(ctx, span)
```

### Finding a Trace
1. **By Trace ID**: [How to search]
2. **By Service**: [How to filter]
3. **By Duration**: [How to find slow traces]
4. **By Error**: [How to find failed traces]

### Common Debugging Scenarios
| Scenario | How to Use Tracing |
|----------|-------------------|
| Slow request | Find trace, identify slow spans |
| Failed request | Find trace, look for error tags |
| Missing data | Trace the write path |
| Intermittent issues | Compare successful vs failed traces |

### Trace Retention
- **Duration**: [Fill in]
- **Storage location**: [Fill in]

---

## Resources

- Tracing Documentation: [Link]
- Trace UI: [Link]
- Instrumentation Guide: [Link]

---

## Notes

*Add your notes here as you learn:*
