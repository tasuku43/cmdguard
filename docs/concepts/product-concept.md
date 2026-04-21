---
title: "Product Concept: Invocation Policy Proxy"
status: proposed
date: 2026-04-19
---

# Product Concept

## 1. Purpose

This document defines the target product concept for `cmdproxy`.

`cmdproxy` is not primarily a command blocker. It is a local policy proxy that
ensures approved CLIs are invoked in organization-approved ways before a
downstream permission system evaluates the request.

## 2. Problem

In AI-assisted command execution, many operational mistakes come from invocation
drift rather than raw capability drift.

Typical failures look like:

1. the right CLI invoked with the wrong credential shape
2. the right CLI invoked with an unsafe flag or wrapper
3. the right CLI invoked in a way that defeats the caller's permission intent
4. a shell wrapper causing the true command shape to become opaque

Permission systems and sandboxes often operate too low in the stack to express
these rules clearly. They can tell that `aws`, `git`, or `kubectl` ran, but not
whether the invocation respected the team's calling conventions.

## 3. Product Thesis

`cmdproxy` should preserve permission intent by normalizing or rejecting CLI
invocations before they reach the caller's final allow / ask / deny layer.

The key design idea is:

- `CLAUDE.md` or equivalent docs teach the preferred invocation shape
- `cmdproxy` enforces or normalizes that invocation shape
- the downstream runtime keeps final permission authority

## 4. Primary Persona

**Operators of AI-agent shell execution**

- run Claude Code, shell hooks, CI wrappers, or similar systems
- want consistent invocation conventions for approved CLIs
- want to reduce accidental permission prompts caused by malformed command shape
- need a reviewable local tool rather than ad-hoc shell glue

## 5. Core Value Proposition

`cmdproxy` should make approved commands conform to policy-approved invocation
shape so downstream permission systems keep their intended meaning.

That value appears in three concrete ways:

1. **Canonicalization**
   Rewrite valid-but-noncanonical invocations into the approved form.
2. **Intent Preservation**
   Prevent wrapper and flag usage from weakening caller-side permission rules.
3. **Reviewability**
   Keep invocation policy declarative, testable, and portable across runtimes.

## 6. Operating Model

`cmdproxy` runs as a local CLI in front of command execution.

- the caller provides a requested invocation, usually as a raw command string
- `cmdproxy` parses that invocation internally
- ordered rules apply directives such as `rewrite` or `reject`
- the resulting invocation is either passed downstream or rejected

The mental model is closer to `nginx` for CLI invocations than to a deny-list
filter.

## 7. Directive Model

The target directive model is:

- `rewrite`: transform an invocation into a canonical, policy-approved form
- `reject`: stop an invocation that must not pass through unchanged
- implicit `pass`: if nothing matches, forward the original invocation

The primary long-term behavior is `rewrite`, not `reject`.

Initial rewrite primitives should stay narrow and typed. The first useful
examples are:

- `unwrap_shell_dash_c`
- `move_flag_to_env`
- `move_env_to_flag`
- `unwrap_wrapper`

These primitives are intended as policy-preserving canonicalization, not as
free-form command templating.

## 7.1 Support Tiers

`cmdproxy` should expose support depth in tiers rather than imply identical
semantics for every CLI.

- Tier 1: built-in semantic contracts with explicit meaning-preserving mappings
- Tier 2: wrapper and shell normalization only
- Tier 3: no rewrite contract yet; pass or reject only

Early Tier 1 candidates include `aws` and `gh`.
Early Tier 2 candidates include `git`, `docker`, `kubectl`, `npm`, `pnpm`,
`yarn`, `terraform`, and `go`.

## 8. Non-goals

1. Replacing downstream permission engines with a full authorization system
2. Acting as a general shell interpreter or full shell AST executor
3. Providing arbitrary command macros or free-form text transformation
4. Hosting policy centrally as a network control plane
5. Solving every low-level escape path that should instead be handled by
   sandboxing or runtime permissions
