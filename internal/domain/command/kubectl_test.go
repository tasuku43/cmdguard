package command

import "testing"

func TestKubectlParserExtractsSemanticFields(t *testing.T) {
	tests := []struct {
		name             string
		raw              string
		wantVerb         string
		wantSubverb      string
		wantResourceType string
		wantResourceName string
		wantNamespace    string
		wantContext      string
		wantFilename     string
		wantAllNS        bool
		wantDryRun       bool
		wantForce        bool
		wantContainer    string
	}{
		{name: "get pods", raw: "kubectl get pods", wantVerb: "get", wantResourceType: "pods"},
		{name: "get pod name", raw: "kubectl get pod foo", wantVerb: "get", wantResourceType: "pod", wantResourceName: "foo"},
		{name: "get slash resource", raw: "kubectl get pod/foo", wantVerb: "get", wantResourceType: "pod", wantResourceName: "foo"},
		{name: "namespace short", raw: "kubectl -n prod get pods", wantVerb: "get", wantResourceType: "pods", wantNamespace: "prod"},
		{name: "namespace and context equals", raw: "kubectl --namespace=prod --context=prod-cluster get pods", wantVerb: "get", wantResourceType: "pods", wantNamespace: "prod", wantContext: "prod-cluster"},
		{name: "delete force", raw: "kubectl delete deployment/foo --force", wantVerb: "delete", wantResourceType: "deployment", wantResourceName: "foo", wantForce: true},
		{name: "apply filename dry run", raw: "kubectl apply -f deployment.yaml --dry-run=server", wantVerb: "apply", wantFilename: "deployment.yaml", wantDryRun: true},
		{name: "all namespaces", raw: "kubectl get pods -A", wantVerb: "get", wantResourceType: "pods", wantAllNS: true},
		{name: "logs container", raw: "kubectl logs pod/foo -c app", wantVerb: "logs", wantResourceType: "pod", wantResourceName: "foo", wantContainer: "app"},
		{name: "rollout restart", raw: "kubectl rollout restart deployment/foo", wantVerb: "rollout", wantSubverb: "restart", wantResourceType: "deployment", wantResourceName: "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := Parse(tt.raw)
			if len(plan.Commands) != 1 {
				t.Fatalf("len(Commands) = %d, want 1", len(plan.Commands))
			}
			cmd := plan.Commands[0]
			if cmd.Parser != "kubectl" || cmd.SemanticParser != "kubectl" || cmd.Kubectl == nil {
				t.Fatalf("parser state = (%q, %q, %v), want kubectl semantic", cmd.Parser, cmd.SemanticParser, cmd.Kubectl)
			}
			got := cmd.Kubectl
			if got.Verb != tt.wantVerb ||
				got.Subverb != tt.wantSubverb ||
				got.ResourceType != tt.wantResourceType ||
				got.ResourceName != tt.wantResourceName ||
				got.Namespace != tt.wantNamespace ||
				got.Context != tt.wantContext ||
				got.AllNamespaces != tt.wantAllNS ||
				got.Force != tt.wantForce ||
				got.Container != tt.wantContainer {
				t.Fatalf("Kubectl = %+v, want verb=%q subverb=%q resource=%q/%q namespace=%q context=%q allNS=%v force=%v container=%q",
					got, tt.wantVerb, tt.wantSubverb, tt.wantResourceType, tt.wantResourceName, tt.wantNamespace, tt.wantContext, tt.wantAllNS, tt.wantForce, tt.wantContainer)
			}
			if tt.wantDryRun {
				if got.DryRun == nil || !*got.DryRun {
					t.Fatalf("DryRun = %v, want true", got.DryRun)
				}
			} else if got.DryRun != nil {
				t.Fatalf("DryRun = %v, want nil", got.DryRun)
			}
			if tt.wantFilename != "" {
				if len(got.Filenames) != 1 || got.Filenames[0] != tt.wantFilename {
					t.Fatalf("Filenames = %#v, want [%q]", got.Filenames, tt.wantFilename)
				}
			}
		})
	}
}
