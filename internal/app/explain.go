package app

import (
	"errors"
	"sort"
	"strings"

	"github.com/tasuku43/cc-bash-guard/internal/adapter/claude"
	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
	"github.com/tasuku43/cc-bash-guard/internal/domain/policy"
	configrepo "github.com/tasuku43/cc-bash-guard/internal/infra/config"
)

type ExplainResult struct {
	Command        string                `json:"command"`
	Parsed         ExplainParsed         `json:"parsed"`
	Policy         ExplainSourceDecision `json:"policy"`
	ClaudeSettings ExplainSourceDecision `json:"claude_settings"`
	Final          ExplainFinalDecision  `json:"final"`
	Trace          []policy.TraceStep    `json:"trace,omitempty"`
}

type ExplainParsed struct {
	Shape          string                 `json:"shape"`
	ShapeFlags     []string               `json:"shape_flags,omitempty"`
	Diagnostics    []string               `json:"diagnostics,omitempty"`
	Segments       []ExplainSegment       `json:"segments"`
	Normalized     []ExplainNormalization `json:"normalized,omitempty"`
	EvaluatedInner *ExplainSegment        `json:"evaluated_inner,omitempty"`
	EvaluationOnly bool                   `json:"evaluation_only_normalization,omitempty"`
}

type ExplainNormalization struct {
	ProgramToken string `json:"program_token"`
	CommandName  string `json:"command_name"`
	Reason       string `json:"reason"`
}

type ExplainSegment struct {
	CommandName  string         `json:"command_name,omitempty"`
	ProgramToken string         `json:"program_token,omitempty"`
	Parser       string         `json:"parser"`
	Semantic     map[string]any `json:"semantic,omitempty"`
	Raw          string         `json:"raw,omitempty"`
}

type ExplainSourceDecision struct {
	Outcome     string            `json:"outcome"`
	MatchedRule *ExplainRuleMatch `json:"matched_rule,omitempty"`
	Matched     any               `json:"matched"`
}

type ExplainRuleMatch struct {
	Name    string `json:"name,omitempty"`
	Source  string `json:"source,omitempty"`
	Bucket  string `json:"bucket,omitempty"`
	Index   int    `json:"index"`
	Message string `json:"message,omitempty"`
}

type ExplainFinalDecision struct {
	Outcome string `json:"outcome"`
	Reason  string `json:"reason"`
}

type ExplainWhyNotResult struct {
	Command          string              `json:"command"`
	RequestedOutcome string              `json:"requested_outcome"`
	Actual           ExplainWhyNotActual `json:"actual"`
	MatchedRule      *ExplainRuleMatch   `json:"matched_rule,omitempty"`
	Parsed           ExplainParsed       `json:"parsed"`
	Reasons          []ExplainWhyNotItem `json:"reasons"`
	Suggestions      []ExplainWhyNotItem `json:"suggestions,omitempty"`
	Trace            []policy.TraceStep  `json:"trace,omitempty"`
}

type ExplainWhyNotActual struct {
	Policy         string `json:"policy"`
	ClaudeSettings string `json:"claude_settings"`
	Final          string `json:"final"`
}

type ExplainWhyNotItem struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func RunExplain(command string, env Env) (ExplainResult, error) {
	result, _, err := EvaluateForCommand(command, env, false)
	return result, err
}

func RunExplainWhyNot(command string, requested string, env Env) (ExplainWhyNotResult, error) {
	result, err := RunExplain(command, env)
	return ExplainWhyNot(result, requested), err
}

