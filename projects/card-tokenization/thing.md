# Build a Card-Tokenization Lab in Go

## What you’re building

Build a Go service that exchanges a synthetic card number for a token. Store the card in an encrypted SQLite vault, then use the token to make fake payments. Send requests with curl; everything runs locally and costs nothing.

The encryption code is provided. You’ll write the HTTP handlers and database code, using the request examples and tests in each step.

By the end, the merchant app will store tokens and masked card details. Only the vault and fake processors will need the full card number after collection.

### How this maps to Basis Theory and Spreedly

Basis Theory and Spreedly both store cards and let merchants use them through tokens. The comparison below uses their documentation as of October 23, 2026.

| Capability | Basis Theory documentation | Spreedly documentation | What you build |
| --- | --- | --- | --- |
| Store sensitive data behind a reference | Tokens, encrypted data, masked views | Payment-method tokens and masked card metadata | Random reference, encrypted synthetic PAN, masked metadata |
| Restrict access | Applications, permissions, access rules, isolated tenants | Environment-scoped API access and secure tokenization mechanisms | Merchant isolation plus separate create/read/use/revoke permissions |
| Send cards to processors | Proxy resolves tokens and forwards data to a PSP | Built-in gateway integrations and Payment Method Distribution to receivers | Send credentials to two fixed local fake processors |
| Separate collection from long-term storage | Short-lived Token Intents can convert into persistent Tokens | Cached methods can be retained for future use | Temporary collection records that expire unless retained |
| Identify repeated card data | Fingerprints and configurable deduplication | Environment-scoped card fingerprints | Optional keyed, merchant-scoped fingerprint experiment |
| Manage the credential lifecycle | Update/delete/TTL and other token operations | Retention, redaction, account updater, vault migration | Revocation in the core; expiry/retention in the lifecycle exercise |
| Maintain network credentials | Separate network-token capabilities | Advanced Vault network tokens and account updater | Outside this project |

A vault decides who can collect, view, retain, and use a card. Encryption protects the stored data; authorization controls what callers can do with it.

Skip dashboards, billing, bulk migrations, 3DS, fraud systems, account updater, and network-token provisioning.

You’ll start with one active AES key. Basis Theory documents per-token key protection; this project uses a simpler key setup.

### Keep it local and use fake cards

- Use only the two allowlisted synthetic numbers below. The service must reject everything else.
- Never enter a real card, even locally. A format or Luhn check cannot tell you whether a card is synthetic.
- Collect no CVC, real names, billing addresses, or expiration dates in this lab.
- Bind services to `127.0.0.1`; do not deploy or expose them through a tunnel.
- Do not send synthetic credentials to any payment processor or network.
- Do not use this service for real payments or claim PCI compliance.

The allowlist prevents accidental real-card entry. Anyone changing the source can bypass it.

## 1. Draw the system before writing handlers

Use two Go processes and two SQLite databases:

```text
curl: synthetic fixtures only
  |
  | POST /cards (JSON + local API credential)
  | synthetic PAN in the request body
  v
Merchant app :8080
  | owns: wallet metadata and purchase requests
  | holds: merchant-scoped vault API credential
  | must not persist: PAN or vault encryption key
  |
  | POST /v1/tokens
  | Authorization: Bearer <merchant credential>
  v
Vault service :8081
  | owns: tokens, encrypted PANs, authorization, revocation
  | holds: encryption key and credential-to-merchant mapping
  |
  +--> vault.db: ciphertext + token metadata
  |
  +--> internal fake processor: deterministic approve/decline
       no network access and no actual money movement

Merchant app --> merchant.db: token + last4 + display label
```

PAN means primary account number: the card number. The merchant backend receives it here. With real cards, forwarding the number without storing it still brings the backend into PCI scope.

curl keeps browser code out of the project. Another design puts the vault first, so the merchant receives only a token. Basis Theory documents this as an inbound proxy. Compare those two request paths once yours works.

Start with a Go function that returns approval or decline. In step 13A, move it into a third process. You’ll then have to handle HTTP timeouts and crashes between services.

Both processes run under your laptop account. Separate processes do not protect them from someone who controls that account.

**Check:** mark which components see the full card number, encrypted data, or just a token.

## 2. Create the Go project

Use a currently supported Go release, SQLite through `database/sql`, and the pure-Go SQLite driver:

```bash
mkdir go-tokenization-lab
cd go-tokenization-lab
go mod init example.com/tokenlab
go get modernc.org/sqlite
mkdir -p cmd/merchant cmd/vault internal/vault internal/merchant
```

