---
title: "cc-bash-guard init"
status: implemented
date: 2026-04-28
---

# cc-bash-guard init

## Purpose

`cc-bash-guard init` bootstraps a local `cc-bash-guard` setup without destructively
modifying existing user configuration.

## Responsibilities

`cc-bash-guard init` should:

- create a starter user-wide config when one does not exist
- create a built-in profile config when `--profile <name>` is specified
- list built-in profiles with `--list-profiles`
- explain where the user-wide config lives
- detect compatible Claude Code settings files
- print the hook snippet needed to register `cc-bash-guard hook`

Supported profiles:

- `balanced`
- `strict`
- `git-safe`
- `aws-k8s`
- `argocd`

`init` without `--profile` preserves the legacy starter config.

## Starter Config Goal

The starter config should reflect the new product identity.

It should:

- use the current schema shape without requiring an in-file version field
- demonstrate semantic `command.name` and `command.semantic` matchers where
  parser support exists
- include examples that show the intended rule effect
- be valid under `cc-bash-guard verify`
- include rule-local tests and top-level tests

## Safety Principle

`init` should remain conservative and idempotent.

- never overwrite an existing user config silently
- prefer showing status and next steps over mutating non-trivial caller config
- keep the generated starter config small and explanatory
