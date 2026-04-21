---
title: "Verified Hook Artifact"
status: proposed
date: 2026-04-21
---

# Verified Hook Artifact

## Purpose

`cmdproxy` should not execute hook-time policy directly from the human-edited
YAML source config.

Instead, `cmdproxy verify` compiles the current config into a machine-only JSON
artifact. `cmdproxy hook claude` reads only that artifact.

## Required Fields

The runtime artifact must carry at least:

- `version`
- `source_path`
- `source_hash`
- `cmdproxy_version`
- `verified_at`
- `compiled_rules`

## Runtime Gate

`cmdproxy hook claude` should:

1. hash the current source config
2. look for the artifact matching that hash
3. reject execution if the artifact is missing or stale
4. evaluate only the compiled rules from that artifact

## Non-Goals

- human readability
- user editing
- long-term compatibility across arbitrary schema generations

The artifact is an internal compiled runtime format, not a public config
surface.

## Contract Metadata

The artifact does not need to expose human-readable support tiers, but the
compiled rules inside it may depend on built-in contracts from multiple tiers.

At minimum, the runtime should preserve the already-validated rewrite specs.
Future versions may include contract metadata when that materially improves
runtime diagnostics.
