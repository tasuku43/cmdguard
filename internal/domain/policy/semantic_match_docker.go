package policy

import commandpkg "github.com/tasuku43/cc-bash-guard/internal/domain/command"

func init() {
	registerSemanticHandler(semanticHandler{
		command:  "docker",
		match:    func(s SemanticMatchSpec, cmd commandpkg.Command) bool { return s.Docker().matches(cmd) },
		validate: ValidateDockerSemanticMatchSpec,
	})
}

func (s DockerSemanticSpec) matches(cmd commandpkg.Command) bool {
	if cmd.SemanticParser != "docker" || cmd.Docker == nil {
		return false
	}
	d := cmd.Docker
	if s.Verb != "" && d.Verb != s.Verb {
		return false
	}
	if len(s.VerbIn) > 0 && !containsString(s.VerbIn, d.Verb) {
		return false
	}
	if s.Subverb != "" && d.Subverb != s.Subverb {
		return false
	}
	if len(s.SubverbIn) > 0 && !containsString(s.SubverbIn, d.Subverb) {
		return false
	}
	if s.ComposeCommand != "" && d.ComposeCommand != s.ComposeCommand {
		return false
	}
	if len(s.ComposeCommandIn) > 0 && !containsString(s.ComposeCommandIn, d.ComposeCommand) {
		return false
	}
	if s.Image != "" && d.Image != s.Image {
		return false
	}
	if len(s.ImageIn) > 0 && !containsString(s.ImageIn, d.Image) {
		return false
	}
	if s.Container != "" && d.Container != s.Container {
		return false
	}
	if s.Service != "" && d.Service != s.Service {
		return false
	}
	if s.Context != "" && d.Context != s.Context {
		return false
	}
	if len(s.ContextIn) > 0 && !containsString(s.ContextIn, d.Context) {
		return false
	}
	if s.Host != "" && d.Host != s.Host {
		return false
	}
	if s.HostPrefix != "" && !hasStringPrefix(d.Host, s.HostPrefix) {
		return false
	}
	if s.File != "" && !containsString(d.Files, s.File) && d.File != s.File {
		return false
	}
	if len(s.FileIn) > 0 && !containsAnyString(append(d.Files, d.File), s.FileIn) {
		return false
	}
	if s.FilePrefix != "" && !containsPrefix(append(d.Files, d.File), s.FilePrefix) {
		return false
	}
	if s.ProjectName != "" && d.ProjectName != s.ProjectName {
		return false
	}
	if len(s.ProjectNameIn) > 0 && !containsString(s.ProjectNameIn, d.ProjectName) {
		return false
	}
	if s.Profile != "" && !containsString(d.Profiles, s.Profile) && d.Profile != s.Profile {
		return false
	}
	if len(s.ProfileIn) > 0 && !containsAnyString(append(d.Profiles, d.Profile), s.ProfileIn) {
		return false
	}
	if !matchBool(s.DryRun, d.DryRun) || !matchBool(s.Detach, d.Detach) || !matchBool(s.Interactive, d.Interactive) ||
		!matchBool(s.Tty, d.Tty) || !matchBool(s.RM, d.RM) || !matchBool(s.Force, d.Force) ||
		!matchBool(s.Privileged, d.Privileged) || !matchBool(s.NetworkHost, d.NetworkHost) ||
		!matchBool(s.PIDHost, d.PIDHost) || !matchBool(s.IPCHost, d.IPCHost) || !matchBool(s.UTSHost, d.UTSHost) ||
		!matchBool(s.Device, d.Device) || !matchBool(s.HostMount, d.HostMount) || !matchBool(s.RootMount, d.RootMount) ||
		!matchBool(s.DockerSocketMount, d.DockerSocketMount) || !matchBool(s.PublishAll, d.PublishAll) ||
		!matchBool(s.NoCache, d.NoCache) || !matchBool(s.All, d.All) || !matchBool(s.VolumesFlag, d.VolumesFlag) ||
		!matchBool(s.Prune, d.Prune) || !matchBool(s.AllResources, d.AllResources) || !matchBool(s.RemoveOrphans, d.RemoveOrphans) {
		return false
	}
	if s.User != "" && d.User != s.User || s.Workdir != "" && d.Workdir != s.Workdir ||
		s.Entrypoint != "" && d.Entrypoint != s.Entrypoint || s.Network != "" && d.Network != s.Network ||
		s.PID != "" && d.PID != s.PID || s.IPC != "" && d.IPC != s.IPC || s.UTS != "" && d.UTS != s.UTS ||
		s.Pull != "" && d.Pull != s.Pull || s.Platform != "" && d.Platform != s.Platform {
		return false
	}
	for _, check := range []struct {
		want []string
		got  []string
	}{
		{s.CapAddContains, d.CapAdd}, {s.CapDropContains, d.CapDrop}, {s.SecurityOptContains, d.SecurityOpt},
		{s.DevicesContains, d.Devices}, {s.MountsContains, d.Mounts}, {s.VolumesContains, d.Volumes},
		{s.EnvFilesContains, d.EnvFiles}, {s.EnvKeysContains, d.EnvKeys}, {s.PortsContains, d.Ports},
		{s.BuildArgKeysContains, d.BuildArgKeys}, {s.FlagsContains, d.Flags},
	} {
		for _, value := range check.want {
			if !containsString(check.got, value) {
				return false
			}
		}
	}
	for _, prefix := range s.FlagsPrefixes {
		if !containsPrefix(d.Flags, prefix) {
			return false
		}
	}
	return true
}

func matchBool(want *bool, got bool) bool {
	return want == nil || *want == got
}

func hasStringPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}
