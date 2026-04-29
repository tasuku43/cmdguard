package policy

import commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "helm",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.Helm().matches(cmd) },
		validate: ValidateHelmSemanticMatchSpec,
	})
}

func (s HelmSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "helm" || cmd.Helm == nil {
		return false
	}
	h := cmd.Helm
	if s.Verb != "" && h.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, h.Verb) {
		return false
	}
	if s.Subverb != "" && h.Subverb != s.Subverb {
		return false
	}
	if len(s.SubverbIn) > 0 && !containsString(s.SubverbIn, h.Subverb) {
		return false
	}
	if s.Release != "" && h.Release != s.Release {
		return false
	}
	if s.Chart != "" && h.Chart != s.Chart {
		return false
	}
	if len(s.ChartIn) > 0 && !containsString(s.ChartIn, h.Chart) {
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
	if s.Kubeconfig != "" && h.Kubeconfig != s.Kubeconfig {
		return false
	}
	if s.DryRun != nil && h.DryRun != *s.DryRun {
		return false
	}
	if s.Force != nil && h.Force != *s.Force {
		return false
	}
	if s.Atomic != nil && h.Atomic != *s.Atomic {
		return false
	}
	if s.Wait != nil && h.Wait != *s.Wait {
		return false
	}
	if s.WaitForJobs != nil && h.WaitForJobs != *s.WaitForJobs {
		return false
	}
	if s.Install != nil && h.Install != *s.Install {
		return false
	}
	if s.ReuseValues != nil && h.ReuseValues != *s.ReuseValues {
		return false
	}
	if s.ResetValues != nil && h.ResetValues != *s.ResetValues {
		return false
	}
	if s.ResetThenReuseValues != nil && h.ResetThenReuseValues != *s.ResetThenReuseValues {
		return false
	}
	if s.CleanupOnFail != nil && h.CleanupOnFail != *s.CleanupOnFail {
		return false
	}
	if s.CreateNamespace != nil && h.CreateNamespace != *s.CreateNamespace {
		return false
	}
	if s.DependencyUpdate != nil && h.DependencyUpdate != *s.DependencyUpdate {
		return false
	}
	if s.Devel != nil && h.Devel != *s.Devel {
		return false
	}
	if s.KeepHistory != nil && h.KeepHistory != *s.KeepHistory {
		return false
	}
	if s.Cascade != "" && h.Cascade != s.Cascade {
		return false
	}
	if len(s.CascadeIn) > 0 && !containsString(s.CascadeIn, h.Cascade) {
		return false
	}
	if s.ValuesFile != "" && !containsString(h.ValuesFiles, s.ValuesFile) {
		return false
	}
	if len(s.ValuesFileIn) > 0 && !containsAnyString(h.ValuesFiles, s.ValuesFileIn) {
		return false
	}
	for _, value := range s.ValuesFilesContains {
		if !containsString(h.ValuesFiles, value) {
			return false
		}
	}
	for _, key := range s.SetKeysContains {
		if !containsString(h.SetKeys, key) {
			return false
		}
	}
	for _, key := range s.SetStringKeysContains {
		if !containsString(h.SetStringKeys, key) {
			return false
		}
	}
	for _, key := range s.SetFileKeysContains {
		if !containsString(h.SetFileKeys, key) {
			return false
		}
	}
	if s.RepoName != "" && h.RepoName != s.RepoName {
		return false
	}
	if s.RepoURL != "" && h.RepoURL != s.RepoURL {
		return false
	}
	if s.Registry != "" && h.Registry != s.Registry {
		return false
	}
	if s.PluginName != "" && h.PluginName != s.PluginName {
		return false
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
