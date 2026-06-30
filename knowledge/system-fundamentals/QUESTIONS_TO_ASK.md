# Questions to Ask Teammates

A running list of questions to ask during 1:1s, team meetings, or Slack conversations. Check them off as you get answers and note who helped you.

## How to Use This File

1. **Before a 1:1**: Pick 2-3 questions relevant to that person's expertise
2. **During conversations**: Add new questions that come up
3. **After getting answers**: Check off the question and update the relevant topic file
4. **Track who knows what**: Note the person who answered - they're your go-to for follow-ups

---

## Architecture Questions

### Overview & Big Picture
- [ ] Can you draw me the high-level architecture of Weave on a whiteboard?
- [ ] What are the core services and how do they interact?
- [ ] What's the request flow from a user action to database and back?
- [ ] How many services do we have in production?

### Service Communication
- [ ] How do services discover each other?
- [ ] What protocols do we use for inter-service communication (gRPC, REST, etc.)?
- [ ] Do we use synchronous or asynchronous communication? When each?
- [ ] What's our approach to API versioning?

### Architecture Decisions
- [ ] What are the biggest architectural tradeoffs we've made?
- [ ] What would you change about our architecture if you could start over?
- [ ] Why did we choose a microservices architecture vs monolith?
- [ ] What's our strategy for service boundaries?

---

## Infrastructure Questions

### Kubernetes
- [ ] How many K8s clusters do we run? What's the purpose of each?
- [ ] How are namespaces organized?
- [ ] What's our pod resource allocation strategy?
- [ ] How do we handle K8s upgrades?
- [ ] What's our approach to node pools?

### Deployment
- [ ] Walk me through a deployment from PR merge to production
- [ ] How long does a typical deployment take?
- [ ] What's our rollback procedure?
- [ ] How do we handle database migrations during deployments?
- [ ] What triggers a deployment? (manual, auto on merge, scheduled?)

### Load Balancing
- [ ] What load balancer do we use? (GCP LB, nginx, envoy, etc.)
- [ ] How is traffic routed to services?
- [ ] How do we handle traffic spikes?
- [ ] What's our approach to canary deployments?

### Networking
- [ ] How is our VPC structured?
- [ ] How do services in different clusters communicate?
- [ ] What's our DNS setup?
- [ ] How do we handle external API calls?

---

## Observability Questions

### Monitoring
- [x] During [incident X], we looked at these dashboards but they didn't point us to the root cause. Can you walk me through what the dashboards showed during that incident, and what we *should* have been looking at instead?
  - [ ] the dashboards didn't show the error beacuse it was a nil pointer dereference. a dashboard that would've helped us find the error is measeuring started payments against completed payments. this dashboard would've shown started payments as high, and completed payments suddenly dropping off
- [ ] I've noticed we get alerts for panic recoveries, but I haven't seen them visualized in these dashboards. Is there a way to see panic recovery trends over time, or do we only track them through alerts?
  - [ ] you can see these in google cloud console. if you filter out the useless logs (which is something i'm going to work on doin with a pr, so you don't have to do it manually), then that is a good place to se panic recoveries and nil pointer dereference errors.
- [ ] Can we look at the metrics during our last outage

- [ ] How do we handle metric cardinality issues?

### Alerting
- [ ] What triggers our critical alerts?
- [ ] What's the on-call rotation like?
- [ ] Walk me through a recent incident response
- [ ] How do we prevent alert fatigue?
- [ ] What's our escalation policy?

### Logging
- [ ] Where do logs go? How do I query them?
- [ ] What's our log retention policy?
- [ ] What should I log vs what's too verbose?
- [ ] How do we correlate logs across services?

### Tracing
- [ ] Do we use distributed tracing? What tool?
- [ ] How do I trace a request across services?
- [ ] What's the sampling rate for traces?
- [ ] How do we debug performance issues?

---

## Reliability Questions

### Scaling
- [ ] How do we scale services? (HPA, VPA, manual?)
- [ ] What's our approach to database scaling?
- [ ] How do we handle traffic spikes?
- [ ] What are our current scaling limits?

### Resilience
- [ ] Do we use circuit breakers? Where?
- [ ] What's our retry strategy?
- [ ] How do we handle downstream service failures?
- [ ] What happens if the database goes down?

### Disaster Recovery
- [ ] What's our backup strategy?
- [ ] How often do we test disaster recovery?
- [ ] What's our RTO/RPO?
- [ ] What's the procedure for a major outage?

### Security
- [ ] How do services authenticate with each other?
- [ ] How do we manage secrets?
- [ ] What's our approach to least privilege?
- [ ] How do we handle security vulnerabilities?

---

## Meta Questions

- [ ] What's the best way to learn our system quickly?
- [ ] What documentation should I read?
- [ ] What's the most common production issue?
- [ ] What do you wish you knew when you started?

---

## Answered Questions Log

| Question | Answer Summary | Answered By | Date |
|----------|---------------|-------------|------|
| *Example: How many K8s clusters?* | *3 clusters: dev, staging, prod* | *@teammate* | *2026-04-15* |