Commit `go.mod` and `go.sum`. HTTP, crypto, JSON, and testing come from the standard library; SQLite needs a driver.

Suggested layout:

```text
cmd/merchant/main.go         composition and HTTP server
cmd/vault/main.go            composition and HTTP server
internal/vault/crypto.go     encryption and token generation
internal/vault/crypto_test.go
internal/vault/store.go      SQL transactions
internal/vault/http.go       authentication and handlers
internal/merchant/http.go    JSON API and wallet
```

Ignore `.env`, local key files, `*.db`, `*.db-wal`, and `*.db-shm`. Never commit credentials or database snapshots.

Limit each SQLite connection pool to one connection and enable foreign keys on it. That keeps connection settings and writes easier to follow. Return errors when the database is locked or a query fails.

## 3. Accept only the two test cards

Start with a pure function in the vault that accepts exactly:

```text
4242424242424242 -> simulated approval
4000000000000002 -> simulated decline
```

Treat these as local fixtures, not real authorization behavior. Reject spaces, punctuation, empty strings, and every other value. Do not add general card validation yet.

Define JSON requests with only the fields needed for each operation. Use `http.MaxBytesReader` with a small limit such as 4 KiB, `json.Decoder.DisallowUnknownFields()`, and require EOF after the first JSON value. Require the intended content type.

Do not echo rejected input in a response or error. For example:

```json
{"error":"unsupported_synthetic_card"}
```

Return `400` for malformed JSON, `413` for an oversized request, `415` for an unsupported content type, and `422` for a well-formed but disallowed synthetic card.

**Try it:** submit `4111111111111111`. It passes a Luhn check but is not on your allowlist. Expect `422`.

## 4. Implement token generation and authenticated encryption

A token is a random reference, not a hash of a card and not the encrypted card:

```text
tok_lab_<32 random bytes encoded as base64url>
                     |
                     v
database lookup -> encrypted synthetic PAN
```

Do not derive the token from the PAN. Card numbers have structure and a constrained search space; hashing them does not turn them into strong secrets.

Put the following in `internal/vault/crypto.go`:

```go
package vault

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
)

type Cipher struct {
    aead cipher.AEAD
}

func NewCipher(key []byte) (*Cipher, error) {
    if len(key) != 32 {
        return nil, errors.New("vault key must be exactly 32 bytes")
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("create AES cipher: %w", err)
    }
    aead, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("create GCM: %w", err)
    }
    return &Cipher{aead: aead}, nil
}

func NewToken() (string, error) {
    raw := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, raw); err != nil {
        return "", fmt.Errorf("generate token: %w", err)
    }
    return "tok_lab_" + base64.RawURLEncoding.EncodeToString(raw), nil
}

func associatedData(merchantID, token, keyID string) ([]byte, error) {
    return json.Marshal([]string{"tokenlab:v1", merchantID, token, keyID})
}

func (c *Cipher) Encrypt(
    merchantID, token, keyID, pan string,
) (nonce, ciphertext []byte, err error) {
    aad, err := associatedData(merchantID, token, keyID)
    if err != nil {
        return nil, nil, fmt.Errorf("encode associated data: %w", err)
    }
    nonce = make([]byte, c.aead.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, nil, fmt.Errorf("generate nonce: %w", err)
    }
    ciphertext = c.aead.Seal(nil, nonce, []byte(pan), aad)
    return nonce, ciphertext, nil
}

func (c *Cipher) Decrypt(
    merchantID, token, keyID string, nonce, ciphertext []byte,
) (string, error) {
    if len(nonce) != c.aead.NonceSize() {
        return "", errors.New("invalid nonce size")
    }
    aad, err := associatedData(merchantID, token, keyID)
    if err != nil {
        return "", fmt.Errorf("encode associated data: %w", err)
    }
    plaintext, err := c.aead.Open(nil, nonce, ciphertext, aad)
    if err != nil {
        return "", fmt.Errorf("decrypt vault record: %w", err)
    }
    return string(plaintext), nil
}
```

AES-GCM encrypts the data and detects tampering. `Seal` includes the authentication tag in its output. Each encryption needs a fresh nonce; reusing one with the same key is unsafe. These random 96-bit nonces suit the lab’s low volume. A production design also needs per-key usage limits.

Associated data binds a ciphertext to its merchant, token, key version, and format version. It is authenticated but not encrypted. Swapping a ciphertext into a different row should fail decryption.

Write these tests before adding HTTP handlers:

