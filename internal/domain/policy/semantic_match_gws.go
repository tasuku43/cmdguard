package policy

import (
	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "gws",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.Gws().matches(cmd) },
		validate: ValidateGwsSemanticMatchSpec,
	})
}

func (s GwsSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "gws" || cmd.Gws == nil {
		return false
	}
	gws := cmd.Gws
	if s.Service != "" && gws.Service != s.Service {
		return false
	}
	if len(s.ServiceIn) > 0 && !containsString(s.ServiceIn, gws.Service) {
		return false
	}
	if len(s.ResourcePath) > 0 && !equalStrings(gws.ResourcePath, s.ResourcePath) {
		return false
	}
	for _, resource := range s.ResourcePathContains {
		if !containsString(gws.ResourcePath, resource) {
			return false
		}
	}
	if s.Method != "" && gws.Method != s.Method {
		return false
	}
	if len(s.MethodIn) > 0 && !containsString(s.MethodIn, gws.Method) {
		return false
	}
	if s.Helper != nil && gws.Helper != *s.Helper {
		return false
	}
	if s.Mutating != nil && gws.Mutating != *s.Mutating {
		return false
	}
	if s.Destructive != nil && gws.Destructive != *s.Destructive {
		return false
	}
	if s.ReadOnly != nil && gws.ReadOnly != *s.ReadOnly {
		return false
	}
	if s.DryRun != nil && gws.DryRun != *s.DryRun {
		return false
	}
	if s.PageAll != nil && gws.PageAll != *s.PageAll {
		return false
	}
	if s.Upload != nil && gws.Upload != *s.Upload {
		return false
	}
	if s.Sanitize != nil && gws.Sanitize != *s.Sanitize {
		return false
	}
	if s.Params != nil && gws.Params != *s.Params {
		return false
	}
	if s.JSONBody != nil && gws.JSONBody != *s.JSONBody {
		return false
	}
	if s.Unmasked != nil && gws.Unmasked != *s.Unmasked {
		return false
	}
	for _, scope := range s.Scopes {
		if !containsString(gws.Scopes, scope) {
			return false
		}
	}
	for _, flag := range s.FlagsContains {
		if !containsString(gws.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(gws.Flags, prefix) {
			return false
		}
	}
	return true
}
