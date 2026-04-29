package policy

import (
	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "helmfile",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.Helmfile().matches(cmd) },
		validate: ValidateHelmfileSemanticMatchSpec,
	})
}

func (s HelmfileSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "helmfile" || cmd.Helmfile == nil {
		return false
	}
	h := cmd.Helmfile
	if s.Verb != "" && h.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, h.Verb) {
		return false
	}
	if s.Environment != "" && h.Environment != s.Environment {
		return false
	}
	if len(s.EnvironmentIn) > 0 && !containsString(s.EnvironmentIn, h.Environment) {
		return false
	}
	if s.EnvironmentMissing != nil && (h.Environment == "") != *s.EnvironmentMissing {
		return false
	}
	if s.File != "" && !containsString(h.Files, s.File) {
		return false
	}
	if len(s.FileIn) > 0 && !containsAnyString(h.Files, s.FileIn) {
		return false
	}
	if s.FilePrefix != "" && !containsPrefix(h.Files, s.FilePrefix) {
		return false
	}
	if s.FileMissing != nil && (len(h.Files) == 0) != *s.FileMissing {
		return false
	}
	if s.Namespace != "" && h.Namespace != s.Namespace {
		return false
	}
	if len(s.NamespaceIn) > 0 && !containsString(s.NamespaceIn, h.Namespace) {
		return false
	}
	if s.NamespaceMissing != nil && (h.Namespace == "") != *s.NamespaceMissing {
		return false
	}
	if s.KubeContext != "" && h.KubeContext != s.KubeContext {
		return false
	}
	if len(s.KubeContextIn) > 0 && !containsString(s.KubeContextIn, h.KubeContext) {
		return false
	}
	if s.KubeContextMissing != nil && (h.KubeContext == "") != *s.KubeContextMissing {
		return false
	}
	if s.Selector != "" && !containsString(h.Selectors, s.Selector) {
		return false
	}
	if len(s.SelectorIn) > 0 && !containsAnyString(h.Selectors, s.SelectorIn) {
		return false
	}
	for _, value := range s.SelectorContains {
		if !containsSubstring(h.Selectors, value) {
			return false
		}
	}
	if s.SelectorMissing != nil && (len(h.Selectors) == 0) != *s.SelectorMissing {
		return false
	}
	if s.Interactive != nil && h.Interactive != *s.Interactive {
		return false
	}
	if s.DryRun != nil {
		if h.DryRun == nil || *h.DryRun != *s.DryRun {
			return false
		}
	}
	if s.Wait != nil && h.Wait != *s.Wait {
		return false
	}
	if s.WaitForJobs != nil && h.WaitForJobs != *s.WaitForJobs {
		return false
	}
	if s.SkipDiff != nil && h.SkipDiff != *s.SkipDiff {
		return false
	}
	if s.SkipNeeds != nil && h.SkipNeeds != *s.SkipNeeds {
		return false
	}
	if s.IncludeNeeds != nil && h.IncludeNeeds != *s.IncludeNeeds {
		return false
	}
	if s.IncludeTransitiveNeeds != nil && h.IncludeTransitiveNeeds != *s.IncludeTransitiveNeeds {
		return false
	}
	if s.Purge != nil && h.Purge != *s.Purge {
		return false
	}
	if s.Cascade != "" && h.Cascade != s.Cascade {
		return false
	}
	if len(s.CascadeIn) > 0 && !containsString(s.CascadeIn, h.Cascade) {
		return false
	}
	if s.DeleteWait != nil && h.DeleteWait != *s.DeleteWait {
		return false
	}
	if s.StateValuesFile != "" && !containsString(h.StateValuesFiles, s.StateValuesFile) {
		return false
	}
	if len(s.StateValuesFileIn) > 0 && !containsAnyString(h.StateValuesFiles, s.StateValuesFileIn) {
		return false
	}
	for _, key := range s.StateValuesSetKeysContains {
		if !containsString(h.StateValuesSetKeys, key) {
			return false
		}
	}
	for _, key := range s.StateValuesSetStringKeysContains {
		if !containsString(h.StateValuesSetStringKeys, key) {
			return false
		}
	}
	for _, flag := range s.FlagsContains {
		if !containsString(h.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(h.Flags, prefix) {
			return false
		}
	}
	return true
}
