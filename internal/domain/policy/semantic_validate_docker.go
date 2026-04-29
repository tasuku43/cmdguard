package policy

import "strings"

func ValidateDockerSemanticMatchSpec(prefix string, semantic SemanticMatchSpec) []string {
	return ValidateDockerSemanticSpec(prefix, semantic.Docker())
}

func ValidateDockerSemanticSpec(prefix string, semantic DockerSemanticSpec) []string {
	var issues []string
	if IsZeroDockerSemanticSpec(semantic) {
		issues = append(issues, prefix+" must not be empty")
	}
	stringFields := map[string]string{
		"verb": semantic.Verb, "subverb": semantic.Subverb, "compose_command": semantic.ComposeCommand,
		"image": semantic.Image, "container": semantic.Container, "service": semantic.Service,
		"context": semantic.Context, "host": semantic.Host, "host_prefix": semantic.HostPrefix,
		"file": semantic.File, "file_prefix": semantic.FilePrefix, "project_name": semantic.ProjectName,
		"profile": semantic.Profile, "user": semantic.User, "workdir": semantic.Workdir,
		"entrypoint": semantic.Entrypoint, "network": semantic.Network, "pid": semantic.PID,
		"ipc": semantic.IPC, "uts": semantic.UTS, "pull": semantic.Pull, "platform": semantic.Platform,
	}
	for name, value := range stringFields {
		if strings.TrimSpace(value) == "" && value != "" {
			issues = append(issues, prefix+"."+name+" must be non-empty")
		}
	}
	issues = append(issues, validateNonEmptyStrings(prefix+".verb_in", semantic.VerbIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".subverb_in", semantic.SubverbIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".compose_command_in", semantic.ComposeCommandIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".image_in", semantic.ImageIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".context_in", semantic.ContextIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".file_in", semantic.FileIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".project_name_in", semantic.ProjectNameIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".profile_in", semantic.ProfileIn)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".cap_add_contains", semantic.CapAddContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".cap_drop_contains", semantic.CapDropContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".security_opt_contains", semantic.SecurityOptContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".devices_contains", semantic.DevicesContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".mounts_contains", semantic.MountsContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".volumes_contains", semantic.VolumesContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".env_files_contains", semantic.EnvFilesContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".env_keys_contains", semantic.EnvKeysContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".ports_contains", semantic.PortsContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".build_arg_keys_contains", semantic.BuildArgKeysContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".flags_contains", semantic.FlagsContains)...)
	issues = append(issues, validateNonEmptyStrings(prefix+".flags_prefixes", semantic.FlagsPrefixes)...)
	return issues
}
