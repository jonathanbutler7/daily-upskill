# Money Movement

How funds move between customers, merchants, processors, banks, and networks.

## Core Concepts

### The Players

- **Cardholder** - Person paying with a card
- **Merchant** - Business accepting payment
- **Acquirer** - Bank that processes payments for the merchant
- **Issuer** - Bank that issued the cardholder's card
- **Card Network** - Visa, Mastercard, etc. Routes transactions between acquirer and issuer
- **Processor/Gateway** - Software layer that connects merchant to acquirer (Stripe, Adyen, etc.)
- **Payment Facilitator (PayFac)** - Aggregates merchants under one acquirer relationship

### The Flow

1. **Authorization** - "Can this card pay $X?" Issuer approves or declines. No money moves yet.
2. **Capture** - "Actually charge the $X." Merchant confirms they want the money.
3. **Clearing** - Networks batch transactions and calculate who owes whom.
4. **Settlement** - Actual money moves between banks.
5. **Funding** - Acquirer deposits funds into merchant's bank account.
6. **Payout** - If using a PayFac, they transfer funds to the sub-merchant.

### Timing

Authorization is instant. Settlement takes 1-3 business days. Funding depends on the merchant's agreement.

A customer sees "payment successful" at authorization. The merchant doesn't have the money until funding. This gap causes confusion and bugs.

## What Can Go Wrong

- **Auth expires** - Authorizations have a window (7-30 days). If you don't capture in time, the auth expires and you can't charge.
- **Partial capture** - You can capture less than authorized (e.g., item out of stock). You can't capture more.
- **Settlement mismatch** - Your records say $100, the settlement report says $98. Where's the $2? (Fees, chargebacks, adjustments.)
- **Funding delay** - Merchant expects money Tuesday, it arrives Thursday. Support tickets.

## How This Connects to My Projects

**payer-sync**: The VCC (virtual credit card) is a card authorization. The ERA (remittance advice) is the settlement report. Reconciling them means matching authorizations to settlements.

**ledger-db**: The ledger tracks internal state. Settlement reports are external state. Reconciliation finds mismatches.

## Questions to Ask About Real Systems

- What's the authorization window for our payment methods?
- How do we handle partial captures?
- What's the typical settlement delay?
- How do we reconcile our records against settlement reports?
- What happens when settlement amounts don't match our expectations?

## Keywords

Authorization, capture, clearing, settlement, funding, payout, acquirer, issuer, card network, payment facilitator, PayFac, auth expiry, partial capture