| Test | Expected result |
| --- | --- |
| Encrypt then decrypt | Original synthetic PAN |
| Encrypt the same input twice | Different nonces and ciphertext |
| Change one ciphertext byte | Decryption fails |
| Change merchant, token, or key ID | Decryption fails |
| Use another encryption key | Decryption fails |
| Supply a malformed nonce | Error, not panic |
| Generate 1,000 tokens | Correct prefix, no observed duplicates |

The token test catches implementation mistakes; it does not prove randomness mathematically.

Run:

```bash
go test ./internal/vault
```

## 5. Separate the key from the database

Generate a local 32-byte key:

```bash
umask 077
openssl rand -base64 32 > vault-key-v1.b64
```

Load it in the vault process only. Trim surrounding whitespace, decode standard base64, and call `NewCipher`. Missing, malformed, or wrong-length keys must fail startup.

Use a keyring mapping `v1` to that cipher, with a separately configured active key ID. Persist the key ID on each encrypted row.

Do not generate a replacement key automatically when the file is missing. That would turn a recoverable configuration issue into inaccessible records.

**Experiment:** close the vault, copy its database to a separate fixture directory, and attempt decryption with another key. It must fail.

Someone with only the database should not be able to decrypt the cards. Someone who controls the vault process can. A local key file does not provide the access controls or auditing of a key management service (KMS) or hardware security module (HSM).

## 6. Define the database invariants

Create the vault schema:

```sql
CREATE TABLE tokens (
    token TEXT PRIMARY KEY,
    merchant_id TEXT NOT NULL,
    key_id TEXT NOT NULL,
    nonce BLOB NOT NULL CHECK(length(nonce) = 12),
    ciphertext BLOB NOT NULL CHECK(length(ciphertext) >= 16),
    last4 TEXT NOT NULL CHECK(length(last4) = 4),
    state TEXT NOT NULL CHECK(state IN ('active', 'revoked')),
    created_at INTEGER NOT NULL
);

CREATE INDEX tokens_merchant ON tokens(merchant_id);

CREATE TABLE tokenization_requests (
    merchant_id TEXT NOT NULL,
    request_id TEXT NOT NULL,
    token TEXT NOT NULL REFERENCES tokens(token),
    PRIMARY KEY (merchant_id, request_id),
    UNIQUE (token)
);

CREATE TABLE payments (
    merchant_id TEXT NOT NULL,
    payment_id TEXT NOT NULL,
    token TEXT NOT NULL REFERENCES tokens(token),
    amount_minor INTEGER NOT NULL CHECK(amount_minor > 0),
    currency TEXT NOT NULL CHECK(currency = 'usd'),
    outcome TEXT NOT NULL CHECK(outcome IN ('approved', 'declined')),
    created_at INTEGER NOT NULL,
    PRIMARY KEY (merchant_id, payment_id)
);
```

Never store PAN or request bodies in an idempotency table. A tokenization replay can compare the submitted synthetic PAN against the original token's decrypted value internally; do not return that value.

Create a smaller merchant database:

```sql
CREATE TABLE wallet (
    token TEXT PRIMARY KEY,
    last4 TEXT NOT NULL,
    label TEXT NOT NULL
);
```

Start with one merchant and one demo customer. Add the second merchant in the authorization tests. Multiple customers within a merchant would need another ownership check.

Write these invariants at the top of your store tests:

1. A merchant cannot read, charge, or revoke another merchant's token.
2. A token is active or revoked; revocation does not reverse.
3. One merchant/request ID produces at most one committed token.
4. One merchant/payment ID produces at most one committed simulated outcome.
5. Ciphertext cannot be moved to another token or merchant without failing authentication.
6. No HTTP response returns a PAN, ciphertext, nonce, key, or request body.

## 7. Authenticate the service caller

Generate separate random API credentials for `merchant_a` and `merchant_b`. Load a credential-to-principal mapping in the vault; give each merchant process only its own credentials. A principal contains a merchant ID and a set of allowed operations.

Define `token:create`, `token:read`, `token:use`, and `token:revoke`. For merchant A, create a collection credential with create permission, a metadata-only credential with read permission, and a backend credential with purchase and revocation permissions.

Check permission before accessing a resource. A valid credential lacking the operation returns `403`. Even a credential with use permission cannot access another merchant's token. No principal has a public "return the plaintext PAN" operation.

For every vault request:

```text
Authorization header
       |
       v
Validate configured credential
       |
       v
Attach merchant identity to request context
       |
       v
Query WHERE merchant_id = authenticated merchant AND token = supplied token
```

Use constant-time comparison for the small fixed set of same-length lab credentials, such as Go's `crypto/subtle.ConstantTimeCompare`. Never log the header or accept the merchant identity from JSON as authority.

