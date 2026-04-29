package policy

import "strings"

func ValidateHelmSemanticMatchSpec(prefix string, semantic SemanticMatchSpec) []string {
	return ValidateHelmSemanticSpec(prefix, semantic.Helm())
}

func ValidateHelmSemanticSpec(prefix string, semantic HelmSemanticSpec) []string {
	var issues []string
	if IsZeroHelmSemanticSpec(semantic) {
		issues = append(issues, prefix+" must not be empty")
	}
	for name, value := range map[string]string{
		"verb":         semantic.Verb,
		"subverb":      semantic.Subverb,
		"release":      semantic.Release,
		"chart":        semantic.Chart,
		"namespace":    semantic.Namespace,
		"kube_context": semantic.KubeContext,
		"kubeconfig":   semantic.Kubeconfig,
		"cascade":      semantic.Cascade,
		"values_file":  semantic.ValuesFile,
		"repo_name":    semantic.RepoName,
		"repo_url":     semantic.RepoURL,
		"registry":     semantic.Registry,
		"plugin_name":  semantic.PluginName,
	} {
		if strings.TrimSpace(value) == "" && value != "" {
			issues = append(issues, prefix+"."+name+" must be non-empty")
		}
	}
	issues = append(issues, validateNonEmptyStrings(prefix+".verb_in", semantic.VerbIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".subverb_in", semantic.SubverbIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".chart_in", semantic.ChartIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".namespace_in", semantic.NamespaceIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".kube_context_in", semantic.KubeContextIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".cascade_in", semantic.CascadeIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".values_file_in", semantic.ValuesFileIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".values_files", semantic.ValuesFilesContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".set_keys_contains", semantic.SetKeysContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".set_string_keys_contains", semantic.SetStringKeysContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".set_file_keys_contains", semantic.SetFileKeysContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".flags_contains", semantic.FlagsContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".flags_prefixes", semantic.FlagsPrefixes)...)
	return issues
}
