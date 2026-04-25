package command

import "strings"

type KubectlParser struct{}

func (KubectlParser) Program() string {
	return "kubectl"
}

func (KubectlParser) Parse(base Command) (Command, bool) {
	if base.Program != "kubectl" {
		return Command{}, false
	}

	cmd := base
	cmd.Parser = KubectlParser{}.Program()
	cmd.SemanticParser = KubectlParser{}.Program()
	cmd.Args = []string{}

	semantic := &KubectlSemantic{}
	var positionals []string

	for i := 0; i < len(base.RawWords); i++ {
		word := base.RawWords[i]
		switch {
		case kubectlOptionWithValue(word, "-n", "--namespace"):
			value, consumed := kubectlOptionValue(word, "--namespace", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: kubectlOptionName(word, "-n", "--namespace"), Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Namespace = value
				i += kubectlConsumedWords(word)
			}
		case kubectlOptionWithValue(word, "", "--context"):
			value, consumed := kubectlOptionValue(word, "--context", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--context", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Context = value
				i += kubectlConsumedWords(word)
			}
		case kubectlOptionWithValue(word, "", "--kubeconfig"):
			value, consumed := kubectlOptionValue(word, "--kubeconfig", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--kubeconfig", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Kubeconfig = value
				i += kubectlConsumedWords(word)
			}
		case kubectlOptionWithValue(word, "-f", "--filename"):
			value, consumed := kubectlOptionValue(word, "--filename", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: kubectlOptionName(word, "-f", "--filename"), Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Filenames = append(semantic.Filenames, value)
				i += kubectlConsumedWords(word)
			}
		case kubectlOptionWithValue(word, "-l", "--selector"):
			value, consumed := kubectlOptionValue(word, "--selector", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: kubectlOptionName(word, "-l", "--selector"), Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Selectors = append(semantic.Selectors, value)
				i += kubectlConsumedWords(word)
			}
		case kubectlOptionWithValue(word, "-c", "--container"):
			value, consumed := kubectlOptionValue(word, "--container", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: kubectlOptionName(word, "-c", "--container"), Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Container = value
				i += kubectlConsumedWords(word)
			}
		case word == "-A" || word == "--all-namespaces":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.AllNamespaces = true
		case word == "--dry-run" || strings.HasPrefix(word, "--dry-run="):
			option := parseOptionWord(word, i)
			cmd.Options = append(cmd.Options, option)
			v := true
			semantic.DryRun = &v
		case word == "--force":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Force = true
		case word == "-R" || word == "--recursive":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Recursive = true
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.Options = append(cmd.Options, parseOptionWord(word, i))
		default:
			positionals = append(positionals, word)
		}
	}

	semantic.Verb, semantic.Subverb, semantic.ResourceType, semantic.ResourceName = kubectlAction(positionals)
	cmd.ActionPath = kubectlActionPath(semantic.Verb, semantic.Subverb)
	cmd.Args = kubectlArgs(positionals, len(cmd.ActionPath))
	semantic.Flags = normalizedKubectlFlags(cmd.Options, semantic.Subverb)
	cmd.Kubectl = semantic
	cmd.Namespace = semantic.Namespace
	cmd.ResourceType = semantic.ResourceType
	cmd.ResourceName = semantic.ResourceName
	return cmd, true
}

func kubectlOptionWithValue(word string, short string, long string) bool {
	if short != "" && word == short {
		return true
	}
	return word == long || strings.HasPrefix(word, long+"=")
}

func kubectlOptionName(word string, short string, long string) string {
	if short != "" && word == short {
		return short
	}
	return long
}

func kubectlOptionValue(word string, long string, words []string, i int) (string, bool) {
	if value, ok := strings.CutPrefix(word, long+"="); ok {
		return value, true
	}
	if i+1 >= len(words) {
		return "", false
	}
	return words[i+1], true
}

func kubectlConsumedWords(word string) int {
	if strings.Contains(word, "=") {
		return 0
	}
	return 1
}

func kubectlAction(positionals []string) (string, string, string, string) {
	if len(positionals) == 0 {
		return "", "", "", ""
	}
	verb := positionals[0]
	offset := 1
	subverb := ""
	if verb == "rollout" && len(positionals) > 1 {
		subverb = positionals[1]
		offset = 2
	}
	if len(positionals) <= offset {
		return verb, subverb, "", ""
	}
	resourceType, resourceName := splitKubectlResource(positionals[offset])
	if resourceName == "" && len(positionals) > offset+1 {
		resourceName = positionals[offset+1]
	}
	return verb, subverb, resourceType, resourceName
}

func splitKubectlResource(word string) (string, string) {
	resourceType, resourceName, ok := strings.Cut(word, "/")
	if !ok {
		return word, ""
	}
	return resourceType, resourceName
}

func kubectlActionPath(verb string, subverb string) []string {
	var action []string
	if verb != "" {
		action = append(action, verb)
	}
	if subverb != "" {
		action = append(action, subverb)
	}
	return action
}

func kubectlArgs(positionals []string, actionLen int) []string {
	if len(positionals) <= actionLen {
		return []string{}
	}
	return append([]string(nil), positionals[actionLen:]...)
}

func normalizedKubectlFlags(options []Option, subverb string) []string {
	flags := make([]string, 0, len(options)+1)
	if subverb != "" {
		flags = append(flags, subverb)
	}
	for _, option := range options {
		flags = append(flags, option.Name)
		if option.HasValue {
			flags = append(flags, option.Name+"="+option.Value)
		}
	}
	return flags
}
