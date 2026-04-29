package buildinfo

import "testing"

func TestReadIncludesReleaseLdflags(t *testing.T) {
	oldVersion, oldCommit, oldDate := version, commit, date
	t.Cleanup(func() {
		version, commit, date = oldVersion, oldCommit, oldDate
	})

	version = "1.2.3"
	commit = "abc123"
	date = "2026-04-29T00:00:00Z"

	got := Read()
	if got.Version != "1.2.3" {
		t.Fatalf("Version = %q, want %q", got.Version, "1.2.3")
	}
	if got.VCSRevision != "abc123" {
		t.Fatalf("VCSRevision = %q, want %q", got.VCSRevision, "abc123")
	}
	if got.BuildDate != "2026-04-29T00:00:00Z" {
		t.Fatalf("BuildDate = %q, want %q", got.BuildDate, "2026-04-29T00:00:00Z")
	}
}
