# Orchestration

How do you coordinate when to kick off work?

Options considered

1. Polling DB for updates to stateful columns
   1. Drawbacks:
      1. Slow, requires running code when you might not need to
      2. Manual handling of retries, backoffs, failures
2. Message queue
   1. Drawbacks:
      1. Too much overhead and complexity for the requirements
      2. Main advantage is subsecond latency, whcih doesn't matter right now
3. CDC like debezium
   1. Drawbacks:
      1. Adds overhead and complexity, not worth it now
4. PostgreSQL listen/notify
   1. Drawbacks
      1. Not reliable for production 
      2. Potential for dropped messages
      3. No transactional guarantees
      4. Doesn't scale beyond a few replicas

Decision:

Postgresql polling.

Why? Main drawback is not an issue here. Provides transactional guarantees and also scales better than listen/notify