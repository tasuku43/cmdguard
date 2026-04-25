---
title: "Evaluation Model"
status: implemented
date: 2026-04-26
---

# Evaluation Model

`cc-bash-proxy` is a permission-only policy proxy.

Evaluation flow:

1. receive the original command string
2. parse it into a `CommandPlan`
3. apply parser-backed normalization for evaluation only
4. evaluate permission buckets in order: `deny`, `ask`, `allow`
5. return `allow`, `ask`, or `deny`

The command string is not rewritten before execution. `Decision.Command` remains
the original command unless an explicitly external compatibility path such as
`--rtk` is enabled after policy evaluation.

Evaluation-only normalization includes:

- shell `-c` wrapper inspection, including `bash`, `sh`, `zsh`, `dash`, `ksh`,
  `/bin/bash`, `env bash`, `command bash`, `exec sh`, `sudo bash`, `nohup`,
  `timeout`, and `busybox sh`
- basename command matching for absolute command paths
- command-specific semantic parsing, including AWS `--profile`,
  `--profile=value`, and `AWS_PROFILE`

Unsafe shell shapes, parse errors, redirects, background execution, subshells,
command substitution, process substitution, and unknown shapes fail closed and
must not broaden to `allow`.

Compound commands are evaluated through `CommandPlan.Commands`. If any inner
command is denied, the whole command is denied. If all inner commands are
allowed and the composition shape is allowable, the whole command may allow;
otherwise it falls back to `ask`.

Raw regex matching is always `patterns`. Semantic parser support is reserved
for higher-risk command families such as `git`, `gh`, `aws`, `kubectl`, and
`helmfile`. Commands without semantic parsers should be covered with
`patterns`.
