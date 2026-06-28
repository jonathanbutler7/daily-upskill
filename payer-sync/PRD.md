# **Toy PayerSync PRD**

## **1\. Summary**

Build a toy version of **PayerSync**, a product that automates ingestion of healthcare remittance data, processes payer-issued virtual credit cards (VCCs), posts resulting payments into a practice management system (PMS), and notifies offices when processing is complete.

This product is intended to simulate the core operational loop of post-adjudication insurance payment handling:

1. ingest ERA packets and VCC payloads from SFTP  
2. reconcile the two data sources  
3. process the VCC through a payment processor  
4. write the financial result back to the PMS ledger  
5. notify the office  
6. retain raw source records for legal/compliance purposes

This is not meant to reproduce every edge of real-world RCM. It is a constrained, implementation-oriented clone focused on the core workflow and the system design tradeoffs.

---

## **2\. Problem**

Healthcare offices often receive insurance reimbursements in a fragmented way:

* remittance data arrives as **EDI X12 835 / ERA**  
* card funding details arrive separately as a **VCC CSV**  
* the two files may not arrive at the same time  
* one VCC payload may cover multiple claims  
* one office payment may need to be allocated across multiple claims or service lines  
* payment posting and office communication are often manual

The result is a slow, error-prone workflow with weak observability and poor reconciliation.

The toy PayerSync product should reduce manual work by turning these asynchronous files into a single, traceable processing pipeline.

---

## **3\. Goals**

### **Business goals**

* reduce office effort required to process payer-issued VCC insurance payments  
* automatically convert ERA \+ VCC files into posted PMS ledger entries  
* provide an auditable trail for every payment from ingestion through posting  
* support operational scaling with minimal manual intervention

### **Product goals**

* reliably ingest ERA and VCC payloads from SFTP  
* reconcile ERA payments to VCC funding instruments using a deterministic matching model  
* handle out-of-order arrival between ERA and VCC files  
* process matched VCCs using a Stripe-like payment processor abstraction  
* write back posted payments into the PMS with idempotency guarantees  
* notify offices of successful and failed processing  
* preserve raw source artifacts for a legally reasonable retention period

### **Non-goals for MVP**

* payer portal integrations  
* EFT / ACH payment handling  
* patient messaging workflows  
* appeals, denials management, or claim adjudication logic  
* full PMS read APIs for balance discovery or claim lookup  
* chargeback/dispute workflows  
* refunds or reversals beyond internal design notes  
* advanced UI workflows beyond a minimal internal/admin queue and status view

---

## **4\. Users**

### **Primary users**

**Office billing staff**

* want insurance payments processed and posted with minimal effort  
* need simple status visibility and email notifications

**Operations / support staff**

* need tools to investigate mismatches, retries, and exceptions  
* need to answer “where is this payment?” and “why did posting fail?”

### **Secondary users**

**Engineering / platform admins**

* need observability, replay, retention, and operational controls

---

## **5\. Key assumptions**

1. ERA packets are received as **EDI X12 835** files via **SFTP**.  
2. VCC payloads are received as **CSV files** via the same SFTP location.  
3. Production environments are already **PCI compliant**; the product may process raw PAN/CVV in the PCI environment.  
4. We have access to a **dummy processor** with Stripe-like semantics.  
5. We have access to a **PMS writeback system** with simple ledger posting methods.  
6. Offices are notified by **direct SMTP email**.  
7. Raw ERA and VCC source files must be retained for **7 years**.  
8. ERA and VCC files may arrive **out of order**.  
9. The primary linking concept between ERA and VCC payloads is the **trace number**:  
   * ERA: `TRN`  
   * VCC CSV: `trace_id`

---

## **6\. Source formats**

### **ERA packet**

Assumed structure:

* ISA / GS / ST headers  
* BPR → payment info (amount, method)  
* TRN → trace number  
* N1 → payer/provider info  
* CLP → claim-level payment  
* SVC → service line details  
* CAS → adjustments  
* SE / GE / IEA trailers

### **VCC payload CSV**

