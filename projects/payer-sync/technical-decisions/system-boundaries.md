# System Boundaries

This diagram is meant to show the system boundaries for the payer sync project. 

## Internal Modules

The primary modules of the system. The DB is the source of truth and the orchestrator.

**Decision:** use DB as an orchestration mechanism instead of a message queue or event bus. 

The system is state a large state machine, so the state transitions will be the events we need to pay attention to

1. This project will be deployed as a single application
   1. Rationale: So the modules don't need to communicate using an event bus, they can all easily communicate with the DB
2. The DB will require polling for updates
   1. Rationale: we don't need the near real time speed of a message queue
3. Speed is not a non functional requirement. 
   1. Rationale: It is acceptable for ERA and VCC packets to be unmatched for up to 5 days 
   2. Receiving payments, especially when coming from insurance, speed is not necessary
4. With DB polling, state updates and transitions are atomic. 
   1. No need to handle state updates *and* publishing messages. 
   2. No need to handle the complexity of what happens when a DB transaction succeeds, but a message publish fails, for example
