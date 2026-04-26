package command

import "testing"

func TestArgoCDParserExtractsSemanticFields(t *testing.T) {
	tests := []struct {
		name         string
		raw          string
		wantVerb     string
		wantAppName  string
		wantProject  string
		wantRevision string
	}{
		{name: "app list project", raw: "argocd app list --project prod", wantVerb: "app list", wantProject: "prod"},
		{name: "app get", raw: "argocd app get payments", wantVerb: "app get", wantAppName: "payments"},
		{name: "app diff revision", raw: "argocd app diff payments --revision abc123", wantVerb: "app diff", wantAppName: "payments", wantRevision: "abc123"},
		{name: "app rollback positional revision", raw: "argocd app rollback payments 42", wantVerb: "app rollback", wantAppName: "payments", wantRevision: "42"},
		{name: "app sync", raw: "argocd app sync payments", wantVerb: "app sync", wantAppName: "payments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := singleParsedCommand(t, tt.raw)
			if cmd.Parser != "argocd" || cmd.SemanticParser != "argocd" || cmd.ArgoCD == nil {
				t.Fatalf("parser state = (%q, %q, %v), want argocd semantic", cmd.Parser, cmd.SemanticParser, cmd.ArgoCD)
			}
			got := cmd.ArgoCD
			if got.Verb != tt.wantVerb || got.AppName != tt.wantAppName || got.Project != tt.wantProject || got.Revision != tt.wantRevision {
				t.Fatalf("ArgoCD=%+v, want verb=%q app=%q project=%q revision=%q", got, tt.wantVerb, tt.wantAppName, tt.wantProject, tt.wantRevision)
			}
		})
	}
}
