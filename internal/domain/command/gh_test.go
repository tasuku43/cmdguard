package command

import "testing"

func TestGhParserExtractsAPISemanticFields(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		wantMethod   string
		wantEndpoint string
		wantField    string
		wantRawField string
		wantHeader   string
		wantPaginate bool
		wantInput    bool
	}{
		{name: "relative endpoint", raw: "gh api repos/OWNER/REPO/pulls", wantMethod: "GET", wantEndpoint: "/repos/OWNER/REPO/pulls"},
		{name: "absolute endpoint", raw: "gh api /repos/OWNER/REPO/pulls", wantMethod: "GET", wantEndpoint: "/repos/OWNER/REPO/pulls"},
		{name: "delete method", raw: "gh api -X DELETE repos/OWNER/REPO/issues/1", wantMethod: "DELETE", wantEndpoint: "/repos/OWNER/REPO/issues/1"},
		{name: "post method", raw: "gh api --method POST repos/OWNER/REPO/dispatches", wantMethod: "POST", wantEndpoint: "/repos/OWNER/REPO/dispatches"},
		{name: "fields and header", raw: `gh api -F title=hello -f body=world -H "Accept: application/vnd.github+json" repos/OWNER/REPO/issues`, wantMethod: "GET", wantEndpoint: "/repos/OWNER/REPO/issues", wantField: "title", wantRawField: "body", wantHeader: "accept"},
		{name: "paginate", raw: "gh api --paginate repos/OWNER/REPO/issues", wantMethod: "GET", wantEndpoint: "/repos/OWNER/REPO/issues", wantPaginate: true},
		{name: "input", raw: "gh api --input payload.json repos/OWNER/REPO/dispatches", wantMethod: "GET", wantEndpoint: "/repos/OWNER/REPO/dispatches", wantInput: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := singleParsedCommand(t, tt.raw)
			if cmd.Parser != "gh" || cmd.SemanticParser != "gh" || cmd.Gh == nil {
				t.Fatalf("parser state = (%q, %q, %v), want gh semantic", cmd.Parser, cmd.SemanticParser, cmd.Gh)
			}
			got := cmd.Gh
			if got.Area != "api" || got.Method != tt.wantMethod || got.Endpoint != tt.wantEndpoint {
				t.Fatalf("Gh = %+v, want area=api method=%q endpoint=%q", got, tt.wantMethod, tt.wantEndpoint)
			}
			if tt.wantField != "" && !containsString(got.FieldKeys, tt.wantField) {
				t.Fatalf("FieldKeys=%#v, want %q", got.FieldKeys, tt.wantField)
			}
			if tt.wantRawField != "" && !containsString(got.RawFieldKeys, tt.wantRawField) {
				t.Fatalf("RawFieldKeys=%#v, want %q", got.RawFieldKeys, tt.wantRawField)
			}
			if tt.wantHeader != "" && !containsString(got.HeaderKeys, tt.wantHeader) {
				t.Fatalf("HeaderKeys=%#v, want %q", got.HeaderKeys, tt.wantHeader)
			}
			if got.Paginate != tt.wantPaginate {
				t.Fatalf("Paginate=%v, want %v", got.Paginate, tt.wantPaginate)
			}
			if got.Input != tt.wantInput {
				t.Fatalf("Input=%v, want %v", got.Input, tt.wantInput)
			}
		})
	}
}