Missing or invalid credentials return `401`. For a token that is absent or owned by another merchant, return the same `404` response. Owning the token string alone is insufficient authority.

Use HTTP only on loopback for this project. Real credentials and card numbers require protected transport and network access controls. Never reuse the local credentials elsewhere.

**Experiment:** copy a token from merchant A and call its metadata endpoint using merchant B's credential. You should receive `404`, not masked details.

**Second experiment:** use A's metadata-only credential to create a payment. Expect `403`, even though both the credential and token belong to A. Authentication, tenant isolation, and operation authorization are three different checks.

## 8. Implement tokenization with retry semantics

Define:

```text
POST /v1/tokens
Authorization: Bearer <credential>
Idempotency-Key: <request ID>
Content-Type: application/json

{"pan":"4242424242424242"}

201 Created
{"token":"tok_lab_...","last4":"4242","state":"active"}
```

Accept IDs of 1–100 ASCII letters, digits, hyphens, or underscores. Reject requests without an ID.

Implement the operation on a dedicated `*sql.Conn` using `BEGIN IMMEDIATE`. Perform every query and write through that same connection until `COMMIT`; roll back on any failure. Do not mix those statements with `db.Exec` calls that may use another connection. An alternative is a driver-supported immediate transaction configuration, but do not assume `database/sql.BeginTx` alone selects it.

Inside the transaction:

1. Look up `(authenticated merchant, request ID)`.
2. If it exists, load its token, verify the original PAN matches internally, and replay the original create response with `200`. Different input returns `409`.
3. If not, generate a token and encrypt the allowlisted PAN.
4. Insert the active token and request mapping together.
5. Commit before returning `201`.

For replay after revocation, return the original create response; it describes the earlier operation, not current usability. The metadata endpoint reports current state. Document this distinction so clients do not treat a replayed create response as a fresh authorization.

Set a bounded SQLite busy timeout. Return `503` if it expires. Retry with the same request ID and unchanged input.

**Experiment:** send ten concurrent requests with the same merchant, request ID, and PAN. After transient lock failures are retried, every successful response must contain the same token, and the database must have one matching token and one request mapping.

Serializing writes keeps these transactions simple, at the cost of throughput.

## 9. Drive the merchant API with curl

Implement `POST /cards` as a JSON endpoint in the merchant app:

```text
POST /cards
Authorization: Bearer <local merchant API credential>
Idempotency-Key: card_001
Content-Type: application/json

{"pan":"4242424242424242"}
```

Authenticate curl with a separate local merchant API credential. Do not expose vault service credentials in the response or reuse them as the merchant's inbound credentials.

Accept JSON only, reject unexpected Host headers and browser origins, do not enable CORS, and require an explicit Authorization header on mutations. There are no cookies or ambient browser credentials, so cookie-based CSRF protection is not part of this curl lab. Loopback binding alone is not sufficient access control.

Add `GET /healthz` to both services. Readiness should check that the expected schema is available; vault startup must already have validated the key configuration. Return status only, never configuration or secrets.

After implementing startup configuration, run each service in a separate terminal:

```bash
go run ./cmd/vault
```

```bash
go run ./cmd/merchant
```

In a third terminal, confirm readiness before submitting fixtures:

```bash
curl --fail-with-body http://127.0.0.1:8081/healthz
curl --fail-with-body http://127.0.0.1:8080/healthz
```

For local credentials, create an ignored, owner-readable curl config file containing an `Authorization` header and `Content-Type: application/json`. Keep it separate from the vault key file. The placeholder below must be replaced locally with the randomly generated merchant API credential:

```text
# merchant-a.curl.conf -- chmod 600; never commit
header = "Authorization: Bearer <LOCAL_MERCHANT_API_CREDENTIAL>"
header = "Content-Type: application/json"
```

Now send the synthetic fixture through stdin:

```bash
curl --fail-with-body \
  --config merchant-a.curl.conf \
  -H 'Idempotency-Key: card_001' \
  --data-binary @- \
  http://127.0.0.1:8080/cards <<'JSON'
{"pan":"4242424242424242"}
JSON
```

The fixture may still appear in shell history; that is acceptable only because it is fixed synthetic data. Never substitute a real PAN, and avoid `curl -v`, tracing, or shell tracing that could expose credentials.

In `POST /cards`:

1. Authenticate, limit the body, and strictly decode JSON.
2. Forward the synthetic PAN to the vault with the merchant's collection credential and caller's request ID.
3. Persist only the returned token and last four digits.
4. Return JSON containing the token and masked metadata. Implement `GET /cards` to list the local wallet under the authenticated demo identity.

