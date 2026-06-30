# Payments Table Definitions

Raw transcript pasted from `psql` (`\d` output), kept verbatim and wrapped for readability.

## payments.invoices

```text
Table "payments.invoices"
           Column           |            Type             | Collation | Nullable | Default
----------------------------+-----------------------------+-----------+----------+-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 id                         | uuid                        |           | not null |
 merchantid                 | uuid                        |           | not null |
 personid                   | uuid                        |           |          | '00000000-0000-0000-0000-000000000000'::uuid
 amount                     | integer                     |           | not null |
 createdat                  | timestamp with time zone    |           | not null |
 uniquelink                 | character varying(36)       |           | not null |
 hasattachment              | boolean                     |           | not null | false
 authattempts               | integer                     |           | not null | 0
 expiresat                  | timestamp with time zone    |           | not null |
 status                     | integer                     |           | not null |
 personfirstname            | character varying(100)      |           |          |
 personlastname             | character varying(100)      |           |          |
 personpmid                 | character varying(100)      |           |          |
 personemail                | character varying(100)      |           |          |
 personmobilephone          | character varying(100)      |           |          |
 personhomephone            | character varying(100)      |           |          |
 skipattachmentauth         | boolean                     |           | not null | false
 provider_name              | text                        |           |          |
 memo                       | text                        |           |          |
 providername               | text                        |           |          |
 personbirthdate            | timestamp without time zone |           |          |
 modifiedat                 | timestamp with time zone    |           |          | CURRENT_TIMESTAMP
 userid                     | text                        |           |          |
 username                   | text                        |           |          |
 ts_search                  | tsvector                    |           |          | generated always as (to_tsvector('english'::regconfig, (((((((((((((((id::text || ' '::text) || amount::text) || ' '::text) || COALESCE(personfirstname, ''::character varying)::text) || ' '::text) || COALESCE(personlastname, ''::character varying)::text) || ' '::text) || personid::text) || ' '::text) || COALESCE(personemail, ' '::character varying)::text) || ' '::text) || COALESCE(personmobilephone, ''::character varying)::text) || ' '::text) || COALESCE(personhomephone, ''::character varying)::text) || ' '::text) || COALESCE(username, (''::text || ' '::text) || COALESCE(userid, ''::text)))) stored
 appliedsurchargepercentage | integer                     |           | not null | 0
 appliedsurchargeamount     | integer                     |           | not null | 0
 desiredsurchargepercentage | integer                     |           | not null | 0
 surchargingenabled         | boolean                     |           | not null | false
 chart_number               | character varying(20)       |           |          |
 patient_identifier         | character varying(20)       |           |          |
 patientidentifier          | character varying(20)       |           |          |
Indexes:
    "invoices_pkey" PRIMARY KEY, btree (id)
    "idx_invoices_merchantid" btree (merchantid)
    "idx_invoices_merchantid_createdat" btree (merchantid, createdat DESC)
    "idx_invoices_pending_status" btree (status) WHERE status = 9
    "idx_invoices_personfirstname" btree (personfirstname)
    "idx_invoices_personid" btree (personid)
    "idx_invoices_personlastname" btree (personlastname)
    "idx_invoices_ts_search" gin (ts_search)
    "idx_invoices_username" btree (username)
    "invoices_createdat_idx" btree (createdat)
    "invoices_merchantid_id_idx" UNIQUE, btree (merchantid, id) REPLICA IDENTITY
    "unq_uniquelink" UNIQUE CONSTRAINT, btree (uniquelink)
Referenced by:
FERENCES payments.invoices(id) ON DELETE CASCADE
NCES payments.invoices(id) ON DELETE CASCADE
id) REFERENCES payments.invoices(id)
Triggers:
pdate_modifiedat()
```

## payments.payment_log

```text
payments=> \d payments.payment_log
                               Table "payments.payment_log"
         Column         |           Type           | Collation | Nullable |    Default
------------------------+--------------------------+-----------+----------+----------------
 id                     | uuid                     |           | not null |
 merchantid             | uuid                     |           |          |
 amount                 | integer                  |           |          |
 weavefee               | integer                  |           |          |
 receiptemail           | character varying(100)   |           |          |
 paymentstatus          | integer                  |           |          |
 statusreason           | text                     |           |          |
 confirmationcode       | character varying(100)   |           |          |
 submittedat            | timestamp with time zone |           |          |
 createdat              | timestamp with time zone |           |          |
 recordedat             | timestamp with time zone |           |          |
 processorid            | uuid                     |           |          |
 paymenttype            | integer                  |           | not null | 0
 pricingid              | character varying(40)    |           |          |
 pricingrate            | integer                  |           |          |
 pricingtransactioncost | integer                  |           |          |
 updatedat              | timestamp with time zone |           |          |
 origin                 | integer                  |           | not null | 0
 popupnotificationsent  | boolean                  |           | not null | false
 userid                 | text                     |           |          |
 expires_at             | timestamp with time zone |           |          |
 processortype          | text                     |           | not null | 'stripe'::text
Indexes:
    "payment_log_pkey" PRIMARY KEY, btree (id)
    "idx_payment_log_expires_at" btree (expires_at) WHERE expires_at IS NOT NULL
    "idx_payment_log_merchantid" btree (merchantid)
    "idx_payment_log_payment_origin" btree (origin)
    "idx_payment_log_payment_status" btree (paymentstatus)
payments=>
```


# payments.invoice_payments

handles linking between invoice record (singular) and payment log records (plural)

```
payments=> \d payments.invoice_payments
             Table "payments.invoice_payments"
      Column       | Type | Collation | Nullable | Default
-------------------+------+-----------+----------+---------
 invoiceid         | uuid |           | not null |
 paymentid         | uuid |           | not null |
 invoice_change_id | uuid |           |          |
Indexes:
    "invoice_payments_pkey" PRIMARY KEY, btree (invoiceid, paymentid)
    "idx_invoice_payments_paymentid" btree (paymentid)
    "invoice_payments_invoiceid_idx" btree (invoiceid)

payments=>
```
