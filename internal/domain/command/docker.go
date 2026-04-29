package command

import (
	"path/filepath"
	"strings"
)

type DockerParser struct{}

func init() {
	RegisterDefaultParser(DockerParser{})
}

func (DockerParser) Program() string {
	return "docker"
}

func (DockerParser) Parse(base Command) (Command, bool) {
	if base.Program != "docker" {
		return Command{}, false
	}
	cmd := base
	cmd.Parser = DockerParser{}.Program()
	cmd.SemanticParser = DockerParser{}.Program()
	cmd.Args = []string{}

	i := parseDockerGlobalOptions(&cmd, base.RawWords, 0)
	if i >= len(base.RawWords) {
		return cmd, true
	}
	cmd.ActionPath, cmd.Options, cmd.Args = splitDockerAction(base.RawWords[i:], i)
	cmd.Docker = buildDockerSemantic(cmd.ActionPath, cmd.GlobalOptions, cmd.Options, cmd.Args)
	return cmd, true
}

func parseDockerGlobalOptions(cmd *Command, words []string, start int) int {
	i := start
	for i < len(words) {
		word := words[i]
		switch {
		case dockerOptionHasValue(word, "--config"):
			i = appendDockerGlobalValue(cmd, words, i, "--config")
		case dockerOptionHasValue(word, "--context"):
			value, consumed, ok := dockerOptionValue(words, i, "--context")
			if !ok {
				return i
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "--context", Value: value, HasValue: true, Position: i})
			i += consumed
		case word == "-c":
			if i+1 >= len(words) {
				return i
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "-c", Value: words[i+1], HasValue: true, Position: i})
			i += 2
		case dockerOptionHasValue(word, "-H"):
			value, consumed, ok := dockerOptionValue(words, i, "-H")
			if !ok {
				return i
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "-H", Value: value, HasValue: true, Position: i})
			i += consumed
		case dockerOptionHasValue(word, "--host"):
			value, consumed, ok := dockerOptionValue(words, i, "--host")
			if !ok {
				return i
			}
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: "--host", Value: value, HasValue: true, Position: i})
			i += consumed
		case dockerOptionHasValue(word, "--log-level"), dockerOptionHasValue(word, "--tlscacert"), dockerOptionHasValue(word, "--tlscert"), dockerOptionHasValue(word, "--tlskey"):
			name := dockerOptionName(word)
			i = appendDockerGlobalValue(cmd, words, i, name)
		case word == "--debug" || word == "-D" || word == "--tls" || word == "--tlsverify":
			cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: word, Position: i})
			i++
		default:
			return i
		}
	}
	return i
}

func appendDockerGlobalValue(cmd *Command, words []string, i int, name string) int {
	value, consumed, ok := dockerOptionValue(words, i, name)
	if !ok {
		return i
	}
	cmd.GlobalOptions = append(cmd.GlobalOptions, Option{Name: name, Value: value, HasValue: true, Position: i})
	return i + consumed
}

func splitDockerAction(words []string, startPosition int) ([]string, []Option, []string) {
	if len(words) == 0 {
		return nil, nil, nil
	}
	actionPath := []string{words[0]}
	options := []Option{}
	args := []string{}
	for i := 1; i < len(words); i++ {
		word := words[i]
		if word == "--" {
			args = append(args, words[i:]...)
			break
		}
		if strings.HasPrefix(word, "-") && word != "-" {
			option := parseDockerOption(words[0], words, i, startPosition+i)
			options = append(options, option)
			if option.HasValue && !strings.Contains(word, "=") && dockerOptionConsumesNext(option.Name) {
				i++
			}
			continue
		}
		args = append(args, word)
	}
	return actionPath, options, args
}

func parseDockerOption(verb string, words []string, i int, position int) Option {
	word := words[i]
	if name, value, ok := strings.Cut(word, "="); ok {
		return Option{Name: name, Value: value, HasValue: true, Position: position}
	}
	if dockerOptionConsumesNextForVerb(verb, word) && i+1 < len(words) {
		return Option{Name: word, Value: words[i+1], HasValue: true, Position: position}
	}
	return parseOptionWord(word, position)
}