Assumed columns:

* `payment_id`  
* `trace_id`  
* `payer_name`  
* `provider_npi`  
* `provider_tax_id`  
* `issue_date`  
* `amount`  
* `card_number`  
* `expiration_date`  
* `cvv`  
* `patient_name`  
* `patient_id`  
* `claim_id`  
* `service_date_start`  
* `service_date_end`

---

## **7\. Product definition**

### **What PayerSync does**

For each office/location, PayerSync will:

1. monitor SFTP for new ERA and VCC files  
2. ingest and persist raw files  
3. parse files into normalized records  
4. build payment candidates from each source  
5. match ERA payment candidates to VCC payment candidates  
6. process the card for matched VCCs  
7. allocate the processed amount to claims using ERA data  
8. write the results into the PMS ledger  
9. notify the office  
10. archive raw files and retain searchable metadata

### **MVP product surface**

#### **Internal/admin queue**

A simple internal queue or status view showing:

* office/location  
* trace number  
* payer  
* BPR amount  
* VCC amount  
* match status  
* processor status  
* PMS writeback status  
* notification status  
* exception reason, if any

#### **Office notification**

Email notifications only in MVP.

---

## **8\. Core concepts and data model**

### **8.1 ERA Remittance**

A raw 835 file plus its parsed representation.

Key fields:

* `era_id`  
* `office_id`  
* `payer_name`  
* `provider_npi`  
* `provider_tax_id`  
* `bpr_amount`  
* `payment_method`  
* `trace_number`  
* `received_at`  
* `file_hash`  
* `raw_storage_key`

### **8.2 ERA Payment Group**

The payment-level unit extracted from the ERA that will be matched to a VCC.

For MVP, assume one payment group per ERA transaction set.

Key fields:

* `era_payment_group_id`  
* `trace_number`  
* `bpr_amount`  
* `claim_count`  
* `claims[]`  
* `adjustments[]`

### **8.3 VCC File**

A raw CSV file plus metadata.

Key fields:

* `vcc_file_id`  
* `office_id`  
* `received_at`  
* `file_hash`  
* `raw_storage_key`  
* `row_count`  
* `source_filename`

### **8.4 VCC Row**

Each CSV record.

Key fields:

* `vcc_row_id`  
* `payment_id`  
* `trace_id`  
* `payer_name`  
* `provider_npi`  
* `provider_tax_id`  
* `issue_date`  
* `amount`  
* `card_fingerprint`  
* `last4`  
* `expiration_date`  
* `patient_id`  
* `claim_id`  
* `service_date_start`  
* `service_date_end`  
* `source_vcc_file_id`

### **8.5 VCC Payment Group**

The logical funding instrument to be run once through the processor.

For MVP, all VCC rows with the same:

* `trace_id`  
* `payment_id`  
* `provider_npi`  
* `provider_tax_id`  
* `card_fingerprint`

are grouped into a single VCC payment group.

The **group amount** is the sum of member row amounts.

### **8.6 Reconciled Payment**

A joined ERA \+ VCC object ready for payment processing.

### **8.7 Ledger Posting**

A normalized record representing the PMS writeback attempt and result.

---

## **9\. Matching and reconciliation rules**

This is the most important part of the product.

### **9.1 Primary matching rule**

An ERA payment group matches a VCC payment group when:

1. `ERA.TRN == VCC.trace_id`  
2. provider identity is consistent  
3. payment amount is consistent

### **9.2 Provider identity consistency**

At least one of the following must match, and none may conflict:

* `provider_npi`  
* `provider_tax_id`

If both are present and one conflicts, the match is invalid.

### **9.3 Amount consistency**

* `ERA.BPR.amount` must equal `sum(VCC row amounts for the VCC payment group)`  
* no tolerance in MVP; exact decimal match required

### **9.4 Payer normalization**

Payer names may vary slightly by formatting. Matching should use:

* exact normalized string match when available  
* payer name is a secondary confidence signal, not the primary key

Normalization examples:

