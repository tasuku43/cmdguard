---
title: "Compatibility and Distribution"
status: implemented
date: 2026-04-18
---

# Compatibility and Distribution

## 1. Scope

This document defines the intended compatibility and distribution stance for
the current `cc-bash-guard` rule model.

## 2. Rule Schema Stability

The current project intentionally favors a single active schema over explicit
in-file schema version numbers.

- rule files do not carry a `version` field
- breaking schema changes should be documented clearly in release notes and
  migration docs
- the implementation should reject unknown or invalid shapes rather than guess
  compatibility behavior

## 3. Runtime Expectations

`cc-bash-guard` is intended to run as:

- a local CLI in developer environments
- a hook target for AI-agent and shell integrations
- a CI-time validation or enforcement command

The implementation should favor:

- static binary distribution
- predictable exit codes
- no runtime service dependency

## 4. Distribution Targets

Planned distribution channels are:

- Homebrew tap
- mise from GitHub Releases
- GitHub Releases
- `go install github.com/tasuku43/cc-bash-guard/cmd/cc-bash-guard@latest`

User-facing quick-start docs should foreground Homebrew and mise. GitHub
Releases and Go source builds remain documented as secondary paths.

## 5. Platform Stance

`cc-bash-guard` should target the major developer platforms used for CLI tooling:

- macOS
- Linux
- Windows

The exact release matrix is an implementation detail, but the user-facing docs
should describe the supported installation paths clearly.
