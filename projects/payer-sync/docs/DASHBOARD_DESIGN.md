# PayerSync FE Dashboard Design

## 1. Wireframe & Layout

### Overall Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          PayerSync Dashboard                               │
│  [Date Range Picker]                                                  [⟳ Refresh]
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         HERO: PAYMENT PIPELINE                             │
│                                                                             │
│  [FILES] → [MATCHING] → [PROCESSING] → [WRITEBACK] → [NOTIFIED]           │
│    45        38 (84%)      35 (92%)       35 (100%)      35 (100%)        │
│   +$127K    +$98K         +$98K          +$98K          +$98K             │
│                                                                             │
│  Real-time flow showing payments moving through each stage with counts      │
│  and cumulative amounts at each step.                                      │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────┬──────────────────────────┬──────────────────────┐
│  AUTO-MATCH RATE        │  AUTO-PROCESS RATE       │  WRITEBACK SUCCESS    │
│  87%                    │  92%                     │  98%                  │
│  +2% ↗ vs yesterday     │  +1% ↗ vs yesterday      │  → flat               │
│  38 of 44 matched       │  32 of 35 processed      │  34 of 35 posted      │
└──────────────────────────┴──────────────────────────┴──────────────────────┘

┌──────────────────────────┬──────────────────────────┬──────────────────────┐
│  MEDIAN TIME TO POST     │  EXCEPTION RATE          │  TOTAL PROCESSED      │
│  2h 14m                  │  0.9 per 1K              │  $1,274,500           │
│  -8m ↘ vs yesterday      │  ↑ +2 today             │  45 payments, 24h     │
│  measured from ERA       │  4 exceptions queued     │                       │
└──────────────────────────┴──────────────────────────┴──────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  LIVE EVENT LOG                                                             │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│  14:32  ✅  MATCHED       TRACE-12847 | UnitedHealth → $5,400             │
│  14:31  ✅  PROCESSING    TRACE-12847 | Processor approved                │
│  14:30  ✅  POSTED        TRACE-12847 | 3 claims + 1 adjustment           │
│  14:29  📧  NOTIFIED      Boston Medical Group                             │
│                                                                             │
│  14:28  ✅  MATCHED       TRACE-12846 | Cigna → $8,200                    │
│  14:26  ⏳  AWAITING_VCC   TRACE-12845 | ERA arrived from Aetna ($2,100)   │
│  14:24  ⏳  AWAITING_ERA   TRACE-12844 | VCC file received ($3,500)        │
│                                                                             │
│  [Load More]                                                               │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  SELECTED PAYMENT JOURNEY                                                  │
│  Trace: TRACE-12847 | Office: Boston Medical Group | Payer: UnitedHealth   │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                             │
│   [ERA ARRIVED] ──2min──> [MATCHED] ──1min──> [PROCESSING] ──2sec──> [✓]  │
│     14:23                 14:25              14:26                 14:26    │
│   Amount: $5,400       Confidence: 100%     Status: SUCCEEDED              │
│                                                                             │
│   [WRITEBACK] ──2min──> [POSTED] ──1min──> [OFFICE NOTIFIED] ──(sent)──> │
│     14:26              14:28              14:29                 ✓          │
│   3 claims posted, 1 adjustment posted                                     │
│                                                                             │
│   [View Details]  [View Raw Files]  [Audit Log]                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────┬──────────────────────────┬──────────────────────┐
│  PARSE HEALTH            │  PROCESSOR HEALTH        │  WRITEBACK HEALTH    │
│  ✅ 94% success          │  ✅ 100% success         │  ✅ 98% success      │
│  Failures: 3 of 50 ERAs  │  Failures: 0 of 35      │  Failures: 1 of 35   │
│  Error types:            │  Processor: Online       │  Error types:        │
│  - Invalid EDI (2)       │  Latency: 1.2s avg      │  - Network (1)       │
│  - Parsing (1)           │                          │                      │
└──────────────────────────┴──────────────────────────┴──────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│  EXCEPTION QUEUE                                                            │
│  ─────────────────────────────────────────────────────────────────────────  │
│  4 items | Sorted by: Age (oldest first)                                   │
│                                                                             │
│  1. AMOUNT_MISMATCH | TRACE-12840 | United → $2,100 | 2h ago              │
│     ERA: $2,000 | VCC: $2,100 | Assigned to: John Smith                   │
│                                                                             │
│  2. PROVIDER_CONFLICT | TRACE-12839 | Cigna → $5,400 | 1h 45m ago         │
│     NPI mismatch in VCC row. Assigned to: Sarah Chen                       │
│                                                                             │
│  3. PARSE_ERROR | ERA_FILE_1784 | Aetna | 50m ago                         │
│     EDI format error on line 15. Assigned to: Unassigned                   │
│                                                                             │
│  4. AWAITING_VCC | TRACE-12838 | UnitedHealth → $8,200 | 28m ago          │
│     Auto-escalates if unmatched after 5 business days. Expires: May 31     │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Database Queries

