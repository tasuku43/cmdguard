package command

import "strings"

type AwsParser struct{}

func init() {
	RegisterDefaultParser(AwsParser{})
}

func (AwsParser) Program() string {
	return "aws"
}

func (AwsParser) Parse(base Command) (Command, bool) {
	if base.Program != "aws" {
		return Command{}, false
	}

	cmd := base
	cmd.Parser = AwsParser{}.Program()
	cmd.Args = []string{}

	i := 0
	for i < len(base.RawWords) {
		word := base.RawWords[i]
		switch {
		case isAWSGlobalOptionWithValue(word, "--profile"):
			value, consumed := awsOptionValue(word, "--profile", base.RawWords, i)
			if !consumed {
				cmd.ActionPath = append(cmd.ActionPath, base.RawWords[i:]...)
				return cmd, true
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "--profile", Value: value, HasValue: true, Position: i})
			i += awsConsumedWords(word)
		case isAWSGlobalOptionWithValue(word, "--region"):
			value, consumed := awsOptionValue(word, "--region", base.RawWords, i)
			if !consumed {
				cmd.ActionPath = append(cmd.ActionPath, base.RawWords[i:]...)
				return cmd, true
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "--region", Value: value, HasValue: true, Position: i})
			i += awsConsumedWords(word)
		case isAWSGlobalOptionWithValue(word, "--endpoint-url"):
			value, consumed := awsOptionValue(word, "--endpoint-url", base.RawWords, i)
			if !consumed {
				cmd.ActionPath = append(cmd.ActionPath, base.RawWords[i:]...)
				return cmd, true
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "--endpoint-url", Value: value, HasValue: true, Position: i})
			i += awsConsumedWords(word)
		case word == "--no-cli-pager":
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: word, Position: i})
			i++
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.ActionPath = append(cmd.ActionPath, base.RawWords[i:]...)
			return cmd, true
		default:
			cmd.ActionPath, cmd.Options, cmd.Args = splitAWSAction(base.RawWords[i:], i)
			cmd.AWS = buildAWSSemantic(cmd.Env, cmd.GlobalOptions, cmd.ActionPath, cmd.Options)
			if cmd.AWS != nil {
				cmd.SemanticParser = AwsParser{}.Program()
			}
			return cmd, true
		}
	}

	return cmd, true
}

func isAWSGlobalOptionWithValue(word string, name string) bool {
	return word == name || strings.HasPrefix(word, name+"=")
}

func awsOptionValue(word string, name string, words []string, i int) (string, bool) {
	if value, ok := strings.CutPrefix(word, name+"="); ok {
		return value, true
	}
	if i+1 >= len(words) {
		return "", false
	}
	return words[i+1], true
}

func awsConsumedWords(word string) int {
	if strings.Contains(word, "=") {
		return 1
	}
	return 2
}

func splitAWSAction(words []string, startPosition int) ([]string, []Option, []string) {
	var actionPath []string
	var options []Option
	var args []string
	for i, word := range words {
		if i < 2 {
			actionPath = append(actionPath, word)
			continue
		}
		if strings.HasPrefix(word, "-") && word != "-" {
			options = append(options, parseOptionWord(word, startPosition+i))
			continue
		}
		args = append(args, word)
	}
	return actionPath, options, args
}

func buildAWSSemantic(env map[string]string, globalOptions []Option, actionPath []string, options []Option) *AWSSemantic {
	if len(actionPath) < 2 {
		return nil
	}

	profile, profileConflict := awsProfile(env, globalOptions)
	region, regionSource := awsRegion(env, globalOptions)
	dryRun := awsBoolOption(options, "--dry-run", "--no-dry-run")
	noCLIPager := awsNoCLIPager(globalOptions, options)

	return &AWSSemantic{
		Service:         actionPath[0],
		Operation:       actionPath[1],
		Profile:         profile,
		Region:          region,
		EndpointURL:     lastOptionValue(globalOptions, "--endpoint-url"),
		DryRun:          dryRun,
		NoCLIPager:      noCLIPager,
		Flags:           normalizedAWSFlags(append(append([]Option{}, globalOptions...), options...)),
		ProfileConflict: profileConflict,
		RegionSource:    regionSource,
	}
}

func awsProfile(env map[string]string, globalOptions []Option) (string, bool) {
	envProfile := env["AWS_PROFILE"]
	cliProfile := lastOptionValue(globalOptions, "--profile")
	if cliProfile == "" {
		return envProfile, false
	}
	return cliProfile, envProfile != "" && envProfile != cliProfile
}

func awsRegion(env map[string]string, globalOptions []Option) (string, string) {
	if cliRegion := lastOptionValue(globalOptions, "--region"); cliRegion != "" {
		return cliRegion, "cli"
	}
	if envRegion := env["AWS_REGION"]; envRegion != "" {
		return envRegion, "AWS_REGION"
	}
	if envDefaultRegion := env["AWS_DEFAULT_REGION"]; envDefaultRegion != "" {
		return envDefaultRegion, "AWS_DEFAULT_REGION"
	}
	return "", ""
}

func awsBoolOption(options []Option, trueName string, falseName string) *bool {
	var value *bool
	for _, option := range options {
		switch option.Name {
		case trueName:
			v := true
			value = &v
		case falseName:
			v := false
			value = &v
		}
	}
	return value
}

func awsNoCLIPager(globalOptions []Option, options []Option) *bool {
	for _, option := range append(append([]Option{}, globalOptions...), options...) {
		if option.Name == "--no-cli-pager" {
			v := true
			return &v
		}
	}
	return nil
}

func lastOptionValue(options []Option, name string) string {
	value := ""
	for _, option := range options {
		if option.Name == name && option.HasValue {
			value = option.Value
		}
	}
	return value
}

func normalizedAWSFlags(options []Option) []string {
	flags := make([]string, 0, len(options))
	for _, option := range options {
		flags = append(flags, option.Name)
	}
	return flags
}
