# Editing Protocol

Rules for maintaining this architecture repository.

## Source of truth

~/AI/architecture/ is the single source of truth for system design.
All architectural decisions, current state, and open questions live here.

## Before any change

1. Read docs/00-overview.md and docs/01-current-state.md
2. Read docs/09-editing-protocol.md (this file)
3. Read any additional docs relevant to the change

## Change process

1. Summarize the relevant current state
2. Identify what part of the architecture is affected
3. Propose the smallest sensible change
4. List exactly which files must be updated
5. Identify risks, tradeoffs, or open questions
6. Plan first, then edit only when explicitly approved

## File update rules

- If the top-level picture changes → update docs/00-overview.md
- If current implementation changes → update docs/01-current-state.md
- If a structural decision is made → add or update an ADR in decisions/
- If issues are unresolved → record in state/open-questions.md
- If follow-up steps exist → record in state/next-actions.md
- Do not let files drift out of sync

## ADR format

Decisions are recorded in decisions/ADR-NNN-short-title.md:
- Title
- Date
- Status (proposed / accepted / superseded)
- Context
- Decision
- Consequences
