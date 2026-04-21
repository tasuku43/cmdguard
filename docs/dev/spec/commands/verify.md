---
title: "cmdproxy verify"
status: proposed
date: 2026-04-20
---

# cmdproxy verify

## Purpose

`cmdproxy verify` is a stricter trust-oriented check than `cmdproxy doctor`.

It exists to answer a narrower question:

**Can the current local `cmdproxy` setup be reasonably trusted as part of the
execution path?**

## Behavior

`cmdproxy verify` should:

- run the same config and rule validation used by `doctor`
- compile and write verified hook artifacts for every configured policy file
- require build metadata to be visible in the current binary
- fail if Claude Code settings exist but do not point at `cmdproxy hook claude`
- fail if Claude Code settings use `cmdproxy hook claude` via PATH lookup
  rather than an absolute binary path
- fail if an absolute Claude Code hook target does not exist or is not
  executable
- fail if Claude Code points at a different `cmdproxy` binary than the one
  currently being verified

It should not require Claude Code to be installed. If no Claude settings file is
present, that condition should remain informational rather than fatal.

## Output

### Human-readable

The default output should include:

- the running version
- the visible VCS revision or an explicit missing marker
- the underlying doctor-style checks
- a final verified true/false result
- the artifact cache paths when verification also produced executable hook artifacts

### JSON

`cmdproxy verify --format json` should expose:

- `verified`
- `build_info`
- `report`
- `failures`
- `artifact_built`
- `artifact_cache`

## Relationship To `doctor`

- `doctor` is broad and diagnostic
- `verify` is narrow and trust-oriented

`doctor` may emit warnings that are acceptable in development. `verify` should
promote a smaller set of trust-critical conditions into failures.

## Hook Relationship

`cmdproxy hook claude` reads only verified artifacts at runtime.

- If a verified artifact exists and matches the current config hash, the hook uses it
- If the config changed and no verified artifact is available, the hook should try an implicit verify once
- If that implicit verify still fails, the hook must return a deny response with `invalid_config`
