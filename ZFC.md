# ZFC (Zero Framework Cognition) Principles

> **Core Principle:** This application is pure orchestration that delegates ALL reasoning to external AI. We build a "thin, safe, deterministic shell" around AI reasoning with strong guardrails and observability.

## Allowed (ZFC-Compliant)

### IO and Plumbing
- Read/write files, list directories, parse JSON, serialize/deserialize
- Persist to stores, watch events, index documents

### Structural Safety Checks
- Schema validation, required fields verification
- Path traversal prevention, timeout enforcement, cancellation handling

### Policy Enforcement
- Budget caps, rate limits, confidence thresholds
- "Don't run without approval" gates

### Mechanical Transforms
- Parameter substitution (e.g., `${param}` replacement)
- Compilation
- Formatting and rendering AI-provided data

### State Management
- Lifecycle tracking, progress monitoring
- Mission journaling, escalation policy execution

### Typed Error Handling
- Use SDK-provided error classes (instanceof checks)
- Avoid message parsing

## Forbidden (ZFC-Violations)

### Ranking/Scoring/Selection
- Any algorithm that chooses among alternatives based on heuristics or weights

### Plan/Composition/Scheduling
- Decisions about dependencies, ordering, parallelization, retry policies

### Semantic Analysis
- Inferring complexity, scope, file dependencies
- Determining "what should be done next"

### Heuristic Classification
- Keyword-based routing
- Fallback decision trees
- Domain-specific rules

### Quality Judgment
- Opinionated validation beyond structural safety
- Recommendations like "test-first recommended"

## The Correct Flow

1. **Gather Raw Context** (IO only)
   - User intent, project files, constraints, mission state

2. **Call AI for Decisions**
   - Classification, selection, composition
   - Ordering, validation, next steps

3. **Validate Structure**
   - Schema conformance
   - Safety checks
   - Policy enforcement

4. **Execute Mechanically**
   - Run AI's decisions without modification
