# Service Mesh & Communication

How services discover and communicate with each other at Weave.

## Service Communication Diagram

---

## Questions to Ask

- [ ] Do we use a service mesh? Which one (Istio, Linkerd, Envoy, custom)?
- [ ] How do services discover each other?
- [ ] What protocol do we use for service-to-service communication?
- [ ] Do we use mTLS between services?
- [ ] How do we handle service versioning and routing?
- [ ] What's our timeout/retry policy for inter-service calls?

---

## Key Concepts (System Design Interview)

### Service Discovery
- **Client-side discovery**: Client queries registry, picks instance
- **Server-side discovery**: Load balancer queries registry, routes request
- **DNS-based**: Services resolve via DNS (Kubernetes default)
- **Weave's approach**: [Fill in]

### Communication Protocols
| Protocol | Use Case | Pros | Cons |
|----------|----------|------|------|
| REST/HTTP | External APIs, simple CRUD | Universal, debuggable | Verbose, no streaming |
| gRPC | Internal services, high perf | Fast, typed, streaming | Binary, harder to debug |
| GraphQL | Flexible queries | Client flexibility | Complexity, caching |
| **Weave uses**: | [Fill in] | | |

### Service Mesh Benefits
- **Traffic management**: Routing, load balancing, retries
- **Security**: mTLS, authorization policies
- **Observability**: Metrics, tracing, logging
- **Does Weave use these?**: [Fill in]

### Sidecar Pattern
- Proxy runs alongside each service
- Handles cross-cutting concerns
- Examples: Envoy, Linkerd proxy
- **Weave's sidecar setup**: [Fill in]

---

## Weave Implementation

### Service Discovery
- **Method**: [Kubernetes DNS / Consul / Custom / etc.]
- **Service naming convention**: [e.g., service-name.namespace.svc.cluster.local]
- **How to find a service**: [Fill in]

### Communication Patterns
| Pattern | When Used | Example |
|---------|-----------|---------|
| Sync Request/Response | [Fill in] | [Fill in] |
| Async Message Queue | [Fill in] | [Fill in] |
| Event-Driven | [Fill in] | [Fill in] |

### Inter-Service Authentication
- **Method**: [mTLS / JWT / API Keys / etc.]
- **How it works**: [Fill in]
- **How to set up a new service**: [Fill in]

### Timeouts & Retries
- **Default timeout**: [Fill in]
- **Retry policy**: [Fill in]
- **Circuit breaker**: [Fill in]

---

## Resources

- Service Mesh Docs: [Link]
- Service Registry: [Link]
- Example Service Implementation: [Link]

---

## Notes

*Add your notes here as you learn:*
