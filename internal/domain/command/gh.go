package command

import "strings"

type GhParser struct{}

func init() {
	RegisterDefaultParser(GhParser{})
}

func (GhParser) Program() string {
	return "gh"
}

func (GhParser) Parse(base Command) (Command, bool) {
	if base.Program != "gh" {
		return Command{}, false
	}

	cmd := base
	cmd.Parser = GhParser{}.Program()
	cmd.SemanticParser = GhParser{}.Program()
	cmd.Args = []string{}

	var positionals []string
	for i := 0; i < len(base.RawWords); i++ {
		word := base.RawWords[i]
		switch {
		case ghOptionWithValue(word, "-R", "--repo"):
			value, consumed := ghOptionValue(word, "--repo", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-R", "--repo"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--hostname"):
			value, consumed := ghOptionValue(word, "--hostname", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--hostname", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-o", "--org"):
			value, consumed := ghOptionValue(word, "--org", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-o", "--org"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-e", "--env"):
			value, consumed := ghOptionValue(word, "--env", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-e", "--env"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--env-name"):
			value, consumed := ghOptionValue(word, "--env-name", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--env-name", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--ref"):
			value, consumed := ghOptionValue(word, "--ref", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--ref", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-X", "--method"):
			value, consumed := ghOptionValue(word, "--method", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-X", "--method"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-F", "--field"):
			value, consumed := ghOptionValue(word, "--field", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-F", "--field"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-f", "--raw-field"):
			value, consumed := ghOptionValue(word, "--raw-field", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-f", "--raw-field"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-H", "--header"):
			value, consumed := ghOptionValue(word, "--header", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-H", "--header"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--input"):
			value, consumed := ghOptionValue(word, "--input", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--input", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--base"):
			value, consumed := ghOptionValue(word, "--base", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--base", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--head"):
			value, consumed := ghOptionValue(word, "--head", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--head", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--job"):
			value, consumed := ghOptionValue(word, "--job", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--job", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "", "--state"):
			value, consumed := ghOptionValue(word, "--state", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: "--state", Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-l", "--label"):
			value, consumed := ghOptionValue(word, "--label", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-l", "--label"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-a", "--assignee"):
			value, consumed := ghOptionValue(word, "--assignee", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-a", "--assignee"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-t", "--title"):
			value, consumed := ghOptionValue(word, "--title", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-t", "--title"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case ghOptionWithValue(word, "-b", "--body"):
			value, consumed := ghOptionValue(word, "--body", base.RawWords, i)
			cmd.Options = append(cmd.Options, Option{Name: ghOptionName(word, "-b", "--body"), Value: value, HasValue: consumed, Position: i})
			if consumed && !strings.Contains(word, "=") {
				i++
			}
		case word == "-w" || word == "--web" ||
			word == "--paginate" || word == "--silent" || word == "-i" || word == "--include" ||
			word == "--draft" || word == "--fill" || word == "--force" ||
			word == "--prerelease" || word == "--latest" ||
			word == "--merge" || word == "-m" || word == "--squash" || word == "-s" || word == "--rebase" || word == "-r" ||
			word == "--delete-branch" || word == "--admin" || word == "--auto" ||
			word == "--failed" || word == "--debug" || word == "--exit-status":
			cmd.Options = append(cmd.Options, Option{Name: word, Position: i})
		case strings.HasPrefix(word, "-") && word != "-":
			cmd.Options = append(cmd.Options, parseOptionWord(word, i))
		default:
			positionals = append(positionals, word)
		}
	}

	cmd.ActionPath = ghActionPath(positionals)
	cmd.Args = ghArgs(positionals)
	cmd.Gh = buildGhSemantic(cmd.Env, cmd.ActionPath, cmd.Args, cmd.Options)
	return cmd, true
}

func ghOptionWithValue(word string, short string, long string) bool {
	if short != "" && word == short {
		return true
	}
	return word == long || strings.HasPrefix(word, long+"=")
}

func ghOptionName(word string, short string, long string) string {
	if short != "" && word == short {
		return short
	}
	return long
}

func ghOptionValue(word string, long string, words []string, i int) (string, bool) {
	if value, ok := strings.CutPrefix(word, long+"="); ok {
		return value, true
	}
	if i+1 >= len(words) {
		return "", false
	}
	return words[i+1], true
}

func ghActionPath(positionals []string) []string {
	if len(positionals) == 0 {
		return []string{}
	}
	if len(positionals) == 1 || positionals[0] == "api" {
		return append([]string(nil), positionals[:1]...)
	}
	return append([]string(nil), positionals[:2]...)
}

func ghArgs(positionals []string) []string {
	if len(positionals) == 0 {
		return []string{}
	}
	if positionals[0] == "api" {
		return append([]string(nil), positionals[1:]...)
	}
	if len(positionals) <= 2 {
		return []string{}
	}
	return append([]string(nil), positionals[2:]...)
}

func buildGhSemantic(env map[string]string, actionPath []string, args []string, options []Option) *GhSemantic {
	semantic := &GhSemantic{
		Hostname: env["GH_HOST"],
		Flags:    normalizedGhFlags(options),
	}
	if len(actionPath) > 0 {
		semantic.Area = actionPath[0]
	}
	if len(actionPath) > 1 {
		semantic.Verb = actionPath[1]
	}
	if repo := lastGhOptionValue(options, "-R", "--repo"); repo != "" {
		semantic.Repo = repo
	}
	if hostname := lastGhOptionValue(options, "--hostname"); hostname != "" {
		semantic.Hostname = hostname
	}
	semantic.Org = lastGhOptionValue(options, "-o", "--org")
	semantic.EnvName = lastGhOptionValue(options, "-e", "--env", "--env-name")
	semantic.Ref = lastGhOptionValue(options, "--ref")
	semantic.Web = ghHasAnyOption(options, "-w", "--web")

	switch semantic.Area {
	case "api":
		fillGhAPISemantic(semantic, args, options)
	case "pr":
		fillGhPRSemantic(semantic, args, options)
	case "issue":
		fillGhIssueSemantic(semantic, args, options)
	case "repo":
		fillGhRepoSemantic(semantic, args)
	case "release":
		fillGhReleaseSemantic(semantic, args, options)
	case "secret":
		fillGhSecretSemantic(semantic, args)
	case "search":
		fillGhSearchSemantic(semantic, args)
	case "workflow":
		fillGhWorkflowSemantic(semantic, args)
	case "auth":
		fillGhAuthSemantic(semantic, args)
	case "run":
		fillGhRunSemantic(semantic, args, options)
	}
	return semantic
}

func fillGhAPISemantic(semantic *GhSemantic, args []string, options []Option) {
	semantic.Method = "GET"
	if method := lastGhOptionValue(options, "-X", "--method"); method != "" {
		semantic.Method = strings.ToUpper(method)
	}
	if len(args) > 0 {
		semantic.Endpoint = normalizeGhEndpoint(args[0])
	}
	semantic.Paginate = ghHasAnyOption(options, "--paginate")
	semantic.Input = ghHasAnyOption(options, "--input")
	semantic.Silent = ghHasAnyOption(options, "--silent")
	semantic.IncludeHeaders = ghHasAnyOption(options, "-i", "--include")
	for _, value := range ghOptionValues(options, "-F", "--field") {
		if key := ghAssignmentKey(value); key != "" {
			semantic.FieldKeys = append(semantic.FieldKeys, key)
		}
	}
	for _, value := range ghOptionValues(options, "-f", "--raw-field") {
		if key := ghAssignmentKey(value); key != "" {
			semantic.RawFieldKeys = append(semantic.RawFieldKeys, key)
		}
	}
	for _, value := range ghOptionValues(options, "-H", "--header") {
		if key := ghHeaderKey(value); key != "" {
			semantic.HeaderKeys = append(semantic.HeaderKeys, key)
		}
	}
}

func fillGhPRSemantic(semantic *GhSemantic, args []string, options []Option) {
	if len(args) > 0 {
		semantic.PRNumber = args[0]
	}
	semantic.Base = lastGhOptionValue(options, "--base")
	semantic.Head = lastGhOptionValue(options, "--head")
	semantic.Draft = ghHasAnyOption(options, "--draft")
	semantic.Fill = ghHasAnyOption(options, "--fill")
	semantic.Admin = ghHasAnyOption(options, "--admin")
	semantic.Auto = ghHasAnyOption(options, "--auto")
	semantic.DeleteBranch = ghHasAnyOption(options, "--delete-branch")
	if semantic.Verb == "checkout" {
		semantic.Force = ghHasAnyOption(options, "--force", "-f")
	}
	switch {
	case ghHasAnyOption(options, "--merge", "-m"):
		semantic.MergeStrategy = "merge"
	case ghHasAnyOption(options, "--squash", "-s"):
		semantic.MergeStrategy = "squash"
	case ghHasAnyOption(options, "--rebase", "-r"):
		semantic.MergeStrategy = "rebase"
	}
}

func fillGhIssueSemantic(semantic *GhSemantic, args []string, options []Option) {
	if len(args) > 0 {
		semantic.IssueNumber = args[0]
	}
	semantic.State = lastGhOptionValue(options, "--state")
	semantic.Labels = append(semantic.Labels, ghOptionValues(options, "-l", "--label")...)
	semantic.Assignees = append(semantic.Assignees, ghOptionValues(options, "-a", "--assignee")...)
	semantic.Title = lastGhOptionValue(options, "-t", "--title")
	semantic.Body = lastGhOptionValue(options, "-b", "--body")
}

func fillGhRepoSemantic(semantic *GhSemantic, args []string) {
	if semantic.Repo == "" && len(args) > 0 {
		semantic.Repo = args[0]
	}
}

func fillGhReleaseSemantic(semantic *GhSemantic, args []string, options []Option) {
	if len(args) > 0 {
		semantic.Tag = args[0]
	}
	semantic.Prerelease = ghHasAnyOption(options, "--prerelease")
	semantic.Draft = ghHasAnyOption(options, "--draft")
	semantic.Latest = ghHasAnyOption(options, "--latest")
}

func fillGhSecretSemantic(semantic *GhSemantic, args []string) {
	if len(args) > 0 {
		semantic.SecretName = args[0]
	}
}

func fillGhSearchSemantic(semantic *GhSemantic, args []string) {
	semantic.SearchType = semantic.Verb
	if len(args) > 0 {
		semantic.Query = strings.Join(args, " ")
	}
}

func fillGhWorkflowSemantic(semantic *GhSemantic, args []string) {
	if len(args) > 0 {
		semantic.WorkflowName = args[0]
		semantic.WorkflowID = args[0]
	}
}

func fillGhAuthSemantic(semantic *GhSemantic, args []string) {
	if semantic.Verb == "" && len(args) > 0 {
		semantic.Verb = args[0]
	}
}

func fillGhRunSemantic(semantic *GhSemantic, args []string, options []Option) {
	if len(args) > 0 {
		semantic.RunID = args[0]
	}
	semantic.Failed = ghHasAnyOption(options, "--failed")
	semantic.Job = lastGhOptionValue(options, "--job")
	semantic.Debug = ghHasAnyOption(options, "--debug")
	semantic.Force = ghHasAnyOption(options, "--force")
	semantic.ExitStatus = ghHasAnyOption(options, "--exit-status")
}

func normalizeGhEndpoint(endpoint string) string {
	if endpoint == "" || strings.HasPrefix(endpoint, "/") {
		return endpoint
	}
	return "/" + endpoint
}

func ghAssignmentKey(value string) string {
	key, _, _ := strings.Cut(value, "=")
	return strings.TrimSpace(key)
}

func ghHeaderKey(value string) string {
	key, _, ok := strings.Cut(value, ":")
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(key))
}

func ghHasAnyOption(options []Option, names ...string) bool {
	for _, option := range options {
		for _, name := range names {
			if option.Name == name {
				return true
			}
		}
	}
	return false
}

func lastGhOptionValue(options []Option, names ...string) string {
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

func ghOptionValues(options []Option, names ...string) []string {
	var values []string
	for _, option := range options {
		if !option.HasValue {
			continue
		}
		for _, name := range names {
			if option.Name == name {
				values = append(values, option.Value)
			}
		}
	}
	return values
}

func normalizedGhFlags(options []Option) []string {
	flags := make([]string, 0, len(options)*2)
	for _, option := range options {
		flags = append(flags, option.Name)
		if option.HasValue {
			flags = append(flags, option.Name+"="+option.Value)
		}
	}
	return flags
}