If the vault succeeds but the merchant database write fails, return an error and preserve the request ID. A retry should recover the same token. If the caller never retries, the vault will have a record missing from the merchant wallet. Note how you would find and reconcile those records.

Use `Cache-Control: no-store` for API responses and keep PANs out of URLs. Do not use body-dumping middleware.

**Checkpoint:** curl -> merchant -> vault carries the synthetic PAN. The merchant database contains only the token and masked metadata. Retry the exact command with `card_001`; it must return the same token.

## 10. Use the token without returning the card

Add:

```text
POST /v1/payments
Idempotency-Key: purchase_001

{"token":"tok_lab_...","amount_minor":1000,"currency":"usd"}

200 OK
{"payment_id":"purchase_001","outcome":"approved"}
```

In an immediate transaction:

1. Look up the merchant/payment ID first.
2. For an existing payment, compare token, amount, and currency. Replay its saved outcome when identical; otherwise return `409`.
3. For a new payment, find an active token owned by the caller.
4. Decrypt internally and call a pure Go function implementing the two fixture outcomes.
5. Persist the outcome, commit, and return only the payment ID and outcome.

Return `409 token_revoked` for a new payment using an owned revoked token. A decline is a successfully processed simulation with an explicit `declined` outcome, not an infrastructure error.

Give the merchant app a `POST /purchases` action accepting a token and Idempotency-Key. It selects a fixed $10 amount itself and persists the purchase ID before contacting the vault so retries reuse it. Do not accept caller-supplied amount fields in this fixed-price merchant endpoint.

```bash
curl --fail-with-body \
  --config merchant-a.curl.conf \
  -H 'Idempotency-Key: purchase_001' \
  --data-binary '{"token":"<TOKEN_FROM_CARD_RESPONSE>"}' \
  http://127.0.0.1:8080/purchases
```

**Experiment:** simulate a decline, then retry with the same payment ID. You must see the same recorded decline, not another logical purchase.

The fake processor is a function with no external side effects, so it can run inside the transaction. An HTTP call to a real processor cannot: your database cannot roll back an external charge. Step 13A adds persisted delivery states and recovery for that case.

## 11. Revoke, then test the race

Implement:

```text
POST /v1/tokens/{token}/revoke
GET  /v1/tokens/{token}
```

Revocation verifies ownership and changes `active` to `revoked` transactionally. Repeating it returns the same revoked state. Metadata returns token, last four digits, and current state only.

Expose revocation through an authenticated `POST /cards/{token}/revoke` in the merchant app. Preserve the vault tombstone; remove the card from the merchant wallet after a successful vault response.

Define ordering precisely:

```text
Payment commits before revocation -> payment outcome remains valid.
Revocation commits before new payment -> payment rejected.
Replay of an older payment after revocation -> original outcome returned.
```

Both operations use immediate transactions, so one finishes before the other can write. Return success only after commit.

**Experiment:** race a new payment and revocation repeatedly using a fresh token each time. Each outcome must match one of the first two orderings above.

Revocation is not erasure. Ciphertext may remain in the row, database pages, WAL, snapshots, and backups. Secure deletion and retention require a separate lifecycle design.

## 12. Make failures observable without leaking inputs

Log structured allowlisted fields with `log/slog`, for example:

```text
operation=tokenize request_id=req_123 merchant_id=merchant_a
outcome=created duration_ms=4
```

Do not log PANs, complete tokens, credentials, crypto material, raw request/response bodies, or arbitrary error values that might embed those fields. Return stable error codes and a correlation ID. Classify errors before logging.

Configure HTTP server header/read/write/idle timeouts and client timeouts between the merchant and vault. A vault timeout is an unknown outcome until a retry with the same ID resolves it.

Use these fault-injection checks:

| Fault | Expected behavior | Recovery |
| --- | --- | --- |
| Kill vault after commit but before response | Merchant sees failure/timeout | Same ID replays committed result |
| Lock database until timeout | Explicit `503`; no success claim | Bounded retry with same ID |
| Remove key file before startup | Startup fails | Restore correct key, not a fresh one |
| Corrupt ciphertext in a disposable fixture | Operation fails closed, no PAN returned | Restore known-good fixture/backup |
| Supply wrong merchant credential | `401` or indistinguishable token `404`, as applicable | Correct caller configuration |
| Fail merchant write after vault success | Visible partial failure | Replay same tokenization ID |

Search your logs and merchant database dump for both synthetic PANs. Neither should appear. Check crash dumps and tracing separately if you enable them.

Go strings and HTTP parsing can retain copies in managed memory. Clearing a buffer does not guarantee every plaintext copy has been erased. Do not promise secure memory erasure in this lab.

