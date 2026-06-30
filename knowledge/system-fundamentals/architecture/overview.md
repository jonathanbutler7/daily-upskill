# Architecture Overview

High-level view of Weave's system architecture.

## System Diagram

---

## Questions to Ask

- [ ] What are the core services and what does each one do?
- [ ] How does a typical user request flow through the system?
- [ ] What's the most critical path in our architecture?
- [ ] Which services have the most dependencies?
- [ ] What services are stateless vs stateful?

---

## Key Concepts (System Design Interview)

### Microservices vs Monolith
- **Microservices**: Independent deployment, technology diversity, team autonomy
- **Monolith**: Simpler deployment, easier debugging, lower latency
- **Weave's approach**: [Fill in]

### Service Boundaries
- Domain-driven design (DDD)
- Single responsibility
- Data ownership
- **How Weave defines boundaries**: [Fill in]

### Communication Patterns
- Synchronous (REST, gRPC)
- Asynchronous (message queues, events)
- **Weave's primary pattern**: [Fill in]

---

## Weave Implementation

### Core Services
| Service | Purpose | Dependencies | Owner Team |
|---------|---------|--------------|------------|
| [Service 1] | [What it does] | [What it depends on] | [Team] |
| [Service 2] | [What it does] | [What it depends on] | [Team] |
| [Service 3] | [What it does] | [What it depends on] | [Team] |

### Request Flow Example
*Describe a typical user action (e.g., "User sends a message") and trace it through the system:*

1. User action: [Fill in]
2. Hits load balancer: [Fill in]
3. Routes to service: [Fill in]
4. Database interaction: [Fill in]
5. Response path: [Fill in]

### Key Numbers
- Total number of services: [Fill in]
- Average request latency: [Fill in]
- Peak requests per second: [Fill in]
- Database size: [Fill in]

---

## Resources

- Internal Architecture Docs: [Link]
- Service Catalog: [Link]
- Architecture Decision Records (ADRs): [Link]

---

## Notes

*Add your notes here as you learn:*
