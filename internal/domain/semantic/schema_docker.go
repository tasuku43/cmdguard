package semantic

var dockerSchema = Schema{
	Command:     "docker",
	order:       85,
	Description: "Docker CLI and docker compose commands with high-risk flag detection.",
	Parser:      "docker",
	Fields: []Field{
		stringField("verb", "Top-level docker command, such as run, exec, ps, images, system, or compose."),
		stringListField("verb_in", "Allowed docker verbs."),
		stringField("subverb", "Second action token for command groups such as system prune or compose down."),
		stringListField("subverb_in", "Allowed docker subverbs."),
		stringField("compose_command", "Command after docker compose, such as up, down, run, exec, logs, ps, or config."),
		stringListField("compose_command_in", "Allowed docker compose commands."),
		stringField("image", "Image name for docker run, pull, push, or build -t when straightforward."),
		stringListField("image_in", "Allowed Docker images."),
		stringField("container", "Container name or id for commands such as exec, logs, inspect, rm, and stop when straightforward."),
		stringField("service", "Compose service for docker compose run or exec when straightforward."),
		stringField("context", "Docker context from --context or -c before the verb."),
		stringListField("context_in", "Allowed Docker contexts."),
		stringField("host", "Docker daemon host from -H or --host before the verb."),
		stringField("host_prefix", "Required prefix for Docker daemon host."),
		stringField("file", "Compose file or Dockerfile flag value when straightforward."),
		stringListField("file_in", "Allowed file values."),
		stringField("file_prefix", "Required file path prefix."),
		stringField("project_name", "Compose project name from -p or --project-name."),
		stringListField("project_name_in", "Allowed compose project names."),
		stringField("profile", "Compose profile from --profile."),
		stringListField("profile_in", "Allowed compose profiles."),
		boolField("dry_run", "True for docker compose --dry-run."),
		boolField("detach", "True when -d or --detach is present."),
		boolField("interactive", "True when -i, --interactive, or -it is present."),
		boolField("tty", "True when -t, --tty, or -it is present."),
		boolField("rm", "True when --rm is present."),
		boolField("force", "True when -f or --force is present."),
		boolField("privileged", "True when --privileged is present."),
		stringField("user", "User from -u or --user."),
		stringField("workdir", "Working directory from -w or --workdir."),
		stringField("entrypoint", "Entrypoint from --entrypoint."),
		stringField("network", "Network mode from --network or --net."),
		boolField("network_host", "True when network mode is host."),
		stringField("pid", "PID namespace from --pid."),
		boolField("pid_host", "True when --pid=host is present."),
		stringField("ipc", "IPC namespace from --ipc."),
		boolField("ipc_host", "True when --ipc=host is present."),
		stringField("uts", "UTS namespace from --uts."),
		boolField("uts_host", "True when --uts=host is present."),
		stringListField("cap_add_contains", "Required values from --cap-add."),
		stringListField("cap_drop_contains", "Required values from --cap-drop."),
		stringListField("security_opt_contains", "Required values from --security-opt."),
		boolField("device", "True when --device is present."),
		stringListField("devices_contains", "Required --device values."),
		stringListField("mounts_contains", "Required --mount values."),
		stringListField("volumes_contains", "Required -v or --volume values."),
		boolField("host_mount", "True when a bind mount source is an absolute host path."),
		boolField("root_mount", "True when a bind mount source is /."),
		boolField("docker_socket_mount", "True when a mount references common Docker socket paths."),
		stringListField("env_files_contains", "Required --env-file values."),
		stringListField("env_keys_contains", "Required keys from -e/--env KEY=VALUE."),
		stringListField("ports_contains", "Required -p/--publish values."),
		boolField("publish_all", "True when -P or --publish-all is present."),
		stringField("pull", "Value from --pull."),
		boolField("no_cache", "True when --no-cache is present."),
		stringListField("build_arg_keys_contains", "Required keys from --build-arg KEY=VALUE."),
		stringField("platform", "Platform from --platform."),
		boolField("all", "True when -a or --all is present."),
		boolField("volumes_flag", "True when --volumes is present."),
		boolField("prune", "True for docker system/image/volume/builder prune forms."),
		boolField("all_resources", "True for prune -a/--all or compose --all-resources."),
		boolField("remove_orphans", "True when --remove-orphans is present."),
		stringListField("flags_contains", "Parser-recognized docker option tokens that must be present."),
		stringListField("flags_prefixes", "Parser-recognized docker option tokens that must start with these prefixes."),
	},
	Examples: []Example{
		{Title: "Allow read-only Docker commands", YAML: `permission:
  allow:
    - command:
        name: docker
        semantic:
          verb_in: [ps, images, inspect, logs, version, info]`},
		{Title: "Deny privileged containers", YAML: `permission:
  deny:
    - command:
        name: docker
        semantic:
          verb: run
          privileged: true`},
		{Title: "Deny Docker socket mounts", YAML: `permission:
  deny:
    - command:
        name: docker
        semantic:
          docker_socket_mount: true`},
	},
	Notes: []string{
		"Docker semantics are syntactic only; the parser does not inspect images, Dockerfiles, compose files, containers, or daemon state.",
		"`host_mount`, `root_mount`, and `docker_socket_mount` are best-effort checks over command-line mount flags.",
	},
}

func init() {
	RegisterSchema(dockerSchema)
}
