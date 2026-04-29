package ccbashguard_test

import (
	"os"
	"strings"
	"testing"
)

func TestInstallTrustDocsCoverVerificationBoundaries(t *testing.T) {
	body := readProjectFile(t, "INSTALL.md")
	required := []string{
		"`cc-bash-guard` is part of your shell execution trust boundary",
		"Checksum verification confirms that the downloaded archive matches the digest",
		"Checksums alone do not",
		"independently prove source provenance",
		"GitHub artifact attestations provide provenance signals",
		"`tasuku43/cc-bash-guard`",
		"It does not prove runtime behavior is safe",
		"does not prove the policy",
		"engine is bug-free",
		"gh attestation verify \"$ARCHIVE\" -R tasuku43/cc-bash-guard",
		"cc-bash-guard version --format json",
		"cc-bash-guard verify --format json",
		"Install trust boundary checklist",
		"Treat Homebrew tap, mise plugin behavior, and source builds as part of the",
	}
	for _, want := range required {
		if !strings.Contains(body, want) {
			t.Fatalf("INSTALL.md missing %q", want)
		}
	}
}

func TestSecurityDocsHavePrivateReportingPath(t *testing.T) {
	body := readProjectFile(t, "SECURITY.md")
	required := []string{
		"Please do not open public issues for unpatched vulnerabilities.",
		"GitHub private vulnerability reporting is currently not enabled",
		"TODO: enable GitHub private vulnerability reporting",
		"private maintainer contact address",
	}
	for _, want := range required {
		if !strings.Contains(body, want) {
			t.Fatalf("SECURITY.md missing %q", want)
		}
	}
}

func readProjectFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
