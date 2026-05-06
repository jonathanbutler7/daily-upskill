# System Design Cheat Sheet

<details>
<summary><strong>Databases</strong></summary>

<details>
<summary><strong>SQL</strong></summary>

- **When:** Relationships, complex joins, ACID transactions, 
- **Tradeoff:** Horizontal scaling is complex and expensive

</details>
<details>
<summary><strong>NoSQL</strong></summary>

- **When:** Flexible schema, high write throughput
- **Tradeoff:** Limited joins, eventual consistency

</details>
<details>
<summary><strong>Redis</strong></summary>

- **When:** Caching, sessions, rate limiting
- **Tradeoff:** Volatile by default, memory-bound

</details>
</details>

---

<details>
<summary><strong>Architecture</strong></summary>

<details>
<summary><strong>Monolith</strong></summary>

- **When:** Small teams, simple domains
- **Tradeoff:** Tight coupling, single point of failure

</details>
<details>
<summary><strong>Microservices</strong></summary>

- **When:** Large orgs, independent scaling & deployment
- **Tradeoff:** Distributed systems complexity (latency, failures, tracing)

</details>
<details>
<summary><strong>Serverless</strong></summary>

- **When:** Variable/sporadic traffic, event-driven workloads
- **Tradeoff:** Cold starts, vendor lock-in, tail latency

</details>
</details>

---

<details>
<summary><strong>Communication</strong></summary>

<details>
<summary><strong>REST</strong></summary>

- **When:** Public APIs, browser compatibility, caching
- **Tradeoff:** Over/under-fetching

</details>
<details>
<summary><strong>gRPC</strong></summary>

- **When:** Internal services, low latency, streaming
- **Tradeoff:** Poor browser support (requires gRPC-Web)

</details>
<details>
<summary><strong>Message Queue</strong></summary>

- **When:** Decoupling, traffic spikes, async processing
- **Tradeoff:** Added latency & operational complexity

</details>
</details>

---

<details>
<summary><strong>Scaling</strong></summary>

<details>
<summary><strong>Vertical</strong></summary>

- **When:** Simple, quick scaling (single node)
- **Tradeoff:** Hardware limits, single point of failure

</details>
<details>
<summary><strong>Horizontal</strong></summary>

- **When:** High availability & massive scale
- **Tradeoff:** Requires stateless design (or distributed state)

</details>
<details>
<summary><strong>Read Replicas</strong></summary>

- **When:** Read-heavy workloads
- **Tradeoff:** Replication lag (eventual consistency)

</details>
<details>
<summary><strong>Sharding</strong></summary>

- **When:** Massive write scale
- **Tradeoff:** Cross-shard queries are complex & expensive

</details>
</details>

---

<details>
<summary><strong>Consistency</strong></summary>

<details>
<summary><strong>Strong Consistency</strong></summary>

- **When:** Financial transactions, inventory, reservations
- **Tradeoff:** Higher latency, lower availability under partitions

</details>
<details>
<summary><strong>Eventual Consistency</strong></summary>

- **When:** Social feeds, analytics, recommendations
- **Tradeoff:** Temporary inconsistencies (stale reads)

</details>
<details>
<summary><strong>ACID</strong></summary>

- Atomicity, Consistency, Isolation, and Durability
- **When:** Data integrity is critical
- **Tradeoff:** Harder to scale horizontally

</details>
<details>
<summary><strong>BASE</strong></summary>

- Basically Available, Soft state, Eventual consistency
- **When:** Availability > immediate consistency
- **Tradeoff:** Application must handle inconsistency

</details>
</details>

---

<details>
<summary><strong>Deployments</strong></summary>

<details>
<summary><strong>Blue-Green</strong></summary>

- **When:** Zero-downtime deployments, instant rollback
- **Tradeoff:** Double resource usage during cutover

</details>
<details>
<summary><strong>Canary</strong></summary>

- **When:** Gradual rollout with real user traffic
- **Tradeoff:** Complex monitoring & rollout logic

</details>
<details>
<summary><strong>Rolling</strong></summary>

- **When:** Resource-constrained environments
- **Tradeoff:** Mixed versions running simultaneously

</details>
</details>

---

<details>
<summary><strong>Reliability</strong></summary>

<details>
<summary><strong>Circuit Breaker</strong></summary>

- **When:** Protecting against flaky downstream services
- **Tradeoff:** Requires careful tuning & monitoring

</details>
<details>
<summary><strong>Retry + Backoff</strong></summary>

- **When:** Handling transient failures
- **Tradeoff:** Risk of retry storms / thundering herd

</details>
<details>
<summary><strong>Timeout</strong></summary>

- **When:** Everywhere
- **Tradeoff:** Poor tuning leads to cascading failures or false errors

</details>
</details>
