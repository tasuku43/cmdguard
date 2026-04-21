---
title: "cmdproxy hook"
status: proposed
date: 2026-04-21
---

# cmdproxy hook

## Purpose

`cmdproxy hook claude` is the Claude Code hook entrypoint. It reads the
Claude Code `PreToolUse` Bash payload from stdin, applies directive-driven
policy evaluation, and emits Claude Code hook JSON on stdout.

## Input Sources

`cmdproxy hook claude` supports:

- Claude Code `PreToolUse` Bash payloads

Unsupported or malformed input is converted into a deny response for the hook
caller.

## Runtime Behavior

The target flow is:

1. Read stdin fully
2. Parse Claude Code hook JSON
3. Normalize the Bash command into an invocation request
4. Load the verified runtime artifact for the current config hash
5. Parse the invocation internally
6. Evaluate rules using first-match directive semantics, including
   `rewrite.continue`
7. Emit Claude Code hook JSON:
   - no output for `pass`
   - `updatedInput` for `rewrite`
   - deny decision for `reject`
   - deny decision for `error`

## Implemented Rewrite Support

The current implementation already supports rewrite outcomes for:

- `rewrite.unwrap_shell_dash_c`
- `rewrite.move_flag_to_env`
- `rewrite.move_env_to_flag`
- `rewrite.unwrap_wrapper`

If a rewrite primitive matches but cannot safely rewrite the invocation,
evaluation continues and the original command may still pass unless a later
`reject` rule matches.

If the source config was edited after the last successful `cmdproxy verify`, the
hook should emit a deny response instructing the caller to run
`cmdproxy verify`.

## Notes

- `check` remains the local interactive command for inspecting one command
  string at a time
- downstream permission systems remain the final execution authority
