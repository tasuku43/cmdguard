package policy

import (
	"strings"

	commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"
)

type semanticHandler struct {
	command  string
	match    func(SemanticMatchSpec, commandpkg.Command) bool
	validate func(string, SemanticMatchSpec) []string
}

var semanticHandlers []semanticHandler

func registerSemanticHandler(handler semanticHandler) {
	if strings.TrimSpace(handler.command) == "" || handler.match == nil || handler.validate == nil {
		return
	}
	semanticHandlers = append(semanticHandlers, handler)
}

func lookupSemanticHandler(command string) (semanticHandler, bool) {
	command = strings.TrimSpace(command)
	for _, handler := range semanticHandlers {
		if handler.command == command {
			return handler, true
		}
	}
	return semanticHandler{}, false
}

func permissionSemanticMatches(command string, semantic SemanticMatchSpec, cmd commandpkg.Command) bool {
	handler, ok := lookupSemanticHandler(command)
	if !ok {
		return false
	}
	return handler.match(semantic, cmd)
}

func validateSemanticMatchSpec(command string, prefix string, semantic SemanticMatchSpec) []string {
	handler, ok := lookupSemanticHandler(command)
	if !ok {
		return nil
	}
	return handler.validate(prefix, semantic)
}
