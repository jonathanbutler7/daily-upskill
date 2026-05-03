# System Design Cheat Sheet

<details>
<summary><strong>Databases</strong></summary>

<details>
<summary><strong>SQL</strong></summary>

- **When:** ACID, joins, relationships
- **Tradeoff:** Hard to scale horizontally

</details>
<details>
<summary><strong>NoSQL</strong></summary>

- **When:** Flexible schema, scale writes
- **Tradeoff:** No joins, eventual consistency

</details>
<details>
<summary><strong>Redis</strong></summary>

- **When:** Caching, sessions
- **Tradeoff:** Data loss risk, memory-bound

</details>
</details>

---

<details>
<summary><strong>Architecture</strong></summary>

<details>
<summary><strong>Monolith</strong></summary>

- **When:** Small team, simple domain
- **Tradeoff:** One bug = full outage

</details>
<details>
<summary><strong>Microservices</strong></summary>

- **When:** Large teams, independent scaling
- **Tradeoff:** Network complexity

</details>
<details>
<summary><strong>Serverless</strong></summary>

- **When:** Sporadic traffic, event-driven
- **Tradeoff:** Cold starts, vendor lock-in

</details>
</details>

---

<details>
<summary><strong>Communication</strong></summary>

<details>
<summary><strong>REST</strong></summary>

- **When:** Public APIs, caching
- **Tradeoff:** Over/under-fetching

</details>
<details>
<summary><strong>gRPC</strong></summary>

- **When:** Service-to-service, low latency
- **Tradeoff:** Not browser-native

</details>
<details>
<summary><strong>Message Queue</strong></summary>

- **When:** Decoupling, spikes
- **Tradeoff:** Added latency

</details>
</details>

---

<details>
<summary><strong>Scaling</strong></summary>

<details>
<summary><strong>Vertical</strong></summary>

- **When:** Quick fix, not distributed
- **Tradeoff:** Hardware limits

</details>
<details>
<summary><strong>Horizontal</strong></summary>

- **When:** High availability
- **Tradeoff:** Stateless required

</details>
<details>
<summary><strong>Read Replicas</strong></summary>

- **When:** Read-heavy
- **Tradeoff:** Replication lag

</details>
<details>
<summary><strong>Sharding</strong></summary>

- **When:** Scale writes
- **Tradeoff:** Cross-shard queries expensive

</details>
</details>

---

<details>
<summary><strong>Consistency</strong></summary>

<details>
<summary><strong>Strong</strong></summary>

- **When:** Financial, inventory
- **Tradeoff:** Higher latency

</details>
<details>
<summary><strong>Eventual</strong></summary>

- **When:** Social feeds, analytics
- **Tradeoff:** Stale reads possible

</details>
<details>
<summary><strong>ACID</strong></summary>

- **When:** Data integrity critical
- **Tradeoff:** Harder to scale

</details>
<details>
<summary><strong>BASE</strong></summary>

- **When:** Availability > consistency
- **Tradeoff:** App handles inconsistency

</details>
</details>

---

<details>
<summary><strong>Deployments</strong></summary>

<details>
<summary><strong>Blue-Green</strong></summary>

- **When:** Instant rollback
- **Tradeoff:** 2x resources

</details>
<details>
<summary><strong>Canary</strong></summary>

- **When:** Test with real traffic
- **Tradeoff:** Monitoring overhead

</details>
<details>
<summary><strong>Rolling</strong></summary>

- **When:** Resource-constrained
- **Tradeoff:** Mixed versions

</details>
</details>

---

<details>
<summary><strong>Reliability</strong></summary>

<details>
<summary><strong>Circuit Breaker</strong></summary>

- **When:** Flaky downstream
- **Tradeoff:** Needs tuning

</details>
<details>
<summary><strong>Retry + Backoff</strong></summary>

- **When:** Transient failures
- **Tradeoff:** Retry storms

</details>
<details>
<summary><strong>Timeout</strong></summary>

- **When:** Always
- **Tradeoff:** Too short/long = problems

</details>
</details>