## 13. Practice key rotation and recovery

Only start after the earlier tests pass:

1. Generate `v2`, retain `v1`, and configure both in the vault keyring.
2. Make `v2` active for new writes.
3. For each old row, decrypt using its stored key ID.
4. Re-encrypt with a fresh nonce and associated data containing `v2`.
5. Update ciphertext, nonce, and key ID atomically inside a transaction.
6. Verify every active record decrypts under its recorded key.
7. Restart and verify old and new tokens still work.

Stop if any record cannot decrypt. Do not overwrite it with a placeholder or silently skip the failure.

Do not remove `v1` just because current rows are migrated. Old backups may still require it. Losing the key loses access to those records; retaining it indefinitely has exposure costs. Recovery and retention policy determine when retirement is safe.

**Try it:** restore an old database snapshot into a disposable fixture with only `v2` available. Rows encrypted with `v1` should fail to decrypt.

## 13A. Send cards to a fake processor

Basis Theory’s Proxy and Spreedly’s gateway and receiver integrations can send a vaulted card to another service. Add that behavior now: the merchant chooses a destination, and the vault sends the card without returning it to the merchant.

Add `cmd/fakeprocessor`. It runs on `127.0.0.1:8082` with two routes representing independent destinations:

```text
Merchant has token, amount, destination ID
              |
              v
Vault authenticates and authorizes use
              |
              +--> destination processor_a -> /a/payments
              |
              +--> destination processor_b -> /b/payments
                     |
                     v
            receives synthetic PAN, not vault token
            returns processor reference + outcome, never PAN
```

Run both fake destinations in one process.

### Add the delivery endpoint

Add `POST /v1/deliveries`:

```json
{
  "token": "tok_lab_...",
  "destination": "processor_a",
  "amount_minor": 1000,
  "currency": "usd"
}
```

Require an Idempotency-Key and `token:use`. Authorize destination access as well as token ownership. A use credential may allow A but not B.

Map destination IDs to fixed loopback URLs, backend credentials, and request adapters in server configuration. Reject caller-supplied URLs, headers, templates, and HTTP methods. Disable redirects and environment-proxy use in this HTTP client. Otherwise, a caller could make the vault send card data to an endpoint they control.

The vault resolves the token and constructs a request containing the synthetic PAN only inside its authorized delivery path. The fake processor requires a destination-specific credential. It returns a small allowlisted response schema, for example:

```json
{"processor_reference":"fake_a_001","outcome":"approved"}
```

Strictly decode and validate responses. Never reflect raw upstream bodies, headers, or error strings back to the caller. A deliberately broken processor fixture that echoes a PAN must produce a sanitized error, not a leaked response.

**Experiment:** deliver the same vault token to A and B using separate delivery IDs. Both accept the resolved synthetic fixture; neither needs to understand `tok_lab_...`. The merchant never receives the PAN back.

The token works with both destinations because your vault knows how to call each one. Neither processor accepts the vault token directly.

### Replace the in-transaction simulation with a durable workflow

Do not make HTTP requests while holding the SQLite write transaction. Introduce a separate `deliveries` table keyed by `(merchant_id, delivery_id)` containing token, destination, amount, currency, state, downstream idempotency ID, and sanitized result.

Use these states:

```text
pending -> dispatching -> approved
                      -> declined
                      -> unknown -> reconcile -> approved / declined
                                             -> unknown if still unresolved
```

1. In a short transaction, validate ownership, active state, permissions, and request shape; create or replay the delivery record. A reused ID with different fields returns `409`.
2. A single local worker claims a pending row. Atomically check that the token is still active and change the row to `dispatching`. If revoked before claim, persist a terminal `rejected` result instead; include `rejected` in the table's allowed states.
3. After commit, decrypt and call the configured destination outside the database transaction.
4. Persist the validated outcome in another short transaction. Terminal replays return the saved outcome; in-flight replays return `202` with a delivery ID and status URL.
5. On a timeout or ambiguous result, record `unknown`. On restart, treat leftover `dispatching` records as potentially sent and reconcile them, not as fresh unsent requests.

Turn off the old `/v1/payments` route and have `/purchases` call `/v1/deliveries`. Keep the earlier simulation tests and add tests for this HTTP workflow. Persist the selected destination with each purchase; changing it on retry is a conflict.

Build a durable idempotency ledger in the fake processor too. Key it by the authenticated destination account and downstream operation ID; store a keyed request fingerprint and outcome, never PAN. The fingerprint must use a separate local secret. Same ID and same fields replay; different fields return `409`. Add an authenticated status lookup that returns outcomes only.

