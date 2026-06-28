# Reconciler

## Overview 

At this stage we will look at files that have been ingestion, parsed, and normalized by the ingester. 

The happy path is to match ERA file to a VCC fie and move on to the processing phase.

## The problem

Once an insurer has approved a payment, the payer will generate 2 files. 

1. The ERA packet explains what the payer is paying for. Think of it as a check that may cover several bills at once.
2. The VCC (virtual credit card) is the funding instrument that the payer issues for the office will use to collect money.

## Job

This job is kicked off after the ingester job finishes. It must run sequentially because if the ingester and reconciler jobs are running simultaneously, the reconciler may check for data that exists, the ingester just hasn't finished processing yet.

It has 3 phases

1. Check unmatched records for matches in the db
   1. Mark newly ingested unmatched records as `AWAITING_ERA` OR `AWAITING_VCC`
   2. Mark matched records by updating `AWAITING_*`  to `MATCHED`
2. Hand off matched records to the payment processor to process the VCC
3. Expire records that go unmatched for >5 days, update to `EXCEPTION_UNMATCHED` send alert to operations

### Questions
1. Can we assume that every ERA and VCC file should have a match?
   1. Let's assume that the percentage of unmatched records will be low
2. After a record is either matched or expired, do we need to hold on to it?
   1. Implication is that the table could become bloated with useless records
3. What if a match arrives after the system has determined that it is unmatched?
   1. If we move processed records into cold storage, we would need to provide a utility to match records in cold storage

Cold storage is an optimization that will not be prioritized for the MVP iteration.

## Match

What constitutes a match?

The primary linking concept between ERA and VCC payloads is the **trace number**:  
- ERA: `TRN`  
- VCC: `trace_id`
- Amount

## DB Queries

To avoid full table scan for each attempt to find a match, use indexes on the trace number for ERA and VCC.

@todo add query sample here when data models are defined

## Concurrency

The handoff of matched records and expiring unmatched records can be run concurrently, if an optimization is necessary in the future.

@todo what if the length of the cron runs longer than the interval at which you poll?

## Handoff

Once a record is matched, the VCC can be handed off to the payment processor

## Idempotency

An idempotency key is not needed. In the event of a match job failing, the reconciler can just run the entire job again. The match job is necessarily filtered to unmatched records.

## Open question

What if a user wants to run an ad hoc job at any stage (ingest, reconcile, etc.)? Do we allow that, which stages do we allow/not allow?