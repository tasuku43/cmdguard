---
title: "cc-bash-guard hook"
status: proposed
date: 2026-04-26
---

# cc-bash-guard hook

`cc-bash-guard hook` is the Claude Code hook entrypoint. It reads the Claude
hook payload from stdin, evaluates permission policy, and emits hook JSON on
stdout.

Runtime flow:

1. parse Claude Code `PreToolUse` Bash payload
2. load the verified effective policy artifact
3. parse the original command string into a `CommandPlan`
4. evaluate `cc-bash-guard` permission policy
5. merge `cc-bash-guard` policy with Claude settings as permission sources
   using `deny > ask > allow > abstain`
6. when `--rtk` is enabled and the merged decision is not `deny`, invoke
   external `rtk rewrite` once and apply the returned command as
   `updatedInput.command`
7. emit `allow`, `ask`, `deny`, or error output

`abstain` means a source had no matching rule. The final fallback is `ask` only
when all sources abstain.

`cc-bash-guard` does not emit `updatedInput.command` for policy evaluation.
Parser-backed normalization is evaluation-only.

`cc-bash-guard hook` does not rewrite commands itself. `--rtk` is an explicit
integration path for installations that use RTK rewriting: cc-bash-guard
evaluates permissions first, then delegates rewriting to external RTK in the
same hook invocation.
