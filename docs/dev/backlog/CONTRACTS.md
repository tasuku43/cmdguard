---
title: "CONTRACTS backlog"
status: implemented
date: 2026-04-21
---

# CONTRACTS Backlog

This backlog tracks built-in command contract expansion for `cmdproxy`.

The goal is to grow support conservatively: first by widening wrapper/shell
normalization coverage, then by adding narrow semantic mappings only where
meaning preservation is stable and documented.

## P1: Near-term command coverage

- [x] CONTRACT-001: Tier 1 baseline for `aws`
  - What: support `--profile <-> AWS_PROFILE` as a built-in semantic mapping.
  - Specs:
    - `internal/contract/contract.go`

- [x] CONTRACT-002: Tier 1 baseline for `gh`
  - What: support `--repo <-> GH_REPO` as a built-in semantic mapping.
  - Specs:
    - `internal/contract/contract.go`

- [x] CONTRACT-003: Tier 2 baseline for `docker`
  - What: support wrapper and shell normalization only.
  - Specs:
    - `internal/contract/contract.go`

- [x] CONTRACT-004: Tier 2 baseline for `kubectl`
  - What: support wrapper and shell normalization only.
  - Specs:
    - `internal/contract/contract.go`

- [x] CONTRACT-005: Tier 2 baseline for `npm`, `pnpm`, and `yarn`
  - What: support wrapper and shell normalization only.
  - Specs:
    - `internal/contract/contract.go`

- [x] CONTRACT-006: Tier 2 baseline for `terraform` and `go`
  - What: support wrapper and shell normalization only.
  - Specs:
    - `internal/contract/contract.go`

## P2: Next-wave semantic mappings

- [ ] CONTRACT-007: Evaluate `kubectl` context and namespace mappings
  - What: determine whether any documented `kubectl` flag/env mappings are
    stable enough for Tier 1 semantic contracts.
  - Guardrail:
    - do not add undocumented or loosely equivalent mappings

- [ ] CONTRACT-008: Evaluate `aws --region` mappings
  - What: decide whether `--region` can safely become a built-in semantic
    mapping in the same way as `--profile`.
  - Guardrail:
    - confirm the exact env contract before enabling rewrite

- [ ] CONTRACT-009: Add high-impact operation profiles for `docker` and `terraform`
  - What: classify commands such as `docker run`, `docker exec`,
    `terraform apply`, and `terraform destroy` for stricter refusal or review.
  - Guardrail:
    - do not conflate operation risk profiles with meaning-preserving rewrites