### Query 1: Dashboard Headline Metrics (Real-time)

```sql
-- Queries for the top 6 KPI cards
-- Aggregates over last 24 hours by default; parameterized for date range

WITH time_window AS (
  SELECT 
    COALESCE($1::timestamp, NOW() - INTERVAL '24 hours') as start_time,
    COALESCE($2::timestamp, NOW()) as end_time
)
SELECT
  -- Auto-match rate
  (
    SELECT 
      ROUND(100.0 * COUNT(CASE WHEN status = 'MATCHED' THEN 1 END) / 
        NULLIF(COUNT(*), 0), 2) as pct
    FROM reconciled_payments
    WHERE created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as auto_match_rate,
  
  -- Auto-process rate (matched payments that got processed without exception)
  (
    SELECT 
      ROUND(100.0 * COUNT(CASE WHEN status IN ('PAYMENT_SUCCEEDED', 'POSTING', 'POSTED', 'NOTIFIED') THEN 1 END) / 
        NULLIF(COUNT(CASE WHEN status = 'MATCHED' THEN 1 END), 0), 2) as pct
    FROM reconciled_payments
    WHERE created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as auto_process_rate,
  
  -- Writeback success rate
  (
    SELECT 
      ROUND(100.0 * COUNT(CASE WHEN status IN ('POSTED', 'NOTIFIED') THEN 1 END) / 
        NULLIF(COUNT(CASE WHEN status IN ('PAYMENT_SUCCEEDED', 'POSTING', 'POSTED', 'NOTIFIED') THEN 1 END), 0), 2) as pct
    FROM reconciled_payments
    WHERE created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as writeback_success_rate,
  
  -- Median time from ERA arrival to PMS posting (in minutes)
  (
    SELECT PERCENTILE_CONT(0.5) WITHIN GROUP (
      ORDER BY EXTRACT(EPOCH FROM (posted_at - era_received_at)) / 60
    )::int
    FROM reconciled_payments
    WHERE era_received_at IS NOT NULL
      AND posted_at IS NOT NULL
      AND created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as median_time_to_post_minutes,
  
  -- Exception rate per 1000 payments
  (
    SELECT 
      ROUND(1000.0 * COUNT(CASE WHEN status = 'EXCEPTION' THEN 1 END) / 
        NULLIF(COUNT(*), 0), 2) as rate
    FROM reconciled_payments
    WHERE created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as exception_rate_per_1k,
  
  -- Total amount processed
  (
    SELECT COALESCE(SUM(matched_amount), 0)
    FROM reconciled_payments
    WHERE status IN ('POSTED', 'NOTIFIED')
      AND created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as total_amount_processed_cents,
  
  -- Count of payments processed
  (
    SELECT COUNT(*)
    FROM reconciled_payments
    WHERE status IN ('POSTED', 'NOTIFIED')
      AND created_at >= (SELECT start_time FROM time_window)
      AND created_at <= (SELECT end_time FROM time_window)
  ) as payment_count;
```

### Query 2: Pipeline Flow Counts (for the hero section)

