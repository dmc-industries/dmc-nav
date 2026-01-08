# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Required Protocols

### Session Workflow (BEADS.md)
Uses `bd` CLI for issue tracking.
- **Session start**: Run `bd stats` and `bd ready`
- **Session end**: Commit, push, verify `git status` shows "up to date with origin"
- Work is NOT complete until `git push` succeeds

### Architecture Principles (ZFC.md)
Follow Zero Framework Cognition when writing code:
- **DO**: Pure orchestration, IO, schema validation, policy enforcement, mechanical transforms
- **DON'T**: Local heuristics, ranking/scoring, semantic analysis, plan composition, quality judgment
- Delegate ALL reasoning to AI; build thin deterministic shells

## Repository Overview

<!-- TODO: Describe this repository's purpose and structure -->

## Build & Development

<!-- TODO: Add build commands, test commands, development setup -->

## Code Style & Conventions

<!-- TODO: Document coding standards, patterns used -->
