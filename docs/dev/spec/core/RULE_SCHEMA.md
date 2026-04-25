---
title: "Pipeline Schema"
status: implemented
date: 2026-04-26
---

# Pipeline Schema

Current configs are permission-only.

Top-level keys:

- `claude_permission_merge_mode`
- `permission`
- `test`

Top-level `rewrite` is no longer supported. If present, verification fails with:

```text
top-level rewrite is no longer supported; cc-bash-proxy no longer rewrites commands. Use permission.command / env / patterns, and rely on parser-backed normalization for evaluation.
```

Permission rules use only `command`, `env`, and `patterns`. Singular
`pattern` and permission `match` are not supported.

```yaml
permission:
  deny:
    - name: git force push
      command:
        name: git
        semantic:
          verb: push
          force: true

  allow:
    - name: aws identity
      command:
        name: aws
        semantic:
          service: sts
          operation: get-caller-identity
      env:
        requires:
          - AWS_PROFILE

  ask:
    - name: helm upgrade fallback
      patterns:
        - "^helm\\s+upgrade\\b"
      env:
        requires:
          - KUBECONFIG
```

Rules may combine `command + env` or `patterns + env`. `command + patterns` is
invalid.

Top-level `test` asserts final permission decision only:

```yaml
test:
  - in: "git status"
    decision: allow
```

`rewritten` is not supported because `cc-bash-proxy` does not rewrite commands.
