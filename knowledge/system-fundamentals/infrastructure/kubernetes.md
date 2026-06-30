# Kubernetes

How Weave uses Kubernetes for container orchestration and deployment.

## Cluster Architecture Diagram

---

## Questions to Ask

- [ ] How many K8s clusters do we run? What's the purpose of each?
- [ ] How are namespaces organized?
- [ ] What's our pod resource allocation strategy (requests/limits)?
- [ ] How do we handle K8s upgrades?
- [ ] What's our approach to node pools?
- [ ] Do we use any K8s operators?
- [ ] How do we manage K8s manifests? (Helm, Kustomize, raw YAML?)

---

## Key Concepts (System Design Interview)

### Core K8s Objects
| Object | Purpose | Example Use |
|--------|---------|-------------|
| Pod | Smallest deployable unit | Single container or sidecar pattern |
| Deployment | Manages ReplicaSets, rolling updates | Stateless services |
| StatefulSet | Ordered, persistent pods | Databases, caches |
| Service | Stable network endpoint | Load balancing to pods |
| Ingress | External HTTP routing | API gateway |
| ConfigMap | Configuration data | Environment variables |
| Secret | Sensitive data | API keys, passwords |

### Scaling Strategies
- **HPA (Horizontal Pod Autoscaler)**: Scale pods based on CPU/memory/custom metrics
- **VPA (Vertical Pod Autoscaler)**: Adjust resource requests/limits
- **Cluster Autoscaler**: Add/remove nodes based on demand
- **Weave's approach**: [Fill in]

### Resource Management
- **Requests**: Guaranteed resources (used for scheduling)
- **Limits**: Maximum resources (enforced at runtime)
- **QoS Classes**: Guaranteed, Burstable, BestEffort
- **Weave's defaults**: [Fill in]

---

## Weave Implementation

### Cluster Inventory
| Cluster | Purpose | Region | Node Count | K8s Version |
|---------|---------|--------|------------|-------------|
| [Cluster 1] | [Production/Dev/etc.] | [Region] | [Count] | [Version] |
| [Cluster 2] | [Purpose] | [Region] | [Count] | [Version] |

### Namespace Organization
| Namespace | Purpose | Services |
|-----------|---------|----------|
| [namespace-1] | [Purpose] | [Services] |
| [namespace-2] | [Purpose] | [Services] |

### Resource Defaults
```yaml
# Standard resource configuration
resources:
  requests:
    cpu: [Fill in]
    memory: [Fill in]
  limits:
    cpu: [Fill in]
    memory: [Fill in]
```

### Common kubectl Commands
```bash
# Commands you'll use frequently at Weave
kubectl get pods -n [namespace]
kubectl logs [pod-name] -n [namespace]
kubectl describe pod [pod-name] -n [namespace]
kubectl exec -it [pod-name] -n [namespace] -- /bin/sh
```

---

## Resources

- K8s Cluster Access: [Link]
- Namespace Documentation: [Link]
- Resource Guidelines: [Link]

---

## Notes

*Add your notes here as you learn:*
