# System Design Patterns Reference

Quick-reference guide for common system design decisions and their tradeoffs.

---

## Data Storage Patterns

### Relational (SQL) vs NoSQL

**Relational (PostgreSQL, MySQL)**:
- Use when: You need ACID transactions, complex joins, or structured data with relationships
- Tradeoff: Harder to scale horizontally; schema changes can be painful

**NoSQL (MongoDB, DynamoDB, Cassandra)**:
- Use when: You need flexible schemas, horizontal scaling, or high write throughput
- Tradeoff: Limited join support; eventual consistency in many cases

### NoSQL Subtypes

| Type | Examples | Use When | Tradeoff |
|------|----------|----------|----------|
| **Document** | MongoDB, CouchDB | Flexible, nested data structures | No joins; denormalization required |
| **Key-Value** | Redis, DynamoDB | Simple lookups, caching, sessions | No complex queries |
| **Column-Family** | Cassandra, HBase | Time-series, write-heavy workloads | Complex data modeling |
| **Graph** | Neo4j, Neptune | Relationship-heavy queries (social networks) | Not great for non-graph queries |

### Caching Strategies

**Cache-Aside (Lazy Loading)**:
- Use when: Read-heavy workloads with tolerance for stale data
- Tradeoff: Cache miss = slow first request; potential for stale data

**Write-Through**:
- Use when: You need strong consistency between cache and DB
- Tradeoff: Higher write latency; every write hits both cache and DB

**Write-Back (Write-Behind)**:
- Use when: Write-heavy workloads where you can tolerate data loss risk
- Tradeoff: Data loss if cache fails before persisting to DB

---

## Architecture Patterns

### Monolith vs Microservices

**Monolith**:
- Use when: Small team, early-stage product, simple domain
- Tradeoff: One bug can take down everything; harder to scale individual components

**Microservices**:
- Use when: Large teams, independent deployment needs, different scaling requirements per service
- Tradeoff: Network complexity; distributed system challenges (latency, partial failures)

### Event-Driven vs Request-Response

**Request-Response (Synchronous)**:
- Use when: You need immediate responses; simple request/reply patterns
- Tradeoff: Tight coupling; caller blocks waiting for response

**Event-Driven (Asynchronous)**:
- Use when: Loose coupling needed; fire-and-forget operations; complex workflows
- Tradeoff: Harder to debug; eventual consistency; message ordering challenges

### Serverless vs Containers vs VMs

**Serverless (Lambda, Cloud Functions)**:
- Use when: Sporadic traffic, event-driven workloads, minimal ops overhead
- Tradeoff: Cold starts; vendor lock-in; limited execution time

**Containers (Kubernetes, ECS)**:
- Use when: Consistent environments, microservices, need control over runtime
- Tradeoff: Operational complexity; need to manage orchestration

**VMs (EC2, Compute Engine)**:
- Use when: Legacy apps, full OS control, compliance requirements
- Tradeoff: Slower scaling; more resource overhead

---

## Communication Patterns

### REST vs GraphQL vs gRPC

**REST**:
- Use when: Simple CRUD APIs, broad client compatibility, caching important
- Tradeoff: Over-fetching/under-fetching; multiple round trips for related data

**GraphQL**:
- Use when: Clients need flexible queries, multiple data sources, mobile apps with bandwidth concerns
- Tradeoff: Complexity; caching harder; potential for expensive queries

**gRPC**:
- Use when: Service-to-service communication, low latency, streaming needed
- Tradeoff: Not browser-native; requires protobuf schema management

### Message Queues vs Direct Calls

**Direct Calls (HTTP/gRPC)**:
- Use when: Immediate response needed; simple request/reply
- Tradeoff: Tight coupling; caller fails if callee is down

**Message Queues (Kafka, RabbitMQ, SQS)**:
- Use when: Decoupling services; handling traffic spikes; guaranteed delivery
- Tradeoff: Added latency; message ordering complexity; operational overhead

### Pub/Sub vs Point-to-Point

**Pub/Sub**:
- Use when: Multiple consumers need the same message; event broadcasting
- Tradeoff: No guarantee of processing order across consumers

**Point-to-Point (Queue)**:
- Use when: Single consumer per message; work distribution
- Tradeoff: Scaling consumers requires careful coordination

---

## Scaling Patterns

### Horizontal vs Vertical Scaling

**Vertical Scaling (Scale Up)**:
- Use when: Quick fix; application isn't designed for distribution
- Tradeoff: Hardware limits; single point of failure; expensive at scale

**Horizontal Scaling (Scale Out)**:
- Use when: Need high availability; traffic exceeds single machine capacity
- Tradeoff: Requires stateless design; distributed system complexity