func EvaluateForCommand(command string, env Env, autoVerify bool) (ExplainResult, policy.Decision, error) {
	loaded, err := loadVerifiedPipelineForEvaluation(env, autoVerify)
	plan := commandpkg.Parse(command)
	result := ExplainResult{
		Command: command,
		Parsed:  explainParsed(plan),
	}
	if err != nil {
		result.Policy = ExplainSourceDecision{Outcome: "error", Matched: nil}
		result.ClaudeSettings = ExplainSourceDecision{Outcome: "abstain", Matched: nil}
		result.Final = ExplainFinalDecision{Outcome: "deny", Reason: err.Error() + "; run cc-bash-guard verify"}
		return result, policy.Decision{}, err
	}

	policyDecision, err := policy.Evaluate(loaded.Pipeline, command)
	if err != nil {
		result.Policy = ExplainSourceDecision{Outcome: "error", Matched: nil}
		result.ClaudeSettings = ExplainSourceDecision{Outcome: "abstain", Matched: nil}
		result.Final = ExplainFinalDecision{Outcome: "ask", Reason: err.Error()}
		return result, policy.Decision{}, err
	}

	claudeDecision := claude.ExplainCommand(command, env.Cwd, env.Home)
	finalDecision := claude.ApplyPermissionBridge(claude.Tool, policyDecision, env.Cwd, env.Home)

	result.Policy = explainPolicyDecision(policyDecision)
	result.ClaudeSettings = explainClaudeDecision(claudeDecision)
	result.Final = ExplainFinalDecision{Outcome: finalDecision.Outcome, Reason: finalReason(policyDecision, claudeDecision.Outcome, finalDecision)}
	result.Trace = finalDecision.Trace
	return result, finalDecision, nil
}

func loadVerifiedPipelineForEvaluation(env Env, autoVerify bool) (configrepo.Loaded, error) {
	loaded := configrepo.LoadEffectiveForHookTool(env.Cwd, env.Home, env.XDGConfigHome, env.XDGCacheHome, claude.Tool)
	if len(loaded.Errors) == 0 {
		return loaded, nil
	}
	if shouldAttemptImplicitVerify(loaded.Errors) {
		if !autoVerify {
			return loaded, errors.New("verified artifact missing or stale; run cc-bash-guard verify")
		}
		if err := ensureVerifiedArtifacts(env, claude.Tool); err == nil {
			loaded = configrepo.LoadEffectiveForHookTool(env.Cwd, env.Home, env.XDGConfigHome, env.XDGCacheHome, claude.Tool)
		} else {
			return loaded, err
		}
	}
	if len(loaded.Errors) > 0 {
		return loaded, errors.New(strings.Join(policy.ErrorStrings(loaded.Errors), "; "))
	}
	return loaded, nil
}

func explainParsed(plan commandpkg.CommandPlan) ExplainParsed {
	parsed := ExplainParsed{
		Shape:      string(plan.Shape.Kind),
		ShapeFlags: plan.Shape.Flags(),
		Segments:   make([]ExplainSegment, 0, len(plan.Commands)),
	}
	for _, diagnostic := range plan.Diagnostics {
		parsed.Diagnostics = append(parsed.Diagnostics, diagnostic.Message)
	}
	for _, cmd := range plan.Commands {
		parsed.Segments = append(parsed.Segments, explainSegment(cmd))
	}
	for _, normalized := range plan.Normalized {
		parsed.Normalized = append(parsed.Normalized, ExplainNormalization{
			ProgramToken: normalized.OriginalToken,
			CommandName:  normalized.CommandName,
			Reason:       normalized.Reason,
		})
		if normalized.Reason == "shell_dash_c" {
			parsed.Shape = "shell_c"
			parsed.EvaluationOnly = true
			if len(parsed.Segments) > 0 {
				inner := parsed.Segments[0]
				parsed.EvaluatedInner = &inner
			}
		}
	}
	return parsed
}

func explainSegment(cmd commandpkg.Command) ExplainSegment {
	parser := cmd.Parser
	if parser == "" {
		parser = "generic"
	}
	return ExplainSegment{
		CommandName:  cmd.Program,
		ProgramToken: cmd.ProgramToken,
		Parser:       parser,
		Semantic:     semanticMap(cmd),
		Raw:          cmd.Raw,
	}
}

