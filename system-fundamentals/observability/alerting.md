# Alerting

How Weave detects and responds to system issues.

## Alerting Flow Diagram

---

## Questions to Ask

- [ ] What triggers our critical alerts?
- [ ] What's the on-call rotation like?
- [ ] Walk me through a recent incident response
- [ ] How do we prevent alert fatigue?
- [ ] What's our escalation policy?
- [ ] Where are runbooks stored?
- [ ] How do I add a new alert?

---

## Key Concepts (System Design Interview)

### Alert Severity Levels
| Level | Description | Response Time | Example |
|-------|-------------|---------------|---------|
| Critical/P1 | Service down, data loss | Immediate | Database unreachable |
| High/P2 | Degraded service | < 1 hour | High error rate |
| Medium/P3 | Potential issue | < 4 hours | Disk filling up |
| Low/P4 | Informational | Next business day | Certificate expiring |
| **Weave's levels**: | [Fill in] | | |

### Alert Design Principles
- **Actionable**: Every alert should have a clear action
- **Relevant**: Alert on symptoms, not causes
- **Timely**: Detect issues before users do
- **Deduplicated**: Group related alerts

### On-Call Best Practices
- Clear escalation paths
- Runbooks for common issues
- Post-incident reviews
- Alert fatigue prevention
- **Weave's practices**: [Fill in]

### Incident Management
- **Detection**: How we find out about issues
- **Triage**: Assess severity and impact
- **Mitigation**: Stop the bleeding
- **Resolution**: Fix the root cause
- **Post-mortem**: Learn and prevent recurrence

---

## Weave Implementation

### Alert Routing
| Alert Type | Destination | Escalation |
|------------|-------------|------------|
| Critical | [PagerDuty / etc.] | [Escalation path] |
| Warning | [Slack / etc.] | [Escalation path] |
| Info | [Slack / etc.] | [None] |

### On-Call Rotation
- **Schedule**: [Weekly / Bi-weekly / etc.]
- **Tool**: [PagerDuty / OpsGenie / etc.]
- **Handoff process**: [Fill in]

### Key Alerts to Know
| Alert | Meaning | First Response |
|-------|---------|----------------|
| [Alert 1] | [What it means] | [What to do] |
| [Alert 2] | [What it means] | [What to do] |
| [Alert 3] | [What it means] | [What to do] |

### Escalation Policy
1. **Level 1**: [Who / When]
2. **Level 2**: [Who / When]
3. **Level 3**: [Who / When]

### Runbook Location
- **Where**: [Confluence / GitHub / etc.]
- **How to find**: [Fill in]
- **How to update**: [Fill in]

### Creating a New Alert
```yaml
# Example alert rule
# [Fill in with Weave's actual format]
groups:
- name: example
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High error rate detected
      runbook: [link]
```

---

## Resources

- Alerting Documentation: [Link]
- On-Call Schedule: [Link]
- Runbooks: [Link]
- Incident Response Guide: [Link]

---

## Notes

*Add your notes here as you learn:*
