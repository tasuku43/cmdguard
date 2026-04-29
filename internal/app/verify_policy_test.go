package app

import (
	"strings"
	"testing"

	"github.com/tasuku43/cc-bash-guard/internal/domain/policy"
	configrepo "github.com/tasuku43/cc-bash-guard/internal/infra/config"
)

func TestVerifyPolicyFailuresSemanticAllowSubsumedByBroadAllow(t *testing.T) {
	truePtr := func(v bool) *bool { return &v }
	tests := []struct {
		name    string
		allow   []policy.PermissionRuleSpec
		command string
	}{
		{
			name: "git semantic allow plus command name allow",
			allow: []policy.PermissionRuleSpec{
				{Name: "git status only", Command: policy.PermissionCommandSpec{Name: "git", Semantic: &policy.SemanticMatchSpec{Verb: "status"}}},
				{Name: "all git", Command: policy.PermissionCommandSpec{Name: "git"}},
			},
			command: "git",
		},
		{
			name: "aws semantic allow plus broad aws pattern",
			allow: []policy.PermissionRuleSpec{
				{Name: "aws identity only", Command: policy.PermissionCommandSpec{Name: "aws", Semantic: &policy.SemanticMatchSpec{Service: "sts", Operation: "get-caller-identity"}}},
				{Name: "all aws", Patterns: []string{`^aws\s+.*$`}},
			},
			command: "aws",
		},
		{
			name: "kubectl semantic allow plus name_in containing kubectl",
			allow: []policy.PermissionRuleSpec{
				{Name: "kubectl read only", Command: policy.PermissionCommandSpec{Name: "kubectl", Semantic: &policy.SemanticMatchSpec{VerbIn: []string{"get", "describe"}}}},
				{Name: "broad infra tools", Command: policy.PermissionCommandSpec{NameIn: []string{"kubectl", "helm", "terraform"}}},
			},
			command: "kubectl",
		},
		{
			name: "terraform semantic allow plus broad terraform pattern",
			allow: []policy.PermissionRuleSpec{
				{Name: "terraform plan", Command: policy.PermissionCommandSpec{Name: "terraform", Semantic: &policy.SemanticMatchSpec{Subcommand: "plan"}}},
				{Name: "broad terraform", Patterns: []string{`^terraform\s+.*$`}},
			},
			command: "terraform",
		},
		{
			name: "docker semantic allow plus broad docker pattern",
			allow: []policy.PermissionRuleSpec{
				{Name: "docker ps", Command: policy.PermissionCommandSpec{Name: "docker", Semantic: &policy.SemanticMatchSpec{Verb: "ps"}}},
				{Name: "broad docker", Patterns: []string{`^\s*docker\s+.*$`}},
			},
			command: "docker",
		},
		{
			name: "argocd semantic allow plus command name allow",
			allow: []policy.PermissionRuleSpec{
				{Name: "argocd app get", Command: policy.PermissionCommandSpec{Name: "argocd", Semantic: &policy.SemanticMatchSpec{Verb: "app get"}}},
				{Name: "all argocd", Command: policy.PermissionCommandSpec{Name: "argocd"}},
			},
			command: "argocd",
		},
		{
			name: "semantic allow plus env only allow",
			allow: []policy.PermissionRuleSpec{
				{Name: "aws identity only", Command: policy.PermissionCommandSpec{Name: "aws", Semantic: &policy.SemanticMatchSpec{Service: "sts", Operation: "get-caller-identity"}}},
				{Name: "any command with AWS_PROFILE", Env: policy.PermissionEnvSpec{Requires: []string{"AWS_PROFILE"}}},
			},
			command: "aws",
		},
		{
			name: "semantic allow plus boundary broad pattern",
			allow: []policy.PermissionRuleSpec{
				{Name: "docker ps", Command: policy.PermissionCommandSpec{Name: "docker", Semantic: &policy.SemanticMatchSpec{Verb: "ps", Privileged: truePtr(false)}}},
				{Name: "broad docker boundary", Patterns: []string{`^docker(\s|$).*`}},
			},
			command: "docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := verifyPolicyFailures(loadedWithRules(stampAllowSources(tt.allow), nil, nil))
			diag, ok := findDiagnostic(diags, "semantic_allow_subsumed_by_broad_allow")
			if !ok {
				t.Fatalf("missing semantic subsumption diagnostic; got %+v", diags)
			}
			if diag.Command != tt.command {
				t.Fatalf("Command = %q, want %q", diag.Command, tt.command)
			}
			if diag.First == nil || diag.First.Bucket != "allow" || diag.First.Index != 0 || diag.First.Name == "" {
				t.Fatalf("First source = %+v, want narrow allow source", diag.First)
			}
			if diag.Second == nil || diag.Second.Bucket != "allow" || diag.Second.Index != 1 || diag.Second.Name == "" {
				t.Fatalf("Second source = %+v, want broad allow source", diag.Second)
			}
			for _, want := range []string{"semantic allow", "broader allow", "permission.ask", "command.semantic"} {
				if !strings.Contains(diag.Message+" "+diag.Hint+" "+diag.SaferAlternative, want) {
					t.Fatalf("diagnostic missing %q: %+v", want, diag)
				}
			}
		})
	}
}

