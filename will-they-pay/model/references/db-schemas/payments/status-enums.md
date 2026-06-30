# Status Enums

Enum values pulled out of schema to indicate the possible statuses and their semantic meanings for payments and invoice records.

```proto
enum PaymentOrigin {
  UNKNOWN = 0;
  LOCATION = 1; // manual card entry
  PAYMENT_PORTAL = 2; // text to pay
  TERMINAL = 3; // terminal
  LOCATION_PORTAL = 4; // 24/7 payment link
  MOBILE_TAP_TO_PAY = 5; // mobile tap to pay
  PAYMENT_PLAN = 6; // Payment plan scheduled payment
  ONLINE_SCHEDULING = 7; // online scheduling
}

enum InvoiceStatus {
  reserved 1, 2, 3;

  STATUS_UNKNOWN = 0;
  PAID = 4;
  PARTIALLY_PAID = 5;
  UNPAID = 6;
  SCHEDULED = 7;
  CANCELED = 8;
  INVOICE_STATUS_PENDING = 9; // PENDING is already defined at package level, so can't be used again
}

enum PaymentStatus {
    STATUS_UNKNOWN = 0;
    SUCCEEDED = 1;
    FAILED = 2;
    PENDING = 3;
}
```
