package policy

import "testing"

func TestEvaluateDockerSemanticRules(t *testing.T) {
	trueValue := true
	p := NewPipeline(PipelineSpec{
		Permission: PermissionSpec{
			Deny: []PermissionRuleSpec{
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{Verb: "run", Privileged: &trueValue}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{NetworkHost: &trueValue}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{PIDHost: &trueValue}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{DockerSocketMount: &trueValue}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{RootMount: &trueValue}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{Verb: "system", Subverb: "prune", AllResources: &trueValue, VolumesFlag: &trueValue}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{Verb: "compose", ComposeCommand: "down", VolumesFlag: &trueValue}}},
			},
			Ask: []PermissionRuleSpec{
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{Verb: "run"}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{Verb: "compose", ComposeCommandIn: []string{"up", "run", "exec"}}}},
			},
			Allow: []PermissionRuleSpec{
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{VerbIn: []string{"ps", "images", "inspect", "logs", "version", "info"}}}},
				{Command: PermissionCommandSpec{Name: "docker", Semantic: &SemanticMatchSpec{Verb: "compose", ComposeCommandIn: []string{"ps", "logs", "config", "images", "ls"}, File: "docker-compose.prod.yml", ProjectName: "prod"}}},
			},
		},
	}, Source{})

	tests := []struct {
		command string
		want    string
	}{
		{command: "docker ps", want: "allow"},
		{command: "docker run --rm alpine echo hello", want: "ask"},
		{command: "docker run --privileged alpine", want: "deny"},
		{command: "docker run --network host alpine", want: "deny"},
		{command: "docker run --pid=host alpine", want: "deny"},
		{command: "docker run -v /var/run/docker.sock:/var/run/docker.sock alpine", want: "deny"},
		{command: "docker run -v /:/host alpine", want: "deny"},
		{command: "docker system prune -a --volumes", want: "deny"},
		{command: "docker compose up -d", want: "ask"},
		{command: "docker compose down --volumes", want: "deny"},
		{command: "docker compose -f docker-compose.prod.yml -p prod ps", want: "allow"},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got, err := Evaluate(p, tt.command)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if got.Outcome != tt.want {
				t.Fatalf("Outcome = %q, want %q; decision=%+v", got.Outcome, tt.want, got)
			}
		})
	}
}
