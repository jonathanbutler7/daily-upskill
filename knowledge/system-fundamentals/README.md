# System Fundamentals

A knowledge base for understanding Weave's production infrastructure, designed to prepare you for system design interviews through real-world examples.

## How to Use This Section

1. **Pick a topic** you want to understand (e.g., "How does monitoring work?")
2. **Review the question prompts** in that file
3. **Ask teammates** and fill in the answers
4. **Update diagrams** as you understand the architecture
5. **Reference during interviews** - you'll have real production examples!

## Structure

```
system-fundamentals/
├── README.md                    # You are here
├── QUESTIONS_TO_ASK.md          # Running list of questions for teammates
├── architecture/                # System design & service interactions
│   ├── overview.md              # High-level Weave architecture
│   ├── service-mesh.md          # How services communicate
│   ├── data-flow.md             # Data pipelines & flow
│   └── tradeoffs.md             # Architecture decisions & tradeoffs
├── infrastructure/              # K8s, deployment, networking
│   ├── kubernetes.md            # K8s setup, clusters, namespaces
│   ├── deployment-pipeline.md   # CI/CD, how services get deployed
│   ├── load-balancing.md        # Load balancers, ingress, traffic routing
│   └── networking.md            # VPCs, subnets, service discovery
├── observability/               # Monitoring, logging, alerting, tracing
│   ├── monitoring.md            # Metrics, dashboards
│   ├── alerting.md              # Alert rules, on-call, incident response
│   ├── logging.md               # Log aggregation, querying
│   └── tracing.md               # Distributed tracing
└── reliability/                 # Scaling, resilience, DR, security
    ├── scaling.md               # Horizontal/vertical scaling strategies
    ├── resilience.md            # Circuit breakers, retries, timeouts
    ├── disaster-recovery.md     # Backup, failover, DR procedures
    └── security.md              # Auth, secrets management, compliance
```

## Progress Tracker

| Category | Topic | Status |
|----------|-------|--------|
| Architecture | Overview | ⬜ Not Started |
| Architecture | Service Mesh | ⬜ Not Started |
| Architecture | Data Flow | ⬜ Not Started |
| Architecture | Tradeoffs | ⬜ Not Started |
| Infrastructure | Kubernetes | ⬜ Not Started |
| Infrastructure | Deployment Pipeline | ⬜ Not Started |
| Infrastructure | Load Balancing | ⬜ Not Started |
| Infrastructure | Networking | ⬜ Not Started |
| Observability | Monitoring | ⬜ Not Started |
| Observability | Alerting | ⬜ Not Started |
| Observability | Logging | ⬜ Not Started |
| Observability | Tracing | ⬜ Not Started |
| Reliability | Scaling | ⬜ Not Started |
| Reliability | Resilience | ⬜ Not Started |
| Reliability | Disaster Recovery | ⬜ Not Started |
| Reliability | Security | ⬜ Not Started |

**Status Legend:**
- ⬜ Not Started
- 🟡 In Progress (have some notes)
- ✅ Complete (diagram + notes filled in)

## Tips for Learning

- **Start with Architecture Overview** - Get the big picture first
- **Follow the data** - Understanding data flow helps connect all the pieces
- **Ask "why"** - Don't just learn what exists, understand the tradeoffs
- **Draw it yourself** - Recreating diagrams from memory reinforces learning
- **Connect to interviews** - For each topic, think "How would I explain this in an interview?"
