---
title: "Product Concept: Bash Permission Policy"
status: proposed
date: 2026-04-22
---

# Product Concept

## 1. Purpose

This document describes the product positioning for `cc-bash-guard`.
User-facing current behavior remains defined by `README.md` and `docs/user/`.

`cc-bash-guard` is declarative, testable Bash permission policy for Claude
Code. It evaluates commands against YAML policy and returns `allow`, `ask`, or
`deny`.

## 2. Problem

AI coding agents can express the same intended command in multiple shell forms.
Prefix and wildcard permissions can miss equivalent forms or become broad
enough to allow more than intended.

Examples:

1. the right CLI invoked through an absolute path
2. the right CLI invoked through a shell `-c` wrapper
3. the right CLI invoked with global flags before the subcommand
4. a raw wildcard rule covering destructive subcommands by accident

Runtime permission systems often sit too low in the stack to express these
rules clearly. They can tell that `aws`, `git`, or `kubectl` ran, but not
whether the invocation matched the intended operation.

## 3. Product Thesis

`cc-bash-guard` should be policy-as-code for Claude Code Bash permissions.

The key design idea is:

- users describe permission policy in declarative YAML
- semantic parsers expose command meaning for supported CLIs
- examples are tests that must pass with `cc-bash-guard verify`
- `cc-bash-guard explain` makes decisions inspectable
- hook execution relies on verified artifacts

Default policy evaluation does not rewrite commands. Parser-backed
normalization is evaluation-only. Command rewriting exists only through the
explicit `cc-bash-guard hook --rtk` bridge to external RTK rewriting.

## 4. Primary Persona

**Operators of AI-agent shell execution**

- run Claude Code, shell hooks, CI wrappers, or similar systems
- want flexible local permission policy without ad-hoc shell glue
- want semantic command policy that can distinguish read-only and destructive
  operations
- need policies that humans and coding agents can test, explain, and maintain

## 5. Core Value Proposition

`cc-bash-guard` makes Claude Code Bash permissions declarative, semantic, and
testable.

That value appears in five concrete ways:

1. **Semantic Matching**
   Match command meaning such as `git` verb or `aws` service and operation.
2. **Fail-Closed Permission Evaluation**
   Decide `deny`, `ask`, `allow`, or `abstain` without making ambiguity
   permissive.
3. **Test-First Policy**
   Use YAML examples with `verify` to pin safe commands and near misses.
4. **Explainability**
   Inspect parse shape, semantic fields, matched rules, Claude settings, and
   final decisions.
5. **Verified Artifacts**
   Use reviewed, verified effective policy at hook time.

## 6. Operating Model

`cc-bash-guard` runs as a local CLI in front of Claude Code Bash execution.

- the caller provides a requested command string
- `cc-bash-guard` parses that command internally
- permission buckets are evaluated in order `deny -> ask -> allow`
- no-match is `abstain`
- permission sources merge as `deny > ask > allow > abstain`
- the final fallback is `ask` when all sources abstain
- the resulting decision is returned to the caller

The mental model is a local permission policy layer, not a deny-list filter and
not a command rewriting engine.

## 7. Non-Goals

1. Acting as a general shell interpreter or full shell AST executor
2. Providing malware detection
3. Providing OS, filesystem, network, or credential sandboxing
4. Rewriting commands in default hook mode
5. Hosting policy centrally as a network control plane
