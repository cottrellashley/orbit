# ADR-001: Bootstrap architecture repository

**Date:** 2026-03-07
**Status:** Accepted

## Context

Ashley reset his local AI/workspace setup on 2026-03-07, archiving the
previous Claude Code + Codex + AshleyDB setup into ~/Archive/. A new
minimal directory structure was created with two domains (executive,
engineering) and placeholder configs.

There was no canonical place to describe the system design, track
decisions, or record open questions.

## Decision

Create ~/AI/architecture/ as a git repository containing:
- docs/ — numbered design documents
- decisions/ — architectural decision records
- state/ — open questions and next actions

This repo is documentation only. It contains no code, secrets, or
runtime configuration.

## Consequences

- All future architecture changes are proposed and tracked here
- The editing protocol (docs/09-editing-protocol.md) governs how
  changes are made
- This is a fresh start; no design debt is carried from the previous system
