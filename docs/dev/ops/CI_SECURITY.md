# CI Security Checks

`cc-bash-guard` gates shell permission decisions before execution, so CI treats
parser regressions, dependency risk, and release artifact integrity as
security-relevant.

## Pull Requests And Main

- `go test ./...` keeps the normal regression suite fast and unchanged.
- `go test -race ./...` runs in a separate job to catch data races without
  slowing the primary test job.
- `Staticcheck` runs through `task lint` with a pinned Go tool version.
  Stylecheck (`ST*`) and `U1000` are temporarily disabled to avoid code cleanup
  in this hardening-only change.
- `govulncheck` scans reachable Go vulnerability data.
- `CodeQL` analyzes Go code on pull requests, pushes to `main`, and weekly.
- `Dependency Review` runs on pull requests and fails only for high-or-higher
  severity dependency changes or explicitly denied copyleft licenses.
- `Dependabot` opens weekly pull requests for Go modules and GitHub Actions
  updates so dependency-review changes stay visible.
- `task smoke` builds a temporary binary and runs:
  - `cc-bash-guard version`
  - `cc-bash-guard help`
  - `cc-bash-guard help semantic`

## Scheduled Security Scans

`Security Nightly` runs `govulncheck` and OpenSSF Scorecard on `main` pushes,
weekly schedules, and manual dispatch. Scorecard uploads SARIF to GitHub code
scanning and publishes aggregate results without using repository secrets.

## Action Pinning

Existing first-party and release-critical actions are pinned to immutable SHAs.
New security actions use pinned major or release tags where SHA pinning needs a
separate review of upstream release digests. Follow-up work should pin:

- `github/codeql-action/*`
- `actions/dependency-review-action`
- `ossf/scorecard-action`
