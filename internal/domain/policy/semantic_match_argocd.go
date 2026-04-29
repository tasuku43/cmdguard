package policy

import (
	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "argocd",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.ArgoCD().matches(cmd) },
		validate: ValidateArgoCDSemanticMatchSpec,
	})
}

func (s ArgoCDSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "argocd" || cmd.ArgoCD == nil {
		return false
	}
	a := cmd.ArgoCD
	if s.Verb != "" && a.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, a.Verb) {
		return false
	}
	if s.AppName != "" && a.AppName != s.AppName {
		return false
	}
	if len(s.AppNameIn) > 0 && !containsString(s.AppNameIn, a.AppName) {
		return false
	}
	if s.Project != "" && a.Project != s.Project {
		return false
	}
	if len(s.ProjectIn) > 0 && !containsString(s.ProjectIn, a.Project) {
		return false
	}
	if s.Revision != "" && a.Revision != s.Revision {
		return false
	}
	for _, flag := range s.FlagsContains {
		if !containsString(a.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(a.Flags, prefix) {
			return false
		}
	}
	return true
}
