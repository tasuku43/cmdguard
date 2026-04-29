package command

import (
	"strconv"
	"strings"
)

type TerraformParser struct{}

func init() {
	RegisterDefaultParser(TerraformParser{})
}

func (TerraformParser) Program() string {
	return "terraform"
}

func (TerraformParser) Parse(base Command) (Command, bool) {
	if base.Program != "terraform" {
		return Command{}, false
	}

	cmd := base
	cmd.Parser = TerraformParser{}.Program()
	cmd.SemanticParser = TerraformParser{}.Program()
	cmd.Args = []string{}
	semantic := &TerraformSemantic{}

	i := 0
	for i < len(base.RawWords) {
		word := base.RawWords[i]
		switch {
		case word == "-version" || word == "--version":
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: word, Position: i})
			semantic.Subcommand = "version"
			cmd.ActionPath = []string{"version"}
			i++
		case terraformOptionWithValue(word, "-chdir"):
			value, consumed := terraformOptionValue(word, "-chdir", base.RawWords, i)
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "-chdir", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.GlobalChdir = value
				cmd.WorkingDirectory = value
				i += terraformConsumedWords(word)
			}
			i++
		case word == "-help" || word == "--help":
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: word, Position: i})
			i++
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.GlobalOptions = append(cmd.GlobalOptions, parseOptionWord(word, i))
			i++
		default:
			semantic.Subcommand = word
			cmd.ActionPath = []string{word}
			i++
			goto subcommand
		}
		if semantic.Subcommand == "version" {
			break
		}
	}

subcommand:
	for ; i < len(base.RawWords); i++ {
		word := base.RawWords[i]
		if terraformOptionWithValue(word, "-target") {
			value, consumed := terraformOptionValue(word, "-target", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "-target", Value: value, HasValue: consumed, Position: i})
			semantic.Target = true
			if consumed {
				semantic.Targets = append(semantic.Targets, value)
				i += terraformConsumedWords(word)
			}
			continue
		}
		if terraformOptionWithValue(word, "-replace") {
			value, consumed := terraformOptionValue(word, "-replace", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "-replace", Value: value, HasValue: consumed, Position: i})
			semantic.Replace = true
			if consumed {
				semantic.Replaces = append(semantic.Replaces, value)
				i += terraformConsumedWords(word)
			}
			continue
		}
		if terraformOptionWithValue(word, "-out") {
			value, consumed := terraformOptionValue(word, "-out", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "-out", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.Out = value
				i += terraformConsumedWords(word)
			}
			continue
		}
		if terraformOptionWithValue(word, "-var-file") {
			value, consumed := terraformOptionValue(word, "-var-file", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "-var-file", Value: value, HasValue: consumed, Position: i})
			if consumed {
				semantic.VarFiles = append(semantic.VarFiles, value)
				i += terraformConsumedWords(word)
			}
			continue
		}
		if terraformOptionWithValue(word, "-var") {
			value, consumed := terraformOptionValue(word, "-var", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "-var", Value: value, HasValue: consumed, Position: i})
			semantic.Vars = true
			if consumed {
				i += terraformConsumedWords(word)
			}
			continue
		}
		if parsed, ok := terraformBoolOption(word, "-input", base.RawWords, i); ok {
			cmd.Options = append(cmd.Options, parsed.option)
			if parsed.option.HasValue {
				semantic.Input = &parsed.value
				i += terraformConsumedWords(word)
			}
			continue
		}
		if parsed, ok := terraformBoolOption(word, "-lock", base.RawWords, i); ok {
			cmd.Options = append(cmd.Options, parsed.option)
			if parsed.option.HasValue {
				semantic.Lock = &parsed.value
				i += terraformConsumedWords(word)
			}
			continue
		}
		if parsed, ok := terraformBoolOption(word, "-refresh", base.RawWords, i); ok {
			cmd.Options = append(cmd.Options, parsed.option)
			if parsed.option.HasValue {
				semantic.Refresh = &parsed.value
				i += terraformConsumedWords(word)
			}
			continue
		}
		if parsed, ok := terraformBoolOption(word, "-backend", base.RawWords, i); ok {
			cmd.Options = append(cmd.Options, parsed.option)
			if parsed.option.HasValue {
				semantic.Backend = &parsed.value
				i += terraformConsumedWords(word)
			}
			continue
		}
		switch {
		case word == "-destroy":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Destroy = true
		case word == "-auto-approve":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.AutoApprove = true
		case word == "-refresh-only":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.RefreshOnly = true
		case word == "-upgrade":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Upgrade = true
		case word == "-reconfigure":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Reconfigure = true
		case word == "-migrate-state":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.MigrateState = true
		case word == "-recursive":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Recursive = true
		case word == "-check":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Check = true
		case word == "-json":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.JSON = true
		case word == "-force":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
			semantic.Force = true
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.Options = append(cmd.Options, parseOptionWord(word, i))
		default:
			cmd.Args = append(cmd.Args, word)
			if len(cmd.ActionPath) < 2 && terraformGroupSubcommand(semantic, word) {
				cmd.ActionPath = append(cmd.ActionPath, word)
			}
		}
	}

	if semantic.Subcommand == "destroy" {
		semantic.Destroy = true
	}
	if semantic.Subcommand == "apply" && len(cmd.Args) > 0 {
		semantic.PlanFile = cmd.Args[0]
	}
	semantic.Flags = normalizedTerraformFlags(cmd.GlobalOptions, cmd.Options)
	cmd.Terraform = semantic
	return cmd, true
}

type terraformParsedBool struct {
	option Option
	value  bool
}

func terraformBoolOption(word, name string, words []string, i int) (terraformParsedBool, bool) {
	if word != name && !strings.HasPrefix(word, name+"=") {
		return terraformParsedBool{}, false
	}
	value, consumed := terraformOptionValue(word, name, words, i)
	option := Option{Name: name, Value: value, HasValue: consumed, Position: i}
	if !consumed {
		return terraformParsedBool{option: option}, true
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return terraformParsedBool{option: option}, true
	}
	return terraformParsedBool{option: option, value: parsed}, true
}

func terraformGroupSubcommand(semantic *TerraformSemantic, word string) bool {
	switch semantic.Subcommand {
	case "workspace":
		semantic.WorkspaceSubcommand = word
	case "state":
		semantic.StateSubcommand = word
	case "providers":
		semantic.ProvidersSubcommand = word
	case "metadata":
		semantic.MetadataSubcommand = word
	default:
		return false
	}
	return true
}

func terraformOptionWithValue(word string, name string) bool {
	return word == name || strings.HasPrefix(word, name+"=")
}

func terraformOptionValue(word string, name string, words []string, i int) (string, bool) {
	if value, ok := strings.CutPrefix(word, name+"="); ok {
		return value, true
	}
	if i+1 >= len(words) {
		return "", false
	}
	return words[i+1], true
}

func terraformConsumedWords(word string) int {
	if strings.Contains(word, "=") {
		return 0
	}
	return 1
}

func normalizedTerraformFlags(globalOptions, options []Option) []string {
	flags := make([]string, 0, (len(globalOptions)+len(options))*2)
	for _, option := range append(append([]Option(nil), globalOptions...), options...) {
		flags = append(flags, option.Name)
		if option.HasValue {
			flags = append(flags, option.Name+"="+option.Value)
		}
	}
	return flags
}
