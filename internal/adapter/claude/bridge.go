package claude

import (
	"strings"

	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
	"github.com/tasuku43/cc-bash-guard/internal/domain/policy"
)

const Tool = "claude"

func Supported(tool string) bool {
	switch strings.TrimSpace(tool) {
	case Tool:
		return true
	default:
		return false
	}
}

func ApplyPermissionBridge(tool string, decision policy.Decision, cwd string, home string) policy.Decision {
	switch strings.TrimSpace(tool) {
	case Tool:
		return applyPermissionBridge(decision, cwd, home)
	default:
		return decision
	}
}

func applyPermissionBridge(decision policy.Decision, cwd string, home string) policy.Decision {
	verdict := CheckCommand(decision.Command, cwd, home)
	claudeOutcome := permissionVerdictOutcome(verdict)
	decision.Trace = append(decision.Trace, ccBashGuardPolicyTrace(decision))
	decision.Trace = append(decision.Trace, policy.TraceStep{
		Action:  "permission",
		Name:    "claude_settings",
		Effect:  claudeOutcome,
		Message: claudeSettingsTraceMessage(claudeOutcome),
	})
	if merged, ok := mergeCompositionPermissionSources(decision, cwd, home, claudeOutcome); ok {
		return merged
	}
	return mergePermissionSources(decision, claudeOutcome)
}

func ccBashGuardPolicyTrace(decision policy.Decision) policy.TraceStep {
	outcome := strings.TrimSpace(decision.Outcome)
	if outcome == "" {
		outcome = "abstain"
	}
	reason := "explicit"
	if outcome == "abstain" {
		reason = "abstain"
	}
	step := policy.TraceStep{
		Action: "permission",
		Name:   "cc_bash_guard_policy",
		Effect: outcome,
		Reason: reason,
	}
	if matched := matchedPolicyTraceName(decision.Trace); matched != "" {
		step.Message = "matched rule: " + matched
	}
	return step
}

func matchedPolicyTraceName(trace []policy.TraceStep) string {
	for i := len(trace) - 1; i >= 0; i-- {
		step := trace[i]
		if step.Action != "permission" {
			continue
		}
		if step.Name == "no_match" || step.Name == "fail_closed" || step.Name == "composition" || step.Name == "composition.command" {
			continue
		}
		if strings.TrimSpace(step.Name) != "" {
			return step.Name
		}
	}
	return ""
}

func permissionVerdictOutcome(verdict PermissionVerdict) string {
	switch verdict {
	case PermissionDeny:
		return "deny"
	case PermissionAsk:
		return "ask"
	case PermissionAllow:
		return "allow"
	default:
		return "abstain"
	}
}

func claudeSettingsTraceMessage(outcome string) string {
	switch outcome {
	case "deny":
		return "Claude settings deny matched"
	case "ask":
		return "Claude settings ask matched"
	case "allow":
		return "Claude settings allow matched"
	default:
		return "Claude settings did not define a matching permission"
	}
}

func mergePermissionSources(decision policy.Decision, claudeOutcome string) policy.Decision {
	ccOutcome := strings.TrimSpace(decision.Outcome)
	if ccOutcome == "" {
		ccOutcome = "abstain"
	}

	switch {
	case ccOutcome == "deny" || claudeOutcome == "deny":
		if claudeOutcome == "deny" {
			decision.Outcome = "deny"
			decision.Explicit = true
			decision.Reason = "claude_settings"
			if strings.TrimSpace(decision.Message) == "" {
				decision.Message = "blocked by Claude settings"
			}
		}
		decision.Trace = append(decision.Trace, finalMergeTrace("deny", "source denied"))
	case ccOutcome == "ask" || claudeOutcome == "ask":
		if ccOutcome != "ask" && claudeOutcome == "ask" {
			decision.Outcome = "ask"
			decision.Explicit = true
			decision.Reason = "claude_settings"
		}
		decision.Trace = append(decision.Trace, finalMergeTrace("ask", "source asked"))
	case ccOutcome == "allow" || claudeOutcome == "allow":
		if ccOutcome != "allow" && claudeOutcome == "allow" {
			decision.Outcome = "allow"
			decision.Explicit = true
			decision.Reason = "claude_settings"
		}
		decision.Trace = append(decision.Trace, finalMergeTrace("allow", "source allowed"))
	default:
		decision.Outcome = "ask"
		decision.Explicit = false
		decision.Reason = "default_fallback"
		decision.Trace = append(decision.Trace, finalMergeTrace("ask", "all sources abstained; fallback ask"))
	}
	return decision
}

