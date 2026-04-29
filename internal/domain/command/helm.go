package command

import "strings"

type HelmParser struct{}

func init() {
	RegisterDefaultParser(HelmParser{})
}

func (HelmParser) Program() string {
	return "helm"
}

func (HelmParser) Parse(base Command) (Command, bool) {
	if base.Program != "helm" {
		return Command{}, false
	}

	cmd := base
	cmd.Parser = HelmParser{}.Program()
	cmd.SemanticParser = HelmParser{}.Program()
	cmd.Args = []string{}

	semantic := &HelmSemantic{}
	var positionals []string

	for i := 0; i < len(base.RawWords); i++ {
		word := base.RawWords[i]
		switch {
		case helmOptionWithValue(word, "-n", "--namespace"):
			value, consumed := helmOptionValue(word, "--namespace", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: helmOptionName(word, "-n", "--namespace"), Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Namespace = value
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "", "--kube-context"):
			value, consumed := helmOptionValue(word, "--kube-context", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--kube-context", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.KubeContext = value
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "", "--kubeconfig"):
			value, consumed := helmOptionValue(word, "--kubeconfig", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--kubeconfig", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Kubeconfig = value
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "-f", "--values"):
			value, consumed := helmOptionValue(word, "--values", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: helmOptionName(word, "-f", "--values"), Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.ValuesFiles = append(semantic.ValuesFiles, value)
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "", "--set"):
			value, consumed := helmOptionValue(word, "--set", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--set", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.SetKeys = append(semantic.SetKeys, helmSetKey(value))
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "", "--set-string"):
			value, consumed := helmOptionValue(word, "--set-string", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--set-string", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.SetStringKeys = append(semantic.SetStringKeys, helmSetKey(value))
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "", "--set-file"):
			value, consumed := helmOptionValue(word, "--set-file", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--set-file", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.SetFileKeys = append(semantic.SetFileKeys, helmSetKey(value))
				i += helmConsumedWords(word)
			}
		case helmOptionWithValue(word, "", "--cascade"):
			value, consumed := helmOptionValue(word, "--cascade", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--cascade", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Cascade = value
				i += helmConsumedWords(word)
			}
		case helmKnownValueOption(word):
			value, consumed := helmOptionValue(word, helmLongName(word), base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: helmLongName(word), Value: value, HasValue: consumed, Position: i})
			if consumed {
				i += helmConsumedWords(word)
			}
		case word == "--dry-run" || strings.HasPrefix(word, "--dry-run="):
			cmd.Options = append(cmd.Options, parseOptionWord(word, i))
			semantic.DryRun = true
		case word == "--force":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Force = true
		case word == "--atomic":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Atomic = true
		case word == "--wait":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Wait = true
		case word == "--wait-for-jobs":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.WaitForJobs = true
		case word == "--install" || word == "-i":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Install = true
		case word == "--reuse-values":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.ReuseValues = true
		case word == "--reset-values":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.ResetValues = true
		case word == "--reset-then-reuse-values":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.ResetThenReuseValues = true
		case word == "--cleanup-on-fail":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.CleanupOnFail = true
		case word == "--create-namespace":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.CreateNamespace = true
		case word == "--dependency-update":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.DependencyUpdate = true
		case word == "--devel":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Devel = true
		case word == "--keep-history":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.KeepHistory = true
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.Options = append(cmd.Options, parseOptionWord(word, i))
		default:
			positionals = append(positionals, word)
		}
	}

	populateHelmAction(semantic, positionals)
	cmd.ActionPath = helmActionPath(semantic.Verb, semantic.Subverb)
	cmd.Args = helmArgs(positionals, len(cmd.ActionPath))
	semantic.Flags = normalizedHelmFlags(cmd.Options)
	cmd.Helm = semantic
	cmd.Namespace = semantic.Namespace
	return cmd, true
}

func helmOptionWithValue(word string, short string, long string) bool {
	if short != "" && word == short {
		return true
	}
	return word == long || strings.HasPrefix(word, long+"=")
}

func helmOptionName(word string, short string, long string) string {
	if short != "" && word == short {
		return short
	}
	return long
}

func helmOptionValue(word string, long string, words []string, i int) (string, bool) {
	if value, ok := strings.CutPrefix(word, long+"="); ok {
		return value, true
	}
	if i+1 >= len(words) {
		return "", false
	}
	return words[i+1], true
}

func helmConsumedWords(word string) int {
	if strings.Contains(word, "=") {
		return 0
	}
	return 1
}

func helmLongName(word string) string {
	name, _, _ := strings.Cut(word, "=")
	return name
}

func helmKnownValueOption(word string) bool {
	name := helmLongName(word)
	switch name {
	case "--registry-config", "--repository-cache", "--repository-config", "--burst-limit", "--qps":
		return word == name || strings.HasPrefix(word, name+"=")
	default:
		return false
	}
}

func helmSetKey(value string) string {
	key, _, ok := strings.Cut(value, "=")
	if !ok {
		return value
	}
	return key
}

func populateHelmAction(h *HelmSemantic, positionals []string) {
	if len(positionals) == 0 {
		return
	}
	h.Verb = positionals[0]
	offset := 1
	if helmGroupedVerb(h.Verb) && len(positionals) > 1 {
		h.Subverb = positionals[1]
		offset = 2
	}
	args := positionals[offset:]
	switch h.Verb {
	case "install":
		if len(args) > 0 {
			h.Release = args[0]
		}
		if len(args) > 1 {
			h.Chart = args[1]
		}
	case "upgrade", "template":
		if len(args) > 0 {
			h.Release = args[0]
		}
		if len(args) > 1 {
			h.Chart = args[1]
		}
	case "status", "history", "uninstall", "rollback", "test":
		if len(args) > 0 {
			h.Release = args[0]
		}
	case "get":
		if len(args) > 0 {
			h.Release = args[0]
		}
	case "package", "pull", "push", "verify", "lint":
		if len(args) > 0 {
			h.Chart = args[0]
		}
	case "repo":
		if len(args) > 0 && (h.Subverb == "add" || h.Subverb == "remove") {
			h.RepoName = args[0]
		}
		if len(args) > 1 && h.Subverb == "add" {
			h.RepoURL = args[1]
		}
	case "registry":
		if len(args) > 0 && (h.Subverb == "login" || h.Subverb == "logout") {
			h.Registry = args[0]
		}
	case "plugin":
		if len(args) > 0 {
			h.PluginName = args[0]
		}
	}
}

func helmGroupedVerb(verb string) bool {
	switch verb {
	case "repo", "registry", "plugin", "dependency", "get", "show", "search", "completion":
		return true
	default:
		return false
	}
}

func helmActionPath(verb string, subverb string) []string {
	var action []string
	if verb != "" {
		action = append(action, verb)
	}
	if subverb != "" {
		action = append(action, subverb)
	}
	return action
}

func helmArgs(positionals []string, actionLen int) []string {
	if len(positionals) <= actionLen {
		return []string{}
	}
	return append([]string(nil), positionals[actionLen:]...)
}

func normalizedHelmFlags(options []Option) []string {
	flags := make([]string, 0, len(options)*2)
	for _, option := range options {
		flags = append(flags, option.Name)
		if option.HasValue {
			flags = append(flags, option.Name+"="+option.Value)
		}
	}
	return flags
}
