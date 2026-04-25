package doctoring

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tasuku43/cc-bash-proxy/internal/adapter/claude"
	"github.com/tasuku43/cc-bash-proxy/internal/domain/policy"
	"github.com/tasuku43/cc-bash-proxy/internal/infra/buildinfo"
	configrepo "github.com/tasuku43/cc-bash-proxy/internal/infra/config"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Status   Status `json:"status"`
	Message  string `json:"message"`
}

type Report struct {
	Tool                       string              `json:"tool,omitempty"`
	ClaudePermissionMergeMode  string              `json:"claude_permission_merge_mode,omitempty"`
	ConfigSources              []configrepo.Source `json:"config_sources,omitempty"`
	SettingsPaths              []string            `json:"settings_paths,omitempty"`
	EffectiveFingerprint       string              `json:"effective_fingerprint,omitempty"`
	VerifiedArtifactExists     bool                `json:"verified_artifact_exists"`
	VerifiedArtifactCompatible bool                `json:"verified_artifact_compatible"`
	Checks                     []Check             `json:"checks"`
}

func Run(loaded configrepo.Loaded, tool string, cwd string, home string) Report {
	var checks []Check

	if len(loaded.Errors) == 0 {
		checks = append(checks,
			Check{ID: "config.parse", Category: "config", Status: StatusPass, Message: "configuration files parsed"},
			Check{ID: "config.schema", Category: "config", Status: StatusPass, Message: "configuration schema is valid"},
			Check{ID: "permission.tests-present", Category: "permission", Status: StatusPass, Message: "permission tests are present"},
			Check{ID: "test.e2e-present", Category: "test", Status: StatusPass, Message: "end-to-end tests are present"},
		)
	} else {
		msg := strings.Join(policy.ErrorStrings(loaded.Errors), "; ")
		checks = append(checks,
			Check{ID: "config.parse", Category: "config", Status: StatusFail, Message: msg},
			Check{ID: "config.schema", Category: "config", Status: StatusFail, Message: msg},
			Check{ID: "permission.tests-present", Category: "permission", Status: StatusFail, Message: msg},
			Check{ID: "test.e2e-present", Category: "test", Status: StatusFail, Message: msg},
		)
		if strings.Contains(msg, "top-level rewrite is no longer supported") {
			checks = append(checks, Check{
				ID:       "config.rewrite-migration",
				Category: "config",
				Status:   StatusFail,
				Message:  "rewrite migration: unwrap_shell_dash_c is now built-in CommandPlan parsing; strip_command_path is basename command normalization; move_flag_to_env is no longer supported, use AWS semantic profile matching and document preferred command style in CLAUDE.md or user docs",
			})
		}
	}

	if len(loaded.Errors) == 0 {
		if err := testsPass(loaded.Pipeline, tool, cwd, home); err != nil {
			checks = append(checks, Check{ID: "tests.pass", Category: "test", Status: StatusFail, Message: err.Error()})
		} else {
			checks = append(checks, Check{ID: "tests.pass", Category: "test", Status: StatusPass, Message: "permission and end-to-end tests match expectations"})
		}
	} else {
		checks = append(checks, Check{ID: "tests.pass", Category: "test", Status: StatusFail, Message: "skipped because configuration is invalid"})
	}

	if ids := envOnlyAllowNames(loaded.Pipeline); len(ids) > 0 {
		checks = append(checks, Check{ID: "permission.env-only-allow", Category: "permission", Status: StatusWarn, Message: "env-only allow rules are broad: " + strings.Join(ids, ", ")})
	} else {
		checks = append(checks, Check{ID: "permission.env-only-allow", Category: "permission", Status: StatusPass, Message: "no env-only allow rules"})
	}

	mergeMode := claudePermissionMergeMode(loaded.Pipeline)
	if tool == claude.Tool && mergeMode == claude.MergeModeMigrationCompat {
		checks = append(checks, Check{ID: "permission.claude-merge-mode", Category: "permission", Status: StatusWarn, Message: "Claude permission merge mode is migration_compat; use strict for security-first behavior"})
	} else if tool == claude.Tool {
		checks = append(checks, Check{ID: "permission.claude-merge-mode", Category: "permission", Status: StatusPass, Message: "Claude permission merge mode: " + mergeMode})
	}

	if path, err := exec.LookPath("cc-bash-proxy"); err == nil {
		checks = append(checks, Check{ID: "install.binary-on-path", Category: "install", Status: StatusPass, Message: "cc-bash-proxy found on PATH at " + path})
	} else {
		checks = append(checks, Check{ID: "install.binary-on-path", Category: "install", Status: StatusWarn, Message: "cc-bash-proxy not found on PATH"})
	}

	if exe, err := os.Executable(); err == nil {
		checks = append(checks, Check{ID: "install.binary-executable", Category: "install", Status: StatusPass, Message: "running binary: " + exe})
	} else {
		checks = append(checks, Check{ID: "install.binary-executable", Category: "install", Status: StatusWarn, Message: "running binary path could not be determined"})
	}

	bi := buildinfo.Read()
	if bi.VCSRevision != "" {
		msg := "build metadata available"
		if bi.VCSModified != "" {
			msg += " (vcs.modified=" + bi.VCSModified + ")"
		}
		checks = append(checks, Check{ID: "install.binary-build-info", Category: "install", Status: StatusPass, Message: msg})
	} else {
		checks = append(checks, Check{ID: "install.binary-build-info", Category: "install", Status: StatusWarn, Message: "build metadata missing; prefer binaries built with VCS info embedded"})
	}

	if tool == claude.Tool {
		claudeSettings := filepath.Join(home, ".claude", "settings.json")
		checks = append(checks, claudeHookRegistrationCheck(claudeSettings))
	}

	return Report{ClaudePermissionMergeMode: mergeMode, Checks: checks}
}