Pass the same downstream ID on all attempts. Reconciliation first checks status; if no operation is recorded, reissue only with that same ID to the same destination. The fake processor's unique constraint and transaction handle a race with an earlier attempt. Never fail over an unknown A attempt to B: A may already have approved.

This recovery works because the fake processor supports durable idempotency and status lookup. Check those guarantees before using the same approach with a real processor.

Define revocation at this new boundary:

```text
Revocation commits before worker claim -> delivery rejected.
Worker claim commits first -> delivery may complete after revocation.
Terminal delivery replay -> original result, even after revocation.
```

Revocation cannot recall a card number already sent. These rules replace the step 11 ordering once payments cross an HTTP boundary.

**Fault injection:** make A commit an approval but delay its response past the vault client timeout. Restart the vault and reconcile. There must be one fake processor operation and one eventual approved delivery, not two charges or a guessed decline.

## 13B. Expire cards that haven’t been saved

Basis Theory has short-lived Token Intents that can become persistent Tokens. Spreedly distinguishes cached and retained payment methods. Add a similar two-step flow: collect temporarily, then save only when requested.

Keep temporary collections separate from active/revoked tokens:

```text
POST /v1/collections -> temporary encrypted collection, five-minute lab TTL
POST /v1/collections/{id}/retain -> new persistent vault token

pending collection -> converted
                   -> expired
```

The five-minute TTL is a lab choice, not either provider's production default.

Use `collection:create` and `collection:retain` permissions. The collection-only credential cannot create persistent tokens; remove `token:create` from it when switching the merchant to this flow. Keep `/v1/tokens` limited to the privileged backend credential for direct-create exercises. This prevents the collection role from bypassing retention policy.

Bind collection ciphertext to its merchant, collection ID, and key version. During retention, generate a new token and re-encrypt using the persistent token's associated data. Do not copy ciphertext bound to the collection ID into a token row.

Atomically check expiry and pending state, create the persistent token, record the conversion mapping, and clear the temporary ciphertext. Retain requests have their own idempotency key; concurrent retries must resolve to one persistent token. A previously converted collection can return its original token mapping to an authorized caller. An expired, unconverted collection returns `410`.

Change `/cards` to collect first. Retain with the backend credential only when the request includes `save_for_future_use: true`, and record that choice locally. Real customer consent would also need terms explaining the intended use.

Enforce expiry on access even when cleanup has not run. Use a fake clock in tests, not five-minute sleeps. A cleanup job clears expired temporary ciphertext and retains only the minimal tombstone needed for replay/state checks. Clearing a SQL column does not prove physical erasure from WAL or backups.

**Experiments:** expired collection cannot be retained; a collection-only key cannot retain; concurrent retain calls create one token; skipped save consent creates no persistent wallet entry.

Do not accept collection IDs on payment endpoints. Basis Theory supports more operations on Token Intents; you don’t need those for this step.

## 13C. Optional: fingerprinting versus idempotency

Both providers expose a way to recognize repeated underlying card data. That solves a different problem from replaying one API request:

```text
Same request ID + same data -> same operation result (idempotency).
Different request IDs + same data -> potentially different tokens,
                                    same scoped fingerprint.
```

Implement a fingerprint using HMAC-SHA-256 with a separate secret and an unambiguous encoding of merchant ID and synthetic PAN. Return it only to authorized callers. Including the merchant ID prevents the same card from having a shared fingerprint across merchants. Neither provider’s internal fingerprint algorithm is assumed here.

Do not use a plain PAN hash, expose an unauthenticated fingerprint oracle, or treat a fingerprint as permission to use a card. Do not deduplicate automatically yet: lifecycle, ownership, consent, and deletion behavior make that a separate decision.

**Experiment:** two create requests with different IDs produce different tokens but the same fingerprint for A; the same fixture under B has a different fingerprint.

## 14. What real cards would require

| Area | This lab implements | Production adds |
| --- | --- | --- |
| Card collection | Allowlisted synthetic input through curl and Go JSON APIs | PCI-scoped collection environment, plus applicable browser controls if a payment page is added |
| Transport | Loopback HTTP only | TLS, service identity, network segmentation and access controls |
| Storage | AES-256-GCM and key IDs | Governed key lifecycle, managed key protection, usage limits and recovery controls |
| Authorization | Static merchant identities, operation scopes, destination authorization | Credential rotation, customer ownership, policy governance, operational access controls |
| Reliability | SQLite transactions and persisted replay | Capacity planning, HA, durable external-operation workflows and reconciliation |
| Audit | Structured local events | Protected audit trails, monitoring, incident response and access review |
| Data lifecycle | Revocation, temporary collections, retention and expiry | Governed backups, deletion policy, recovery and legal requirements |
| Payments | Internal simulator, then durable delivery to two local fake destinations | Approved processor connectivity, real idempotency contracts, authorization, clearing, settlement and disputes |
| Compliance | No real cardholder data | Scope and validation determined with the relevant acquirer/payment brands and qualified guidance |

