package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runCLIHelpTest(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := Run(args, Streams{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}, Env{Cwd: ".", Home: "."})
	return code, stdout.String(), stderr.String()
}

func TestHelpMatchExplainsSemantic(t *testing.T) {
	code, stdout, stderr := runCLIHelpTest("help", "match")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"semantic",
		"cc-bash-proxy help semantic",
		"Permission rules do not use match or pattern",
		"patterns",
		"command.name",
		"semantic.flags_contains",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help match missing %q:\n%s", want, stdout)
		}
	}
}

func TestHelpSemanticListsSupportedCommands(t *testing.T) {
	code, stdout, stderr := runCLIHelpTest("help", "semantic")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{"git", "gh", "aws", "kubectl", "helmfile"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help semantic missing %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "argocd") {
		t.Fatalf("help semantic listed unimplemented argocd:\n%s", stdout)
	}
}

func TestHelpSemanticGitShowsSchema(t *testing.T) {
	code, stdout, stderr := runCLIHelpTest("help", "semantic", "git")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"Semantic schema: git",
		"verb",
		"force",
		"--force-with-lease",
		"Examples:",
		"Validation rules:",
		"command.semantic requires exact command.name",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("help semantic git missing %q:\n%s", want, stdout)
		}
	}
}

func TestHelpSemanticUnknownFails(t *testing.T) {
	code, _, stderr := runCLIHelpTest("help", "semantic", "unknown")
	if code == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(stderr, "unknown semantic command") || !strings.Contains(stderr, "git") {
		t.Fatalf("stderr=%s", stderr)
	}
}

func TestSemanticSchemaJSON(t *testing.T) {
	code, stdout, stderr := runCLIHelpTest("semantic-schema", "--format", "json")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	var payload struct {
		Schemas []struct {
			Command string `json:"command"`
			Fields  []struct {
				Name string `json:"name"`
			} `json:"fields"`
		} `json:"schemas"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	if len(payload.Schemas) == 0 {
		t.Fatalf("missing schemas")
	}
	foundGit := false
	for _, schema := range payload.Schemas {
		if schema.Command == "git" {
			foundGit = true
			if len(schema.Fields) == 0 {
				t.Fatalf("git fields empty")
			}
		}
	}
	if !foundGit {
		t.Fatalf("git schema missing: %+v", payload.Schemas)
	}
}
