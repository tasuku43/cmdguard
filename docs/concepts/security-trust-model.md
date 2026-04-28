---
title: "Security Trust Model"
status: proposed
date: 2026-04-20
---

# Security Trust Model

## 1. Why This Matters

`cc-bash-guard` participates in the local execution trust boundary for Claude
Code Bash tool calls. It evaluates commands against permission policy before
Claude Code decides whether to allow, ask for confirmation, or deny execution.

Default `cc-bash-guard` policy evaluation does not rewrite commands. Parser
normalization is evaluation-only. Rewriting is only possible through the
explicit `cc-bash-guard hook --rtk` bridge, which delegates to external RTK
after permission evaluation and never for `deny`.

If a malicious change lands in the distributed binary, the risk is incorrect
permission decisions, weakened fail-closed behavior, stale artifact acceptance,
or misleading explain/verify output. For that reason, users should treat
`cc-bash-guard` as part of their local execution trust boundary.

## 2. Primary Threat Model

The highest-priority threat for `cc-bash-guard` is **binary or implementation
tampering**.

This includes:

1. a malicious contribution that changes permission evaluation
2. a compromised release process that ships different code than the reviewed
   repository state
3. a local replacement of the expected binary after installation
4. a change that makes parser uncertainty, stale artifacts, or unknown shell
   shapes more permissive

This document does **not** treat user-authored policy rules as the main threat.
Dangerous local rules are still possible, but the primary publication question
is whether users can trust the shipped tool itself.

## 3. Security Principles

`cc-bash-guard` should follow these principles:

1. **Policy-only default**
   Default hook execution evaluates permissions and does not rewrite commands.
2. **Deterministic behavior**
   A reviewed rule set and reviewed binary should produce predictable results.
3. **Fail closed**
   Permission ambiguity should become `ask` or `deny`, never broad `allow`.
4. **Explainable decisions**
   `verify` and `explain` should expose parse shape, semantic fields, matched
   rules, Claude settings, and final decisions.
5. **Verifiable distribution**
   Users should be able to verify what binary they installed and what source
   revision it came from.

## 4. Required Publication Controls

Before wider publication, the project should adopt or preserve the following
baseline:

### Repository and Review

- protect the default branch
- require PR review before merge
- add `CODEOWNERS` for security-sensitive paths
- treat changes in hook handling, parser behavior, config loading, verified
  artifacts, and rule evaluation as security-sensitive

### Release Integrity

- publish release checksums for every binary artifact
- expose build metadata such as version and commit
- prefer signed releases or artifact attestations over unsigned binaries
- document the recommended verified installation path

### Runtime Verification

- make it easy to inspect the installed binary identity
- make it easy to inspect the hook command Claude Code is actually executing
- require `cc-bash-guard verify` to produce hook artifacts before default hook
  execution trusts policy
- keep `explain` output useful for policy review

## 5. Non-Goals

This trust model does not attempt to solve:

- low-level runtime sandboxing
- credential compromise outside `cc-bash-guard`
- arbitrary shell escape paths that should be handled by the downstream runtime
- centralized remote policy enforcement
- malware detection

`cc-bash-guard` should remain a local Bash permission policy proxy that users
can reasonably trust as part of their execution boundary.