* trim whitespace  
* collapse repeated spaces  
* uppercase  
* remove common punctuation

### **9.5 Claim-level consistency**

When possible, claim-level signals should be checked but not required for a match:

* VCC `claim_id` exists in ERA `CLP` set  
* service date windows overlap

If claim-level data conflicts heavily, send to exception queue.

### **9.6 Out-of-order arrival**

Files may arrive in either order.

#### **ERA arrives first**

* create ERA payment group in `AWAITING_VCC`  
* store parsed metadata and raw file  
* do not process payment  
* retry matching whenever new VCC files arrive

#### **VCC arrives first**

* create VCC payment group in `AWAITING_ERA`  
* do not process payment  
* retry matching whenever new ERAs arrive

### **9.7 Waiting window**

A payment group may remain unmatched for up to **5 business days**.

After that:

* state becomes `EXCEPTION_UNMATCHED`  
* operations alert is generated  
* office is not emailed automatically

### **9.8 Reprocessing policy**

Every newly ingested ERA or VCC file triggers a reconciliation pass against unmatched candidates for the same office.

### **9.9 Duplicate file handling**

Files are deduplicated using:

* office/location  
* source filename  
* file hash

If a file is byte-for-byte identical to a previously ingested file, treat it as a duplicate and do not reprocess it.

### **9.10 Duplicate row handling**

Rows are deduplicated using a natural key of:

* `payment_id`  
* `trace_id`  
* `claim_id`  
* `amount`  
* `provider_npi`  
* `provider_tax_id`  
* `issue_date`

### **9.11 Corrections and updates to VCC payloads**

Raw files are immutable after ingest. We never edit raw packets in place.

If a later VCC CSV arrives for a previously seen `trace_id`, treat it as one of three cases:

#### **Case A: exact duplicate**

Same natural keys and same card fingerprint.

* mark duplicate  
* ignore for downstream processing

#### **Case B: safe correction before processing**

Same `trace_id`, prior VCC payment group is not yet processed, and new file changes non-funding metadata only.  
Examples:

* patient name formatting  
* service date corrections  
* row ordering changes

Behavior:

* new version supersedes prior parsed representation  
* raw file still retained  
* latest version used for matching/writeback

#### **Case C: material correction or conflicting funding data**

Same `trace_id`, but card details, amount, or provider identity differ.

Behavior:

* move to exception queue  
* do not auto-process  
* require manual resolution

### **9.12 Constraints on VCC updates**

To keep the toy system deterministic:

* a single `trace_id` should normally resolve to one VCC payment group  
* if a payer sends multiple files for the same `trace_id`, that is treated as abnormal unless it is an exact duplicate  
* the system should prefer safety over automation in conflict scenarios

---

## **10\. Payment processing behavior**

### **10.1 Processing unit**

For MVP, process **one card transaction per matched ERA/VCC payment group**.

That means:

* use the VCC once  
* charge the **total matched amount**  
* allocate internally across claims using the ERA data

This avoids charging once per claim row and better fits the fact that the ERA describes one remittance payment event.

### **10.2 Dummy processor model**

The dummy processor should behave like a Stripe-style API:

#### **Objects**

* `PaymentMethod`  
* `PaymentIntent`  
* `Charge`

#### **Required semantics**

* tokenization / payment method creation from raw PAN in PCI environment  
* create payment intent with amount and metadata  
* confirm payment intent  
* automatic capture in MVP  
* idempotency keys  
* stable processor IDs  
* decline codes / failure reasons  
* test responses for success, insufficient\_funds, expired\_card, invalid\_cvc, processor\_unavailable

#### **Suggested API shape**

* `create_payment_method(card_number, exp, cvv, billing_details)`  
* `create_payment_intent(amount, currency, payment_method_id, metadata, idempotency_key)`  
* `confirm_payment_intent(payment_intent_id)`  
* `retrieve_payment_intent(payment_intent_id)`

#### **Suggested statuses**

* `requires_payment_method`  
* `requires_confirmation`  
* `processing`  
* `succeeded`  
* `failed`  
* `canceled`