func TestVerifyPolicyFailuresBroadAllowRules(t *testing.T) {
	tests := []struct {
		name    string
		allow   []policy.PermissionRuleSpec
		command string
		reason  string
	}{
		{
			name:    "env only allow rejected",
			allow:   []policy.PermissionRuleSpec{{Name: "any aws profile", Env: policy.PermissionEnvSpec{Requires: []string{"AWS_PROFILE"}}}},
			reason:  "env-only allow",
			command: "",
		},
		{
			name:    "supported command name only rejected",
			allow:   []policy.PermissionRuleSpec{{Name: "all git", Command: policy.PermissionCommandSpec{Name: "git"}}},
			command: "git",
			reason:  "command.name allows the whole git command namespace without semantic constraints",
		},
		{
			name:    "name_in with supported semantic command rejected",
			allow:   []policy.PermissionRuleSpec{{Name: "infra tools", Command: policy.PermissionCommandSpec{NameIn: []string{"ls", "kubectl", "pwd"}}}},
			command: "kubectl",
			reason:  "command.name_in includes supported semantic command kubectl without semantic constraints",
		},
		{
			name:    "script runner command name rejected",
			allow:   []policy.PermissionRuleSpec{{Name: "all npm", Command: policy.PermissionCommandSpec{Name: "npm"}}},
			command: "npm",
			reason:  "script runner or generic interpreter npm",
		},
		{
			name:    "name_in with interpreter rejected",
			allow:   []policy.PermissionRuleSpec{{Name: "shells", Command: policy.PermissionCommandSpec{NameIn: []string{"bash", "zsh"}}}},
			command: "bash",
			reason:  "script runner or generic interpreter bash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := verifyPolicyFailures(loadedWithRules(stampAllowSources(tt.allow), nil, nil))
			diag, ok := findDiagnostic(diags, "broad_allow_rule")
			if !ok {
				t.Fatalf("missing broad allow rule diagnostic; got %+v", diags)
			}
			if diag.Command != tt.command {
				t.Fatalf("Command = %q, want %q", diag.Command, tt.command)
			}
			if !strings.Contains(diag.Reason, tt.reason) {
				t.Fatalf("Reason = %q, want containing %q", diag.Reason, tt.reason)
			}
			for _, want := range []string{"semantic", "permission.ask", "narrow anchored"} {
				if !strings.Contains(diag.Message+" "+diag.Hint+" "+diag.SaferAlternative, want) {
					t.Fatalf("diagnostic missing %q: %+v", want, diag)
				}
			}
		})
	}
}

func TestVerifyPolicyFailuresBroadAllowPatternRules(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "unanchored allow regex rejected", pattern: `git\s+status`},
		{name: "wildcard crossing shell metacharacters rejected", pattern: `^git\s+status.*$`},
		{name: "whole git namespace rejected", pattern: `^git\s+.*$`},
		{name: "whole aws namespace rejected", pattern: `^aws\s+.*$`},
		{name: "whole kubectl namespace rejected", pattern: `^kubectl\s+.*$`},
		{name: "whole terraform namespace rejected", pattern: `^terraform\s+.*$`},
		{name: "whole docker namespace rejected", pattern: `^docker\s+.*$`},
		{name: "script runner namespace rejected", pattern: `^npm\s+.*$`},
		{name: "interpreter namespace rejected", pattern: `^python3\s+.*$`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := verifyPolicyFailures(loadedWithRules(stampAllowSources([]policy.PermissionRuleSpec{
				{Name: tt.name, Patterns: []string{tt.pattern}},
			}), nil, nil))
			diag, ok := findDiagnostic(diags, "broad_allow_pattern")
			if !ok {
				t.Fatalf("missing broad allow pattern diagnostic; got %+v", diags)
			}
			if diag.Pattern != tt.pattern {
				t.Fatalf("Pattern = %q, want %q", diag.Pattern, tt.pattern)
			}
			for _, want := range []string{"narrower regex", "anchored", "shell metacharacters"} {
				if !strings.Contains(diag.Message+" "+diag.Hint+" "+diag.SaferAlternative, want) {
					t.Fatalf("diagnostic missing %q: %+v", want, diag)
				}
			}
		})
	}
}

