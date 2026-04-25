# User Documentation

Start here if your goal is to use `cc-bash-proxy` in local workflows.

## Entry

- `docs/user/START_HERE.md`

## Current Focus

- user-wide config at `~/.config/cc-bash-proxy/cc-bash-proxy.yml`
- rule editing followed by `cc-bash-proxy verify`
- hook integration via `cc-bash-proxy hook`
- semantic match discovery with `cc-bash-proxy help semantic` and
  `cc-bash-proxy help semantic <command>`

## Semantic Rule Help

Permission rules use `command`, `env`, and `patterns`. `pattern` does not
exist, permission `match` does not exist, and top-level `rewrite` is not a
user-facing feature. Permission `command` does not support `command_in`.

`cc-bash-proxy` never rewrites the command that is executed. Parser-backed
normalization is evaluation-only: shell `-c` wrappers are inspected as inner
commands, absolute command paths match by basename, and AWS profile flags are
read by the AWS semantic parser.

`command.semantic` is command-specific and selected by exact `command.name`.
Supported semantic commands are exposed by:

```sh
cc-bash-proxy help semantic
cc-bash-proxy help semantic git
cc-bash-proxy semantic-schema --format json
```

Use `patterns` for raw command regex matching. Use
`semantic.flags_contains` / `semantic.flags_prefixes` for flags recognized by a
command-specific semantic parser.

## AWS Command Style

Prefer documenting AWS commands with environment-prefixed profiles:

```sh
AWS_PROFILE=myprof aws eks list-clusters
```

Avoid relying on cc-bash-proxy to convert:

```sh
aws --profile myprof eks list-clusters
```

`cc-bash-proxy` does not perform that conversion. It evaluates AWS
`profile`, `service`, and `operation` through the semantic parser. Ambiguous or
dangerous command styles should be handled with `ask` or `deny` policy.

`rtk` integration is outside this guide.

## Intended Guide Set

- `RULES.md`: writing directive-based rules safely
- `CLAUDE_CODE.md`: Claude Code hook usage and permission layering
- `SHELL.md`: shell and CI integration patterns

These guides are not written yet, but this directory remains the intended home
for user-facing documentation.
