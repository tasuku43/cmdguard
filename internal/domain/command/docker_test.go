package command

import "testing"

func TestDockerParserExtractsSemanticFields(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want DockerSemantic
	}{
		{name: "ps", raw: "docker ps", want: DockerSemantic{Verb: "ps"}},
		{name: "images", raw: "docker images", want: DockerSemantic{Verb: "images"}},
		{name: "inspect container", raw: "docker inspect my-container", want: DockerSemantic{Verb: "inspect", Container: "my-container"}},
		{name: "logs container", raw: "docker logs my-container", want: DockerSemantic{Verb: "logs", Container: "my-container"}},
		{name: "run rm image", raw: "docker run --rm alpine echo hello", want: DockerSemantic{Verb: "run", Image: "alpine", RM: true}},
		{name: "privileged", raw: "docker run --privileged alpine", want: DockerSemantic{Verb: "run", Image: "alpine", Privileged: true}},
		{name: "host network", raw: "docker run --network host alpine", want: DockerSemantic{Verb: "run", Image: "alpine", Network: "host", NetworkHost: true}},
		{name: "host pid", raw: "docker run --pid=host alpine", want: DockerSemantic{Verb: "run", Image: "alpine", PID: "host", PIDHost: true}},
		{name: "root volume", raw: "docker run -v /:/host alpine", want: DockerSemantic{Verb: "run", Image: "alpine", HostMount: true, RootMount: true}},
		{name: "socket volume", raw: "docker run -v /var/run/docker.sock:/var/run/docker.sock alpine", want: DockerSemantic{Verb: "run", Image: "alpine", HostMount: true, DockerSocketMount: true}},
		{name: "root mount", raw: "docker run --mount type=bind,source=/,target=/host alpine", want: DockerSemantic{Verb: "run", Image: "alpine", HostMount: true, RootMount: true}},
		{name: "exec it", raw: "docker exec -it my-container sh", want: DockerSemantic{Verb: "exec", Container: "my-container", Interactive: true, Tty: true}},
		{name: "rm force", raw: "docker rm -f my-container", want: DockerSemantic{Verb: "rm", Container: "my-container", Force: true}},
		{name: "system prune", raw: "docker system prune -a --volumes", want: DockerSemantic{Verb: "system", Subverb: "prune", Prune: true, All: true, AllResources: true, VolumesFlag: true}},
		{name: "image prune", raw: "docker image prune -a", want: DockerSemantic{Verb: "image", Subverb: "prune", Prune: true, All: true, AllResources: true}},
		{name: "build", raw: "docker build -t myapp:latest --no-cache --build-arg TOKEN=abc .", want: DockerSemantic{Verb: "build", Image: "myapp:latest", NoCache: true, BuildArgKeys: []string{"TOKEN"}}},
		{name: "compose dry run up", raw: "docker compose --dry-run up --build -d", want: DockerSemantic{Verb: "compose", Subverb: "up", ComposeCommand: "up", DryRun: true, Detach: true}},
		{name: "compose file project", raw: "docker compose -f docker-compose.prod.yml -p prod up -d", want: DockerSemantic{Verb: "compose", Subverb: "up", ComposeCommand: "up", File: "docker-compose.prod.yml", ProjectName: "prod", Detach: true}},
		{name: "compose down volumes", raw: "docker compose down --volumes --remove-orphans", want: DockerSemantic{Verb: "compose", Subverb: "down", ComposeCommand: "down", VolumesFlag: true, RemoveOrphans: true}},
		{name: "compose run service", raw: "docker compose run --rm web bash", want: DockerSemantic{Verb: "compose", Subverb: "run", ComposeCommand: "run", Service: "web", RM: true}},
		{name: "compose exec service", raw: "docker compose exec web sh", want: DockerSemantic{Verb: "compose", Subverb: "exec", ComposeCommand: "exec", Service: "web"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := Parse(tt.raw)
			if len(plan.Commands) != 1 {
				t.Fatalf("commands = %d, want 1", len(plan.Commands))
			}
			got := plan.Commands[0]
			if got.Parser != "docker" || got.SemanticParser != "docker" || got.Docker == nil {
				t.Fatalf("parser state = (%q, %q, %v), want docker semantic", got.Parser, got.SemanticParser, got.Docker)
			}
			assertDockerSemantic(t, *got.Docker, tt.want)
		})
	}
}

func assertDockerSemantic(t *testing.T, got DockerSemantic, want DockerSemantic) {
	t.Helper()
	if want.Verb != "" && got.Verb != want.Verb || want.Subverb != "" && got.Subverb != want.Subverb ||
		want.ComposeCommand != "" && got.ComposeCommand != want.ComposeCommand || want.Image != "" && got.Image != want.Image ||
		want.Container != "" && got.Container != want.Container || want.Service != "" && got.Service != want.Service ||
		want.File != "" && got.File != want.File || want.ProjectName != "" && got.ProjectName != want.ProjectName ||
		want.Network != "" && got.Network != want.Network || want.PID != "" && got.PID != want.PID {
		t.Fatalf("semantic = %+v, want matching %+v", got, want)
	}
	if got.RM != want.RM || got.Force != want.Force || got.Privileged != want.Privileged ||
		got.NetworkHost != want.NetworkHost || got.PIDHost != want.PIDHost ||
		got.HostMount != want.HostMount || got.RootMount != want.RootMount || got.DockerSocketMount != want.DockerSocketMount ||
		got.Interactive != want.Interactive || got.Tty != want.Tty || got.Prune != want.Prune ||
		got.All != want.All || got.AllResources != want.AllResources || got.VolumesFlag != want.VolumesFlag ||
		got.NoCache != want.NoCache || got.DryRun != want.DryRun || got.Detach != want.Detach ||
		got.RemoveOrphans != want.RemoveOrphans {
		t.Fatalf("semantic = %+v, want booleans matching %+v", got, want)
	}
	for _, key := range want.BuildArgKeys {
		if !containsString(got.BuildArgKeys, key) {
			t.Fatalf("build args = %v, want %q", got.BuildArgKeys, key)
		}
	}
}