func AddVerifiedArtifactCheck(report Report, status configrepo.EffectiveArtifactStatus) Report {
	report.VerifiedArtifactExists = status.Exists
	report.VerifiedArtifactCompatible = status.Compatible
	if status.Compatible {
		report.Checks = append(report.Checks, Check{ID: "artifact.evaluation-semantics", Category: "artifact", Status: StatusPass, Message: status.Message})
		return report
	}
	if status.Exists {
		report.Checks = append(report.Checks, Check{ID: "artifact.evaluation-semantics", Category: "artifact", Status: StatusFail, Message: status.Message})
		return report
	}
	report.Checks = append(report.Checks, Check{ID: "artifact.evaluation-semantics", Category: "artifact", Status: StatusWarn, Message: status.Message})
	return report
}

func claudePermissionMergeMode(p policy.Pipeline) string {
	switch strings.TrimSpace(p.ClaudePermissionMergeMode) {
	case claude.MergeModeMigrationCompat:
		return claude.MergeModeMigrationCompat
	case claude.MergeModeStrict:
		return claude.MergeModeStrict
	case claude.MergeModeCCBashProxyAuthoritative:
		return claude.MergeModeCCBashProxyAuthoritative
	default:
		return claude.MergeModeStrict
	}
}

func HasFailures(report Report) bool {
	for _, check := range report.Checks {
		if check.Status == StatusFail {
			return true
		}
	}
	return false
}

func testsPass(p policy.Pipeline, tool string, cwd string, home string) error {
	checkPermission := func(scope string, rules []policy.PermissionRuleSpec, effect string) error {
		for i, rule := range rules {
			var expect []string
			switch effect {
			case "deny":
				expect = rule.Test.Deny
			case "ask":
				expect = rule.Test.Ask
			case "allow":
				expect = rule.Test.Allow
			}
			for _, ex := range expect {
				if !permissionRuleMatchesEffect(rule, ex, effect) {
					return &exampleError{Scope: scope, Name: scopeName(scope, i), Kind: "expect", Example: ex}
				}
			}
			for _, ex := range rule.Test.Pass {
				if permissionRuleMatchesEffect(rule, ex, effect) {
					return &exampleError{Scope: scope, Name: scopeName(scope, i), Kind: "pass", Example: ex}
				}
			}
		}
		return nil
	}
	if err := checkPermission("permission.deny", p.Permission.Deny, "deny"); err != nil {
		return err
	}
	if err := checkPermission("permission.ask", p.Permission.Ask, "ask"); err != nil {
		return err
	}
	if err := checkPermission("permission.allow", p.Permission.Allow, "allow"); err != nil {
		return err
	}

	for i, ex := range p.Test {
		decision, err := policy.Evaluate(p, ex.In)
		if err != nil {
			return err
		}
		decision = claude.ApplyPermissionBridgeWithMode(tool, decision, cwd, home, p.ClaudePermissionMergeMode)
		if decision.Outcome != ex.Decision {
			return &exampleError{Scope: "test", Name: scopeName("e2e", i), Kind: "decision", Example: ex.In}
		}
	}
	return nil
}

