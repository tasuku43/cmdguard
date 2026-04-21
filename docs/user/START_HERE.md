# Start Here

`cmdproxy` is a local CLI that sits in front of command execution and enforces
policy-approved invocation shape.

Its main job is to normalize command shape so downstream permission systems
evaluate the invocation you intended, not a drifted wrapper-heavy form.

## Quick Start

1. Create the user config

```sh
cmdproxy init
```

2. Edit `~/.config/cmdproxy/cmdproxy.yml`

3. Verify the config after each change

```sh
cmdproxy verify
cmdproxy doctor --format json
```

4. Spot-check individual commands

```sh
cmdproxy check aws --profile read-only-profile s3 ls
cmdproxy check bash -c 'git status'
```

5. Register `cmdproxy hook claude` in your hook runner

## Verifying an Installed Binary

If you install `cmdproxy` from a release artifact, verify it before relying on
it in your command path.

1. Check the downloaded file against `checksums.txt`
2. Verify the release provenance with GitHub attestation data
3. Inspect the binary metadata
4. Run `cmdproxy verify`

Example:

```sh
shasum -a 256 -c checksums.txt
gh attestation verify path/to/cmdproxy_<tag>_<os>_<arch>.tar.gz -R tasuku43/cmdguard
cmdproxy version --format json
cmdproxy verify --format json
```

## Claude Code

For Claude Code, add `cmdproxy hook claude --rtk` as a `PreToolUse` Bash hook.

```json
{
  "matcher": "Bash",
  "hooks": [
    { "type": "command", "command": "cmdproxy hook claude --rtk" }
  ]
}
```

`cmdproxy hook claude --rtk` does not depend on Bash hook ordering. It rewrites
with `cmdproxy`, evaluates Claude permissions against that rewritten command,
then applies the final `rtk` rewrite before returning `updatedInput`.

## Current Rule Model

- rules use `match` or `pattern`
- rules use one directive: `rewrite` or `reject`
- rewrite rules may opt into `strict: false` for relaxed built-in contracts
- tests live under the directive
- `rewrite.test.expect` uses `in` / `out`
- `reject.test.expect` uses string inputs
- both directive kinds use `test.pass`

If you are contributing to the implementation, start from
`docs/dev/README.md` instead.
