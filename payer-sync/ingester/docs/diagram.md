# Ingester Flow

```mermaid
flowchart TD
	Cron["Cron polls remote server for ERA and VCC files"] --> Detect["Identify file type and extract metadata"]

	Detect --> PlainHash["Generate plaintext hash"]

	PlainHash --> AuthoritativeDedup{"Duplicate by plaintext hash?"}
	AuthoritativeDedup -- Yes --> RecordDupe["Log dupe"] --> EndDup["Stop: exact duplicate"] 
	AuthoritativeDedup -- No --> Raw["Store encrypted file and create file record in `RECEIVED_RAW`"]

	Raw --> Decrypt["Decrypt in memory"]
	Decrypt --> Valid{"Valid and parseable format?"}
	Valid -- No --> RecordFailure["Update file record to `EXCEPTION_PARSE_FAILED` and append audit transition"]
	Valid -- Yes --> Parse["Persist normalized records in DB, update file record to `PARSED`"]
	Parse --> Reconcile["Trigger reconciliation module"]
