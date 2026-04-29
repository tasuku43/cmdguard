package policy

import (
	"strings"

	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "aws",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.AWS().matches(cmd) },
		validate: ValidateAWSSemanticMatchSpec,
	})
}

func (s AWSSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "aws" || cmd.AWS == nil {
		return false
	}
	aws := cmd.AWS
	if s.Service != "" && aws.Service != s.Service {
		return false
	}
	if len(s.ServiceIn) > 0 && !containsString(s.ServiceIn, aws.Service) {
		return false
	}
	if s.Operation != "" && aws.Operation != s.Operation {
		return false
	}
	if len(s.OperationIn) > 0 && !containsString(s.OperationIn, aws.Operation) {
		return false
	}
	if s.Profile != "" && aws.Profile != s.Profile {
		return false
	}
	if len(s.ProfileIn) > 0 && !containsString(s.ProfileIn, aws.Profile) {
		return false
	}
	if s.Region != "" && aws.Region != s.Region {
		return false
	}
	if len(s.RegionIn) > 0 && !containsString(s.RegionIn, aws.Region) {
		return false
	}
	if s.EndpointURL != "" && aws.EndpointURL != s.EndpointURL {
		return false
	}
	if s.EndpointURLPrefix != "" && !strings.HasPrefix(aws.EndpointURL, s.EndpointURLPrefix) {
		return false
	}
	if s.DryRun != nil {
		if aws.DryRun == nil || *aws.DryRun != *s.DryRun {
			return false
		}
	}
	if s.NoCLIPager != nil {
		if aws.NoCLIPager == nil || *aws.NoCLIPager != *s.NoCLIPager {
			return false
		}
	}
	for _, flag := range s.FlagsContains {
		if !containsString(aws.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(aws.Flags, prefix) {
			return false
		}
	}
	return true
}
