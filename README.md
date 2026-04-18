# cmdguard

Declarative, testable command-string policy engine for AI agents and shells.

> **Status:** design phase. See [`docs/dev/spec/README.md`](docs/dev/spec/README.md)
> for the v1.0 implementation contracts and
> [`docs/concepts/product-concept.md`](docs/concepts/product-concept.md) for the
> product concept.

## What it does

`cmdguard` is a tiny hook that decides whether a shell command is allowed to
run. It is called from Claude Code `PreToolUse`, `zsh` `preexec`,
`pre-commit`, CI, or anywhere else a command-string policy is useful.

Rules are declared in YAML. Every rule ships with block/allow examples, and
`cmdguard test` runs them as unit tests — so a rule change that would let
through a command it used to block fails CI, not production.

```yaml
# .cmdguard.yml
version: 1
rules:
  - id: no-git-dash-c
    pattern: '^\s*git\s+-C\b'
    message: "git -C は禁止。cd で移動してから実行してください。"
    block_examples:
      - "git -C repos/foo status"
    allow_examples:
      - "git status"
      - "# git -C in comment"
```

## Non-goals

- LLM-assisted rule authoring and transcript mining live in a separate
  [`cmdguard-claude-plugin`] repository, so the core CLI has no LLM
  dependency.
- Non-`exec` action types (`write`, `fetch`, `mcp_call`) are post-v1.

See [`docs/README.md`](docs/README.md) for the current documentation map.

## Install

Not yet released. Once v1 ships:

```sh
brew install tasuku43/tap/cmdguard
# or
go install github.com/tasuku43/cmdguard/cmd/cmdguard@latest
```

## License

MIT.
