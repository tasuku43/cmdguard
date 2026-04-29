package semantic

var helmSchema = Schema{
	Command:     "helm",
	order:       75,
	Description: "Helm verb, subverb, release, chart, namespace, kube context, values, set keys, and safety flag matching.",
	Parser:      "helm",
	Fields: []Field{
		stringField("verb", "Top-level Helm command such as install, upgrade, uninstall, status, get, repo, registry, plugin, dependency, package, pull, or push."),
		stringListField("verb_in", "Allowed top-level Helm commands."),
		stringField("subverb", "Second action token for grouped commands such as repo add, registry login, plugin install, dependency update, get values, or show chart."),
		stringListField("subverb_in", "Allowed Helm grouped command subverbs."),
		stringField("release", "Release name for release-oriented commands when it can be identified."),
		stringField("chart", "Chart argument for install, upgrade, template, package, pull, push, verify, and lint when it can be identified."),
		stringListField("chart_in", "Allowed chart arguments."),
		stringField("namespace", "Namespace selected by -n or --namespace."),
		stringListField("namespace_in", "Allowed namespaces."),
		boolField("namespace_missing", "True when no namespace was selected."),
		stringField("kube_context", "Kube context selected by --kube-context."),
		stringListField("kube_context_in", "Allowed kube contexts."),
		boolField("kube_context_missing", "True when no kube context was selected."),
		stringField("kubeconfig", "Kubeconfig selected by --kubeconfig."),
		boolField("dry_run", "True when --dry-run is present."),
		boolField("force", "True when --force is present."),
		boolField("atomic", "True when --atomic is present."),
		boolField("wait", "True when --wait is present."),
		boolField("wait_for_jobs", "True when --wait-for-jobs is present."),
		boolField("install", "True when helm upgrade uses --install or -i."),
		boolField("reuse_values", "True when --reuse-values is present."),
		boolField("reset_values", "True when --reset-values is present."),
		boolField("reset_then_reuse_values", "True when --reset-then-reuse-values is present."),
		boolField("cleanup_on_fail", "True when --cleanup-on-fail is present."),
		boolField("create_namespace", "True when --create-namespace is present."),
		boolField("dependency_update", "True when --dependency-update is present."),
		boolField("devel", "True when --devel is present."),
		boolField("keep_history", "True when uninstall --keep-history is present."),
		stringField("cascade", "Cascade value selected by --cascade."),
		stringListField("cascade_in", "Allowed cascade values."),
		stringField("values_file", "Values file selected by -f or --values."),
		stringListField("values_file_in", "Allowed values files."),
		stringListField("values_files", "Values files that must be present."),
		stringListField("set_keys_contains", "Keys selected by --set that must be present."),
		stringListField("set_string_keys_contains", "Keys selected by --set-string that must be present."),
		stringListField("set_file_keys_contains", "Keys selected by --set-file that must be present."),
		stringField("repo_name", "Repository name for helm repo add/remove."),
		stringField("repo_url", "Repository URL for helm repo add."),
		stringField("registry", "Registry host or URL for helm registry login/logout."),
		stringField("plugin_name", "Plugin argument for helm plugin install/update/uninstall/list when straightforward."),
		stringListField("flags_contains", "Parser-recognized Helm option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized Helm option tokens that must start with these prefixes; this depends on the Helm parser."),
	},
	Examples: []Example{
		{Title: "Allow Helm inspection", YAML: `permission:
  allow:
    - command:
        name: helm
        semantic:
          verb_in: [list, status, history, get, show, search, template, lint]`},
		{Title: "Ask before install or upgrade", YAML: `permission:
  ask:
    - command:
        name: helm
        semantic:
          verb_in: [install, upgrade]`},
	},
}

func init() {
	RegisterSchema(helmSchema)
}