### **10.3 Processor metadata**

Every payment intent must include metadata:

* office/location id  
* trace number  
* era id  
* vcc payment group id  
* payer name

### **10.4 Idempotency**

Processor calls must use an idempotency key derived from:

* office\_id  
* trace\_number  
* matched\_amount  
* vcc\_payment\_group\_version

This prevents duplicate charges during retries.

### **10.5 Failure behavior**

If processor call fails:

* mark reconciled payment `PROCESSING_FAILED`  
* retry transient failures with backoff  
* do not write back to PMS  
* do not notify office of success  
* optionally notify office only after retry exhaustion, depending on configuration

### **10.6 PCI boundary**

Only the PCI environment may handle:

* raw PAN  
* CVV  
* full expiration date

The rest of the platform should use only:

* token / payment method id  
* last4  
* fingerprint  
* processor reference IDs

CVV must never be stored after processor use.

---

## **11\. PMS writeback behavior**

### **11.1 PMS integration model**

Assume a simple PMS writeback service with ledger-focused methods.

### **11.2 Required writeback methods**

#### **`post_insurance_payment(...)`**

Posts a payment against a patient/claim ledger entry.

Inputs:

* office\_id  
* patient\_id  
* claim\_id  
* amount  
* service\_date\_start  
* service\_date\_end  
* payer\_name  
* trace\_number  
* processor\_transaction\_id  
* external\_reference\_id  
* posted\_at

#### **`post_adjustment(...)`**

Posts contractual/patient responsibility adjustments derived from ERA CAS segments.

Inputs:

* office\_id  
* patient\_id  
* claim\_id  
* adjustment\_group\_code  
* adjustment\_reason\_code  
* amount  
* trace\_number  
* external\_reference\_id

#### **`post_unapplied_insurance_credit(...)`**

Fallback when the payment is valid but cannot be attributed at claim granularity.

Inputs:

* office\_id  
* amount  
* payer\_name  
* trace\_number  
* external\_reference\_id  
* reason

### **11.3 Writeback allocation strategy**

For MVP:

1. allocate payment using ERA claim-level `CLP` amounts  
2. if service-line data is needed, use `SVC`  
3. post adjustments from `CAS`  
4. if claim-level attribution is impossible, post unapplied insurance credit instead of dropping the payment

### **11.4 Writeback idempotency**

Every writeback call must include `external_reference_id`.

Suggested format:

* payment posting: `payersync:{office_id}:{trace_number}:{claim_id}:payment`  
* adjustment posting: `payersync:{office_id}:{trace_number}:{claim_id}:cas:{index}`

The PMS must treat identical `external_reference_id` values as idempotent.

### **11.5 Writeback retry policy**

Retry transient failures only.

Suggested policy:

* 3 retries  
* exponential backoff  
* move to `WRITEBACK_FAILED` after exhaustion

### **11.6 Ordering rule**

Do not send office success email until:

1. processor charge succeeded  
2. all intended PMS writes have either succeeded or been safely degraded to unapplied credit

---

## **12\. Notifications**

### **12.1 Delivery method**

Direct SMTP.

### **12.2 Notification types**

#### **Success email**

Send when payment was processed and written back successfully.

Contents:

* office/location name  
* payer name  
* trace number  
* total amount  
* number of claims posted  
* posting date/time  
* masked payment reference only

#### **Partial success email**

Send when payment processed successfully but some writebacks degraded to unapplied credit or adjustments failed.

#### **Failure email**

Optional for MVP, but recommended after retry exhaustion for:

* processor failure  
* unmatched timeout  
* writeback failure

### **12.3 PHI/HIPAA email constraint**

Email should avoid unnecessary PHI.

Do not include in email body by default:

* full patient names  
  n- full claim details  
* PAN or card details

Prefer summary data plus a secure reference/trace number.

---

## **13\. Storage and retention**

### **13.1 Retention policy**

Retain raw ERA and VCC source files for **7 years**.

### **13.2 Storage tiers**

#### **Hot storage**

