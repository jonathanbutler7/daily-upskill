# Data Flow

How data moves through Weave's systems - from user input to storage and back.
---

## Questions to Ask

- [ ] What's the journey of data from user input to database?
- [ ] How do we handle data validation?
- [ ] Do we use event sourcing or CQRS?
- [ ] How does data get to our analytics/reporting systems?
- [ ] What's our data retention policy?
- [ ] How do we handle data migrations?

---

## Key Concepts (System Design Interview)

### Data Consistency Models
| Model | Description | Use Case |
|-------|-------------|----------|
| Strong Consistency | All reads see latest write | Financial transactions |
| Eventual Consistency | Reads may see stale data temporarily | Social feeds, caches |
| Causal Consistency | Related operations are ordered | Messaging systems |
| **Weave's model**: | [Fill in] | |

### CQRS (Command Query Responsibility Segregation)
- Separate read and write models
- Optimized for different access patterns
- **Does Weave use CQRS?**: [Fill in]

### Event Sourcing
- Store events, not current state
- Rebuild state by replaying events
- **Does Weave use event sourcing?**: [Fill in]

### Data Pipeline Patterns
- **ETL**: Extract, Transform, Load (batch)
- **ELT**: Extract, Load, Transform (modern data warehouses)
- **Streaming**: Real-time processing (Kafka, etc.)
- **Weave's approach**: [Fill in]

---

### Read Path
*Trace a read operation (e.g., "User fetches their data"):*

1. Request received: [Fill in]
2. Cache check: [Fill in]
3. Database query: [Fill in]
4. Response assembly: [Fill in]

### Data Stores
| Store | Purpose | Technology | Retention |
|-------|---------|------------|-----------|
| Primary DB | [Fill in] | [PostgreSQL?] | [Fill in] |
| Cache | [Fill in] | [Redis?] | [Fill in] |
| Search | [Fill in] | [Elasticsearch?] | [Fill in] |
| Analytics | [Fill in] | [BigQuery?] | [Fill in] |

### Event/Message Flow
- **Message broker**: [Kafka / RabbitMQ / Pub/Sub / etc.]
- **Event types**: [Fill in]
- **Consumers**: [Fill in]

---

## Resources

- Data Architecture Docs: [Link]
- Database Schema: [Link]
- Event Catalog: [Link]

---

## Notes

*Add your notes here as you learn:*
