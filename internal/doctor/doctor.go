package doctor

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tasuku43/cmdproxy/internal/buildinfo"
	"github.com/tasuku43/cmdproxy/internal/config"
	"github.com/tasuku43/cmdproxy/internal/domain/policy"
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
	Checks []Check `json:"checks"`
}

func Run(loaded config.Loaded, home string) Report {
	var checks []Check

	if len(loaded.Errors) == 0 {
		checks = append(checks,
			Check{ID: "config.parse", Category: "config", Status: StatusPass, Message: "configuration files parsed"},
			Check{ID: "config.schema", Category: "config", Status: StatusPass, Message: "configuration schema is valid"},
			Check{ID: "rules.unique-id", Category: "rules", Status: StatusPass, Message: "rule IDs are unique"},
			Check{ID: "rules.matcher-validate", Category: "rules", Status: StatusPass, Message: "rule matchers are valid"},
			Check{ID: "rules.tests-present", Category: "rules", Status: StatusPass, Message: "directive tests are present"},
		)
	} else {
		msg := strings.Join(policy.ErrorStrings(loaded.Errors), "; ")
		checks = append(checks,
			Check{ID: "config.parse", Category: "config", Status: StatusFail, Message: msg},
			Check{ID: "config.schema", Category: "config", Status: StatusFail, Message: msg},
			Check{ID: "rules.unique-id", Category: "rules", Status: StatusFail, Message: msg},
			Check{ID: "rules.matcher-validate", Category: "rules", Status: StatusFail, Message: msg},
			Check{ID: "rules.tests-present", Category: "rules", Status: StatusFail, Message: msg},
		)
	}

	if len(loaded.Errors) == 0 {
		if err := testsPass(loaded.Rules); err != nil {
			checks = append(checks, Check{ID: "rules.tests-pass", Category: "rules", Status: StatusFail, Message: err.Error()})
		} else {
			checks = append(checks, Check{ID: "rules.tests-pass", Category: "rules", Status: StatusPass, Message: "directive tests match expectations"})
		}
	} else {
		checks = append(checks, Check{ID: "rules.tests-pass", Category: "rules", Status: StatusFail, Message: "skipped because configuration is invalid"})
	}

	if ids := relaxedRuleIDs(loaded.Rules); len(ids) > 0 {
		checks = append(checks, Check{ID: "rules.relaxed-contracts", Category: "rules", Status: StatusWarn, Message: "relaxed rewrite contracts enabled: " + strings.Join(ids, ", ")})
	} else {
		checks = append(checks, Check{ID: "rules.relaxed-contracts", Category: "rules", Status: StatusPass, Message: "all rewrite contracts use strict validation"})
	}

	if warning := broadnessWarning(loaded.Rules); warning != "" {
		checks = append(checks, Check{ID: "rules.pattern-broadness", Category: "diagnostics", Status: StatusWarn, Message: warning})
	} else {
		checks = append(checks, Check{ID: "rules.pattern-broadness", Category: "diagnostics", Status: StatusPass, Message: "patterns are not obviously broad"})
	}

	if warning := shadowingWarning(loaded.Rules); warning != "" {
		checks = append(checks, Check{ID: "rules.shadowing", Category: "diagnostics", Status: StatusWarn, Message: warning})
	} else {
		checks = append(checks, Check{ID: "rules.shadowing", Category: "diagnostics", Status: StatusPass, Message: "no obvious shadowing detected"})
	}

	if path, err := exec.LookPath("cmdproxy"); err == nil {
		checks = append(checks, Check{ID: "install.binary-on-path", Category: "install", Status: StatusPass, Message: "cmdproxy found on PATH at " + path})
	} else {
		checks = append(checks, Check{ID: "install.binary-on-path", Category: "install", Status: StatusWarn, Message: "cmdproxy not found on PATH"})
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

	claudeSettings := filepath.Join(home, ".claude", "settings.json")
	if _, err := os.Stat(claudeSettings); err == nil {
		data, readErr := os.ReadFile(claudeSettings)
		if readErr == nil && strings.Contains(string(data), "cmdproxy hook claude") && strings.Contains(string(data), "\"matcher\": \"Bash\"") {
			checks = append(checks, Check{ID: "install.claude-registered", Category: "install", Status: StatusPass, Message: "Claude Code hook registration detected"})
			hookCommand, absolutePath := extractClaudeHookCommand(string(data))
			if absolutePath {
				checks = append(checks, Check{ID: "install.claude-hook-path", Category: "install", Status: StatusPass, Message: "Claude Code hook appears to use an absolute cmdproxy path"})
				if path, ok := hookBinaryPath(hookCommand); ok {
					if stat, statErr := os.Stat(path); statErr != nil {
						checks = append(checks, Check{ID: "install.claude-hook-target", Category: "install", Status: StatusWarn, Message: "Claude Code hook target does not exist: " + path})
					} else if stat.Mode()&0o111 == 0 {
						checks = append(checks, Check{ID: "install.claude-hook-target", Category: "install", Status: StatusWarn, Message: "Claude Code hook target is not executable: " + path})
					} else {
						checks = append(checks, Check{ID: "install.claude-hook-target", Category: "install", Status: StatusPass, Message: "Claude Code hook target exists and is executable: " + path})
						if exe, exeErr := os.Executable(); exeErr == nil {
							if sameExecutable(exe, path) {
								checks = append(checks, Check{ID: "install.claude-hook-binary-match", Category: "install", Status: StatusPass, Message: "Claude Code hook target matches the running cmdproxy binary"})
							} else {
								checks = append(checks, Check{ID: "install.claude-hook-binary-match", Category: "install", Status: StatusWarn, Message: "Claude Code hook targets a different cmdproxy binary than the one being verified"})
							}
						}
					}
				}
			} else {
				checks = append(checks, Check{ID: "install.claude-hook-path", Category: "install", Status: StatusWarn, Message: "Claude Code hook uses PATH lookup; prefer an absolute cmdproxy path"})
			}
		} else {
			checks = append(checks, Check{ID: "install.claude-registered", Category: "install", Status: StatusWarn, Message: "Claude Code settings found but cmdproxy hook claude not detected"})
		}
	} else {
		checks = append(checks, Check{ID: "install.claude-registered", Category: "install", Status: StatusWarn, Message: "Claude Code settings.json not found"})
	}

	return Report{Checks: checks}
}

func HasFailures(report Report) bool {
	for _, check := range report.Checks {
		if check.Status == StatusFail {
			return true
		}
	}
	return false
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
		if command, ok := v["command"].(string); ok && strings.Contains(command, "cmdproxy hook claude") {
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

func hookBinaryPath(command string) (string, bool) {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "", false
	}
	if strings.HasPrefix(fields[0], "/") {
		return fields[0], true
	}
	return "", false
}

func sameExecutable(a, b string) bool {
	left, err := filepath.EvalSymlinks(a)
	if err != nil {
		left = a
	}
	right, err := filepath.EvalSymlinks(b)
	if err != nil {
		right = b
	}
	return left == right
}

func testsPass(rules []policy.Rule) error {
	for _, r := range rules {
		if strings.TrimSpace(r.Reject.Message) != "" {
			for _, ex := range r.Reject.Test.Expect {
				decision, err := policy.Evaluate([]policy.Rule{r}, ex)
				if err != nil {
					return err
				}
				if decision.Outcome != "reject" {
					return &exampleError{RuleID: r.ID, Kind: "reject", Example: ex}
				}
			}
			for _, ex := range r.Reject.Test.Pass {
				decision, err := policy.Evaluate([]policy.Rule{r}, ex)
				if err != nil {
					return err
				}
				if decision.Outcome != "pass" {
					return &exampleError{RuleID: r.ID, Kind: "pass", Example: ex}
				}
			}
			continue
		}
		for _, ex := range r.Rewrite.Test.Expect {
			decision, err := policy.Evaluate([]policy.Rule{r}, ex.In)
			if err != nil {
				return err
			}
			if decision.Outcome != "rewrite" || decision.Command != ex.Out {
				return &exampleError{RuleID: r.ID, Kind: "rewrite", Example: ex.In}
			}
		}
		for _, ex := range r.Rewrite.Test.Pass {
			decision, err := policy.Evaluate([]policy.Rule{r}, ex)
			if err != nil {
				return err
			}
			if decision.Outcome != "pass" {
				return &exampleError{RuleID: r.ID, Kind: "pass", Example: ex}
			}
		}
	}
	return nil
}

type exampleError struct {
	RuleID  string
	Kind    string
	Example string
}

func (e *exampleError) Error() string {
	return "rule " + e.RuleID + " has failing " + e.Kind + " example: " + e.Example
}

func broadnessWarning(rules []policy.Rule) string {
	for _, r := range rules {
		if r.Pattern == "" {
			continue
		}
		if r.Pattern == ".*" || r.Pattern == "^.*$" || r.Pattern == ".+" || r.Pattern == "^.+$" {
			return "rule " + r.ID + " pattern is extremely broad"
		}
	}
	return ""
}

func shadowingWarning(rules []policy.Rule) string {
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if strings.TrimSpace(rules[j].Reject.Message) != "" {
				for _, ex := range rules[j].Reject.Test.Expect {
					matched, err := rules[i].Match(ex)
					if err != nil {
						continue
					}
					if matched {
						return "rule " + rules[i].ID + " likely shadows later rule " + rules[j].ID
					}
				}
				continue
			}
			for _, ex := range rules[j].Rewrite.Test.Expect {
				matched, err := rules[i].Match(ex.In)
				if err != nil {
					continue
				}
				if matched {
					return "rule " + rules[i].ID + " likely shadows later rule " + rules[j].ID
				}
			}
		}
	}
	return ""
}

func relaxedRuleIDs(rules []policy.Rule) []string {
	var ids []string
	for _, rule := range rules {
		if !policy.IsZeroRewriteSpec(rule.Rewrite) && !policy.RewriteStrict(rule.Rewrite) {
			ids = append(ids, rule.ID)
		}
	}
	return ids
}
