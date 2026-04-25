---
title: "cc-bash-proxy hook"
status: proposed
date: 2026-04-26
---

# cc-bash-proxy hook

`cc-bash-proxy hook` is the Claude Code hook entrypoint. It reads the Claude
hook payload from stdin, evaluates permission policy, and emits hook JSON on
stdout.

Runtime flow:

1. parse Claude Code `PreToolUse` Bash payload
2. load the verified effective policy artifact
3. parse the original command string into a `CommandPlan`
4. evaluate `cc-bash-proxy` permission policy
5. merge with Claude settings according to `claude_permission_merge_mode`
6. emit `allow`, `ask`, `deny`, or error output

`cc-bash-proxy` does not emit `updatedInput.command` for policy evaluation,
because it does not rewrite commands. Parser-backed normalization is
evaluation-only.

If `--rtk` is enabled, `rtk rewrite` is invoked after `cc-bash-proxy` permission
evaluation. rtk behavior is a compatibility path outside this permission-only
policy model.