```sql
-- Shows counts and amounts at each stage of the pipeline
WITH time_window AS (
  SELECT 
    COALESCE($1::timestamp, NOW() - INTERVAL '24 hours') as start_time,
    COALESCE($2::timestamp, NOW()) as end_time
)
SELECT 
  'FILES_RECEIVED' as stage,
  COUNT(DISTINCT era_id) + COUNT(DISTINCT vcc_payment_group_id) as count,
  COALESCE(SUM(
    CASE WHEN status IN ('RECEIVED_RAW', 'PARSED', 'AWAITING_MATCH', 'MATCHED', 
                         'PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED', 'WRITING_BACK', 
                         'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED') 
         THEN matched_amount ELSE 0 END
  ), 0)::bigint as total_amount_cents
FROM reconciled_payments
WHERE created_at >= (SELECT start_time FROM time_window)
  AND created_at <= (SELECT end_time FROM time_window)

UNION ALL

SELECT 
  'MATCHED' as stage,
  COUNT(*) as count,
  COALESCE(SUM(CASE WHEN status IN ('MATCHED', 'PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED', 
                                      'WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED') 
                     THEN matched_amount ELSE 0 END), 0)::bigint as total_amount_cents
FROM reconciled_payments
WHERE status IN ('MATCHED', 'PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED', 
                 'WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED')
  AND created_at >= (SELECT start_time FROM time_window)
  AND created_at <= (SELECT end_time FROM time_window)

UNION ALL

SELECT 
  'PROCESSING' as stage,
  COUNT(*) as count,
  COALESCE(SUM(CASE WHEN status IN ('PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED', 
                                      'WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED') 
                     THEN matched_amount ELSE 0 END), 0)::bigint as total_amount_cents
FROM reconciled_payments
WHERE status IN ('PROCESSING_PAYMENT', 'PAYMENT_SUCCEEDED', 'WRITING_BACK', 
                 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED')
  AND created_at >= (SELECT start_time FROM time_window)
  AND created_at <= (SELECT end_time FROM time_window)

UNION ALL

SELECT 
  'WRITEBACK' as stage,
  COUNT(*) as count,
  COALESCE(SUM(CASE WHEN status IN ('WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED') 
                     THEN matched_amount ELSE 0 END), 0)::bigint as total_amount_cents
FROM reconciled_payments
WHERE status IN ('WRITING_BACK', 'POSTED', 'PARTIALLY_POSTED', 'NOTIFIED')
  AND created_at >= (SELECT start_time FROM time_window)
  AND created_at <= (SELECT end_time FROM time_window)

UNION ALL

SELECT 
  'NOTIFIED' as stage,
  COUNT(*) as count,
  COALESCE(SUM(CASE WHEN status = 'NOTIFIED' THEN matched_amount ELSE 0 END), 0)::bigint as total_amount_cents
FROM reconciled_payments
WHERE status = 'NOTIFIED'
  AND created_at >= (SELECT start_time FROM time_window)
  AND created_at <= (SELECT end_time FROM time_window)
ORDER BY CASE stage 
  WHEN 'FILES_RECEIVED' THEN 1
  WHEN 'MATCHED' THEN 2
  WHEN 'PROCESSING' THEN 3
  WHEN 'WRITEBACK' THEN 4
  WHEN 'NOTIFIED' THEN 5
END;
```

### Query 3: Live Event Log (paginated, ordered by recency)

```sql
-- Returns recent state transitions for the live event log
-- Combines multiple entity types into a single event stream
SELECT 
  ts,
  event_type,
  trace_number,
  office_name,
  payer_name,
  status_from,
  status_to,
  amount_cents,
  detail_message
FROM (
  -- Payment state transitions
  SELECT 
    audit_timestamp as ts,
    'PAYMENT_STATE_CHANGE' as event_type,
    trace_number,
    (SELECT name FROM offices WHERE id = office_id) as office_name,
    payer_name,
    old_status as status_from,
    new_status as status_to,
    matched_amount as amount_cents,
    new_status || ' - ' || COALESCE(status_reason, '') as detail_message
  FROM payment_audit_log
  WHERE audit_timestamp >= NOW() - INTERVAL '24 hours'
  
  UNION ALL
  
  -- Notifications sent
  SELECT 
    sent_at as ts,
    'NOTIFICATION_SENT' as event_type,
    trace_number,
    office_name,
    payer_name,
    NULL,
    'NOTIFIED',
    matched_amount,
    'Notification delivered to office'
  FROM notifications
  WHERE sent_at >= NOW() - INTERVAL '24 hours'
  
  UNION ALL
  
  -- Exceptions
  SELECT 
    created_at as ts,
    'EXCEPTION' as event_type,
    trace_number,
    (SELECT name FROM offices WHERE id = office_id) as office_name,
    payer_name,
    NULL,
    'EXCEPTION',
    matched_amount,
    exception_code || ': ' || exception_message
  FROM exception_queue
  WHERE created_at >= NOW() - INTERVAL '24 hours'
)
ORDER BY ts DESC
LIMIT $1 OFFSET $2;
```