func buildDockerSemantic(actionPath []string, globals []Option, options []Option, args []string) *DockerSemantic {
	if len(actionPath) == 0 {
		return nil
	}
	d := &DockerSemantic{Verb: actionPath[0], Flags: normalizedDockerFlags(options)}
	d.Context = dockerLastValue(globals, "--context", "-c")
	d.Host = dockerLastValue(globals, "--host", "-H")
	for _, option := range globals {
		d.Flags = append(d.Flags, option.Name)
	}
	if d.Verb == "compose" {
		parseDockerCompose(d, options, args)
	} else {
		parseDockerCommand(d, options, args)
	}
	d.applyRiskFields()
	return d
}

func parseDockerCommand(d *DockerSemantic, options []Option, args []string) {
	pos := dockerPositional(args)
	if isDockerCommandGroup(d.Verb) && len(pos) > 0 {
		d.Subverb = pos[0]
		pos = pos[1:]
	}
	if d.Verb == "build" {
		parseDockerBuildOptions(d, options)
	} else {
		parseDockerCommonOptions(d, options)
	}
	switch d.Verb {
	case "run":
		if len(pos) > 0 {
			d.Image = pos[0]
		}
	case "exec":
		if len(pos) > 0 {
			d.Container = pos[0]
		}
	case "build":
		d.Image = dockerLastValue(options, "-t", "--tag")
	case "pull", "push":
		if len(pos) > 0 {
			d.Image = pos[0]
		}
	case "logs", "inspect", "rm", "stop", "start", "restart", "kill":
		if len(pos) > 0 {
			d.Container = pos[0]
		}
	}
	if d.Subverb != "" {
		switch d.Verb {
		case "container":
			if len(pos) > 0 {
				d.Container = pos[0]
			}
		case "image":
			if len(pos) > 0 && d.Subverb != "prune" {
				d.Image = pos[0]
			}
		}
	}
	d.Prune = d.Subverb == "prune" || d.Verb == "prune"
	d.AllResources = d.All
}

func parseDockerBuildOptions(d *DockerSemantic, options []Option) {
	parseDockerCommonOptions(d, options)
	d.Tty = false
}

func parseDockerCompose(d *DockerSemantic, options []Option, args []string) {
	composeOptions, commandOptions, pos := splitComposeParts(options, args)
	d.Subverb = ""
	for _, option := range composeOptions {
		switch option.Name {
		case "-f", "--file":
			if option.HasValue {
				d.Files = append(d.Files, option.Value)
				d.File = option.Value
			}
		case "-p", "--project-name":
			d.ProjectName = option.Value
		case "--profile":
			if option.HasValue {
				d.Profiles = append(d.Profiles, option.Value)
				d.Profile = option.Value
			}
		case "--env-file":
			if option.HasValue {
				d.EnvFiles = append(d.EnvFiles, option.Value)
			}
		case "--dry-run":
			d.DryRun = true
		case "--all-resources":
			d.AllResources = true
		}
	}
	d.Flags = append(d.Flags, normalizedDockerFlags(composeOptions)...)
	if len(pos) == 0 {
		parseDockerCommonOptions(d, commandOptions)
		return
	}
	d.ComposeCommand = pos[0]
	d.Subverb = d.ComposeCommand
	pos = pos[1:]
	parseDockerCommonOptions(d, commandOptions)
	if (d.ComposeCommand == "run" || d.ComposeCommand == "exec") && len(pos) > 0 {
		d.Service = pos[0]
	}
	if d.ComposeCommand == "down" || d.ComposeCommand == "rm" {
		d.Prune = false
	}
}