* last 90 days  
* optimized for replay, debugging, and operations  
* fast object storage access

#### **Warm metadata store**

* normalized relational records for searchable operational history  
* indexed by office, trace number, payer, date, status

#### **Cold storage**

* raw files older than 90 days moved to lower-cost archival object storage  
* retrievable on demand for audit or support

### **13.3 What is retained**

* raw ERA files  
* raw VCC CSV files  
* parsed metadata  
* reconciliation records  
* processor response metadata  
* PMS writeback results  
* email delivery logs  
* audit log of state transitions

### **13.4 What is not retained**

* CVV after processor use  
* unnecessary full PAN copies outside of PCI storage boundary

### **13.5 Performance/scaling considerations**

To keep the system scalable:

* raw files live in object storage, not the transactional database  
* database stores only normalized/searchable metadata and references to raw blobs  
* large parsing and reconciliation work runs asynchronously through queues/jobs  
* office dashboards query normalized metadata only

---

## **14\. States and lifecycle**

### **Payment lifecycle states**

* `RECEIVED_RAW`  
* `PARSED`  
* `AWAITING_MATCH`  
* `MATCHED`  
* `PROCESSING_PAYMENT`  
* `PAYMENT_SUCCEEDED`  
* `PAYMENT_FAILED`  
* `WRITING_BACK`  
* `POSTED`  
* `PARTIALLY_POSTED`  
* `WRITEBACK_FAILED`  
* `NOTIFIED`  
* `EXCEPTION`  
* `ARCHIVED`

A full audit trail of state transitions is required.

---

## **15\. Operational requirements**

### **15.1 Observability**

Need metrics and logs for:

* files ingested by type  
* parse failures  
* unmatched ERA count  
* unmatched VCC count  
* average match latency  
* payment success rate  
* writeback success rate  
* email delivery rate  
* duplicate rate  
* exception queue depth

### **15.2 Admin controls**

Admin/internal users should be able to:

* replay file parsing  
* rerun reconciliation for a trace number  
* retry processor step when safe  
* retry PMS writeback when safe  
* mark exceptions resolved with an audit note

### **15.3 Alerting**

Alert on:

* SFTP ingest failures  
* parse error spikes  
* unmatched backlog older than SLA  
* processor outage rate spike  
* PMS writeback outage  
* SMTP failure rate spike

---

## **16\. Security and compliance constraints**

### **16.1 PCI**

* all raw PAN/CVV handling remains inside PCI-compliant production environments 
* only tokenized/masked artifacts leave the PCI boundary

### **16.2 HIPAA/minimum necessary**

Even though the core workflow touches claim/patient data, downstream notifications and logs should use minimum necessary data.

### **16.3 Encryption**

* data in transit encrypted  
* raw files encrypted at rest  
* archived files encrypted at rest  
* secrets stored in a managed secret store

### **16.4 Access control**

* office users can see only their own office data  
* support/admin access is audited  
* raw card data access should be highly restricted and ideally absent outside processor pathway

---

## **17\. Edge cases and exception handling**

1. **ERA without VCC for 5 business days**  
   * move to exception queue  
2. **VCC without ERA for 5 business days**  
   * move to exception queue  
3. **Same trace number, conflicting provider identity**  
   * exception  
4. **Same trace number, amount mismatch**  
   * exception  
5. **Processor success, PMS failure**  
   * payment is real; hold in writeback retry state and notify operations  
6. **PMS partial posting**  
   * use unapplied credit fallback where possible; mark partial success  
7. **Duplicate SFTP delivery**  
   * dedupe by file hash / natural keys  
8. **Conflicting VCC correction after payment already processed**  
   * exception; do not auto-reverse in MVP  
9. **Multiple ERAs with same trace number for same office**  
   * exception unless proven duplicate  
10. **SMTP failure after successful posting**  
* retry notification separately; do not affect financial state

---

## **18\. Success metrics**

### **Product KPIs**

* % of ERA/VCC payments auto-matched  
* % of matched payments auto-processed without manual intervention  
* % of processed payments successfully written back  
* median time from first file arrival to PMS posting  
* exception rate per 1,000 payments