func mergeCompositionPermissionSources(decision policy.Decision, cwd string, home string, claudeOutcome string) (policy.Decision, bool) {
	if claudeOutcome == "deny" || claudeOutcome == "ask" {
		return policy.Decision{}, false
	}

	commandSteps := compositionCommandTraceIndexes(decision.Trace)
	if len(commandSteps) == 0 {
		return policy.Decision{}, false
	}

	plan := commandpkg.Parse(decision.Command)
	if plan.Shape.Kind == commandpkg.ShellShapeSimple {
		return policy.Decision{}, false
	}

	firstDeny := -1
	firstAsk := -1
	allAllowed := true
	for _, traceIndex := range commandSteps {
		step := decision.Trace[traceIndex]
		ccOutcome := compositionCommandCCOutcome(step)
		segmentClaudeOutcome := permissionVerdictOutcome(CheckCommand(step.Command, cwd, home))
		mergedOutcome, mergeReason := mergePermissionSourceOutcomes(ccOutcome, segmentClaudeOutcome)
		if mergedOutcome != "allow" {
			allAllowed = false
		}
		if mergedOutcome == "deny" && firstDeny < 0 {
			firstDeny = commandIndexValue(step.CommandIndex)
		}
		if mergedOutcome == "ask" && firstAsk < 0 {
			firstAsk = commandIndexValue(step.CommandIndex)
		}
		step.Effect = mergedOutcome
		step.Reason = mergeReason
		if ccOutcome == "abstain" && segmentClaudeOutcome != "abstain" {
			step.Message = claudeSettingsTraceMessage(segmentClaudeOutcome)
		}
		decision.Trace[traceIndex] = step
	}

	switch {
	case firstDeny >= 0:
		decision.Outcome = "deny"
		decision.Explicit = true
		decision.Reason = "composition"
		updateCompositionFinalTrace(&decision, "deny", "command["+itoa(firstDeny)+"] denied", plan)
		decision.Trace = append(decision.Trace, finalMergeTrace("deny", "source denied"))
	case firstAsk >= 0:
		decision.Outcome = "ask"
		decision.Explicit = true
		decision.Reason = "composition"
		updateCompositionFinalTrace(&decision, "ask", "command["+itoa(firstAsk)+"] asked", plan)
		decision.Trace = append(decision.Trace, finalMergeTrace("ask", "source asked"))
	case allAllowed && claudeCompositionAllows(plan.Shape):
		decision.Outcome = "allow"
		decision.Explicit = true
		decision.Reason = "composition"
		updateCompositionFinalTrace(&decision, "allow", "all commands allowed", plan)
		decision.Trace = append(decision.Trace, finalMergeTrace("allow", "source allowed"))
	default:
		decision.Outcome = "ask"
		decision.Explicit = true
		decision.Reason = "composition"
		updateCompositionFinalTrace(&decision, "ask", unsafeCompositionReason(plan.Shape), plan)
		decision.Trace = append(decision.Trace, finalMergeTrace("ask", "source asked"))
	}
	return decision, true
}

func compositionCommandTraceIndexes(trace []policy.TraceStep) []int {
	var indexes []int
	for i, step := range trace {
		if step.Action == "permission" && step.Name == "composition.command" {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func compositionCommandCCOutcome(step policy.TraceStep) string {
	outcome := strings.TrimSpace(step.Effect)
	if outcome == "" {
		return "abstain"
	}
	if outcome == "ask" && !step.Explicit {
		return "abstain"
	}
	return outcome
}

func mergePermissionSourceOutcomes(ccOutcome string, claudeOutcome string) (string, string) {
	switch {
	case ccOutcome == "deny" || claudeOutcome == "deny":
		return "deny", "source denied"
	case ccOutcome == "ask" || claudeOutcome == "ask":
		return "ask", "source asked"
	case ccOutcome == "allow" || claudeOutcome == "allow":
		return "allow", "source allowed"
	default:
		return "ask", "all sources abstained; fallback ask"
	}
}

func commandIndexValue(index *int) int {
	if index == nil {
		return 0
	}
	return *index
}

func updateCompositionFinalTrace(decision *policy.Decision, outcome string, reason string, plan commandpkg.CommandPlan) {
	for i := len(decision.Trace) - 1; i >= 0; i-- {
		if decision.Trace[i].Name != "composition" {
			continue
		}
		decision.Trace[i].Effect = outcome
		decision.Trace[i].Reason = reason
		decision.Trace[i].Shape = string(plan.Shape.Kind)
		decision.Trace[i].ShapeFlags = plan.Shape.Flags()
		return
	}
	decision.Trace = append(decision.Trace, policy.TraceStep{
		Action:     "permission",
		Name:       "composition",
		Effect:     outcome,
		Reason:     reason,
		Shape:      string(plan.Shape.Kind),
		ShapeFlags: plan.Shape.Flags(),
	})
}

func unsafeCompositionReason(shape commandpkg.ShellShape) string {
	switch {
	case shape.HasProcessSubstitution:
		return "process substitution requires confirmation"
	case shape.HasCommandSubstitution:
		return "command substitution requires confirmation"
	case shape.HasRedirection:
		return "redirection requires confirmation"
	case shape.HasSubshell:
		return "subshell requires confirmation"
	case shape.HasBackground:
		return "background execution requires confirmation"
	case shape.HasPipeline && (shape.HasConditional || shape.HasSequence):
		return "pipeline compound shape requires confirmation"
	case shape.Kind == commandpkg.ShellShapeUnknown:
		return "unknown shell shape"
	default:
		return "unsafe command shape"
	}
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

func finalMergeTrace(outcome string, reason string) policy.TraceStep {
	return policy.TraceStep{
		Action:  "permission",
		Name:    "permission_sources_merge",
		Effect:  outcome,
		Reason:  reason,
		Message: "permission sources merged using deny > ask > allow > abstain",
	}
}
