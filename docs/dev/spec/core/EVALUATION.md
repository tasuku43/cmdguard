---
title: "Evaluation Model"
status: proposed
date: 2026-04-19
---

# Evaluation Model

## 1. Scope

This document defines the target evaluation model for `cmdproxy`.

`cmdproxy` evaluates a requested CLI invocation against ordered rules and
applies the first matching directive.

## 2. Rule Model

The target model is directive-driven rather than deny-only.

- a rule contains a matcher and exactly one directive
- supported directives are `rewrite` and `reject`
- if no rule matches, the invocation is passed through unchanged

This yields three runtime outcomes:

- `pass`: no match, original invocation forwarded
- `rewrite`: matched and normalized into a canonical form
- `reject`: matched and blocked

## 3. Caller Contract

`cmdproxy` keeps caller input intentionally simple.

- the caller provides a requested command invocation
- the primary external form is still a raw command string for `exec`
- internal parsing and normalization complexity stays inside `cmdproxy`

The caller should not need to understand the internal matcher model.

## 4. Internal Normalization

Inside `cmdproxy`, the raw command string is normalized into a parsed
invocation model that may include:

- environment assignments
- executable basename
- argument vector
- subcommand
- a small set of unwrapped launcher-style wrappers

Wrapper unwrapping is heuristic and intentionally limited. It may cover common
forms such as `env`, `command`, `exec`, `sudo`, `nohup`, `timeout`, and
`busybox sh`, but it is not a full shell AST.

## 5. Configuration Source

The current target config location is a single user-wide file:

- `$XDG_CONFIG_HOME/cmdproxy/cmdproxy.yml`
- `~/.config/cmdproxy/cmdproxy.yml` when `XDG_CONFIG_HOME` is not set

Within that file, source order is preserved.

## 6. Evaluation Order

Evaluation order is fixed and deterministic.

1. Load the effective config file
2. Preserve rule order
3. Parse and normalize the input invocation
4. Evaluate rules in order
5. Apply the first matching directive
6. If `rewrite.continue: true`, restart evaluation from the top with the rewritten command
7. If nothing matches, pass the current invocation

## 7. Rewrite Semantics

`rewrite` is policy-preserving canonicalization, not arbitrary transformation.

Target properties:

- deterministic
- local to invocation structure
- safer or more canonical than the original form
- suitable for downstream permission evaluation

Examples:

- move a flag value into a sanctioned environment variable
- unwrap `bash -c 'single command'` into the direct command form
- remove a wrapper that obscures the effective executable

The currently implemented rewrite primitives are:

- `rewrite.move_flag_to_env`
- `rewrite.move_env_to_flag`
- `rewrite.unwrap_shell_dash_c`
- `rewrite.unwrap_wrapper`

If a rewrite directive matches but cannot safely transform the invocation, the
default behavior is to continue scanning later rules. It does not implicitly
become a reject.

`rewrite` may also set:

- `continue: true`

When enabled, a successful rewrite restarts evaluation from the beginning using
the rewritten command. The implementation guards this loop with a small fixed
maximum number of rewrite passes.

## 8. Reject Semantics

`reject` is reserved for invocation shapes that must not pass through unchanged.

Typical examples:

- wrappers that destroy permission intent and cannot be safely normalized
- unsafe shell payloads that cannot be unwrapped into a single direct command
- invocation forms that a team wants to forbid regardless of downstream allow
  settings

## 9. Consequences Of First-Match Selection

Because first-match selection is part of the contract:

- rule order is meaningful
- rewrite chains are explicit rather than implicit
- tests can deterministically assert the applied directive
- one rule can shadow later rules

Shadowing remains a diagnostic concern rather than a runtime error.
