package semantic

var kubectlSchema = Schema{
	Command:     "kubectl",
	order:       30,
	Description: "Kubernetes verb, resource, namespace, context, filename, selector, and container matching.",
	Parser:      "kubectl",
	Fields: []Field{
		stringField("verb", "kubectl verb such as get, apply, delete, or exec."),
		stringListField("verb_in", "Allowed kubectl verbs."),
		stringField("subverb", "Secondary action for compound kubectl commands."),
		stringListField("subverb_in", "Allowed kubectl subverbs."),
		stringField("resource_type", "Kubernetes resource type."),
		stringListField("resource_type_in", "Allowed resource types."),
		stringField("resource_name", "Kubernetes resource name."),
		stringListField("resource_name_in", "Allowed resource names."),
		stringField("namespace", "Namespace selected by -n or --namespace."),
		stringListField("namespace_in", "Allowed namespaces."),
		boolField("namespace_missing", "True when no namespace was selected."),
		stringField("context", "Context selected by --context."),
		stringListField("context_in", "Allowed contexts."),
		stringField("kubeconfig", "Kubeconfig path selected by --kubeconfig."),
		boolField("all_namespaces", "True when -A or --all-namespaces is present."),
		stringField("filename", "Filename selected by -f or --filename."),
		stringListField("filename_in", "Allowed filenames."),
		stringField("filename_prefix", "Filename prefix selected by -f or --filename."),
		stringField("selector", "Selector selected by -l or --selector."),
		stringListField("selector_in", "Allowed selectors."),
		stringListField("selector_contains", "Selectors that must be present."),
		boolField("selector_missing", "True when no selector was selected."),
		stringField("container", "Container selected by -c or --container."),
		boolField("dry_run", "True when --dry-run or a --dry-run value other than none is present; false when --dry-run=none is present; unset when absent."),
		boolField("force", "True when --force is present."),
		boolField("recursive", "True when -R or --recursive is present."),
		stringListField("flags_contains", "Parser-recognized kubectl option tokens that must be present; this does not scan raw argv words."),
		stringListField("flags_prefixes", "Parser-recognized kubectl option tokens that must start with these prefixes; this depends on the kubectl parser."),
	},
	Examples: []Example{
		{Title: "Deny production deletes", YAML: `permission:
  deny:
    - command:
        name: kubectl
        semantic:
          verb: delete
          namespace: prod`},
	},
}

func init() {
	RegisterSchema(kubectlSchema)
}
