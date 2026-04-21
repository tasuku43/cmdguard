---
title: "Compatibility and Distribution"
status: implemented
date: 2026-04-18
---

# Compatibility and Distribution

## 1. Scope

This document defines the intended compatibility and distribution stance for
the current `cmdproxy` rule model.

## 2. Rule Schema Stability

The current project intentionally favors a single active schema over explicit
in-file schema version numbers.

- rule files do not carry a `version` field
- breaking schema changes should be documented clearly in release notes and
  migration docs
- the implementation should reject unknown or invalid shapes rather than guess
  compatibility behavior

## 3. Runtime Expectations

`cmdproxy` is intended to run as:

- a local CLI in developer environments
- a hook target for AI-agent and shell integrations
- a CI-time validation or enforcement command

The implementation should favor:

- static binary distribution
- predictable exit codes
- no runtime service dependency

## 4. Distribution Targets

Planned distribution channels are:

- `go install github.com/tasuku43/cmdproxy/cmd/cmdproxy@latest`
- GitHub Releases
- Homebrew tap

Additional package managers are post-v1.

## 5. Platform Stance

`cmdproxy` should target the major developer platforms used for CLI tooling:

- macOS
- Linux
- Windows

The exact release matrix is an implementation detail, but the user-facing docs
should describe the supported installation paths clearly.

## 6. Contract Surface Stance

`cmdproxy` should distinguish between command presence and contract depth.

- adding a CLI to the contract registry does not imply full semantic rewriting
- Tier 1 command contracts may support narrow documented meaning-preserving
  mappings
- Tier 2 command contracts may support only wrapper and shell normalization
- semantic mappings should expand conservatively and rely on documented CLI
  surface area
