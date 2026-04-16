# Load Balancing

How traffic is distributed across services at Weave.

## Load Balancing Architecture

---

## Questions to Ask

- [ ] What load balancer do we use? (GCP LB, AWS ALB, nginx, envoy, etc.)
- [ ] How is traffic routed to services?
- [ ] How do we handle traffic spikes?
- [ ] What's our approach to canary deployments?
- [ ] Do we use any traffic splitting or A/B testing?
- [ ] How do health checks work?
- [ ] What's our SSL/TLS termination strategy?

---

## Key Concepts (System Design Interview)

### Load Balancing Algorithms
| Algorithm | Description | Use Case |
|-----------|-------------|----------|
| Round Robin | Rotate through servers | Equal capacity servers |
| Least Connections | Route to least busy | Variable request duration |
| IP Hash | Consistent routing by IP | Session affinity |
| Weighted | Proportional distribution | Different server capacities |
| **Weave uses**: | [Fill in] | |

### Layer 4 vs Layer 7
| Layer | Works On | Features | Example |
|-------|----------|----------|---------|
| L4 (Transport) | TCP/UDP | Fast, simple | Network LB |
| L7 (Application) | HTTP/HTTPS | Path routing, headers | ALB, nginx |
| **Weave uses**: | [Fill in] | | |

### Health Checks
- **Liveness**: Is the container running?
- **Readiness**: Is the container ready to serve traffic?
- **Startup**: Has the container started successfully?
- **Weave's health check strategy**: [Fill in]

### Traffic Management
- **Rate limiting**: Prevent abuse
- **Circuit breaking**: Fail fast on unhealthy backends
- **Retries**: Handle transient failures
- **Timeouts**: Prevent hanging requests

---

## Weave Implementation

### Load Balancer Stack
| Layer | Technology | Purpose |
|-------|------------|---------|
| Global LB | [GCP LB / AWS ALB / etc.] | [Purpose] |
| Ingress | [nginx / envoy / etc.] | [Purpose] |
| Service LB | [kube-proxy / etc.] | [Purpose] |

### Ingress Configuration
```yaml
# Example ingress configuration
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: [service-name]
spec:
  rules:
  - host: [hostname]
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: [service]
            port:
              number: [port]
```

### Health Check Configuration
```yaml
# Example health check
livenessProbe:
  httpGet:
    path: [path]
    port: [port]
  initialDelaySeconds: [seconds]
  periodSeconds: [seconds]
readinessProbe:
  httpGet:
    path: [path]
    port: [port]
  initialDelaySeconds: [seconds]
  periodSeconds: [seconds]
```

### Traffic Routing Rules
| Route | Destination | Conditions |
|-------|-------------|------------|
| [Path/Host] | [Service] | [Headers/etc.] |

---

## Resources

- Load Balancer Documentation: [Link]
- Ingress Configuration: [Link]
- Traffic Management Policies: [Link]

---

## Notes

*Add your notes here as you learn:*
