# Ingester Flow

```mermaid
flowchart TD
	Cron["Cron polls remote server for ERA and VCC files"] --> Detect["Identify file type and extract metadata"]

	Detect --> Decrypt
	Decrypt --> Valid{"Valid and parseable format?"}
	Valid -- No --> EndInvalid["Stop: invalid file to exception queue"] --> RecordFailure["Record failures in Audit Log DB"] --> Store["Encrypt  and store in bucket"]
	Valid -- Yes --> PlainHash["Compute canonical plaintext hash"]

	PlainHash --> AuthoritativeDedup{"Duplicate by plaintext hash?"}
	AuthoritativeDedup -- Yes --> RecordDupe["Log dupe"] --> EndDup["Stop: exact duplicate"] 
	AuthoritativeDedup -- No --> Raw["Re-encrypt and store raw file in bucket"]

	Raw --> Parse["Parse and persist normalized records in DB"]
	Parse --> Reconcile["Trigger reconciliation module"]