func splitComposeParts(options []Option, args []string) ([]Option, []Option, []string) {
	composeOptions := []Option{}
	commandOptions := []Option{}
	pos := []string{}
	commandSeen := false
	for _, option := range options {
		if commandSeen {
			commandOptions = append(commandOptions, option)
			continue
		}
		if isComposeGlobalOption(option.Name) {
			composeOptions = append(composeOptions, option)
		} else {
			commandOptions = append(commandOptions, option)
		}
	}
	for _, arg := range dockerPositional(args) {
		if !commandSeen {
			commandSeen = true
		}
		pos = append(pos, arg)
	}
	return composeOptions, commandOptions, pos
}

func parseDockerCommonOptions(d *DockerSemantic, options []Option) {
	for _, option := range options {
		switch option.Name {
		case "-d", "--detach":
			d.Detach = true
		case "-i", "--interactive":
			d.Interactive = true
		case "-t", "--tty":
			d.Tty = true
		case "-it", "-ti":
			d.Interactive = true
			d.Tty = true
		case "--rm":
			d.RM = true
		case "-f", "--force":
			d.Force = true
		case "--privileged":
			d.Privileged = true
		case "-u", "--user":
			d.User = option.Value
		case "-w", "--workdir":
			d.Workdir = option.Value
		case "--entrypoint":
			d.Entrypoint = option.Value
		case "--network", "--net":
			d.Network = option.Value
		case "--pid":
			d.PID = option.Value
		case "--ipc":
			d.IPC = option.Value
		case "--uts":
			d.UTS = option.Value
		case "--cap-add":
			d.CapAdd = appendValue(d.CapAdd, option.Value)
		case "--cap-drop":
			d.CapDrop = appendValue(d.CapDrop, option.Value)
		case "--security-opt":
			d.SecurityOpt = appendValue(d.SecurityOpt, option.Value)
		case "--device":
			d.Device = true
			d.Devices = appendValue(d.Devices, option.Value)
		case "--mount":
			d.Mounts = appendValue(d.Mounts, option.Value)
		case "-v", "--volume":
			d.Volumes = appendValue(d.Volumes, option.Value)
		case "--env-file":
			d.EnvFiles = appendValue(d.EnvFiles, option.Value)
		case "-e", "--env":
			if key := dockerEnvKey(option.Value); key != "" {
				d.EnvKeys = append(d.EnvKeys, key)
			}
		case "-p", "--publish":
			d.Ports = appendValue(d.Ports, option.Value)
		case "-P", "--publish-all":
			d.PublishAll = true
		case "--pull":
			d.Pull = option.Value
		case "--no-cache":
			d.NoCache = true
		case "--build-arg":
			if key := dockerEnvKey(option.Value); key != "" {
				d.BuildArgKeys = append(d.BuildArgKeys, key)
			}
		case "--target":
			d.Target = option.Value
		case "--platform":
			d.Platform = option.Value
		case "-a", "--all":
			d.All = true
		case "--volumes":
			d.VolumesFlag = true
		case "--remove-orphans":
			d.RemoveOrphans = true
		}
	}
}

func (d *DockerSemantic) applyRiskFields() {
	d.NetworkHost = d.Network == "host"
	d.PIDHost = d.PID == "host"
	d.IPCHost = d.IPC == "host"
	d.UTSHost = d.UTS == "host"
	for _, volume := range d.Volumes {
		source := dockerVolumeSource(volume)
		d.markMountSource(source, volume)
	}
	for _, mount := range d.Mounts {
		source, bind := dockerMountSource(mount)
		if bind {
			d.markMountSource(source, mount)
		}
		if dockerReferencesSocket(mount) {
			d.DockerSocketMount = true
		}
	}
	if d.Verb == "system" && d.Subverb == "prune" || d.Subverb == "prune" {
		d.Prune = true
	}
	if d.Prune && d.All {
		d.AllResources = true
	}
}