### Read Replicas vs Sharding

**Read Replicas**:
- Use when: Read-heavy workloads; need to offload reads from primary
- Tradeoff: Replication lag; writes still bottlenecked on primary

**Sharding (Horizontal Partitioning)**:
- Use when: Data too large for single node; need to distribute writes
- Tradeoff: Cross-shard queries expensive; rebalancing is painful; application complexity

### Load Balancing Strategies

| Strategy | Use When | Tradeoff |
|----------|----------|----------|
| **Round Robin** | Servers are homogeneous; simple distribution | Ignores server load |
| **Least Connections** | Requests have varying processing times | Slightly more overhead |
| **IP Hash** | Need session affinity without cookies | Uneven distribution if IPs clustered |
| **Weighted** | Servers have different capacities | Manual configuration needed |

---

## Data Consistency Patterns

### Strong vs Eventual Consistency

**Strong Consistency**:
- Use when: Financial transactions, inventory, anything where stale reads are unacceptable
- Tradeoff: Higher latency; reduced availability during partitions

**Eventual Consistency**:
- Use when: Social feeds, analytics, caching—where slight staleness is acceptable
- Tradeoff: Clients may see stale data; harder to reason about

### ACID vs BASE

**ACID (Atomicity, Consistency, Isolation, Durability)**:
- Use when: Transactions must be all-or-nothing; data integrity critical
- Tradeoff: Harder to scale; locks reduce throughput

**BASE (Basically Available, Soft state, Eventually consistent)**:
- Use when: High availability more important than immediate consistency
- Tradeoff: Application must handle inconsistency; compensating transactions needed

### Single Leader vs Multi-Leader vs Leaderless

**Single Leader**:
- Use when: Strong consistency needed; simpler conflict resolution
- Tradeoff: Write bottleneck; failover complexity

**Multi-Leader**:
- Use when: Geographically distributed writes; offline-capable clients
- Tradeoff: Conflict resolution required; eventual consistency

**Leaderless (Dynamo-style)**:
- Use when: High availability; tolerance for conflicts
- Tradeoff: Read/write quorums add latency; conflict resolution on reads

---

## Deployment Patterns

### Blue-Green vs Canary vs Rolling

**Blue-Green**:
- Use when: Need instant rollback; can afford 2x infrastructure during deploy
- Tradeoff: Resource cost; database migrations tricky

**Canary**:
- Use when: Want to test with real traffic before full rollout
- Tradeoff: Requires traffic splitting; monitoring overhead

**Rolling**:
- Use when: Resource-constrained; gradual rollout acceptable
- Tradeoff: Mixed versions during deploy; slower rollback

### Feature Flags vs Branch Deploys

**Feature Flags**:
- Use when: Decouple deploy from release; A/B testing; gradual rollouts
- Tradeoff: Flag debt accumulates; code complexity

**Branch Deploys**:
- Use when: Testing full features in isolation before merge
- Tradeoff: Environment drift; merge conflicts

---

## Reliability Patterns

### Circuit Breaker vs Retry vs Timeout

**Circuit Breaker**:
- Use when: Downstream service is flaky; prevent cascade failures
- Tradeoff: Adds complexity; needs tuning for thresholds

**Retry with Backoff**:
- Use when: Transient failures expected; idempotent operations
- Tradeoff: Can amplify load during outages (retry storms)

**Timeout**:
- Use when: Always—never wait forever for a response
- Tradeoff: Too short = false failures; too long = resource exhaustion

### Rate Limiting Strategies

| Strategy | Use When | Tradeoff |
|----------|----------|----------|
| **Token Bucket** | Allow bursts while enforcing average rate | Slightly complex to implement |
| **Leaky Bucket** | Smooth, constant output rate | No burst allowance |
| **Fixed Window** | Simple rate limiting | Burst at window boundaries |
| **Sliding Window** | Accurate rate limiting without boundary issues | More memory/computation |

---

## Quick Decision Matrix

| Scenario | Likely Choice | Why |
|----------|---------------|-----|
| Early startup, small team | Monolith + PostgreSQL | Simplicity; iterate fast |
| High read traffic, simple queries | Add Redis cache | Offload DB; sub-ms reads |
| Need to scale writes | Sharding or NoSQL | Distribute write load |
| Service-to-service calls | gRPC | Low latency; type safety |
| Public API | REST | Broad compatibility |
| Decoupling services | Message queue | Async; handle failures gracefully |
| Global users, low latency | CDN + Read replicas | Data closer to users |
| Compliance/audit requirements | Strong consistency + ACID | Data integrity |

---

## Notes

*Add your own patterns and learnings here:*
