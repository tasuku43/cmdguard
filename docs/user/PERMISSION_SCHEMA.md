# Permission Schema

Permission policy is grouped into `deny`, `ask`, and `allow` buckets.

```yaml
include:
  - ./policies/git.yml
  - ./tests/git.yml

permission:
  deny: []
  ask: []
  allow: []
```

## include

Top-level `include` lets you split permission rules and E2E tests across
multiple local YAML files.

```yaml
include:
  - ./policies/git.yml
  - ./policies/aws.yml
  - ./tests/git.yml
```

Rules:

- `include` is top-level only and must be a list of strings.
- Relative paths are resolved from the file that declares the include, not from
  the process working directory.
- Included files may include other files. Cycles fail verification and report
  the include chain.
- Only local files are supported. URLs, shell expansion, environment variable
  expansion, command substitution, and globbing are not supported.
- Empty entries, missing files, and paths that do not resolve to regular files
  are verification errors.
- Symlinks use the same file reading behavior as normal config loading: if the
  operating system resolves the path to a regular file, it is read as that file.

Merge order is deterministic: `include[0]`, `include[1]`, then the current file.
`permission.deny`, `permission.ask`, `permission.allow`, and top-level `test`
lists are concatenated in that order. No deep merge is performed.

`cc-bash-guard verify` resolves includes and writes one effective verified
artifact. The hook uses that artifact for permission evaluation. Included files
are part of the artifact fingerprint, so editing an included file makes the
artifact stale and requires another `cc-bash-guard verify`.

Each rule can use `command`, `env`, and `patterns`.
Each rule may also set `message`; when that rule determines `allow`, `ask`, or
`deny`, the message is returned as Claude Code's permission decision reason.

## command

Use `command` to match a command by name. For commands with semantic parser
support, add `command.semantic`.

```yaml
permission:
  allow:
    - name: git read-only
      command:
        name: git
        semantic:
          verb_in:
            - status
            - diff
            - log
            - show
```

The semantic schema is selected by `command.name`. Inspect supported commands
with:

```sh
cc-bash-guard help semantic
cc-bash-guard semantic-schema --format json
```

## env

Use `env` to require or reject environment variables for the invocation.

```yaml
permission:
  allow:
    - name: AWS identity
      command:
        name: aws
        semantic:
          service: sts
          operation: get-caller-identity
      env:
        requires:
          - AWS_PROFILE
```

`env.requires` means the variable must be present. `env.missing` means the
variable must not be present.

## patterns

Use `patterns` for raw regular expression matching against the original command
string and parsed command elements. Shell `-c` wrappers are unwrapped for
evaluation, so a pattern such as `^aws(\s|$)` also matches
`bash -c 'aws s3 ls'`. This is the fallback for commands without semantic
support.

```yaml
permission:
  allow:
    - name: read-only shell basics
      patterns:
        - "^ls(\\s|$)"
        - "^pwd$"
```

## Valid Combinations

Rules can combine fields as follows:

- `command`
- `command` plus `env`
- `command` plus `semantic`
- `command` plus `semantic` plus `env`
- `patterns`
- `patterns` plus `env`
- `env`

Use semantic matching when a command is listed by `cc-bash-guard help semantic`.
Use `patterns` for commands without semantic support or when raw regex matching
is the intended policy.

## Evaluation Order

`cc-bash-guard` policy and Claude settings permissions are permission sources.
Each source returns `deny`, `ask`, `allow`, or `abstain`.

Decision precedence is:

```text
deny > ask > allow > abstain
```

`abstain` means no matching rule. The final fallback is `ask` only when all
permission sources abstain.

## Command Evaluation

`cc-bash-guard` evaluates commands but does not rewrite them. It only returns a
permission decision: `allow`, `ask`, or `deny`.

Parser-backed normalization is evaluation-only:

- shell `-c` wrappers can be inspected as inner commands
- absolute command paths can match by basename
- command-specific parsers can expose semantic fields such as AWS profile,
  service, and operation

The command returned to Claude Code is not changed by permission evaluation.
