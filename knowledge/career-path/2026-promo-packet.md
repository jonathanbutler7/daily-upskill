# Jonathan Butler — Promotion Packet

**P3-I → P3-II** · **Review period:** January 1–December 31, 2026  
**Time at level:** P3 since September 2023 (Frontend); Backend Payments at P3-I since February 2025

---

## The Case in One Page

In 2026 my work moved from owning CareCredit systems to setting direction for the Payments team. I found problems nobody had assigned, built tools and approaches that other people picked up, and led work across Product, Design, Support, Sales, BI, and Frontend.

I meet **two of the three** P3-II evidence anchors in the [career guide](./career-levels-guide.md): a proposal or pattern I initiated that the team adopted, and cross-team coordination I led.

### Strongest evidence

| When | Example | What I did | What other people did with it | Anchor |
| :--- | :--- | :--- | :--- | :--- |
| [TODO: month] | **Surcharge Calculator** | Turned surcharge rules into a calculator anyone could use | Support used it to check whether customer calls were real bugs. It spread to Slack channels I never shared it in. Product used it to check the math. T2 now troubleshoots surcharge issues without escalating ([TODO: name, role]). | Cross-team |
| Aug | **CareCredit status mapping** | Audited how the UI showed statuses and wrote the CareCredit → Weave status mapping | Product, Design, and Engineering stopped going in circles after weeks. The audit found frontend bugs nobody knew about. | Cross-team |
| Sep | **CareCredit demo for Sales** | Proposed shipping a request-intercepted demo now and deferring full data seeding | Sales runs live CareCredit demos using a guide I wrote with Cassidy Kelch. BI reporting stays clean. | Cross-team |
| Sep | **QuickScreen caching** | Built the proof of concept that set the caching approach | Jan shipped the 60-day production cache using that approach | Adopted pattern |
| Sep–Nov | **Production smoke tests** | Wrote the ADR and designed a Go + Playwright runner, live every 15 minutes | Aligned with Chris Draycott (P5) and the Frontend Stability initiative so frontend engineers can add tests without writing Go | Both |
| Nov | **Deployment sanity tests** | Found an unowned testing gap and built `test-payments-ui` | Jan opened a PR adding simulated terminal readers and single-location support. [TODO: Aditya's contribution] | Adopted pattern |

### Technical credibility

- **Payouts:** cut a 98-location payout summary from **64.2s to 1.3s** by making cross-service calls concurrent, which stopped SLO burn and pages.
- **`stripesyncjob`:** found an hourly job that had run **32 times instead of ~4,000** because of a checkpoint bug stuck on an 8.5-month lookback. I led the decision to retire it.
- **Duplicate payments:** shipped invoice state guards across sync and terminal payment flows through **100% rollout**.

### In progress through December 31

- **Full Payments demo seeder TPD** ([PAY-7523](https://linear.app/getweave/issue/PAY-7523/tpd)): discovery is done. The TPD compares a demo-only integration with a reusable multi-processor router, and it's written for another engineer to implement. [TODO: status as of Dec 31 and who picked it up]
- ***How to do TPDs well*** and **incident-based AI PR review skills**: [TODO: name a TPD or PR where these changed the outcome]

### Not claimed

- **Anchor 3 (an engineer measurably improved under me):** not claimed.
- **Organizational need:** [TODO: manager to describe the need for P3-II scope on Payments]

---

# Appendix: Details

Everything below is supporting detail for the summary above.

- [A. Cross-team leadership](#a-cross-team-leadership)
- [B. Patterns the team adopted](#b-patterns-the-team-adopted)
- [C. Team standards](#c-team-standards)
- [D. Technical delivery](#d-technical-delivery)
- [E. Delivery volume](#e-delivery-volume)

---

## A. Cross-team leadership

### CareCredit status mapping (August)
- **Problem:** Product, Design, and Engineering had spent weeks disagreeing on how CareCredit statuses should map to Weave statuses and show up in the UI.
- **What I did:** Audited how the UI displayed statuses and wrote the [CareCredit Status Alignment Document](https://docs.google.com/document/d/1LMEti1HZHO10cMvI-AfLSjlzIglqxeFy9O82WCDby-c/edit?tab=t.0#heading=h.v4pnpx9a1db6) defining the mapping.
- **What others did:** All three functions used the mapping to settle expected behavior and prioritize fixes.
- **Result:** Ended the back-and-forth and surfaced frontend status bugs before customers reported them.

### Surcharge Calculator ([TODO: month])
- **Problem:** Surcharges depend on several location, state, and payment-method rules. People outside Engineering had no quick way to check what a surcharge should be.
- **What I did:** Built the [Surcharge Calculator](https://docs.google.com/spreadsheets/d/1pbEGxKuHuCOpynu1qqQF9gZ1F91fx-o-oY0YTV4TbFI/edit?gid=92512368#gid=92512368), which documents how a surcharge gets determined.
- **What others did:** Support used it on customer calls and shared it in channels beyond the ones I posted in. Product used it to check the math.
- **Result:** Support could tell a real surcharge bug from expected behavior before escalating to Engineering.
- **Evidence:** "It's been extremely useful for my team and T2's. I've had a lot [less] inquiries make it my way because T2 is able [to] troubleshoot surcharge calculations precisely due to the calculator. It reduces a lot of time I used to spend on surcharge issues." — [TODO: name, role]. [TODO: escalations per month before vs. after]
- **Also:** Slack screenshots of unprompted use. [TODO: attach]

### CareCredit demo environment for Sales (September)
- **Problem:** Sales needed live CareCredit demos. A full seeded-data platform would have taken months.
- **What I did:** On September 3, proposed shipping a deterministic request-intercepted demo ([PAY-7230](https://linear.app/getweave/issue/PAY-7230/add-care-credit-demo-request-interceptor)) and kept demo traffic out of BI ([PAY-7284](https://linear.app/getweave/issue/PAY-7284/prevent-demo-locations-from-being-reported-to-bi)). Full seeding moved to a follow-on project.
- **What others did:** Wrote the Sales demo guide with Cassidy Kelch ([PAY-7297](https://linear.app/getweave/issue/PAY-7297/write-up-guide-for-sales-on-how-to-do-a-cc-demo)), and Sales runs demos from it.
- **Result:** Live in production September 8–9. [TODO: number of demos Sales has run]
- **Project:** [Demo Environment for CareCredit](https://linear.app/getweave/project/demo-environment-for-carecredit-3ada65bf2e75)

### CareCredit decoupled rollout (March–October)
- **Problem:** Kylir, the frontend engineer on the decoupled CareCredit work, left mid-project.
- **What I did:** Took over the frontend work, closed the remaining gaps, and wrote a regression-testing and staged rollout plan covering frontend and backend. I used the same rollout and rollback structure for QuickScreen ([PAY-7186](https://linear.app/getweave/issue/PAY-7186/enhance-quickscreen-and-account-lookup)) and duplicate-payment protection ([PAY-6262](https://linear.app/getweave/issue/PAY-6262/enable-testbeta-location-rollout)).
- **Result:** Released to production with no rollout regressions.

---

## B. Patterns the team adopted

### QuickScreen caching (September)
- **Problem:** CareCredit lookups made repeat external pre-qualification calls.
- **What I did:** Built the caching proof of concept ([care-credit-integrations#120](https://github.com/weave-lab/care-credit-integrations/pull/120)) that set the approach.
- **What others did:** Jan implemented 60-day production caching on that approach ([care-credit-integrations#150](https://linear.app/getweave/review/use-cached-quickscreen-response-for-60-days-be5e8b73a3be)).
- **Result:** [TODO: reduction in external calls or latency]

### Production smoke tests (September–November)
- **Problem:** Payments had no continuous production check of real user flows.
- **What I did:** Wrote the ADR on September 22 comparing setup effort, alerting, persistence, and migration paths. Designed a unified Go job that runs Playwright checks and alerts on failure ([PAY-7324](https://linear.app/getweave/issue/PAY-7324/spin-up-playwright-in-unified-job), [PAY-7326](https://linear.app/getweave/issue/PAY-7326/create-playwright-tests-to-determine-if-job-passes-or-fails), [PAY-7328](https://linear.app/getweave/issue/PAY-7328/implement-alerting-mechanism-for-when-the-tests-fail), [PAY-7334](https://linear.app/getweave/issue/PAY-7334/add-docs)).
- **What others did:** Aligned the design with Chris Draycott (P5) and the Frontend Stability initiative so frontend engineers can write plain Playwright tests without Go. [TODO: the specific decision you and Chris agreed on]
- **Result:** Live October 30, running every 15 minutes across bill-pay, payment requests, and the main app.
- **Project:** [Smoke Tests Implementation](https://linear.app/getweave/project/smoke-tests-implementation-8769c3916009)

### Deployment sanity tests (November)
- **Problem:** Deployment checks across payment flows were manual. Frontend CI covered PR builds and smoke tests covered availability, but nothing exercised full authenticated payment flows against deployed services.
- **What I did:** Built [test-payments-ui](https://github.com/weave-lab/test-payments-ui) and kept v1 small: manually triggered runs against stable using test cards, with CI/CD and extra release channels deferred.
- **What others did:** Jan opened a PR adding simulated terminal reader and single-location user support ([test-payments-ui#2](https://github.com/weave-lab/test-payments-ui/pull/2)). Aditya contributed tests using the same structure. [TODO: link Aditya's work]
- **Result:** [TODO: deploys checked or regressions caught]

### Full Payments demo seeder and multi-processor routing (November–December)
- **Problem:** Terminal demo seeding could be built as a one-off script or as the start of a multi-processor routing layer.
- **What I did:** Audited how other SDS projects handle dependent datasets and separate demo data from real data ([PAY-7527](https://linear.app/getweave/issue/PAY-7527/audit-how-other-sds-projects-handle-dependent-datasets), [PAY-7528](https://linear.app/getweave/issue/PAY-7528/audit-how-other-sds-differentiate-demo-data-from-real-data), [PAY-7548](https://linear.app/getweave/issue/PAY-7548/audit-what-data-is-currently-seeded-for-demo-locations)), set up the repo ([PAY-7544](https://linear.app/getweave/issue/PAY-7544/set-up-sds-payments-repo)), and wrote the TPD comparing both options ([PAY-7523](https://linear.app/getweave/issue/PAY-7523/tpd), [PAY-7536](https://linear.app/getweave/issue/PAY-7536/document-possible-architectures), [PAY-7539](https://linear.app/getweave/issue/PAY-7539/multiprocessor-sds-payment-router)).
- **Why multi-processor:** [TODO: the upcoming requirement that justifies it]
- **What others did:** [TODO: who is implementing and which decisions from the TPD carried through]
- **Project:** [Payments Full Experience Demo Seeder](https://linear.app/getweave/project/payments-full-experience-demo-seeder-terminal-payments-8c51fbff2fef)

---

## C. Team standards

### *How to do TPDs well* (November)
I kept seeing TPDs written to get approval for a solution that had already been picked. The guide covers:
- Start with the problem and customer impact.
- Use a Linear issue when a full TPD is overkill.
- Compare real alternatives and how each one fails.
- Agree up front whether the review is early architecture feedback or final polish.
- Write it so someone other than the author can build it.

[TODO: a TPD that changed because of this guide or your review]

### AI PR review skills
- Built Copilot review skills based on production incident analyses so known failure modes get flagged in code review.
- Reviewed teammates' AI skill files in PRs.
- Demoed AI-assisted development timelines in org-wide demos.

[TODO: a PR where a skill caught a real issue]

---

## D. Technical delivery

### `stripesyncjob` silent failure (January)
- Noticed erratic timing in the hourly job and added diagnostic logging. Over six months it had run **32 times instead of ~4,000** because each run took 16+ hours.
- **Root cause:** a GCS checkpoint (`GetNextSyncStartTime`) overrode the 75-minute lookback. A failure around May 5, 2025 pinned the start time, so every run scanned 8.5 months of data.
- Confirmed what reconciliation the job still did, showed my replacement covered it, and led the decision to retire it. ([payments-crons](https://github.com/weave-lab/payments-crons))

### Payout summary speedup (August)
- Traced slow multi-location payout summaries, SLO burn, and pages to sequential calls from `payouts` to `stripe-integration-service`.
- Switched to bounded concurrency and benchmarked at 1, 50, and 100 locations. A 98-location request went from **64.2s to 1.3s**. ([PAY-7168](https://linear.app/getweave/issue/PAY-7168/fix-payouts-summary-endpoint), [payouts#64](https://linear.app/getweave/review/use-concurrent-requests-to-stripe-integration-service-to-100x-speedup-d4150828ca50))

### Endpoint-level SLO (September)
- `payout-transactions` kept breaching a service-wide latency SLO. Used seven days of production latency data to give it its own target (90% under 5s) and kept tight budgets on the fast endpoints. ([payouts#65](https://linear.app/getweave/review/scope-payout-latency-slo-by-endpoint-072bf2ce9761))

### Duplicate-payment prevention (January–August)
- Payment plan status checks ([PAY-6275](https://linear.app/getweave/issue/PAY-6275/check-pendingpaid-status-to-avoid-duplicate-payment-plans)), invoice set to pending on `StartPayment` ([PAY-6244](https://linear.app/getweave/issue/PAY-6244/update-synchronous-startpayment-to-set-invoice-to-pending)), and revert on terminal failure ([PAY-6246](https://linear.app/getweave/issue/PAY-6246/update-processterminalpayment-workflow-to-revert-pending-invoices)).
- Rolled out to 100% of locations on August 26 and analyzed the remaining pending-invoice edge cases ([PAY-6262](https://linear.app/getweave/issue/PAY-6262/enable-testbeta-location-rollout)).

### CareCredit backend
| Date | Capability | Reference |
| :--- | :--- | :--- |
| Mar 16 | Public webhook endpoints | [PAY-6589](https://linear.app/getweave/issue/PAY-6589/create-public-webhook-endpoints) |
| Mar 20 | Location → MID mapping and API contracts | [PAY-6575](https://linear.app/getweave/issue/PAY-6575/location-id-mid-mapping) |
| Mar 25 | Webhooks to Kafka; transactions table | [PAY-6615](https://linear.app/getweave/issue/PAY-6615/publish-webhooks-to-kafka), [PAY-6537](https://linear.app/getweave/issue/PAY-6537/create-care-credit-transactions-table) |
| Apr 1, 21 | Text-to-pay links and async reconciliation | [PAY-6620](https://linear.app/getweave/issue/PAY-6620/purl-ttp), [PAY-6746](https://linear.app/getweave/issue/PAY-6746/process-ttp-payments) |
| May 14 | CareCredit email receipts | [PAY-6800](https://linear.app/getweave/issue/PAY-6800/email-carecredit-customer-receipt) |
| Jun 2 | Encryption-key tooling and runbook | [PAY-6888](https://linear.app/getweave/issue/PAY-6888/improve-encryption-key-ops) |
| Jul 29 | CDC consumer with validation and retries | [PAY-6970](https://linear.app/getweave/issue/PAY-6970/implement-cdc-consumer) |
| Sep 11 | Account lookup and QuickScreen with rollout/rollback plan | [PAY-7186](https://linear.app/getweave/issue/PAY-7186/enhance-quickscreen-and-account-lookup) |
| Oct 30 | Retryable webhook errors so failures aren't silently acked | [PAY-7467](https://linear.app/getweave/issue/PAY-7467/carecredit-webhooks-define-retryable-errors) |

### CareCredit frontend and full-stack fixes
| Date | Fix | Reference |
| :--- | :--- | :--- |
| Aug 12 | Sunbit option missing when CareCredit was also enabled | [PAY-7172](https://linear.app/getweave/issue/PAY-7172/inquiry-s00125992-sunbit-option-not-rendering) |
| Oct 9 | Closed and conflicting account states, with regression tests | [PAY-7403](https://linear.app/getweave/issue/PAY-7403/handle-closed-carecredit-accounts) |
| Oct 13 | Premature and duplicate promo API calls | [PAY-7432](https://linear.app/getweave/issue/PAY-7432/duplicate-payments-platformcare-credit-integrationpurchasespromos) |
| Oct 14 | Empty person ID in payment-link requests | [PAY-7437](https://linear.app/getweave/issue/PAY-7437/empty-person-id-in-purl-request) |
| Oct 20 | Person IDs misparsed as names in payment history (frontend, schema, `invoice`) | [PAY-7398](https://linear.app/getweave/issue/PAY-7398/fix-listpayments-person-id-filter) |

### Hackathon finalist (June)
- Built a payment-likelihood prediction prototype with Pablo Trujillo ([HACK-222](https://linear.app/getweave/issue/HACK-222/demo-will-they-pay)).

---

## E. Delivery volume

As of November 11. [TODO: refresh at year end]
- 124 completed Linear issues across 11 projects
- 43 Major Enhancement and 32 Inquiry issues
- 11 project status updates across six initiatives
- 5 merged PRs in `weave-iso/frontend`
