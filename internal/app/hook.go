package app

import (
	"fmt"
	"strings"

	"github.com/tasuku43/cc-bash-guard/internal/adapter/claude"
	"github.com/tasuku43/cc-bash-guard/internal/adapter/hookinput"
	"github.com/tasuku43/cc-bash-guard/internal/domain/policy"
	"github.com/tasuku43/cc-bash-guard/internal/infra"
	"github.com/tasuku43/cc-bash-guard/internal/infra/buildinfo"
	configrepo "github.com/tasuku43/cc-bash-guard/internal/infra/config"
)

func RunHook(raw []byte, opts HookOptions, env Env) HookResult {
	req, err := hookinput.Normalize(raw)
	if err != nil {
		return HookResult{Payload: hookErrorPayload(claude.Tool, "invalid_input", err.Error())}
	}

	_, decision, err := EvaluateForCommand(req.Command, env, opts.AutoVerify)
	if err != nil {
		return HookResult{Payload: hookErrorPayload(claude.Tool, "invalid_config", err.Error())}
	}
	if opts.UseRTK && decision.Outcome != "deny" {
		decision = applyRTKRewrite(decision)
	}

	return HookResult{Payload: hookPayload(decision, req.Command)}
}

func evaluateDecision(req hookinput.ExecRequest, env Env, autoVerify bool) (policy.Decision, error) {
	_, decision, err := EvaluateForCommand(req.Command, env, autoVerify)
	return decision, err
}

func shouldAttemptImplicitVerify(errs []error) bool {
	if len(errs) == 0 {
		return false
	}
	for _, msg := range policy.ErrorStrings(errs) {
		if strings.Contains(msg, "verified artifact not found") || strings.Contains(msg, "changed since last verify") {
			return true
		}
	}
	return false
}

func ensureVerifiedArtifacts(env Env, tool string) error {
	info := buildinfo.Read()
	_, err := configrepo.VerifyEffectiveToAllCaches(env.Cwd, env.Home, env.XDGConfigHome, env.XDGCacheHome, tool, info.Version)
	return err
}

func hookPayload(decision policy.Decision, originalCommand string) map[string]any {
	switch decision.Outcome {
	case "allow", "ask":
		hookOutput := map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecisionReason": permissionDecisionReason(decision, "cc-bash-guard permission evaluated"),
		}
		// updatedInput is reserved for the RTK integration bridge. Supported
		// cc-bash-guard policy evaluation never rewrites commands.
		if decision.Command != originalCommand {
			hookOutput["updatedInput"] = map[string]any{"command": decision.Command}
		}
		if decision.Outcome == "allow" {
			hookOutput["permissionDecision"] = "allow"
		}
		payload := map[string]any{
			"hookSpecificOutput": hookOutput,
			"cc-bash-guard": map[string]any{
				"outcome":  decision.Outcome,
				"explicit": decision.Explicit,
				"reason":   decision.Reason,
				"trace":    decision.Trace,
			},
		}
		if message, ok := buildRewriteSystemMessage(decision); ok {
			payload["systemMessage"] = message
		}
		return payload
	case "deny":
		reason := permissionDecisionReason(decision, "cc-bash-guard denied by policy")
		return map[string]any{
			"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": reason,
			},
			"cc-bash-guard": map[string]any{
				"outcome":  "deny",
				"explicit": decision.Explicit,
				"reason":   decision.Reason,
				"trace":    decision.Trace,
			},
		}
	default:
		return hookErrorPayload(claude.Tool, "runtime_error", "unsupported decision outcome")
	}
}

func permissionDecisionReason(decision policy.Decision, fallback string) string {
	if reason := strings.TrimSpace(decision.Message); reason != "" {
		return reason
	}
	return fallback
}

func hookErrorPayload(tool string, code string, message string) map[string]any {
	return map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":            "PreToolUse",
			"permissionDecision":       "deny",
			"permissionDecisionReason": "cc-bash-guard " + tool + " " + code + ": " + message,
		},
	}
}

func applyRTKRewrite(decision policy.Decision) policy.Decision {
	// RTK integration only. cc-bash-guard itself does not rewrite commands; this
	// bridge delegates to the external rtk binary after policy evaluation.
	rewritten, ok := infra.RewriteRTK(decision.Command)
	if !ok || rewritten == decision.Command {
		return decision
	}
	decision.Trace = append(decision.Trace, policy.TraceStep{
		Action: "rewrite",
		Name:   "rtk",
		From:   decision.Command,
		To:     rewritten,
	})
	decision.Command = rewritten
	return decision
}

func buildRewriteSystemMessage(decision policy.Decision) (string, bool) {
	// RTK integration only. Current cc-bash-guard policy behavior does not
	// produce rewrite trace steps.
	ruleIDs := make([]string, 0, len(decision.Trace))
	for _, step := range decision.Trace {
		if step.Action != "rewrite" {
			continue
		}
		ruleIDs = append(ruleIDs, step.Name)
	}
	if len(ruleIDs) == 0 {
		return "", false
	}
	return fmt.Sprintf("cc-bash-guard: rewrote [%s] -> %s", strings.Join(ruleIDs, " -> "), decision.Command), true
}
