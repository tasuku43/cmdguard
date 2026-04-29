package policy

import (
	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "kubectl",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.Kubectl().matches(cmd) },
		validate: ValidateKubectlSemanticMatchSpec,
	})
}

func (s KubectlSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "kubectl" || cmd.Kubectl == nil {
		return false
	}
	k := cmd.Kubectl
	if s.Verb != "" && k.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, k.Verb) {
		return false
	}
	if s.Subverb != "" && k.Subverb != s.Subverb {
		return false
	}
	if len(s.SubverbIn) > 0 && !containsString(s.SubverbIn, k.Subverb) {
		return false
	}
	if s.ResourceType != "" && k.ResourceType != s.ResourceType {
		return false
	}
	if len(s.ResourceTypeIn) > 0 && !containsString(s.ResourceTypeIn, k.ResourceType) {
		return false
	}
	if s.ResourceName != "" && k.ResourceName != s.ResourceName {
		return false
	}
	if len(s.ResourceNameIn) > 0 && !containsString(s.ResourceNameIn, k.ResourceName) {
		return false
	}
	if s.Namespace != "" && k.Namespace != s.Namespace {
		return false
	}
	if len(s.NamespaceIn) > 0 && !containsString(s.NamespaceIn, k.Namespace) {
		return false
	}
	if s.NamespaceMissing != nil && (k.Namespace == "") != *s.NamespaceMissing {
		return false
	}
	if s.Context != "" && k.Context != s.Context {
		return false
	}
	if len(s.ContextIn) > 0 && !containsString(s.ContextIn, k.Context) {
		return false
	}
	if s.Kubeconfig != "" && k.Kubeconfig != s.Kubeconfig {
		return false
	}
	if s.AllNamespaces != nil && k.AllNamespaces != *s.AllNamespaces {
		return false
	}
	if s.DryRun != nil {
		if k.DryRun == nil || *k.DryRun != *s.DryRun {
			return false
		}
	}
	if s.Force != nil && k.Force != *s.Force {
		return false
	}
	if s.Recursive != nil && k.Recursive != *s.Recursive {
		return false
	}
	if s.Filename != "" && !containsString(k.Filenames, s.Filename) {
		return false
	}
	if len(s.FilenameIn) > 0 && !containsAnyString(k.Filenames, s.FilenameIn) {
		return false
	}
	if s.FilenamePrefix != "" && !containsPrefix(k.Filenames, s.FilenamePrefix) {
		return false
	}
	if s.Selector != "" && !containsString(k.Selectors, s.Selector) {
		return false
	}
	if len(s.SelectorIn) > 0 && !containsAnyString(k.Selectors, s.SelectorIn) {
		return false
	}
	for _, value := range s.SelectorContains {
		if !containsSubstring(k.Selectors, value) {
			return false
		}
	}
	if s.SelectorMissing != nil && (len(k.Selectors) == 0) != *s.SelectorMissing {
		return false
	}
	if s.Container != "" && k.Container != s.Container {
		return false
	}
	for _, flag := range s.FlagsContains {
		if !containsString(k.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(k.Flags, prefix) {
			return false
		}
	}
	return true
}
