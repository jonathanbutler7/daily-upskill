# Deployment Pipeline

How code gets from PR merge to production at Weave.

## CI/CD Pipeline Diagram

---

## Questions to Ask

- [ ] Walk me through a deployment from PR merge to production
- [ ] What CI/CD tool do we use? (GitHub Actions, Jenkins, ArgoCD, etc.)
- [ ] How long does a typical deployment take?
- [ ] What's our rollback procedure?
- [ ] How do we handle database migrations during deployments?
- [ ] What triggers a deployment? (manual, auto on merge, scheduled?)
- [ ] Do we use GitOps? How?

---

## Key Concepts (System Design Interview)

### Deployment Strategies
| Strategy | Description | Pros | Cons |
|----------|-------------|------|------|
| Rolling | Gradual replacement | Zero downtime | Slow rollback |
| Blue-Green | Two identical environments | Instant rollback | Double resources |
| Canary | Small % of traffic first | Risk mitigation | Complex routing |
| Recreate | Kill all, deploy new | Simple | Downtime |
| **Weave uses**: | [Fill in] | | |

### GitOps Principles
- Git as single source of truth
- Declarative infrastructure
- Automated reconciliation
- **Does Weave use GitOps?**: [Fill in]

### CI/CD Best Practices
- Fast feedback loops
- Automated testing at every stage
- Immutable artifacts
- Environment parity
- **Weave's practices**: [Fill in]

---

## Weave Implementation

### Pipeline Overview
| Stage | Tool | Duration | What Happens |
|-------|------|----------|--------------|
| Build | [Fill in] | [Fill in] | [Fill in] |
| Test | [Fill in] | [Fill in] | [Fill in] |
| Deploy | [Fill in] | [Fill in] | [Fill in] |

### Deployment Flow
1. **PR Merged**: [What triggers?]
2. **CI Runs**: [What tests?]
3. **Image Built**: [Where stored?]
4. **Staging Deploy**: [How?]
5. **Production Deploy**: [How?]

### Rollback Procedure
1. [Step 1]
2. [Step 2]
3. [Step 3]

### Database Migrations
- **When do they run?**: [Fill in]
- **How are they versioned?**: [Fill in]
- **Rollback strategy**: [Fill in]

### Environment Promotion
| Environment | Purpose | Deploy Trigger | Approval Required? |
|-------------|---------|----------------|-------------------|
| Dev | [Purpose] | [Trigger] | [Yes/No] |
| Staging | [Purpose] | [Trigger] | [Yes/No] |
| Production | [Purpose] | [Trigger] | [Yes/No] |

---

## Resources

- CI/CD Documentation: [Link]
- Deployment Runbook: [Link]
- Pipeline Configuration: [Link]

---

## Notes

*Add your notes here as you learn:*
