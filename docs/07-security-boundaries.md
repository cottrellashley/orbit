# Security Boundaries

## Sensitive data policy

Sensitive data lives in ~/Vault/. This includes:
- Health records and tracking
- Journal and personal reflections
- Financial documents
- OAuth tokens and credentials (beyond what tools manage themselves)
- Identity documents

## What must NOT be in the architecture repo

- Raw secrets, API keys, OAuth tokens
- Financial source documents
- Personal health records
- Identity documents (passport, visa, etc.)

## Access rules (design intent)

| Domain | Can access | Cannot access |
|--------|-----------|---------------|
| Executive profile | ~/Executive/, ~/Vault/ (read), ~/AI/architecture/ (read) | ~/Dev/ |
| Engineering profile | ~/Dev/, ~/AI/architecture/ (read) | ~/Executive/, ~/Vault/ |
| Architecture planner | ~/AI/architecture/ (read/write) | ~/Vault/ contents |

## Not yet decided

- Whether Vault should have internal structure (subdirectories by category)
- How access boundaries are enforced (convention vs tooling)
- Whether the executive assistant gets write access to Vault
