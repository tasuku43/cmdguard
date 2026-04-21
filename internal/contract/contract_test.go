package contract

import (
	"strings"
	"testing"

	"github.com/tasuku43/cmdproxy/internal/domain/policy"
)

func TestValidateRulesAcceptsBuiltInContracts(t *testing.T) {
	rules := []policy.RuleSpec{
		{
			ID: "aws-profile-to-env",
			Matcher: policy.MatchSpec{
				Command: "aws",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "--profile",
					Env:  "AWS_PROFILE",
				},
			},
		},
		{
			ID: "gh-repo-to-env",
			Matcher: policy.MatchSpec{
				Command: "gh",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "--repo",
					Env:  "GH_REPO",
				},
			},
		},
		{
			ID: "unwrap-git-shell",
			Matcher: policy.MatchSpec{
				Command: "git",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapShellDashC: true,
			},
		},
		{
			ID: "unwrap-docker-shell",
			Matcher: policy.MatchSpec{
				Command: "docker",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapShellDashC: true,
			},
		},
		{
			ID: "unwrap-kubectl-wrapper",
			Matcher: policy.MatchSpec{
				Command: "kubectl",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapWrapper: policy.UnwrapWrapperSpec{
					Wrappers: []string{"env", "command", "exec"},
				},
			},
		},
		{
			ID: "unwrap-npm-shell",
			Matcher: policy.MatchSpec{
				Command: "npm",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapShellDashC: true,
			},
		},
		{
			ID: "unwrap-pnpm-shell",
			Matcher: policy.MatchSpec{
				Command: "pnpm",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapShellDashC: true,
			},
		},
		{
			ID: "unwrap-yarn-shell",
			Matcher: policy.MatchSpec{
				Command: "yarn",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapShellDashC: true,
			},
		},
		{
			ID: "unwrap-terraform-wrapper",
			Matcher: policy.MatchSpec{
				Command: "terraform",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapWrapper: policy.UnwrapWrapperSpec{
					Wrappers: []string{"env", "command"},
				},
			},
		},
		{
			ID: "unwrap-go-shell",
			Matcher: policy.MatchSpec{
				Command: "go",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapShellDashC: true,
			},
		},
		{
			ID: "aws-region-to-env",
			Matcher: policy.MatchSpec{
				Command: "aws",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "--region",
					Env:  "AWS_DEFAULT_REGION",
				},
			},
		},
		{
			ID: "kubectl-kubeconfig-relaxed",
			Matcher: policy.MatchSpec{
				Command: "kubectl",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "--kubeconfig",
					Env:  "KUBECONFIG",
				},
				Strict: boolPtr(false),
			},
		},
	}

	issues := ValidateRules(rules)
	if len(issues) != 0 {
		t.Fatalf("issues = %v", issues)
	}
}

func TestValidateRulesRejectsUnknownEnvMapping(t *testing.T) {
	rules := []policy.RuleSpec{
		{
			ID: "bad-aws-profile-to-env",
			Matcher: policy.MatchSpec{
				Command: "aws",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "--profile",
					Env:  "HOGE",
				},
			},
		},
	}

	issues := ValidateRules(rules)
	if len(issues) == 0 {
		t.Fatal("expected issues")
	}
	if !strings.Contains(issues[0], "AWS_PROFILE") {
		t.Fatalf("issues = %v", issues)
	}
}

func TestValidateRulesRejectsRelaxedCandidateWhenStrict(t *testing.T) {
	rules := []policy.RuleSpec{
		{
			ID: "kubectl-kubeconfig-strict",
			Matcher: policy.MatchSpec{
				Command: "kubectl",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "--kubeconfig",
					Env:  "KUBECONFIG",
				},
			},
		},
	}

	issues := ValidateRules(rules)
	if len(issues) == 0 {
		t.Fatal("expected issues")
	}
	if !strings.Contains(issues[0], "no strict") {
		t.Fatalf("issues = %v", issues)
	}
}

func TestValidateRulesRejectsUnsupportedCommandContract(t *testing.T) {
	rules := []policy.RuleSpec{
		{
			ID: "python-rewrite",
			Matcher: policy.MatchSpec{
				Command: "python",
			},
			Rewrite: policy.RewriteSpec{
				MoveFlagToEnv: policy.MoveFlagToEnvSpec{
					Flag: "-m",
					Env:  "PYTHON_MODULE",
				},
			},
		},
	}

	issues := ValidateRules(rules)
	if len(issues) == 0 {
		t.Fatal("expected issues")
	}
	if !strings.Contains(issues[0], "not supported") {
		t.Fatalf("issues = %v", issues)
	}
}

func TestValidateRulesRejectsUnsupportedWrapper(t *testing.T) {
	rules := []policy.RuleSpec{
		{
			ID: "bad-wrapper",
			Matcher: policy.MatchSpec{
				Command: "git",
			},
			Rewrite: policy.RewriteSpec{
				UnwrapWrapper: policy.UnwrapWrapperSpec{
					Wrappers: []string{"sudo"},
				},
			},
		},
	}

	issues := ValidateRules(rules)
	if len(issues) == 0 {
		t.Fatal("expected issues")
	}
	if !strings.Contains(issues[0], "wrapper") {
		t.Fatalf("issues = %v", issues)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