func envOnlyAllowNames(p policy.Pipeline) []string {
	var ids []string
	for i, rule := range p.Permission.Allow {
		if policy.IsZeroPermissionCommandSpec(rule.Command) && len(rule.Patterns) == 0 && !policy.IsZeroPermissionEnvSpec(rule.Env) {
			ids = append(ids, scopeName("permission.allow", i))
		}
	}
	return ids
}

func permissionRuleMatchesEffect(rule policy.PermissionRuleSpec, command string, effect string) bool {
	if effect == "allow" {
		return policy.PermissionAllowRuleMatches(rule, command)
	}
	return policy.PermissionRuleMatches(rule, command)
}

type exampleError struct {
	Scope   string
	Name    string
	Kind    string
	Example string
}

func (e *exampleError) Error() string {
	return e.Scope + " " + e.Name + " has failing " + e.Kind + " example: " + e.Example
}

func scopeName(prefix string, idx int) string {
	return prefix + "[" + fmtInt(idx) + "]"
}

func fmtInt(v int) string {
	return strconv.Itoa(v)
}

func claudeHookRegistrationCheck(path string) Check {
	check := Check{ID: "install.claude-registered", Category: "install", Status: StatusWarn}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			check.Message = "Claude Code settings.json not found"
			return check
		}
		check.Message = "Claude Code settings.json could not be read: " + err.Error()
		return check
	}

	registration, err := inspectClaudeHookRegistration(data)
	if err != nil {
		check.Message = "Claude Code settings.json is malformed JSON: " + err.Error()
		return check
	}
	if registration.BashHookCommand != "" {
		check.Status = StatusPass
		if strings.Contains(registration.BashHookCommand, "--rtk") {
			check.Message = "Claude Code Bash hook registration detected with --rtk"
		} else {
			check.Message = "Claude Code Bash hook registration detected without --rtk"
		}
		return check
	}
	if registration.NonBashHookCommand != "" {
		check.Message = "cc-bash-proxy hook exists but matcher is not Bash"
		return check
	}
	if registration.BashMatcher {
		check.Message = "Bash matcher exists but cc-bash-proxy hook is missing"
		return check
	}
	check.Message = "Claude Code settings found but cc-bash-proxy hook not detected"
	return check
}

type claudeHookRegistration struct {
	BashHookCommand    string
	NonBashHookCommand string
	BashMatcher        bool
}

func inspectClaudeHookRegistration(data []byte) (claudeHookRegistration, error) {
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return claudeHookRegistration{}, err
	}
	return findClaudeHookRegistration(payload, ""), nil
}

func extractClaudeHookCommand(raw string) (string, bool) {
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", false
	}
	command := findHookCommand(payload)
	return command, strings.HasPrefix(command, "/")
}

func findHookCommand(node any) string {
	switch v := node.(type) {
	case map[string]any:
		if command, ok := v["command"].(string); ok && strings.Contains(command, "cc-bash-proxy hook") {
			return command
		}
		for _, value := range v {
			if command := findHookCommand(value); command != "" {
				return command
			}
		}
	case []any:
		for _, value := range v {
			if command := findHookCommand(value); command != "" {
				return command
			}
		}
	}
	return ""
}

func findClaudeHookRegistration(node any, inheritedMatcher string) claudeHookRegistration {
	var result claudeHookRegistration
	switch v := node.(type) {
	case map[string]any:
		matcher := inheritedMatcher
		if value, ok := v["matcher"].(string); ok {
			matcher = value
			if value == "Bash" {
				result.BashMatcher = true
			}
		}
		if command, ok := v["command"].(string); ok && strings.Contains(command, "cc-bash-proxy hook") {
			if matcher == "Bash" {
				result.BashHookCommand = command
			} else {
				result.NonBashHookCommand = command
			}
		}
		for _, value := range v {
			result = mergeClaudeHookRegistration(result, findClaudeHookRegistration(value, matcher))
		}
	case []any:
		for _, value := range v {
			result = mergeClaudeHookRegistration(result, findClaudeHookRegistration(value, inheritedMatcher))
		}
	}
	return result
}

func mergeClaudeHookRegistration(a claudeHookRegistration, b claudeHookRegistration) claudeHookRegistration {
	if a.BashHookCommand == "" {
		a.BashHookCommand = b.BashHookCommand
	}
	if a.NonBashHookCommand == "" {
		a.NonBashHookCommand = b.NonBashHookCommand
	}
	a.BashMatcher = a.BashMatcher || b.BashMatcher
	return a
}
