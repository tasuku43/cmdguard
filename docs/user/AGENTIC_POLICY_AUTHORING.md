# Agentic Policy Authoring

`cc-bash-guard verify` and `cc-bash-guard explain` let humans and coding agents
author policy test-first. The loop is simple: edit YAML, add examples, run
`verify`, inspect `explain`, then refine the policy.

## Prompt Claude Code

Give Claude Code a concrete policy-maintenance task:

```text
Edit ~/.config/cc-bash-guard/cc-bash-guard.yml.

Add a rule that denies destructive Argo CD app deletion, but does not match
argocd app get or argocd app sync.

Add tests for both matching and non-matching commands.
Run cc-bash-guard verify.
Use cc-bash-guard explain to confirm the parsed semantic fields and final decision.
Iterate until verification passes.
```

This works because the policy is declarative and the examples are executable
tests, not prose-only guidance.

## Start With A Small Policy

```yaml
permission:
  allow:
    - name: git status
      command:
        name: git
        semantic:
          verb: status
      test:
        allow:
          - "git status"
        abstain:
          - "git push --force origin main"

test:
  allow:
    - "git status"
    - "/usr/bin/git status"
    - "bash -c 'git status'"
    - "git -C repo status"
  ask:
    - "git push --force origin main"
```

Run:

```sh
cc-bash-guard verify
```

When it passes, the verified artifact is written for hook execution.

## Inspect A Near Miss

Use `explain` without executing the command:

```sh
cc-bash-guard explain "git push --force origin main"
```

The output shows the parsed command, semantic fields such as `verb` and
`force`, matched cc-bash-guard rule if any, Claude settings contribution, and
the final merged decision. If no rule matches and Claude settings also abstain,
the final fallback is `ask`.

## Fix A Failed Test

Suppose you expected force push to be denied:

```yaml
test:
  deny:
    - "git push --force origin main"
```

`verify` fails because the policy above only allows `git status`; it does not
deny force push. Add a semantic deny rule:

```yaml
permission:
  deny:
    - name: block force push
      command:
        name: git
        semantic:
          verb: push
          force: true
      test:
        deny:
          - "git push --force origin main"
        abstain:
          - "git push origin main"
```

Run `cc-bash-guard verify` again. Then inspect the decision:

```sh
cc-bash-guard explain "git push --force origin main"
```

This workflow keeps policy reviewable: the YAML states intent, tests pin
expected behavior, `verify` catches mismatches, and `explain` shows why a
command is allowed, denied, asked, or left as `abstain` by cc-bash-guard.
