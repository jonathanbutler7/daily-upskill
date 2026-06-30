# Risk & Compliance

Enough to ask better questions, not enough to be a compliance expert.

## Core Concepts

### Fraud

Fraud is someone using the payment system in a way it wasn't intended, usually to steal money.

**Card fraud**: Stolen card numbers, account takeover, friendly fraud (customer disputes legitimate charge)

**Merchant fraud**: Fake business, money laundering, bust-out schemes (build trust then disappear with funds)

**Internal fraud**: Employee abuses access to steal or manipulate transactions

### Fraud Prevention

**Velocity checks**: How many transactions from this card/IP/device in the last hour? Unusual spikes trigger review.

**Identity verification**: Does the billing address match? Does the CVV match? Is the device fingerprint known?

**Risk scoring**: ML model assigns a risk score. High scores get declined or sent to manual review.

**3D Secure**: Shifts liability to the issuer. Customer authenticates with their bank before the charge completes.

### Chargebacks

Customer disputes a charge with their bank. Bank reverses the payment and takes the money back from the merchant.

**Chargeback reasons**:
- Fraud (card was stolen)
- Not as described (product didn't match)
- Not received (never got the item)
- Duplicate charge
- Subscription not canceled

**Chargeback process**:
1. Customer disputes with issuer
2. Issuer sends chargeback to acquirer
3. Acquirer debits merchant
4. Merchant can submit evidence to fight it
5. Issuer decides: merchant wins or customer wins

High chargeback rates (>1%) can get you kicked off payment networks.

### KYC / KYB

**Know Your Customer (KYC)**: Verify the identity of individuals. Name, address, date of birth, government ID.

**Know Your Business (KYB)**: Verify the identity of businesses. Business registration, beneficial owners, business address.

Required by law for financial services. Prevents money laundering and terrorist financing.

### AML

**Anti-Money Laundering (AML)**: Detect and report suspicious activity that might indicate money laundering.

**Suspicious Activity Reports (SARs)**: If you see something weird (large cash deposits, structuring to avoid reporting thresholds), you file a SAR with regulators.

### PCI DSS

**Payment Card Industry Data Security Standard**: Rules for handling card data.

**Key requirements**:
- Don't store CVV
- Encrypt card numbers at rest and in transit
- Restrict access to cardholder data
- Regular security audits

**Scope reduction**: Use a payment processor (Stripe, Adyen) to handle card data. They're PCI compliant. You never touch raw card numbers.

### Data Retention

How long do you keep payment data? Depends on:

- **Legal requirements**: Tax records, audit trails
- **Chargeback windows**: Need data to fight disputes (typically 120 days)
- **Privacy regulations**: GDPR, CCPA require deletion on request

Balance: Keep enough to defend yourself, delete enough to reduce risk.

## What Can Go Wrong

- **Fraud losses** - Chargebacks exceed revenue
- **Compliance violations** - Fines, loss of payment processing ability
- **Data breach** - Card numbers stolen, PCI violation, massive liability
- **False positives** - Legitimate customers blocked, revenue lost
- **False negatives** - Fraud gets through, money lost

## How This Connects to My Projects

**will-they-pay**: The ML model predicts payment likelihood. Similar approach to fraud scoring. Same risks: bias, overconfidence, stale data.

**payer-sync**: Audit trail for every state transition. Supports compliance requirements for traceability.

## Questions to Ask About Real Systems

- What's our chargeback rate?
- How do we handle fraud disputes?
- What KYC/KYB do we perform on merchants?
- How do we reduce PCI scope?
- What's our data retention policy?
- Who reviews suspicious activity?

## Keywords

Fraud, chargeback, dispute, KYC, KYB, AML, PCI DSS, 3D Secure, velocity check, risk scoring, SAR, data retention, friendly fraud, account takeover
