# Deployments

PayerSync divides up into 5 modules

1. Ingester
2. Reconciler
3. Notifier
4. Processor
5. Writeback

## Questions

1. Should the modules be deployed as separate services or a single application?
2. Should there be a single replica, or multiple?

## Tradeoffs

### Monolith or microservices

Going with a monolith for now, it's only 5 services, and will allow the deployments to be easier. Can always break them up in the future if needed.

Main drawback of monolith architecture is that if you need to make a change, you need to redeploy the entire application. That is acceptable for this project

Is there a way around the issue that arises if one module breaks then the whole monolith breaks? (Exlixir Supervisors)

### Replicas

If there are no replicas, then deploying a new version would guarantee some downtime for the application. Maybe this is ok, since the application is basically one big cron job

### Deployments

Despite being one big cron job, there are state transitions and retries (especially for writebacks). Must ensure no in-flight work being done before shutting down the process and updating

### State

Modules are stateless and can be redeployed and replicated freely. State all exists in the DB