### Query 4: Single Payment Journey (trace number drill-down)

```sql
-- Detailed timeline of a single payment from ERA arrival to office notification
SELECT 
  trace_number,
  office_name,
  payer_name,
  matched_amount,
  era_received_at,
  vcc_received_at,
  matched_at,
  payment_intent_created_at,
  payment_confirmed_at,
  payment_succeeded_at,
  writeback_started_at,
  posted_at,
  notified_at,
  EXTRACT(EPOCH FROM (matched_at - era_received_at)) / 60 as time_to_match_minutes,
  EXTRACT(EPOCH FROM (payment_succeeded_at - matched_at)) / 60 as time_to_process_minutes,
  EXTRACT(EPOCH FROM (posted_at - payment_succeeded_at)) / 60 as time_to_writeback_minutes,
  EXTRACT(EPOCH FROM (notified_at - posted_at)) / 60 as time_to_notify_minutes,
  EXTRACT(EPOCH FROM (notified_at - era_received_at)) / 60 as total_time_minutes,
  claims_posted_count,
  adjustments_posted_count,
  status as current_status,
  CASE 
    WHEN status = 'EXCEPTION' THEN (SELECT exception_message FROM exception_queue WHERE reconciled_payment_id = rp.id LIMIT 1)
    ELSE NULL 
  END as exception_message
FROM reconciled_payments rp
WHERE trace_number = $1
LIMIT 1;
```

### Query 5: Health Metrics (parse, processor, writeback success)

```sql
-- Three separate queries for health cards
-- PARSE HEALTH
SELECT 
  ROUND(100.0 * COUNT(CASE WHEN parse_status = 'SUCCESS' THEN 1 END) / NULLIF(COUNT(*), 0), 2) as success_rate,
  COUNT(CASE WHEN parse_status = 'SUCCESS' THEN 1 END) as successful_files,
  COUNT(CASE WHEN parse_status = 'FAILED' THEN 1 END) as failed_files,
  json_agg(DISTINCT parse_error_code) as error_types
FROM file_ingestions
WHERE ingested_at >= NOW() - INTERVAL '24 hours'
GROUP BY DATE(ingested_at)
ORDER BY DATE(ingested_at) DESC
LIMIT 1;

-- PROCESSOR HEALTH
SELECT 
  ROUND(100.0 * COUNT(CASE WHEN status = 'PAYMENT_SUCCEEDED' THEN 1 END) / 
    NULLIF(COUNT(CASE WHEN status IN ('PAYMENT_SUCCEEDED', 'PAYMENT_FAILED') THEN 1 END), 0), 2) as success_rate,
  COUNT(CASE WHEN status = 'PAYMENT_SUCCEEDED' THEN 1 END) as successful_charges,
  COUNT(CASE WHEN status = 'PAYMENT_FAILED' THEN 1 END) as failed_charges,
  AVG(processor_latency_ms)::int as avg_latency_ms,
  'ONLINE' as processor_status -- in real implementation, check processor API health
FROM reconciled_payments
WHERE payment_attempted_at >= NOW() - INTERVAL '24 hours';

-- WRITEBACK HEALTH
SELECT 
  ROUND(100.0 * COUNT(CASE WHEN status IN ('POSTED', 'PARTIALLY_POSTED') THEN 1 END) / 
    NULLIF(COUNT(CASE WHEN status IN ('POSTED', 'PARTIALLY_POSTED', 'WRITEBACK_FAILED') THEN 1 END), 0), 2) as success_rate,
  COUNT(CASE WHEN status IN ('POSTED', 'PARTIALLY_POSTED') THEN 1 END) as successful_writebacks,
  COUNT(CASE WHEN status = 'WRITEBACK_FAILED' THEN 1 END) as failed_writebacks,
  json_agg(DISTINCT writeback_error_code) as error_types
FROM reconciled_payments
WHERE writeback_started_at >= NOW() - INTERVAL '24 hours';
```