func (d *DockerSemantic) markMountSource(source, raw string) {
	if source == "" {
		if dockerReferencesSocket(raw) {
			d.DockerSocketMount = true
		}
		return
	}
	if filepath.IsAbs(source) {
		d.HostMount = true
	}
	if source == "/" {
		d.RootMount = true
	}
	if dockerReferencesSocket(source) || dockerReferencesSocket(raw) {
		d.DockerSocketMount = true
	}
}

func dockerOptionHasValue(word, name string) bool {
	return word == name || strings.HasPrefix(word, name+"=")
}

func dockerOptionValue(words []string, i int, name string) (string, int, bool) {
	if value, ok := strings.CutPrefix(words[i], name+"="); ok {
		return value, 1, true
	}
	if words[i] != name || i+1 >= len(words) {
		return "", 1, false
	}
	return words[i+1], 2, true
}

func dockerOptionName(word string) string {
	name, _, _ := strings.Cut(word, "=")
	return name
}

func dockerOptionConsumesNext(name string) bool {
	switch name {
	case "--config", "--context", "-c", "-H", "--host", "--log-level", "--tlscacert", "--tlscert", "--tlskey",
		"-f", "--file", "-p", "--project-name", "--profile", "--project-directory", "--env-file",
		"-u", "--user", "-w", "--workdir", "--entrypoint", "--network", "--net", "--pid", "--ipc", "--uts",
		"--cap-add", "--cap-drop", "--security-opt", "--device", "--mount", "-v", "--volume", "-e", "--env",
		"--publish", "--pull", "--build-arg", "--target", "--platform", "-t", "--tag":
		return true
	default:
		return false
	}
}

func dockerOptionConsumesNextForVerb(verb, name string) bool {
	if verb == "compose" {
		switch name {
		case "-f", "--file", "-p", "--project-name", "--profile", "--project-directory", "--env-file":
			return true
		}
	}
	if verb == "build" || verb == "buildx" {
		switch name {
		case "-f", "--file", "-t", "--tag":
			return true
		}
	}
	if name == "-f" || name == "-t" {
		return false
	}
	return dockerOptionConsumesNext(name)
}

func normalizedDockerFlags(options []Option) []string {
	flags := make([]string, 0, len(options))
	for _, option := range options {
		flags = append(flags, option.Name)
	}
	return flags
}

func dockerLastValue(options []Option, names ...string) string {
	value := ""
	for _, option := range options {
		for _, name := range names {
			if option.Name == name && option.HasValue {
				value = option.Value
			}
		}
	}
	return value
}

func dockerPositional(args []string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--" {
			break
		}
		out = append(out, arg)
	}
	return out
}

func isDockerCommandGroup(verb string) bool {
	switch verb {
	case "container", "image", "network", "volume", "system", "builder", "buildx", "plugin", "context", "swarm", "service", "stack", "secret", "config", "manifest", "trust", "mcp", "model":
		return true
	default:
		return false
	}
}

func isComposeGlobalOption(name string) bool {
	switch name {
	case "-f", "--file", "-p", "--project-name", "--profile", "--project-directory", "--env-file", "--dry-run", "--all-resources":
		return true
	default:
		return false
	}
}

func appendValue(values []string, value string) []string {
	if value == "" {
		return values
	}
	return append(values, value)
}

func dockerEnvKey(value string) string {
	if value == "" {
		return ""
	}
	key, _, _ := strings.Cut(value, "=")
	return key
}

func dockerVolumeSource(value string) string {
	if value == "" {
		return ""
	}
	source, _, ok := strings.Cut(value, ":")
	if !ok {
		return ""
	}
	return source
}

func dockerMountSource(value string) (string, bool) {
	if value == "" {
		return "", false
	}
	parts := strings.Split(value, ",")
	isBind := false
	source := ""
	for _, part := range parts {
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch key {
		case "type":
			isBind = val == "bind"
		case "source", "src":
			source = val
		}
	}
	return source, isBind
}

func dockerReferencesSocket(value string) bool {
	return strings.Contains(value, "/var/run/docker.sock") || strings.Contains(value, "/run/docker.sock")
}