func TestVerifyPolicyFailuresSemanticAllowSubsumptionNonFailures(t *testing.T) {
	tests := []struct {
		name  string
		deny  []policy.PermissionRuleSpec
		ask   []policy.PermissionRuleSpec
		allow []policy.PermissionRuleSpec
	}{
		{
			name: "semantic git status allow plus semantic git diff allow",
			allow: []policy.PermissionRuleSpec{
				{Name: "git status", Command: policy.PermissionCommandSpec{Name: "git", Semantic: &policy.SemanticMatchSpec{Verb: "status"}}},
				{Name: "git diff", Command: policy.PermissionCommandSpec{Name: "git", Semantic: &policy.SemanticMatchSpec{Verb: "diff"}}},
			},
		},
		{
			name:  "semantic git status allow plus broad git ask",
			ask:   []policy.PermissionRuleSpec{{Name: "ask all git", Command: policy.PermissionCommandSpec{Name: "git"}}},
			allow: []policy.PermissionRuleSpec{{Name: "git status", Command: policy.PermissionCommandSpec{Name: "git", Semantic: &policy.SemanticMatchSpec{Verb: "status"}}}},
		},
		{
			name:  "semantic git status allow plus broad git deny",
			deny:  []policy.PermissionRuleSpec{{Name: "deny all git", Command: policy.PermissionCommandSpec{Name: "git"}}},
			allow: []policy.PermissionRuleSpec{{Name: "git status", Command: policy.PermissionCommandSpec{Name: "git", Semantic: &policy.SemanticMatchSpec{Verb: "status"}}}},
		},
		{
			name: "semantic aws allow plus narrow pattern for another command",
			allow: []policy.PermissionRuleSpec{
				{Name: "aws identity", Command: policy.PermissionCommandSpec{Name: "aws", Semantic: &policy.SemanticMatchSpec{Service: "sts", Operation: "get-caller-identity"}}},
				{Name: "npm version", Patterns: []string{`^npm\s+--version$`}},
			},
		},
		{
			name: "semantic kubectl allow plus docker command allow",
			allow: []policy.PermissionRuleSpec{
				{Name: "kubectl get", Command: policy.PermissionCommandSpec{Name: "kubectl", Semantic: &policy.SemanticMatchSpec{Verb: "get"}}},
				{Name: "docker ps", Command: policy.PermissionCommandSpec{Name: "docker", Semantic: &policy.SemanticMatchSpec{Verb: "ps"}}},
			},
		},
		{
			name: "deny and ask broad rules remain allowed",
			deny: []policy.PermissionRuleSpec{{Name: "deny all git", Command: policy.PermissionCommandSpec{Name: "git"}}},
			ask:  []policy.PermissionRuleSpec{{Name: "ask all npm", Command: policy.PermissionCommandSpec{Name: "npm"}}},
			allow: []policy.PermissionRuleSpec{
				{Name: "git status", Command: policy.PermissionCommandSpec{Name: "git", Semantic: &policy.SemanticMatchSpec{Verb: "status"}}},
				{Name: "npm version", Patterns: []string{`^npm\s+--version$`}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := verifyPolicyFailures(loadedWithRules(stampAllowSources(tt.allow), stampRules(tt.ask, "ask"), stampRules(tt.deny, "deny")))
			for _, kind := range []string{"semantic_allow_subsumed_by_broad_allow", "broad_allow_rule", "broad_allow_pattern"} {
				if _, ok := findDiagnostic(diags, kind); ok {
					t.Fatalf("unexpected %s diagnostic: %+v", kind, diags)
				}
			}
		})
	}
}

func TestVerifyPolicyFailuresBroadAllowPatternWithoutSemanticUnchanged(t *testing.T) {
	diags := verifyPolicyFailures(loadedWithRules(stampAllowSources([]policy.PermissionRuleSpec{
		{Name: "all npm", Patterns: []string{`^npm\s+.*$`}},
	}), nil, nil))
	if _, ok := findDiagnostic(diags, "broad_allow_pattern"); !ok {
		t.Fatalf("missing existing broad allow pattern diagnostic: %+v", diags)
	}
	if _, ok := findDiagnostic(diags, "semantic_allow_subsumed_by_broad_allow"); ok {
		t.Fatalf("unexpected semantic subsumption diagnostic without semantic allow: %+v", diags)
	}
}

func loadedWithRules(allow []policy.PermissionRuleSpec, ask []policy.PermissionRuleSpec, deny []policy.PermissionRuleSpec) configrepo.Loaded {
	pipeline := policy.NewPipeline(policy.PipelineSpec{
		Permission: policy.PermissionSpec{Deny: deny, Ask: ask, Allow: allow},
	}, policy.Source{Path: "policy.yml"})
	return configrepo.Loaded{Pipeline: pipeline}
}

func stampAllowSources(rules []policy.PermissionRuleSpec) []policy.PermissionRuleSpec {
	return stampRules(rules, "allow")
}

func stampRules(rules []policy.PermissionRuleSpec, bucket string) []policy.PermissionRuleSpec {
	for i := range rules {
		rules[i].Source = policy.Source{Path: "policy.yml", Section: "permission." + bucket, Index: i}
	}
	return rules
}

func findDiagnostic(diags []VerifyDiagnostic, kind string) (VerifyDiagnostic, bool) {
	for _, diag := range diags {
		if diag.Kind == kind {
			return diag, true
		}
	}
	return VerifyDiagnostic{}, false
}
