# Next Actions

Concrete steps, in rough priority order.

## Phase 1: Bootstrap orbit
1. **Install orbit to PATH** — `go install` or symlink the binary
2. **Create `~/.config/orbit/config.yaml`** — initial config with
   opencode adapter (default) and two roles (executive, engineering)
3. **Verify `orbit list` and `orbit status` work** with the config

## Phase 2: Global OpenCode config
4. **Create `~/.config/opencode/`** — set up global config with API keys,
   default model, and any shared settings
5. **Verify OpenCode launches** with the global config from a bare directory

## Phase 3: Executive environment
6. **`orbit init executive`** — or manually register if already exists
7. **Create `~/Executive/opencode.json`** — executive-specific config
8. **Create `~/Executive/AGENTS.md`** — executive system prompt and rules
9. **Set up `assistant` alias** — `alias assistant="orbit open executive"`
10. **Design executive state schema** — decide format for files in
    `~/Executive/state/`

## Phase 4: Engineering environment
11. **Register engineering workspace** — `orbit init engineering --type workspace --path ~/Dev/active`
12. **`orbit new engineering test-project`** — scaffold a minimal project
    to validate the workflow
13. **Set up `dev` alias** — `alias dev="orbit open engineering"`

## Phase 5: Content and polish
14. **Build executive commands/skills** — `/plan`, `/review-week`, etc.
15. **Build engineering commands/skills** — project-specific helpers
16. **Decide what to migrate from AshleyDB** — review archived vault
17. **Set up `~/Vault/` internal structure** — once sensitive data policy
    is decided
