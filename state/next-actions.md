# Next Actions

Concrete steps, in rough priority order.

## Phase 1: Global config
1. **Create `~/.config/opencode/`** — set up global config with API keys,
   default model, and any shared settings
2. **Verify OpenCode launches** with the global config from a bare directory

## Phase 2: Executive environment
3. **Create `~/Executive/opencode.json`** — executive-specific config
4. **Create `~/Executive/AGENTS.md`** — executive system prompt and rules
5. **Set up `assistant` alias** — simple alias or script to run opencode
   in `~/Executive/`
6. **Design executive state schema** — decide format for files in
   `~/Executive/state/`

## Phase 3: Engineering environment
7. **Set up `dev` alias** — script that presents project list from
   `~/Dev/active/` and launches opencode in the selected project
8. **Create a test project** — scaffold a minimal project in
   `~/Dev/active/` to validate the engineering workflow
9. **Decide on project scaffolding** — whether `dev new` should
   scaffold from a template

## Phase 4: Content and polish
10. **Build executive commands/skills** — `/plan`, `/review-week`, etc.
11. **Build engineering commands/skills** — project-specific helpers
12. **Decide what to migrate from AshleyDB** — review archived vault
13. **Set up `~/Vault/` internal structure** — once sensitive data policy
    is decided
