package semantic

var helmfileSchema = Schema{
	Command:     "helmfile",
	order:       70,
	Description: "Helmfile apply, sync, destroy, diff, environment, file, selector, namespace, and values matching.",
	Parser:      "helmfile",
	Fields: []Field{
		stringField("verb", "helmfile verb such as apply, sync, destroy, or diff."),
		stringListField("verb_in", "Allowed helmfile verbs."),
		stringField("environment", "Environment selected by -e or --environment."),
		stringListField("environment_in", "Allowed environments."),
		boolField("environment_missing", "True when no environment was selected."),
		stringField("file", "State file selected by -f or --file."),
		stringListField("file_in", "Allowed state files."),
		stringField("file_prefix", "State file prefix."),
		boolField("file_missing", "True when no state file was selected."),
		stringField("namespace", "Namespace selected by --namespace."),
		stringListField("namespace_in", "Allowed namespaces."),
		boolField("namespace_missing", "True when no namespace was selected."),
		stringField("kube_context", "Kube context selected by --kube-context."),
		stringListField("kube_context_in", "Allowed kube contexts."),
		boolField("kube_context_missing", "True when no kube context was selected."),
		stringField("selector", "Selector selected by -l or --selector."),
		stringListField("selector_in", "Allowed selectors."),
		stringListField("selector_contains", "Selectors that must be present."),
		boolField("selector_missing", "True when no selector was selected."),
		boolField("interactive", "True when --interactive is present."),
		boolField("dry_run", "True when --dry-run is present."),
		boolField("wait", "True when --wait is present."),
		boolField("wait_for_jobs", "True when --wait-for-jobs is present."),
		boolField("skip_diff", "True when --skip-diff is present."),
		boolField("skip_needs", "True when --skip-needs is present."),
		boolField("include_needs", "True when --include-needs is present."),
		boolField("include_transitive_needs", "True when --include-transitive-needs is present."),
		boolField("purge", "True when --purge is present."),
		stringField("cascade", "Cascade value selected by --cascade."),
		stringListField("cascade_in", "Allowed cascade values."),
		boolField("delete_wait", "True when --delete-wait is present."),
		stringField("state_values_file", "State values file selected by --state-values-file."),
		stringListField("state_values_file_in", "Allowed state values files."),
		stringListField("state_values_set_keys_contains", "Keys selected by --state-values-set that must be present."),
		stringListField("state_values_set_string_keys_contains", "Keys selected by --state-values-set-string that must be present."),
		stringListField("flags_contains", "Parser-recognized helmfile option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized helmfile option tokens that must start with these prefixes; this depends on the helmfile parser."),
	},
	Examples: []Example{
		{Title: "Ask before production destroy", YAML: `permission:
  ask:
    - command:
        name: helmfile
        semantic:
          verb: destroy
          environment: prod`},
	},
}

func init() {
	RegisterSchema(helmfileSchema)
}