func TestGhParserExtractsPRSemanticFields(t *testing.T) {
	tests := []struct {
		name              string
		raw               string
		wantVerb          string
		wantPRNumber      string
		wantBase          string
		wantHead          string
		wantDraft         bool
		wantFill          bool
		wantForce         bool
		wantMergeStrategy string
		wantDeleteBranch  bool
		wantAdmin         bool
		wantAuto          bool
	}{
		{name: "view", raw: "gh pr view 123", wantVerb: "view", wantPRNumber: "123"},
		{name: "checkout force", raw: "gh pr checkout 123 --force", wantVerb: "checkout", wantPRNumber: "123", wantForce: true},
		{name: "create", raw: "gh pr create --base main --head feature --draft --fill", wantVerb: "create", wantBase: "main", wantHead: "feature", wantDraft: true, wantFill: true},
		{name: "merge squash delete branch", raw: "gh pr merge 123 --squash --delete-branch", wantVerb: "merge", wantPRNumber: "123", wantMergeStrategy: "squash", wantDeleteBranch: true},
		{name: "merge admin", raw: "gh pr merge 123 --admin", wantVerb: "merge", wantPRNumber: "123", wantAdmin: true},
		{name: "merge auto", raw: "gh pr merge 123 --auto", wantVerb: "merge", wantPRNumber: "123", wantAuto: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := singleParsedCommand(t, tt.raw).Gh
			if got == nil {
				t.Fatal("Gh semantic is nil")
			}
			if got.Area != "pr" || got.Verb != tt.wantVerb || got.PRNumber != tt.wantPRNumber ||
				got.Base != tt.wantBase || got.Head != tt.wantHead || got.Draft != tt.wantDraft ||
				got.Fill != tt.wantFill || got.Force != tt.wantForce || got.MergeStrategy != tt.wantMergeStrategy ||
				got.DeleteBranch != tt.wantDeleteBranch || got.Admin != tt.wantAdmin || got.Auto != tt.wantAuto {
				t.Fatalf("Gh=%+v, want verb=%q pr=%q base=%q head=%q draft=%v fill=%v force=%v strategy=%q delete=%v admin=%v auto=%v",
					got, tt.wantVerb, tt.wantPRNumber, tt.wantBase, tt.wantHead, tt.wantDraft, tt.wantFill, tt.wantForce, tt.wantMergeStrategy, tt.wantDeleteBranch, tt.wantAdmin, tt.wantAuto)
			}
		})
	}
}

func TestGhParserExtractsRunSemanticFields(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		wantVerb       string
		wantRunID      string
		wantFailed     bool
		wantJob        string
		wantDebug      bool
		wantForce      bool
		wantExitStatus bool
	}{
		{name: "view", raw: "gh run view 123", wantVerb: "view", wantRunID: "123"},
		{name: "rerun failed", raw: "gh run rerun 123 --failed", wantVerb: "rerun", wantRunID: "123", wantFailed: true},
		{name: "rerun job debug", raw: "gh run rerun 123 --job 456 --debug", wantVerb: "rerun", wantRunID: "123", wantJob: "456", wantDebug: true},
		{name: "cancel force", raw: "gh run cancel 123 --force", wantVerb: "cancel", wantRunID: "123", wantForce: true},
		{name: "watch exit status", raw: "gh run watch 123 --exit-status", wantVerb: "watch", wantRunID: "123", wantExitStatus: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := singleParsedCommand(t, tt.raw).Gh
			if got == nil {
				t.Fatal("Gh semantic is nil")
			}
			if got.Area != "run" || got.Verb != tt.wantVerb || got.RunID != tt.wantRunID ||
				got.Failed != tt.wantFailed || got.Job != tt.wantJob || got.Debug != tt.wantDebug ||
				got.Force != tt.wantForce || got.ExitStatus != tt.wantExitStatus {
				t.Fatalf("Gh=%+v, want verb=%q run_id=%q failed=%v job=%q debug=%v force=%v exit_status=%v",
					got, tt.wantVerb, tt.wantRunID, tt.wantFailed, tt.wantJob, tt.wantDebug, tt.wantForce, tt.wantExitStatus)
			}
		})
	}
}

func TestGhParserExtractsCommonSemanticFields(t *testing.T) {
	got := singleParsedCommand(t, "GH_HOST=github.example.com gh --repo owner/repo api --hostname ghe.example.com --web repos/OWNER/REPO/pulls").Gh
	if got == nil {
		t.Fatal("Gh semantic is nil")
	}
	if got.Repo != "owner/repo" || got.Hostname != "ghe.example.com" || !got.Web {
		t.Fatalf("Gh=%+v, want repo, cli hostname, web", got)
	}
}

func singleParsedCommand(t *testing.T, raw string) Command {
	t.Helper()
	plan := Parse(raw)
	if len(plan.Commands) != 1 {
		t.Fatalf("len(Commands)=%d, want 1", len(plan.Commands))
	}
	return plan.Commands[0]
}
