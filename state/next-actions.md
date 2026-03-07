# Next Actions

Concrete implementation steps, in rough priority order.

1. **Configure the executive OpenCode profile** — fill in opencode.jsonc
   with provider, model, and any initial MCP servers
2. **Configure the engineering OpenCode profile** — same as above
3. **Decide on profile invocation method** — launcher scripts, aliases,
   or manual opencode flags
4. **Design executive state schema** — decide format for persistent
   state files in ~/Executive/state/
5. **Decide what to migrate from AshleyDB** — review archived vault,
   determine what (if anything) moves to ~/Vault/ or ~/Executive/
6. **Set up ~/Vault/ internal structure** — once the sensitive data
   policy is decided