### Query 6: Exception Queue

```sql
-- Active exceptions, sorted by age
SELECT 
  eq.id,
  eq.trace_number,
  eq.office_id,
  (SELECT name FROM offices WHERE id = eq.office_id) as office_name,
  eq.payer_name,
  rp.matched_amount,
  eq.exception_code,
  eq.exception_message,
  eq.created_at,
  EXTRACT(EPOCH FROM (NOW() - eq.created_at)) / 3600 as age_hours,
  eq.assigned_to_user_id,
  (SELECT email FROM users WHERE id = eq.assigned_to_user_id) as assigned_to_email,
  eq.escalation_date,
  CASE 
    WHEN eq.exception_code = 'AWAITING_VCC' OR eq.exception_code = 'AWAITING_ERA'
    THEN eq.created_at + INTERVAL '5 days'
    ELSE NULL
  END as escalation_deadline
FROM exception_queue eq
LEFT JOIN reconciled_payments rp ON eq.reconciled_payment_id = rp.id
WHERE eq.resolved_at IS NULL
ORDER BY eq.created_at ASC
LIMIT $1 OFFSET $2;
```

---

## 3. Demo Data Seeding Strategy

### Philosophy

The goal is to create a **realistic, flowing scenario** where:
- Data arrives asynchronously (like in production)
- Out-of-order arrival is demonstrated
- Real-time updates feel natural and constant
- Success rates are high enough to be impressive (~85-90%)
- A few exceptions exist to show safety/handling
- Multiple payers and offices show diversity

### Data Generation Script Outline

```bash
# Run demo data seeding in phases to simulate real-time arrival

# Phase 1: Seed static data (offices, payers, processor configs)
./scripts/seed_offices.sh          # 5 offices
./scripts/seed_payers.sh            # 4 major payers

# Phase 2: Continuous generation loop
# This simulates files arriving throughout the time window
./scripts/demo_payment_flow.sh --duration 24h --speed 10x

# --speed 10x means a 24-hour period compresses to 2.4 hours of real time
# Every 1 minute of real time = 10 minutes of simulation time
```

### Phase 1: Static Seed Data

```sql
-- OFFICES
INSERT INTO offices (id, name, location_id, npi, tax_id, created_at) VALUES
('office_001', 'Boston Medical Group', 'boston', '1234567890', '12-3456789', NOW()),
('office_002', 'Seattle Care Partners', 'seattle', '0987654321', '98-7654321', NOW()),
('office_003', 'Austin Health Systems', 'austin', '5551234567', '55-1234567', NOW()),
('office_004', 'Denver Clinic Network', 'denver', '3035551234', '30-3551234', NOW()),
('office_005', 'Miami Physicians LLC', 'miami', '3055559876', '30-5559876', NOW());

-- PAYERS
INSERT INTO payers (id, name, normalized_name) VALUES
('payer_001', 'United Health Insurance', 'UNITED HEALTH INSURANCE'),
('payer_002', 'Cigna Corporation', 'CIGNA CORPORATION'),
('payer_003', 'Aetna Inc.', 'AETNA INC'),
('payer_004', 'Blue Cross Blue Shield', 'BLUE CROSS BLUE SHIELD');
```

### Phase 2: Continuous Payment Flow

The idea is to generate payments **continuously over the time window** with realistic timing:

