package rule

import (
	"path/filepath"
	"strings"
	"unicode"
)

type ParsedCommand struct {
	EnvAssignments map[string]string
	Command        string
	Subcommand     string
	Args           []string
}

func ParseCommand(command string) ParsedCommand {
	tokens := tokenizeCommand(command)
	envAssignments := map[string]string{}

	i := 0
	for i < len(tokens) && isEnvAssignment(tokens[i]) {
		name, value, _ := strings.Cut(tokens[i], "=")
		envAssignments[name] = value
		i++
	}

	commandToken, args := unwrapCommand(tokens[i:])
	parsed := ParsedCommand{
		EnvAssignments: envAssignments,
		Command:        basenameCommand(commandToken),
		Args:           args,
	}
	if len(args) > 0 {
		parsed.Subcommand = args[0]
	}
	return parsed
}

func tokenizeCommand(command string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range command {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case inSingle:
			if r == '\'' {
				inSingle = false
			} else {
				current.WriteRune(r)
			}
		case inDouble:
			switch r {
			case '"':
				inDouble = false
			case '\\':
				escaped = true
			default:
				current.WriteRune(r)
			}
		default:
			switch {
			case unicode.IsSpace(r):
				flush()
			case r == '\'':
				inSingle = true
			case r == '"':
				inDouble = true
			case r == '\\':
				escaped = true
			default:
				current.WriteRune(r)
			}
		}
	}
	flush()
	return tokens
}

func unwrapCommand(tokens []string) (string, []string) {
	if len(tokens) == 0 {
		return "", nil
	}

	i := 0
	for i < len(tokens) {
		token := basenameCommand(tokens[i])
		switch token {
		case "command", "exec", "nohup":
			i++
			continue
		case "env":
			i++
			for i < len(tokens) && isEnvAssignment(tokens[i]) {
				i++
			}
			continue
		case "sudo":
			i = skipSudoWrapper(tokens, i+1)
			continue
		case "timeout":
			i = skipTimeoutWrapper(tokens, i+1)
			continue
		case "busybox":
			if i+1 < len(tokens) && isShellCommand(tokens[i+1]) {
				return tokens[i+1], append([]string(nil), tokens[i+2:]...)
			}
			return tokens[i], append([]string(nil), tokens[i+1:]...)
		default:
			return tokens[i], append([]string(nil), tokens[i+1:]...)
		}
	}
	return "", nil
}

func skipSudoWrapper(tokens []string, i int) int {
	for i < len(tokens) {
		token := tokens[i]
		if token == "--" {
			return i + 1
		}
		if isEnvAssignment(token) {
			i++
			continue
		}
		if !strings.HasPrefix(token, "-") || token == "-" {
			return i
		}
		if sudoOptionConsumesValue(token) && i+1 < len(tokens) {
			i += 2
			continue
		}
		i++
	}
	return i
}

func sudoOptionConsumesValue(token string) bool {
	switch token {
	case "-u", "--user", "-g", "--group", "-h", "--host", "-p", "--prompt", "-r", "--role", "-t", "--type", "-C", "--close-from", "-D", "--chdir":
		return true
	}
	return false
}

func skipTimeoutWrapper(tokens []string, i int) int {
	for i < len(tokens) {
		token := tokens[i]
		if token == "--" {
			i++
			break
		}
		if !strings.HasPrefix(token, "-") || token == "-" {
			break
		}
		if timeoutOptionConsumesValue(token) && i+1 < len(tokens) {
			i += 2
			continue
		}
		i++
	}
	if i < len(tokens) {
		i++
	}
	return i
}

func timeoutOptionConsumesValue(token string) bool {
	switch token {
	case "-k", "--kill-after", "-s", "--signal":
		return true
	}
	return false
}

func isShellCommand(token string) bool {
	switch basenameCommand(token) {
	case "bash", "sh", "zsh", "dash", "ksh":
		return true
	default:
		return false
	}
}

func basenameCommand(token string) string {
	if token == "" {
		return ""
	}
	return filepath.Base(token)
}

func isEnvAssignment(token string) bool {
	name, value, ok := strings.Cut(token, "=")
	if !ok || name == "" || value == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
