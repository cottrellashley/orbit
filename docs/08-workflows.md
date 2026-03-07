# Workflows

## How the system is used

### Starting an executive session
1. Open terminal
2. Launch OpenCode with the executive profile
3. The agent reads ~/Executive/state/ for context
4. Work proceeds against ~/Executive/ (kanban, planning, etc.)

### Starting an engineering session
1. Open terminal
2. Launch OpenCode with the engineering profile
3. Navigate to or specify the target project in ~/Dev/active/
4. Work proceeds against the project

### Architecture planning
1. Open terminal
2. Launch OpenCode (either profile, or bare)
3. The planner reads ~/AI/architecture/docs/ for context
4. Changes are proposed, reviewed, then committed

## Not yet decided

- Exact invocation commands for each profile
- Whether profiles are launched via wrapper scripts in ~/AI/opencode/bin/
- Session handoff or continuity mechanisms
- How the kanban board is updated (manual vs agent-driven)
