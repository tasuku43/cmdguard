package rule

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileIfPresentValidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cmdguard.yml")
	if err := os.WriteFile(path, []byte("version: 1\nrules: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFileIfPresent(Source{Layer: LayerUser, Path: path})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadEffectiveUsesUserConfig(t *testing.T) {
	cwd := t.TempDir()
	home := t.TempDir()
	userPath := filepath.Join(home, ".config", "cmdguard", "cmdguard.yml")
	if err := os.MkdirAll(filepath.Dir(userPath), 0o755); err != nil {
		t.Fatal(err)
	}
	user := `version: 1
rules:
  - id: user-rule
    pattern: "^echo"
    message: "user"
    block_examples: ["echo hi"]
    allow_examples: ["git status"]
`
	if err := os.WriteFile(userPath, []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded := LoadEffective(cwd, home, "")
	if len(loaded.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", loaded.Errors)
	}
	if len(loaded.Rules) != 1 {
		t.Fatalf("got %d rules", len(loaded.Rules))
	}
	if loaded.Rules[0].ID != "user-rule" {
		t.Fatalf("first rule = %s", loaded.Rules[0].ID)
	}
}

func TestLoadFileForEvalIfPresentSkipsExamplesButValidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cmdguard.yml")
	cachePath := filepath.Join(t.TempDir(), "eval-cache-v1.json")
	body := `version: 1
rules:
  - id: user-rule
    pattern: "^git"
    message: "use a safer alternative instead"
    block_examples:
      - "git status"
      - "git log"
    allow_examples:
      - "echo ok"
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	rules, err := LoadFileForEvalIfPresent(Source{Layer: LayerUser, Path: path}, cachePath)
	if err != nil {
		t.Fatalf("LoadFileForEvalIfPresent() error = %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("got %d rules", len(rules))
	}
	if len(rules[0].BlockExamples) != 0 || len(rules[0].AllowExamples) != 0 {
		t.Fatalf("examples should not be loaded: %+v", rules[0])
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to be written: %v", err)
	}
}
