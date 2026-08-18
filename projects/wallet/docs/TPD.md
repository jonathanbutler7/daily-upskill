# TPD: Double Entry Ledger

| Status | Draft \> In Review \> Implementation \> Done |
| :---- | :---- |
| **Created** | 7/24 |
| **Target Work Start** | Date |
| **Target Deadline** | Date |
| **Author(s)** | Jonathan Butler |
| **Project Lead** | Jonathan Butler |
| **Stakeholders** | Jonathan Butler |
| **Product Manager Review** | Jonathan Butler |
| **Eng Director Review** | Date |
| **Requires Architecture Review** *If “Yes” please share with `@principal-engineers` for review.* | Does this project introduce new concepts or business entities which will be used by other teams?  OR Does this project introduce new technologies that have not been productionized at Weave? Not Yet Answered |
| **Principal Engineer Review** | Date |
| **Requires Security Review** *If “Yes” please complete the security assessment and link below.* | Does this project touch sensitive systems or data such as Auth, Compliance, Audit Logging, PII/PH, OS commands in the desktop client? OR Does this project introduce new vendors, third party dependencies, or make outbound requests to a third party? Not Yet Answered |
| **Security Engineer Review** | Date |
| **Quick Links** | \<PRD\> \<figma designs\> \<linear project\> |

# Summary

*A brief description of the problem to be solved.*

Safely recording money movement so that 

# Job to be Done

*The purpose of this section is to ensure that all stakeholders are aligned to ensure we build the right thing.*

## Assumptions

*What assumptions have been made about this project? Are there concepts that need to be defined here?*

* 

## Must Have

*What needs to be done before the project can be considered complete? This is also a list of all criteria that are used to determine the correct solution. Each solution option should be evaluated according to this list.*

* 

## Nice To Have

*Stretch goals*

* 

## Out of Scope

*What isn't part of this project? Let's be explicit for the sake of clarity.*

* 

## Risks / Prerequisites / Blockers

*What are the inherent risks of this project?  How is the team planning to mitigate these risks? List any questions that are or will prevent work from moving forward.* 

* 

## Open Questions

* 

# Solution

Describe the technical solution in depth. Include things like key decision points (logic), sequence diagrams, architecture diagrams, API endpoints, etc.

# Project Plan

## Milestones

*The major check-in points. This can include estimates to provide further clarity into the scope of the project.*

### \<Milestone name\>

- [ ] 

## 

Dependencies  
*Which teams, projects, or other items outside of your team's direct control is the success of this project dependent on? For example, is there work required by Mobile, PLG, Onboarding, Business Systems, or even another company like an Integration Partner?*

## Testing Plan

| Question | Answer |
| :---- | :---- |
| How disruptive is this code that is being written? Will it require significant code refactoring in order to test the changes? |  |
| What existing functionality will break or be affected if something goes wrong with this project in production? |  |
| Are the changes in this project backwards compatible? |  |
| Does this code have the potential to break upstream/downstream dependencies? |  |
| If upstream or downstream services are affected by this change (e.g. increased load), how will you include those in your testing plan? |  |
| Will this project introduce changes to existing API contracts or deprecate existing APIs? |  |
| If the new functionality needs test data, will a new data seeder be required? |  |

*How will this project be tested? Unit tests? TAS tests? Load testing? Smoke tests? Use the testing assessment in the appendix as a guide when building this section.*

## Support & Observability Plan

| Question | Answer |
| :---- | :---- |
| What metrics, logging, **alerts** and dashboards will be created for this project? |  |
| Which team gets alerted if this project causes an incident or negatively impacts customers? |  |
| Who will own this project long term? |  |
| How will this project be supported internally? |  |
| What regular maintenance is required to keep this project healthy? |  |

## Release Plan

| Question | Answer |
| :---- | :---- |
| How will this project be deployed to production? Are there special considerations in the building and deployment of new services and/or existing services? |  |
| Do some services need to be deployed before others? |  |
| Will this project directly or indirectly create or increase infrastructure costs (e.g., GCP, BW, etc)? |  |
| Will this project change or create additional load on any existing service? Have you reached out to the team that owns the affected service? |  |
| What is your plan for rolling back the change in the case where something goes wrong? |  |
| Will feature flags be used in the roll out? If so, which ones? |  |
| Will the release require a [planned maintenance window](https://docs.google.com/document/d/11ZQKh0QTu9pJc4iMlZAguJon1NQ02arzMD9ycaJxEHk/edit?tab=t.0#heading=h.96jl6g8obz0p)? Have you coordinated the window with your director and with support leadership? |  |
| Are there any migrations needed for this project? How are these migrations being coordinated and managed? |  |

*Describe how you will release this change. Use the release assessment in the appendix as a guide when creating this section.*

---

# Appendix 1: Definitions and Assumptions

---

# Appendix 2: Alternative Solutions Considered