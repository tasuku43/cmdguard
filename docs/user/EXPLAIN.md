# Explain Output

`cc-bash-guard explain` inspects a command without executing it. It uses the
same verified artifact as `cc-bash-guard hook`, so run `cc-bash-guard verify`
after editing policy or included files.

```sh
cc-bash-guard explain "bash -c 'git status'"
cc-bash-guard explain --format json "git push --force origin main"
```

## What To Look For

- parsed shell shape: how the shell command was classified
- shape flags: parser-derived flags such as redirects or unsafe shell shapes
- normalized command names: for example `/usr/bin/git` matching `git`
- evaluated inner command: for supported shell `-c` wrappers
- semantic fields: command-specific fields such as `git` `verb` and `force`
- policy outcome: cc-bash-guard policy result, including `abstain`
- Claude settings outcome: native Claude Code permission contribution
- final outcome: merged hook decision
- trace: why each decision step happened

`abstain` means "no matching rule" or "no opinion" for one permission source.
The final hook output never remains `abstain`: if cc-bash-guard policy and
Claude settings both abstain, the final decision falls back to `ask`.

## Common Reads

When policy allows and Claude settings have no matching rule:

```text
policy: allow
claude_settings: abstain
final: allow
```

When neither policy nor Claude settings has a matching rule:

```text
policy: abstain
claude_settings: abstain
final: ask
```

This is the expected conservative fallback. Add a semantic `allow`, `ask`, or
`deny` rule plus tests when you want the command to have a more specific
decision.

## Shell Wrappers

For supported shell `-c` wrappers, `explain` shows that the inner command was
evaluated for policy. For example, `bash -c 'git status'` can match the same
semantic `git` rule as `git status`.

This is not command rewriting. The default hook still passes the original
command through unchanged.
