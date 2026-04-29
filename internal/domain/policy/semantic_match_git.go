package policy

import (
	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "git",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.Git().matches(cmd) },
		validate: ValidateGitSemanticMatchSpec,
	})
}

func (s GitSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "git" || cmd.Git == nil {
		return false
	}
	git := cmd.Git
	if s.Verb != "" && git.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, git.Verb) {
		return false
	}
	if s.Remote != "" && git.Remote != s.Remote {
		return false
	}
	if len(s.RemoteIn) > 0 && !containsString(s.RemoteIn, git.Remote) {
		return false
	}
	if s.Branch != "" && git.Branch != s.Branch {
		return false
	}
	if len(s.BranchIn) > 0 && !containsString(s.BranchIn, git.Branch) {
		return false
	}
	if s.Ref != "" && git.Ref != s.Ref {
		return false
	}
	if len(s.RefIn) > 0 && !containsString(s.RefIn, git.Ref) {
		return false
	}
	if s.Force != nil && git.Force != *s.Force {
		return false
	}
	if s.ForceWithLease != nil && git.ForceWithLease != *s.ForceWithLease {
		return false
	}
	if s.ForceIfIncludes != nil && git.ForceIfIncludes != *s.ForceIfIncludes {
		return false
	}
	if s.Hard != nil && git.Hard != *s.Hard {
		return false
	}
	if s.Recursive != nil && git.Recursive != *s.Recursive {
		return false
	}
	if s.IncludeIgnored != nil && git.IncludeIgnored != *s.IncludeIgnored {
		return false
	}
	if s.Cached != nil && git.Cached != *s.Cached {
		return false
	}
	if s.Staged != nil && git.Staged != *s.Staged {
		return false
	}
	for _, flag := range s.FlagsContains {
		if !containsString(git.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(git.Flags, prefix) {
			return false
		}
	}
	return true
}
