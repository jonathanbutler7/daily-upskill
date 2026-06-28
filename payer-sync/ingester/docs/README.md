# Ingester

## Overview

What does it do?
1. Poll remote server for new files (VCC and ERA)
2. Avoid duplicates by determining if PayerSync has seen them before
3. Decrypt and parse the raw files into normalized formats, store in DB
4. Reencrypt the raw files with PayerSync keys and store in bucket
5. Kick off the reconciler module

## Poll remote server

Open questions
- How often should we do this?
- Should we ingest batches or individual files?
- How many files can we expect at a time?
- What are assumptions we can make about how the server is organized?
- How do we connect to it?

## Avoid duplicate files

It's important that PayerSync's data is accurate. To that end, the handling of duplicates is sensitive.

Before decrypting a file from the remote server, we can't fully know if the file is a duplicate or not yet. We can get a preliminary indicator based on metadata like filename and timestamps, but a full byte-for-byte comparison is not possible until we can decrypt the file, and compare to a decrypted version within PayerSync.

Assuming the created at date (or similar timestamp) for each file is accurate, then we can get all files for a time range and run them through the ingester using 2 phases of dedupe

1. Pre-decrypt check based on file metadata including location info
2. Once the file is decrypted, generate a unique non-reversible fingerprint based on file values to store in the DB

When the ingester gets a file, it can check against the hashes in the DB to see if it already exists

## Encryption

We will ingest files from the remote server and decrypt based on the server's keys.  Once decrypted, we will need to reencrypt the files at rest using a key store service (KMS or vault) before storing in the bucket

## Failure

What happens if parsing a file fails? It's possible that after we ingest and decrypt a file, it fails. It would be a good idea to hold onto a copy of the decrypted file in memory until the parsing is successful. This would allow the Ingester to record the parse failure in an audit log, but also move forward into the decrypt/store step so that the ingester knows about the failure and is not required to refetch the file from the remote server and start all over again.