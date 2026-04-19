---
title: "Evaluation Model"
status: implemented
date: 2026-04-18
---

# Evaluation Model

## 1. Scope

This document defines how `cmdguard` v1 evaluates input against configured
rules and selects a deny decision.

## 2. Rule Model

v1 uses a deny-only rule model.

- A rule contains either a structured `match` predicate or a regular
  expression `pattern`, plus a deny message
- If a rule matches the command string, the command is denied
- If no rule matches, the command is allowed
- v1 does not support allow rules, explicit exceptions, or rule priority

This keeps the runtime contract small and deterministic. More expressive policy
features are post-v1 work.

`cmdguard` keeps the caller contract simple: callers pass a raw command string.
Inside `cmdguard`, predicate-based rules are evaluated against an internal
parsed representation of that command string.

That internal parsing may unwrap a small set of common launcher-style wrappers
before matcher evaluation, such as `env`, `command`, `exec`, `sudo`, `nohup`,
`timeout`, and `busybox sh`. This is intentionally heuristic rather than a full
shell AST.

## 3. Supported Input

`cmdguard eval` accepts only execution input.

- Generic input:
  - `action` must be `"exec"`
  - `command` must be a non-empty string
- Claude Code adapter input:
  - `tool_name` must be `"Bash"`
  - `tool_input.command` must be a non-empty string

If input does not conform to one of these shapes, evaluation fails with an
error. v1 does not silently allow unknown action types.

## 4. Configuration Source

v1 evaluates rules from a single user-wide configuration file:

- `$XDG_CONFIG_HOME/cmdguard/cmdguard.yml`
- `~/.config/cmdguard/cmdguard.yml` when `XDG_CONFIG_HOME` is not set

Within that file, source order is preserved.

## 5. Evaluation Order

The evaluation order is fixed and deterministic.

1. Load the user-wide config file if present
2. Preserve rule source order
3. Evaluate rules in that order
4. Select the first matching rule

The first matching rule is the decision rule. Later matching rules are ignored
for the purpose of the runtime deny decision.

## 6. Rule Identity

Rule IDs must be unique within the effective configuration file.

- Duplicate IDs within a file are errors

v1 does not define override semantics through repeated IDs.

## 7. Consequences of First-Match Selection

Because first-match selection is part of the contract:

- Runtime deny messages are stable for a given config set
- Tests can assert the selected rule ID deterministically
- Rule order is a meaningful part of configuration behavior

This also means one rule can shadow a later rule. Shadowing is not a runtime
error in v1, but diagnostic tooling should be able to report likely shadowing
as a warning.