Use the PCI DSS documentation and qualified guidance to assess a real deployment; this table is not a compliance checklist.

For real PANs, both storage and transmission matter. A merchant backend that briefly forwards PANs can still be in scope. Sensitive authentication data such as CVC must not be retained after authorization, even encrypted; this lab omits it entirely.

## 15. Know where network tokens begin

Your lab token:

```text
tok_lab_... -> your vault lookup -> synthetic PAN
```

A network token is a network-recognized replacement credential with provisioning, domain restrictions, lifecycle rules, and applicable cryptograms. Your random reference does not become usable on Visa or Mastercard because it maps to a card number.

Network-token provisioning is outside this project. It requires provider access, network-specific rules, and cryptograms that a local vault cannot issue.

## Completion checklist

You have completed the lesson when:

- Only the two synthetic fixtures are accepted.
- Merchant storage and logs contain no PANs.
- Database-only access without the vault key cannot decrypt records.
- Modified ciphertext and mismatched associated data fail authentication.
- Another merchant cannot inspect, use, or revoke a token.
- Concurrent retries create one logical operation.
- Revocation/payment races follow the documented ordering.
- Database errors return failures; missing keys prevent startup.
- Rotation and recovery behave as documented.
- A token works at both configured fake destinations without PAN disclosure to the merchant.
- Unauthorized destinations and caller-supplied URLs are rejected.
- An ambiguous downstream outcome is reconciled without cross-destination failover or duplicate operations.
- Temporary collection and persistent retention are distinct and expiry is enforced on access.
- Metadata-only and collection-only credentials cannot perform privileged operations.
- You can explain why this is vault tokenization, not network tokenization or proof of PCI compliance.

Run the suite after each milestone, then finish with:

```bash
go test ./...
go test -race ./...
go vet ./...
```

The race detector finds Go memory races, not SQL-level logical races. Keep the concurrent database tests.

## References

Provider research, reviewed October 23, 2026:

- [Basis Theory: token storage, masking, fingerprints and lifecycle](https://developers.basistheory.com/docs/concepts/what-are-tokens)
- [Basis Theory: applications, tenants and permissions](https://developers.basistheory.com/docs/concepts/access-controls)
- [Basis Theory: Token Intents](https://developers.basistheory.com/docs/concepts/token-intents)
- [Basis Theory: outbound proxy and charging a card](https://developers.basistheory.com/docs/card-payments/charge-card)
- [Basis Theory: inbound API tokenization](https://developers.basistheory.com/docs/card-payments/receive-cards-api)
- [Spreedly: payment-method use and retention](https://developer.spreedly.com/docs/using-payment-methods)
- [Spreedly: delivery to receivers](https://developer.spreedly.com/docs/payment-method-distribution)
- [Spreedly: environment-scoped fingerprints](https://developer.spreedly.com/docs/credit-card-fingerprints)
- [Spreedly: direct API collection and secure tokenization requirements](https://developer.spreedly.com/docs/direct-api)
- [Spreedly: network tokenization](https://developer.spreedly.com/docs/network-tokenization)

Integration caveat: Spreedly's direct API documentation currently announces deprecation of API tokenization without its secure tokenization parameters on October 31, 2026, with end of life on December 31, 2028. Our bearer-authenticated local API is not a blueprint for its production authentication contract.

No provider calls or subscriptions are required for this lesson. Basis Theory TEST tenants are documented as synthetic-data-only, not PCI-compliant environments. Spreedly's [testing guide](https://developer.spreedly.com/docs/testing) says test calls are metered and billed for customers; do not equate test data with free usage.

Implementation references:

- [Go cryptographic randomness](https://pkg.go.dev/crypto/rand)
- [Go AES and authenticated encryption](https://pkg.go.dev/crypto/cipher)
- [Go SQL transactions](https://go.dev/doc/database/execute-transactions)
- [SQLite transaction behavior](https://www.sqlite.org/lang_transaction.html)
- [PCI SSC: CVC storage restrictions](https://www.pcisecuritystandards.org/faqs/1280/)
- [PCI DSS resources](https://www.pcisecuritystandards.org/standards/pci-dss/)
- [Stripe's explanation of network tokens](https://stripe.com/guides/understanding-benefits-of-network-tokens)