```python
#!/usr/bin/env python3
"""
Demo data generation script.
Generates realistic ERA/VCC file arrival patterns with natural delays and exceptions.
"""

import random
import datetime
import psycopg2
from enum import Enum
import time
import sys

class FlowSpeed(Enum):
    REALTIME = 1
    ACCELERATED_10X = 10
    ACCELERATED_30X = 30

# Command-line: python3 scripts/seed_demo_data.py --duration 24h --speed 10x --offices 5

OFFICES = ['office_001', 'office_002', 'office_003', 'office_004', 'office_005']
PAYERS = ['payer_001', 'payer_002', 'payer_003', 'payer_004']

AMOUNTS = [1200, 2500, 5400, 8200, 3600, 4100, 6700, 2300, 9100, 7500]  # in dollars

def generate_trace_number():
    """Generate realistic trace number."""
    return f"TRACE-{random.randint(10000, 99999)}"

def generate_era_event(office_id, trace_number, amount_cents):
    """
    Simulate ERA file arrival at SFTP.
    Returns dict with arrival time and parsed data.
    """
    return {
        'type': 'ERA',
        'trace_number': trace_number,
        'office_id': office_id,
        'payer_name': random.choice(PAYERS),
        'amount_cents': amount_cents,
        'file_hash': f"sha256_{random.randint(100000, 999999)}",
        'raw_storage_key': f"s3://era-raw/{office_id}/{trace_number}.edi",
        'claim_count': random.randint(1, 5),
    }

def generate_vcc_event(office_id, trace_number, amount_cents):
    """
    Simulate VCC CSV file arrival.
    May arrive before or after ERA.
    """
    return {
        'type': 'VCC',
        'trace_number': trace_number,
        'office_id': office_id,
        'payer_name': random.choice(PAYERS),
        'amount_cents': amount_cents,
        'file_hash': f"sha256_{random.randint(100000, 999999)}",
        'raw_storage_key': f"s3://vcc-raw/{office_id}/{trace_number}.csv",
        'row_count': random.randint(1, 3),
        'last4': str(random.randint(1000, 9999)),
    }

def generate_exception_scenario(prob=0.08):
    """
    8% of payments encounter some exception.
    Types: amount mismatch, provider conflict, parse error, await timeout.
    """
    if random.random() > prob:
        return None
    
    exception_types = [
        {'code': 'AMOUNT_MISMATCH', 'message': 'ERA amount does not match VCC sum'},
        {'code': 'PROVIDER_CONFLICT', 'message': 'NPI mismatch between ERA and VCC'},
        {'code': 'PARSE_ERROR', 'message': 'EDI format validation failed'},
        {'code': 'AWAITING_VCC', 'message': 'ERA received; waiting for VCC (5-day window)'},
    ]
    return random.choice(exception_types)

def simulate_payment_flow(duration_hours=24, speed=FlowSpeed.ACCELERATED_10X):
    """
    Main simulation loop.
    Generates realistic payment flows across all offices and payers.
    """
    now = datetime.datetime.now()
    start_time = now - datetime.timedelta(hours=duration_hours)
    
    current_time = start_time
    payment_count = 0
    
    # Generate ~45 payments spread across 24 hours
    # That's roughly 1.875 payments per hour
    payments_per_hour = 1.875
    interval_between_payments_seconds = 3600 / payments_per_hour
    
    print(f"Generating demo payments for {duration_hours}h window (speed: {speed.value}x)")
    print(f"Start: {start_time}, End: {now}")
    print(f"Target: ~{int(duration_hours * payments_per_hour)} payments\n")
    
    while current_time < now:
        # Decide if a payment arrives in this interval
        if random.random() < (interval_between_payments_seconds / 3600):
            office_id = random.choice(OFFICES)
            amount_dollars = random.choice(AMOUNTS)
            amount_cents = amount_dollars * 100
            trace_number = generate_trace_number()
            
            # Decide payment flow type
            flow_type = random.choice(['era_first', 'vcc_first', 'simultaneous'])
            
            if flow_type == 'era_first':
                # ERA arrives first, VCC follows after 5-30 minutes
                era_event = generate_era_event(office_id, trace_number, amount_cents)
                print(f"[{current_time.strftime('%H:%M')}] ERA arrived: {trace_number} | {office_id} | ${amount_dollars}")
                
                # Insert ERA into DB
                # (database insertion code here)
                
                # VCC arrives 5-30 minutes later
                vcc_delay = random.randint(5, 30)
                vcc_time = current_time + datetime.timedelta(minutes=vcc_delay)
                
                # Check for exceptions
                exception = generate_exception_scenario()
                
                # If no exception, simulate matching and processing
                if not exception:
                    vcc_event = generate_vcc_event(office_id, trace_number, amount_cents)
                    print(f"  +{vcc_delay}min → VCC arrived, MATCHED")
                    
                    # Simulate processor taking 1-3 seconds
                    processor_time = random.uniform(0.5, 3)
                    print(f"  +{processor_time:.1f}s → Payment PROCESSED (success)")
                    
                    # Writeback takes 1-2 minutes
                    writebacks = random.randint(1, 5)
                    print(f"  +{random.randint(1, 2)}m → Writebacks complete ({writebacks} claims posted)")
                    
                    # Notification within 1 minute
                    print(f"  +<1m → Office notified by email")
                else:
                    print(f"  EXCEPTION: {exception['code']}")
            
            elif flow_type == 'vcc_first':
                # VCC arrives first, ERA follows
                vcc_event = generate_vcc_event(office_id, trace_number, amount_cents)
                print(f"[{current_time.strftime('%H:%M')}] VCC arrived: {trace_number} | {office_id} | ${amount_dollars}")
                
                era_delay = random.randint(2, 45)  # Could wait up to 5 days, but demo faster
                print(f"  Waiting for ERA... (arrives in {era_delay}m)")
            
            payment_count += 1
        
        # Advance time by ~30 seconds per iteration (compressed based on speed)
        current_time += datetime.timedelta(seconds=30 / speed.value)
    
    print(f"\n✓ Generated {payment_count} payments")
    return payment_count

if __name__ == '__main__':
    duration = 24  # hours
    speed = FlowSpeed.ACCELERATED_10X
    
    # Parse command-line args
    for arg in sys.argv[1:]:
        if arg.startswith('--duration'):
            duration = int(arg.split('=')[1].replace('h', ''))
        elif arg.startswith('--speed'):
            speed_val = int(arg.split('=')[1].replace('x', ''))
            speed = FlowSpeed(speed_val)
    
    simulate_payment_flow(duration_hours=duration, speed=speed)
```

