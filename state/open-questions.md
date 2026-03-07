# Open Questions

Items that need decisions before the architecture can advance.

## Executive system
- What is the schema/format for durable state files?
- How should calendar be handled? (file, MCP integration, or both)
- How do memory and preferences persist across sessions?
- What MCP servers or tools should the executive profile use?

## Engineering system
- Should there be a project-scaffolding convention under ~/Dev/active/?
- What MCP servers or tools should the engineering profile use?

## Cross-cutting
- How are profile access boundaries enforced? (convention only, or tooling)
- What is the exact invocation method for each profile?
- Should ~/AI/opencode/bin/ contain launcher scripts?
- What internal structure should ~/Vault/ have?
- Should the executive assistant have write access to ~/Vault/?
- What should be migrated (if anything) from the archived AshleyDB?
