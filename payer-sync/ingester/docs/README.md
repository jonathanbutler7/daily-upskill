# Ingester

## Overview

What does it do?
1. Poll remote server for new files (VCC and ERA)
2. Avoid duplicates by determining if PayerSync has seen them before
3. Store a raw-received record and the original encrypted file
4. Decrypt and parse the raw files into normalized formats, then update the DB to `PARSED`
5. Kick off the reconciler module

## Invariants

- Will never store decrypted files at rest
- Only ingest files for a given date range
- Avoid storing duplicate files by using deterministic hashes

## Input Contract

### **VCC File**

A raw CSV file plus metadata.

Key fields:

- `vcc_file_id`  
- `office_id`  
- `received_at`  
- `file_hash`  
- `raw_storage_key`  
- `row_count`  
- `source_filename`
- `status`


### **ERA Remittance**

A raw 835 file plus its parsed representation.

Key fields:

- `era_id`  
- `office_id`  
- `payer_name`  
- `provider_npi`  
- `provider_tax_id`  
- `bpr_amount`  
- `payment_method`  
- `trace_number`  
- `received_at`  
- `file_hash`  
- `raw_storage_key`

## Deduping

It's important that PayerSync's data is accurate. To that end, the handling of duplicates is sensitive.

Before decrypting a file from the remote server, we can't fully know if the file is a duplicate or not yet. We can get a preliminary indicator based on metadata like filename and timestamps, but a full byte-for-byte comparison is not possible until we can decrypt the file, and compare to a decrypted version within PayerSync.

Assuming the created at date (or similar timestamp) for each file is accurate, then we can get all files for a time range and run them through the ingester using 2 phases of dedupe

1. Pre-decrypt check based on file metadata including location info
2. Once the file is decrypted, generate a unique non-reversible fingerprint based on file values to store in the DB

When the ingester gets a file, it can check against the hashes in the DB to see if it already exists

## Encryption

We ingest files from the remote server and keep the original encrypted payload at rest in a repo-local folder (`RAW_STORAGE_DIR`, default `raw-storage/`). This is a temporary local-dev implementation until we replace it with real object storage and PayerSync-managed encryption.

## Failure

It's possible that after we ingest and decrypt a file, the module breaks and processing is halted. It would be a good idea to hold onto a copy of the decrypted file in memory until the parsing is successful. This would allow the Ingester to record the parse failure in an audit log, but also move forward into the decrypt/store step so that the ingester knows about the failure and is not required to refetch the file from the remote server and start all over again.

## Lifecycle

- Raw file metadata is recorded first as `RECEIVED_RAW`
- Successful parse updates the file record to `PARSED`
- Parse failures update the file record to `EXCEPTION_PARSE_FAILED`
- ERA payment groups persist normalized `claims` and `adjustments` JSON used by reconciliation and later writeback

## Poll remote server

## Open questions
- How often should we poll the server this?
- Should we ingest batches or individual files?
- How many files can we expect at a time?
- What are assumptions we can make about how the server is organized?
- How do we connect to it?