### Phase 3: Exception Distribution

```
Out of ~45 payments:
- 38-40 auto-match (84-89%)
- 35-37 auto-process (92-94%)
- 34-35 successfully post (98-100%)
- 3-4 end up in exception queue:
  - 1 amount mismatch (demonstration of safety)
  - 1 provider conflict (detection of data issues)
  - 1 parse error (real-world edge case)
  - 1 awaiting match (demonstrates out-of-order handling)
```

### Phase 4: Real-time Update Simulation

Once data is seeded, use a **change notification system** to push live updates:

```sql
-- Create a trigger that notifies listeners of payment state changes
-- This enables WebSocket push to FE

CREATE OR REPLACE FUNCTION notify_payment_change()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify(
    'payment_state_change',
    json_build_object(
      'trace_number', NEW.trace_number,
      'old_status', OLD.status,
      'new_status', NEW.status,
      'timestamp', NOW()
    )::text
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reconciled_payment_state_change
AFTER UPDATE ON reconciled_payments
FOR EACH ROW
WHEN (OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION notify_payment_change();
```

### Running the Demo

```bash
# 1. Seed static data
psql $DATABASE_URL < scripts/seed_offices_payers.sql

# 2. Run continuous generation (takes ~1-2 hours to simulate 24h at 10x speed)
python3 scripts/seed_demo_data.py --duration 24h --speed 10x

# 3. Dashboard will pick up live updates via WebSocket listeners
# As DB changes, FE receives real-time push notifications
```

### Key Demo Moments to Highlight

1. **Out-of-order arrival** – Show VCC arriving, waiting (AWAITING_ERA), then ERA arriving and immediate match
2. **Real-time processing** – Watch a payment flow from RECEIVED → MATCHED → PROCESSING → POSTED → NOTIFIED in ~5 seconds
3. **Exception handling** – Demonstrate an amount mismatch being caught and routed to the exception queue (not lost, not forced through)
4. **High success rate** – By the end of 24h, show 87% matched, 92% processed, 98% posted
5. **Money flow** – Show the total amount processed climbing in real-time ($127K processed in a day)

---

## Tech Stack Recommendations

- **Backend Query Engine**: PostgreSQL with listen/notify for real-time updates
- **FE Framework**: React 18+ with WebSocket for live updates
- **Real-time Updates**: Socket.io or native WebSocket + server-sent events
- **Charting**: Recharts for the pipeline flow visualization
- **Styling**: Tailwind CSS for dashboard layout

