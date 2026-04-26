package command

import "strings"

type ArgoCDParser struct{}

func (ArgoCDParser) Program() string {
	return "argocd"
}

func (ArgoCDParser) Parse(base Command) (Command, bool) {
	if base.Program != "argocd" {
		return Command{}, false
	}

	cmd := base
	cmd.Parser = ArgoCDParser{}.Program()
	cmd.SemanticParser = ArgoCDParser{}.Program()
	cmd.Args = []string{}

	var positionals []string
	for i := 0; i < len(base.RawWords); i++ {
		word := base.RawWords[i]
		switch {
		case argocdOptionWithValue(word, "", "--project"):
			value, consumed := argocdOptionValue(word, "--project", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--project", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case argocdOptionWithValue(word, "", "--revision"):
			value, consumed := argocdOptionValue(word, "--revision", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--revision", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.Options = append(cmd.Options, parseOptionWord(word, i))
		default:
			positionals = append(positionals, word)
		}
	}

	cmd.ActionPath = argocdActionPath(positionals)
	cmd.Args = argocdArgs(positionals)
	cmd.ArgoCD = buildArgoCDSemantic(cmd.ActionPath, cmd.Args, cmd.Options)
	return cmd, true
}

func argocdActionPath(positionals []string) []string {
	if len(positionals) == 0 {
		return []string{}
	}
	if positionals[0] == "app" && len(positionals) > 1 {
		return append([]string(nil), positionals[:2]...)
	}
	return append([]string(nil), positionals[:1]...)
}

func argocdArgs(positionals []string) []string {
	if len(positionals) == 0 {
		return []string{}
	}
	if positionals[0] == "app" {
		if len(positionals) > 2 {
			return append([]string(nil), positionals[2:]...)
		}
		return []string{}
	}
	if len(positionals) > 1 {
		return append([]string(nil), positionals[1:]...)
	}
	return []string{}
}

func buildArgoCDSemantic(actionPath []string, args []string, options []Option) *ArgoCDSemantic {
	semantic := &ArgoCDSemantic{
		Verb:     strings.Join(actionPath, " "),
		Project:  lastArgoCDOptionValue(options, "--project"),
		Revision: lastArgoCDOptionValue(options, "--revision"),
		Flags:    normalizedArgoCDFlags(options),
	}
	if len(args) > 0 {
		semantic.AppName = args[0]
	}
	if semantic.Verb == "app rollback" && semantic.Revision == "" && len(args) > 1 {
		semantic.Revision = args[1]
	}
	return semantic
}

func argocdOptionWithValue(word string, short string, long string) bool {
	if short != "" && word == short {
		return true
	}
	return word == long || strings.HasPrefix(word, long+"=")
}

func argocdOptionValue(word string, long string, words []string, i int) (string, bool) {
	if value, ok := strings.CutPrefix(word, long+"="); ok {
		return value, true
	}
	if i+1 >= len(words) {
		return "", false
	}
	return words[i+1], true
}

func lastArgoCDOptionValue(options []Option, names ...string) string {
	value := ""
	for _, option := range options {
		if !option.HasValue {
			continue
		}
		for _, name := range names {
			if option.Name == name {
				value = option.Value
			}
		}
	}
	return value
}

func normalizedArgoCDFlags(options []Option) []string {
	flags := make([]string, 0, len(options)*2)
	for _, option := range options {
		flags = append(flags, option.Name)
		if option.HasValue {
			flags = append(flags, option.Name+"="+option.Value)
		}
	}
	return flags
}