### **Operational KPIs**

* duplicate ingest rate  
* parse failure rate  
* processor failure rate  
* writeback retry rate  
* unmatched aging backlog

Suggested MVP targets:

* 95%+ successful raw file ingest  
* 90%+ successful parse rate for in-contract files  
* 85%+ automatic match rate for in-contract data  
* 98%+ idempotent protection against duplicate processing

---

## **19\. Rollout plan**

### **Phase 1: ingestion and parsing**

* SFTP polling  
* raw storage  
* ERA parser  
* VCC CSV parser  
* metadata persistence

### **Phase 2: reconciliation and queueing**

* payment group formation  
* matching engine  
* unmatched handling  
* internal status view

### **Phase 3: payment processing**

* dummy processor integration  
* idempotent payment execution  
* processor result persistence

### **Phase 4: PMS writeback and notifications**

* ledger posting integration  
* adjustment posting  
* unapplied credit fallback  
* office email notifications

### **Phase 5: retention and operational hardening**

* hot/cold tiering  
* replay tools  
* alerting  
* audit exports

---

## **20\. Open questions**

1. Should the toy product support only VCC in MVP, or also model EFT payments that arrive in ERA but do not need processor execution?  
2. Should an office see a user-facing queue, or is email \+ internal admin tooling sufficient for the exercise?  
3. What exact PMS identifiers are available at writeback time if claim attribution is incomplete?  
4. Should payer-specific normalization rules be configurable?  
5. Do we want to support a manual “force match” tool for operations in MVP?  
6. Should unmatched VCC/ERA windows be configurable per payer?  
7. If a VCC file includes multiple trace IDs, do we treat each as an independent payment group within the same file? Recommended answer: yes.

---

## **21\. Recommended MVP decisions**

To keep the exercise implementable, the MVP should commit to the following:

* VCC only, not EFT execution  
* one processor charge per matched trace number  
* exact amount match required  
* one trace number maps to one logical payment group  
* conflicts go to exception, not heuristics-heavy auto-resolution  
* claim allocation is driven by ERA, not by the VCC CSV  
* office communication is email-only  
* raw files retained for 7 years with 90-day hot storage

---

## **22\. Final MVP statement**

The toy PayerSync MVP is a back-office reconciliation and payment-posting product for payer-issued virtual cards. It ingests 835 ERAs and VCC CSVs from SFTP, matches them primarily via trace number, safely handles out-of-order delivery, processes the matched VCC once for the total remittance amount, writes the result into the PMS ledger with idempotency, notifies the office by email, and retains auditable raw source records for 7 years.

The core design principle is **safety over aggressive automation**: exact matches auto-process; ambiguous or conflicting records are retained, surfaced, and queued for resolution rather than guessed.

## **Glossary**

* **ACH**: Automated Clearing House  
* **API**: Application Programming Interface  
* **BPR**: Payment Info (ERA Segment)  
* **CAS**: Adjustments (ERA Segment)  
* **CLP**: Claim-Level Payment (ERA Segment)  
* **CSV**: Comma Separated Values  
* **CVV**: Card Verification Value  
* **EDI X12 835**: Standard for Electronic Remittance Advice (ERA)  
* **EFT**: Electronic Funds Transfer  
* **ERA**: Electronic Remittance Advice  
* **HIPAA**: Health Insurance Portability and Accountability Act  
* **KPI**: Key Performance Indicator  
* **MVP**: Minimum Viable Product  
* **NPI**: National Provider Identifier  
* **PAN**: Primary Account Number  
* **PCI**: Payment Card Industry  
* **PHI**: Protected Health Information  
* **PMS**: Practice Management System  
* **RCM**: Revenue Cycle Management  
* **SFTP**: Secure File Transfer Protocol  
* **SMTP**: Simple Mail Transfer Protocol  
* **SVC**: Service Line Details (ERA Segment)  
* **TRN**: Trace Number  
* **VCC**: Virtual Credit Card

