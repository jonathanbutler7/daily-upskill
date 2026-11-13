# How to do TPDs well

Notes to kick off a team discussion. These are open questions, not a proposal.

## Where TPDs get stuck

### 1. Review turns into approval
- RFCs (request for comment) tend to become RFAs (request for approval). The author shows up with a finished solution and reviewers are asked to sign off.
- How to give feedback isn't clear. Design by committee isn't great either, so where's the middle?
- Who makes the call when reviewers disagree? If nobody owns the decision, the doc stalls or the loudest comment wins.

### 2. Knowing when to write one
- How small is too small for a TPD?
- When is a well-written ticket better?
- Writing a TPD signals the work is important, and it invites more scrutiny. That can be good, or it can slow down work that didn't need it.
- Size may be the wrong test. Reversibility might be better: a hard-to-undo decision (schema, public API, vendor choice) deserves a doc even if the work is small.

### 3. Format and template
- The template is recommended but not required. Which one should people use, and how do we review a doc with a custom format?
- "Alternate solutions considered" is an appendix. Putting it up front would make it clearer that the author weighed options before picking one.
- Business impact isn't a section. If a small-impact problem gets a large engineering solution, the doc should make that visible.

### 4. Who the doc is for
- If someone else might implement it, the doc needs to be much clearer than if the author is building it next sprint.

## Guidance for authors

1. State the problem and who it affects. Include the customer impact.
2. Say what happens if we don't do this.
3. Lay out more than one solution, with tradeoffs, before the recommendation.
4. Offer more than one effort level (for example: a 2-day patch vs. a 3-week redesign). This helps whoever picks it up later weigh it against other priorities.
5. Put a short header at the top:
   - **Phase:** exploring / converging / final
   - **Feedback wanted:** e.g. problem framing, alternatives, writing, data model
   - **Decision owner** and **comment deadline**
   - **Implementer and timing**, if known

The header answers the reviewer questions below before anyone has to ask them.

## Guidance for reviewers

1. **Check the phase.** Early docs are open to foundational feedback. Late docs usually aren't, so focus on risks and gaps.
2. **Match the feedback the author asked for.** Writing, alternatives, consistency/availability tradeoffs, etc. Raise anything else as a separate comment and label it.
3. **Check who's implementing and when.** A rougher doc is fine for small scope, or when the author is building it soon. It needs more detail if it'll sit on the shelf or go to someone else.
4. **Label comment weight.** For example: `blocking`, `question`, `nit`. This keeps an RFC from feeling like an RFA.

## Questions for the team

1. What's our bar for when a TPD is needed?
2. Should alternatives move out of the appendix?
3. Should business impact be a required section?
4. Who decides when reviewers disagree, and how long does review stay open?
5. Do we want a lighter "one-pager" option for small but hard-to-reverse changes?