func semanticMap(cmd commandpkg.Command) map[string]any {
	fields := map[string]any{}
	switch {
	case cmd.Git != nil:
		addString(fields, "verb", cmd.Git.Verb)
		addString(fields, "remote", cmd.Git.Remote)
		addString(fields, "branch", cmd.Git.Branch)
		addString(fields, "ref", cmd.Git.Ref)
		addBool(fields, "force", cmd.Git.Force)
		addBool(fields, "force_with_lease", cmd.Git.ForceWithLease)
		addBool(fields, "force_if_includes", cmd.Git.ForceIfIncludes)
		addBool(fields, "hard", cmd.Git.Hard)
		addBool(fields, "recursive", cmd.Git.Recursive)
		addBool(fields, "include_ignored", cmd.Git.IncludeIgnored)
		addBool(fields, "cached", cmd.Git.Cached)
		addBool(fields, "staged", cmd.Git.Staged)
	case cmd.AWS != nil:
		addString(fields, "service", cmd.AWS.Service)
		addString(fields, "operation", cmd.AWS.Operation)
		addString(fields, "profile", cmd.AWS.Profile)
		addString(fields, "region", cmd.AWS.Region)
	case cmd.Kubectl != nil:
		addString(fields, "verb", cmd.Kubectl.Verb)
		addString(fields, "subverb", cmd.Kubectl.Subverb)
		addString(fields, "resource_type", cmd.Kubectl.ResourceType)
		addString(fields, "resource_name", cmd.Kubectl.ResourceName)
		addString(fields, "namespace", cmd.Kubectl.Namespace)
		addString(fields, "context", cmd.Kubectl.Context)
	case cmd.Gh != nil:
		addString(fields, "area", cmd.Gh.Area)
		addString(fields, "verb", cmd.Gh.Verb)
		addString(fields, "repo", cmd.Gh.Repo)
		addString(fields, "hostname", cmd.Gh.Hostname)
	case cmd.Gws != nil:
		addString(fields, "service", cmd.Gws.Service)
		if len(cmd.Gws.ResourcePath) > 0 {
			fields["resource_path"] = cmd.Gws.ResourcePath
		}
		addString(fields, "method", cmd.Gws.Method)
		addBool(fields, "helper", cmd.Gws.Helper)
		addBool(fields, "mutating", cmd.Gws.Mutating)
		addBool(fields, "destructive", cmd.Gws.Destructive)
		addBool(fields, "read_only", cmd.Gws.ReadOnly)
		addBool(fields, "dry_run", cmd.Gws.DryRun)
		addBool(fields, "page_all", cmd.Gws.PageAll)
		addBool(fields, "upload", cmd.Gws.Upload)
		addBool(fields, "sanitize", cmd.Gws.Sanitize)
		addBool(fields, "params", cmd.Gws.Params)
		addBool(fields, "json_body", cmd.Gws.JSONBody)
		addBool(fields, "unmasked", cmd.Gws.Unmasked)
	case cmd.Helm != nil:
		addString(fields, "verb", cmd.Helm.Verb)
		addString(fields, "subverb", cmd.Helm.Subverb)
		addString(fields, "release", cmd.Helm.Release)
		addString(fields, "chart", cmd.Helm.Chart)
		addString(fields, "namespace", cmd.Helm.Namespace)
		addString(fields, "kube_context", cmd.Helm.KubeContext)
		addBool(fields, "dry_run", cmd.Helm.DryRun)
		addBool(fields, "force", cmd.Helm.Force)
		addBool(fields, "atomic", cmd.Helm.Atomic)
		addBool(fields, "wait", cmd.Helm.Wait)
		addBool(fields, "wait_for_jobs", cmd.Helm.WaitForJobs)
		addBool(fields, "install", cmd.Helm.Install)
		addBool(fields, "create_namespace", cmd.Helm.CreateNamespace)
	case cmd.Helmfile != nil:
		addString(fields, "verb", cmd.Helmfile.Verb)
		addString(fields, "environment", cmd.Helmfile.Environment)
		addString(fields, "namespace", cmd.Helmfile.Namespace)
		addString(fields, "kube_context", cmd.Helmfile.KubeContext)
	case cmd.ArgoCD != nil:
		addString(fields, "verb", cmd.ArgoCD.Verb)
		addString(fields, "app_name", cmd.ArgoCD.AppName)
		addString(fields, "project", cmd.ArgoCD.Project)
		addString(fields, "revision", cmd.ArgoCD.Revision)
	case cmd.Docker != nil:
		addString(fields, "verb", cmd.Docker.Verb)
		addString(fields, "subverb", cmd.Docker.Subverb)
		addString(fields, "compose_command", cmd.Docker.ComposeCommand)
		addString(fields, "image", cmd.Docker.Image)
		addString(fields, "container", cmd.Docker.Container)
		addString(fields, "service", cmd.Docker.Service)
		addString(fields, "context", cmd.Docker.Context)
		addString(fields, "host", cmd.Docker.Host)
		addString(fields, "file", cmd.Docker.File)
		addString(fields, "project_name", cmd.Docker.ProjectName)
		addBool(fields, "privileged", cmd.Docker.Privileged)
		addBool(fields, "network_host", cmd.Docker.NetworkHost)
		addBool(fields, "pid_host", cmd.Docker.PIDHost)
		addBool(fields, "ipc_host", cmd.Docker.IPCHost)
		addBool(fields, "uts_host", cmd.Docker.UTSHost)
		addBool(fields, "host_mount", cmd.Docker.HostMount)
		addBool(fields, "root_mount", cmd.Docker.RootMount)
		addBool(fields, "docker_socket_mount", cmd.Docker.DockerSocketMount)
		addBool(fields, "prune", cmd.Docker.Prune)
		addBool(fields, "all_resources", cmd.Docker.AllResources)
		addBool(fields, "volumes_flag", cmd.Docker.VolumesFlag)
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func addString(fields map[string]any, key string, value string) {
	if strings.TrimSpace(value) != "" {
		fields[key] = value
	}
}

func addBool(fields map[string]any, key string, value bool) {
	if value {
		fields[key] = value
	}
}

func explainPolicyDecision(decision policy.Decision) ExplainSourceDecision {
	outcome := decision.Outcome
	if outcome == "" {
		outcome = "abstain"
	}
	return ExplainSourceDecision{
		Outcome:     outcome,
		MatchedRule: matchedPolicyRule(decision.Trace, outcome),
		Matched:     nil,
	}
}

func matchedPolicyRule(trace []policy.TraceStep, outcome string) *ExplainRuleMatch {
	for i := len(trace) - 1; i >= 0; i-- {
		step := trace[i]
		if step.Action != "permission" || step.Effect != outcome || step.Name == "no_match" || step.Name == "fail_closed" || step.Name == "composition" || step.Name == "composition.command" {
			continue
		}
		bucket := ""
		index := 0
		source := ""
		if step.Source != nil {
			bucket = step.Source.Section
			index = step.Source.Index
			source = step.Source.Path
		}
		if bucket == "" && outcome != "" {
			bucket = "permission." + outcome
		}
		name := strings.TrimSpace(step.Name)
		if name == "" {
			name = bucket + "[" + itoa(index) + "]"
		}
		return &ExplainRuleMatch{Name: name, Source: source, Bucket: bucket, Index: index, Message: step.Message}
	}
	return nil
}

func explainClaudeDecision(decision claude.PermissionExplanation) ExplainSourceDecision {
	outcome := decision.Outcome
	if outcome == "" || outcome == "default" {
		outcome = "abstain"
	}
	var matched any
	if decision.Matched != nil {
		matched = decision.Matched
	}
	return ExplainSourceDecision{Outcome: outcome, Matched: matched}
}

func finalReason(policyDecision policy.Decision, claudeOutcome string, finalDecision policy.Decision) string {
	policyOutcome := policyDecision.Outcome
	if policyOutcome == "" {
		policyOutcome = "abstain"
	}
	if claudeOutcome == "" || claudeOutcome == "default" {
		claudeOutcome = "abstain"
	}
	switch {
	case policyOutcome == "deny":
		return "cc-bash-guard policy denied"
	case claudeOutcome == "deny":
		return "Claude settings denied"
	case policyOutcome == "ask" || claudeOutcome == "ask":
		return "at least one source asked"
	case policyOutcome == "allow" || claudeOutcome == "allow":
		return "at least one source allowed and no source denied or asked"
	default:
		if finalDecision.Reason == "default_fallback" {
			return "all permission sources abstained; fallback ask"
		}
		return "all permission sources abstained; fallback ask"
	}
}

func ExplainHasParseError(result ExplainResult) bool {
	for _, diagnostic := range result.Parsed.Diagnostics {
		if strings.TrimSpace(diagnostic) != "" {
			return true
		}
	}
	return false
}

func ExplainWhyNot(result ExplainResult, requested string) ExplainWhyNotResult {
	analysis := ExplainWhyNotResult{
		Command:          result.Command,
		RequestedOutcome: requested,
		Actual: ExplainWhyNotActual{
			Policy:         result.Policy.Outcome,
			ClaudeSettings: result.ClaudeSettings.Outcome,
			Final:          result.Final.Outcome,
		},
		MatchedRule: firstMatchedRule(result),
		Parsed:      result.Parsed,
		Trace:       result.Trace,
	}

	if result.Final.Outcome == requested {
		analysis.Reasons = append(analysis.Reasons, ExplainWhyNotItem{
			Kind:    "requested_outcome_happened",
			Message: "final outcome is already " + requested,
		})
		return analysis
	}

	if result.Policy.Outcome == "error" {
		analysis.Reasons = append(analysis.Reasons, ExplainWhyNotItem{
			Kind:    "policy_error",
			Message: result.Final.Reason,
		})
		if strings.Contains(result.Final.Reason, "verified artifact") {
			analysis.Suggestions = append(analysis.Suggestions, ExplainWhyNotItem{
				Kind:    "run_verify",
				Message: "Run cc-bash-guard verify after editing policy or included policy files",
			})
		}
		return analysis
	}

	if len(result.Parsed.Diagnostics) > 0 {
		analysis.Reasons = append(analysis.Reasons, ExplainWhyNotItem{
			Kind:    "parse_error",
			Message: "command parser reported diagnostics; permission evaluation fails closed",
		})
	}

	addOutcomeReasons(&analysis, result, requested)
	addTraceReasons(&analysis, result, requested)
	addAbstainReasons(&analysis, result, requested)
	addSemanticReason(&analysis, result, requested)
	addSuggestions(&analysis, result, requested)

	if len(analysis.Reasons) == 0 {
		analysis.Reasons = append(analysis.Reasons, ExplainWhyNotItem{
			Kind:    "outcome_mismatch",
			Message: "final outcome was " + result.Final.Outcome + ", not " + requested,
		})
	}
	return analysis
}

func firstMatchedRule(result ExplainResult) *ExplainRuleMatch {
	if result.Policy.MatchedRule != nil {
		return result.Policy.MatchedRule
	}
	return nil
}

func addOutcomeReasons(analysis *ExplainWhyNotResult, result ExplainResult, requested string) {
	switch requested {
	case "allow":
		if result.Policy.Outcome == "deny" {
			addReason(analysis, "deny_outranks_allow", "matched deny rule outranks allow")
		}
		if result.ClaudeSettings.Outcome == "deny" {
			addReason(analysis, "claude_settings_deny", "Claude settings denied; deny outranks allow")
		}
		if result.Policy.Outcome == "ask" && result.Policy.MatchedRule != nil {
			addReason(analysis, "ask_outranks_allow", "matched ask rule outranks allow")
		}
		if result.ClaudeSettings.Outcome == "ask" {
			addReason(analysis, "claude_settings_ask", "Claude settings asked; ask outranks allow")
		}
	case "ask":
		if result.Policy.Outcome == "deny" {
			addReason(analysis, "deny_outranks_ask", "matched deny rule outranks ask")
		}
		if result.ClaudeSettings.Outcome == "deny" {
			addReason(analysis, "claude_settings_deny", "Claude settings denied; deny outranks ask")
		}
		if result.Final.Outcome == "allow" {
			addReason(analysis, "allowed_without_ask", "at least one source allowed and no source denied or asked")
		}
	case "deny":
		if result.Final.Outcome == "ask" {
			addReason(analysis, "asked_without_deny", "no permission source denied; final outcome was ask")
		}
		if result.Final.Outcome == "allow" {
			addReason(analysis, "allowed_without_deny", "no permission source denied; final outcome was allow")
		}
	}
}

func addTraceReasons(analysis *ExplainWhyNotResult, result ExplainResult, requested string) {
	for _, step := range result.Trace {
		if step.Action != "permission" {
			continue
		}
		if step.Name == "fail_closed" && step.Effect == "ask" {
			addReason(analysis, "unsafe_shell_shape", "shell shape is unsafe for structured allow: "+nonEmptyString(step.Reason, result.Final.Reason))
		}
		if step.Name == "composition" && step.Effect == "ask" && requested == "allow" {
			addReason(analysis, "unsafe_shell_shape", "shell shape is not auto-allowed: "+nonEmptyString(step.Reason, result.Final.Reason))
		}
	}
}

func addAbstainReasons(analysis *ExplainWhyNotResult, result ExplainResult, requested string) {
	if result.Policy.Outcome == "abstain" {
		addReason(analysis, "no_policy_match", "cc-bash-guard policy abstained")
	}
	if result.ClaudeSettings.Outcome == "abstain" {
		addReason(analysis, "claude_settings_abstain", "Claude settings abstained")
	}
	if result.Policy.Outcome == "abstain" && result.ClaudeSettings.Outcome == "abstain" && result.Final.Outcome == "ask" {
		addReason(analysis, "fallback_ask", "all permission sources abstained; final fallback is ask")
	}
	if requested == "deny" && result.Policy.Outcome != "deny" && result.ClaudeSettings.Outcome != "deny" {
		addReason(analysis, "no_deny_match", "no deny rule matched in cc-bash-guard policy or Claude settings")
	}
}

func addSemanticReason(analysis *ExplainWhyNotResult, result ExplainResult, requested string) {
	if result.Policy.Outcome != "abstain" {
		return
	}
	for _, segment := range result.Parsed.Segments {
		if segment.Parser != "" && segment.Parser != "generic" && len(segment.Semantic) > 0 {
			addReason(analysis, "semantic_mismatch", "parsed semantic fields did not match any "+requested+" policy rule")
			return
		}
	}
}

func addSuggestions(analysis *ExplainWhyNotResult, result ExplainResult, requested string) {
	if hasReason(*analysis, "deny_outranks_allow") || hasReason(*analysis, "deny_outranks_ask") {
		addSuggestion(analysis, "review_deny_rule", "Review the matched deny rule; deny has highest priority")
	}
	if hasReason(*analysis, "ask_outranks_allow") {
		addSuggestion(analysis, "review_ask_rule", "Review the matched ask rule; explicit ask is intentionally not overridden by allow")
	}
	if hasReason(*analysis, "no_policy_match") {
		addSuggestion(analysis, "add_policy_rule", "Use cc-bash-guard suggest to generate a starter rule")
	}
	if hasReason(*analysis, "semantic_mismatch") {
		addSuggestion(analysis, "compare_semantic_fields", "Compare the parsed semantic fields with command.semantic in the intended rule")
	}
	if hasReason(*analysis, "unsafe_shell_shape") && requested == "allow" {
		addSuggestion(analysis, "avoid_unsafe_shell_shape", "Avoid redirects, subshells, background execution, and other unsafe shell shapes for automatic allow")
	}
	if requested == "deny" && result.Final.Outcome != "deny" {
		addSuggestion(analysis, "add_deny_rule", "Add a permission.deny rule for this command if it must be blocked")
	}
}

func addReason(analysis *ExplainWhyNotResult, kind string, message string) {
	if hasReason(*analysis, kind) {
		return
	}
	analysis.Reasons = append(analysis.Reasons, ExplainWhyNotItem{Kind: kind, Message: message})
}

func addSuggestion(analysis *ExplainWhyNotResult, kind string, message string) {
	for _, item := range analysis.Suggestions {
		if item.Kind == kind {
			return
		}
	}
	analysis.Suggestions = append(analysis.Suggestions, ExplainWhyNotItem{Kind: kind, Message: message})
}

func hasReason(analysis ExplainWhyNotResult, kind string) bool {
	for _, item := range analysis.Reasons {
		if item.Kind == kind {
			return true
		}
	}
	return false
}

func nonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func SortedSemanticKeys(fields map[string]any) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	n := v
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
