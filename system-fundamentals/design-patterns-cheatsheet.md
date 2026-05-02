# System Design Cheat Sheet

## Databases
| Choice | When | Tradeoff |
|--------|------|----------|
| **SQL** | ACID, joins, relationships | Hard to scale horizontally |
| **NoSQL** | Flexible schema, scale writes | No joins, eventual consistency |
| **Redis** | Caching, sessions | Data loss risk, memory-bound |

## Architecture
| Choice | When | Tradeoff |
|--------|------|----------|
| **Monolith** | Small team, simple domain | One bug = full outage |
| **Microservices** | Large teams, independent scaling | Network complexity |
| **Serverless** | Sporadic traffic, event-driven | Cold starts, vendor lock-in |

## Communication
| Choice | When | Tradeoff |
|--------|------|----------|
| **REST** | Public APIs, caching | Over/under-fetching |
| **gRPC** | Service-to-service, low latency | Not browser-native |
| **Message Queue** | Decoupling, spikes | Added latency |

## Scaling
| Choice | When | Tradeoff |
|--------|------|----------|
| **Vertical** | Quick fix, not distributed | Hardware limits |
| **Horizontal** | High availability | Stateless required |
| **Read Replicas** | Read-heavy | Replication lag |
| **Sharding** | Scale writes | Cross-shard queries expensive |

## Consistency
| Choice | When | Tradeoff |
|--------|------|----------|
| **Strong** | Financial, inventory | Higher latency |
| **Eventual** | Social feeds, analytics | Stale reads possible |
| **ACID** | Data integrity critical | Harder to scale |
| **BASE** | Availability > consistency | App handles inconsistency |

## Deployments
| Choice | When | Tradeoff |
|--------|------|----------|
| **Blue-Green** | Instant rollback | 2x resources |
| **Canary** | Test with real traffic | Monitoring overhead |
| **Rolling** | Resource-constrained | Mixed versions |

## Reliability
| Pattern | When | Tradeoff |
|---------|------|----------|
| **Circuit Breaker** | Flaky downstream | Needs tuning |
| **Retry + Backoff** | Transient failures | Retry storms |
| **Timeout** | Always | Too short/long = problems |

## Quick Picks
| Scenario | Go-To |
|----------|-------|
| Early startup | Monolith + Postgres |
| Read-heavy | Add Redis |
| Scale writes | Shard or NoSQL |
| Service calls | gRPC |
| Public API | REST |
| Decouple services | Message queue |
| Global users | CDN + replicas |
