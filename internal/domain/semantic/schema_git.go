package semantic

var gitSchema = Schema{
	Command:     "git",
	order:       10,
	Description: "Git operations such as push, clean, reset, diff, checkout, switch, and status.",
	Parser:      "git",
	Fields: []Field{
		stringField("verb", "Git verb parsed after global git options."),
		stringListField("verb_in", "Allowed Git verbs."),
		stringField("remote", "Remote positional for commands such as git push."),
		stringListField("remote_in", "Allowed remotes."),
		stringField("branch", "Branch positional for push, checkout, or switch."),
		stringListField("branch_in", "Allowed branches."),
		stringField("ref", "Ref positional for push, reset, checkout, or switch."),
		stringListField("ref_in", "Allowed refs."),
		boolField("force", "For git push, true only when --force or -f is present. For git clean, true when -f or --force is present."),
		boolField("force_with_lease", "For git push, true when --force-with-lease is present."),
		boolField("force_if_includes", "For git push, true when --force-if-includes is present."),
		boolField("hard", "True for git reset --hard."),
		boolField("recursive", "True for git clean -d."),
		boolField("include_ignored", "True for git clean -x or --ignored."),
		boolField("cached", "True for git diff --cached or --staged."),
		boolField("staged", "True for git diff --cached or --staged."),
		stringListField("flags_contains", "Parser-recognized git option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized git option tokens that must start with these prefixes; this depends on the git parser."),
	},
	Examples: []Example{
		{Title: "Deny destructive force pushes", YAML: `permission:
  deny:
    - command:
        name: git
        semantic:
          verb: push
          force: true`},
	},
	Notes: []string{
		"`force`, `force_with_lease`, and `force_if_includes` are separate git push fields; use all three when a policy should cover every force-like push syntax.",
		"`flags_contains` and `flags_prefixes` inspect parser-recognized option tokens, not raw argv words. GenericParser fallback never satisfies semantic flags.",
	},
}

func init() {
	RegisterSchema(gitSchema)
}
