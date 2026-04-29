package semantic

var argoCDSchema = Schema{
	Command:     "argocd",
	order:       50,
	Description: "Argo CD app operations such as app get, list, diff, sync, rollback, and delete.",
	Parser:      "argocd",
	Fields: []Field{
		stringField("verb", "Argo CD action path such as app sync or app rollback."),
		stringListField("verb_in", "Allowed Argo CD action paths."),
		stringField("app_name", "Application name positional for argocd app commands."),
		stringListField("app_name_in", "Allowed Argo CD application names."),
		stringField("project", "Project selected by --project."),
		stringListField("project_in", "Allowed Argo CD projects."),
		stringField("revision", "Revision selected by --revision, or rollback revision positional."),
		stringListField("flags_contains", "Parser-recognized argocd option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized argocd option tokens that must start with these prefixes; this depends on the argocd parser."),
	},
	Examples: []Example{
		{Title: "Ask before syncing an app", YAML: `permission:
  ask:
    - command:
        name: argocd
        semantic:
          verb: app sync`},
	},
}

func init() {
	RegisterSchema(argoCDSchema)
}
