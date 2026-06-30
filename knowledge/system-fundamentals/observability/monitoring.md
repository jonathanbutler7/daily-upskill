# Monitoring

How Weave collects and visualizes system metrics.

## Monitoring Stack Diagram

---

## Questions to Ask

- [ ] What metrics do we collect from each service?
- [ ] What dashboards should I know about?
- [ ] How do we set SLOs/SLIs?
- [ ] What's our metrics retention policy?
- [ ] How do we handle metric cardinality issues?
- [ ] What tool do we use for metrics? (Prometheus, Datadog, etc.)
- [ ] How do I add new metrics to a service?

---

## Key Concepts (System Design Interview)

### The Four Golden Signals
| Signal | Description | Example Metric |
|--------|-------------|----------------|
| Latency | Time to serve a request | p50, p95, p99 response time |
| Traffic | Demand on the system | Requests per second |
| Errors | Rate of failed requests | 5xx error rate |
| Saturation | How "full" the system is | CPU, memory utilization |

### Metric Types
| Type | Description | Use Case |
|------|-------------|----------|
| Counter | Cumulative, only increases | Total requests, errors |
| Gauge | Point-in-time value | Temperature, queue size |
| Histogram | Distribution of values | Request latency buckets |
| Summary | Similar to histogram | Quantiles (p50, p99) |

### SLOs, SLIs, SLAs
- **SLI (Service Level Indicator)**: Metric that measures service quality
- **SLO (Service Level Objective)**: Target value for an SLI
- **SLA (Service Level Agreement)**: Contract with consequences
- **Weave's SLOs**: [Fill in]

### Push vs Pull
- **Pull (Prometheus)**: Scraper fetches metrics from endpoints
- **Push (StatsD, Datadog)**: Services push metrics to collector
- **Weave's approach**: [Fill in]

---

## Weave Implementation

### Metrics Stack
| Component | Technology | Purpose |
|-----------|------------|---------|
| Collection | [Prometheus / Datadog / etc.] | [Purpose] |
| Storage | [Prometheus / Thanos / etc.] | [Purpose] |
| Visualization | [Grafana / Datadog / etc.] | [Purpose] |

### Standard Service Metrics
Every service should expose:
- [ ] Request rate (RPS)
- [ ] Error rate (by status code)
- [ ] Latency (p50, p95, p99)
- [ ] [Add more as you learn]

### Key Dashboards
| Dashboard | Purpose | Link |
|-----------|---------|------|
| [Dashboard 1] | [What it shows] | [Link] |
| [Dashboard 2] | [What it shows] | [Link] |
| [Dashboard 3] | [What it shows] | [Link] |

### SLOs
| Service | SLI | SLO Target | Current |
|---------|-----|------------|---------|
| [Service 1] | [Metric] | [Target] | [Current] |
| [Service 2] | [Metric] | [Target] | [Current] |

### Adding Metrics to a Service
```go
// Example: How to add a metric in Go
// [Fill in with Weave's actual pattern]
```

### Useful Queries
```promql
# Example PromQL queries
# Request rate
rate(http_requests_total[5m])

# Error rate
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# P99 latency
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

---

## Resources

- Monitoring Documentation: [Link]
- Grafana Access: [Link]
- Metrics Guidelines: [Link]

---

## Notes

*Add your notes here as you learn:*
