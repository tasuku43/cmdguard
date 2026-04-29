package policy

import (
	"strings"

	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "gh",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.GH().matches(cmd) },
		validate: ValidateGhSemanticMatchSpec,
	})
}

func (s GHSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "gh" || cmd.Gh == nil {
		return false
	}
	gh := cmd.Gh
	if s.Area != "" && gh.Area != s.Area {
		return false
	}
	if len(s.AreaIn) > 0 && !containsString(s.AreaIn, gh.Area) {
		return false
	}
	if s.Verb != "" && gh.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, gh.Verb) {
		return false
	}
	if s.Repo != "" && gh.Repo != s.Repo {
		return false
	}
	if len(s.RepoIn) > 0 && !containsString(s.RepoIn, gh.Repo) {
		return false
	}
	if s.Org != "" && gh.Org != s.Org {
		return false
	}
	if len(s.OrgIn) > 0 && !containsString(s.OrgIn, gh.Org) {
		return false
	}
	if s.EnvName != "" && gh.EnvName != s.EnvName {
		return false
	}
	if len(s.EnvNameIn) > 0 && !containsString(s.EnvNameIn, gh.EnvName) {
		return false
	}
	if s.Hostname != "" && gh.Hostname != s.Hostname {
		return false
	}
	if len(s.HostnameIn) > 0 && !containsString(s.HostnameIn, gh.Hostname) {
		return false
	}
	if s.Web != nil && gh.Web != *s.Web {
		return false
	}
	if s.Method != "" && gh.Method != strings.ToUpper(s.Method) {
		return false
	}
	if len(s.MethodIn) > 0 && !containsStringFoldUpper(s.MethodIn, gh.Method) {
		return false
	}
	if s.Endpoint != "" && gh.Endpoint != s.Endpoint {
		return false
	}
	if s.EndpointPrefix != "" && !strings.HasPrefix(gh.Endpoint, s.EndpointPrefix) {
		return false
	}
	for _, value := range s.EndpointContains {
		if !strings.Contains(gh.Endpoint, value) {
			return false
		}
	}
	if s.Paginate != nil && gh.Paginate != *s.Paginate {
		return false
	}
	if s.Input != nil && gh.Input != *s.Input {
		return false
	}
	if s.Silent != nil && gh.Silent != *s.Silent {
		return false
	}
	if s.IncludeHeaders != nil && gh.IncludeHeaders != *s.IncludeHeaders {
		return false
	}
	for _, key := range s.FieldKeysContains {
		if !containsString(gh.FieldKeys, key) {
			return false
		}
	}
	for _, key := range s.RawFieldKeysContains {
		if !containsString(gh.RawFieldKeys, key) {
			return false
		}
	}
	for _, key := range s.HeaderKeysContains {
		if !containsString(gh.HeaderKeys, strings.ToLower(key)) {
			return false
		}
	}
	if s.PRNumber != "" && gh.PRNumber != s.PRNumber {
		return false
	}
	if s.IssueNumber != "" && gh.IssueNumber != s.IssueNumber {
		return false
	}
	if s.SecretName != "" && gh.SecretName != s.SecretName {
		return false
	}
	if len(s.SecretNameIn) > 0 && !containsString(s.SecretNameIn, gh.SecretName) {
		return false
	}
	if s.Tag != "" && gh.Tag != s.Tag {
		return false
	}
	if s.WorkflowName != "" && gh.WorkflowName != s.WorkflowName {
		return false
	}
	if s.WorkflowID != "" && gh.WorkflowID != s.WorkflowID {
		return false
	}
	if s.SearchType != "" && gh.SearchType != s.SearchType {
		return false
	}
	if len(s.SearchTypeIn) > 0 && !containsString(s.SearchTypeIn, gh.SearchType) {
		return false
	}
	if s.QueryContains != "" && !strings.Contains(gh.Query, s.QueryContains) {
		return false
	}
	if s.Base != "" && gh.Base != s.Base {
		return false
	}
	if s.Head != "" && gh.Head != s.Head {
		return false
	}
	if s.Ref != "" && gh.Ref != s.Ref {
		return false
	}
	if len(s.RefIn) > 0 && !containsString(s.RefIn, gh.Ref) {
		return false
	}
	if s.State != "" && gh.State != s.State {
		return false
	}
	if len(s.StateIn) > 0 && !containsString(s.StateIn, gh.State) {
		return false
	}
	if len(s.LabelIn) > 0 && !containsAnyString(gh.Labels, s.LabelIn) {
		return false
	}
	if len(s.AssigneeIn) > 0 && !containsAnyString(gh.Assignees, s.AssigneeIn) {
		return false
	}
	if s.TitleContains != "" && !strings.Contains(gh.Title, s.TitleContains) {
		return false
	}
	if s.BodyContains != "" && !strings.Contains(gh.Body, s.BodyContains) {
		return false
	}
	if s.Draft != nil && gh.Draft != *s.Draft {
		return false
	}
	if s.Prerelease != nil && gh.Prerelease != *s.Prerelease {
		return false
	}
	if s.Latest != nil && gh.Latest != *s.Latest {
		return false
	}
	if s.Fill != nil && gh.Fill != *s.Fill {
		return false
	}
	if s.Force != nil && gh.Force != *s.Force {
		return false
	}
	if s.Admin != nil && gh.Admin != *s.Admin {
		return false
	}
	if s.Auto != nil && gh.Auto != *s.Auto {
		return false
	}
	if s.DeleteBranch != nil && gh.DeleteBranch != *s.DeleteBranch {
		return false
	}
	if s.MergeStrategy != "" && gh.MergeStrategy != s.MergeStrategy {
		return false
	}
	if len(s.MergeStrategyIn) > 0 && !containsString(s.MergeStrategyIn, gh.MergeStrategy) {
		return false
	}
	if s.RunID != "" && gh.RunID != s.RunID {
		return false
	}
	if s.Failed != nil && gh.Failed != *s.Failed {
		return false
	}
	if s.Job != "" && gh.Job != s.Job {
		return false
	}
	if s.Debug != nil && gh.Debug != *s.Debug {
		return false
	}
	if s.ExitStatus != nil && gh.ExitStatus != *s.ExitStatus {
		return false
	}
	for _, flag := range s.FlagsContains {
		if !containsString(gh.Flags, flag) {
			return false
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(gh.Flags, prefix) {
			return false
		}
	}
	return true
}
