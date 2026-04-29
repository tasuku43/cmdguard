package command

import "testing"

func TestHelmParserExtractsSemanticFields(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		wantVerb       string
		wantSubverb    string
		wantRelease    string
		wantChart      string
		wantNamespace  string
		wantRepoName   string
		wantRepoURL    string
		wantRegistry   string
		wantPluginName string
		wantSetKey     string
		wantValuesFile string
		wantWait       bool
		wantAtomic     bool
		wantInstall    bool
		wantForce      bool
		wantCreateNS   bool
		wantKeepHist   bool
	}{
		{name: "list", raw: "helm list -n default", wantVerb: "list", wantNamespace: "default"},
		{name: "status", raw: "helm status my-release --namespace prod", wantVerb: "status", wantRelease: "my-release", wantNamespace: "prod"},
		{name: "get values", raw: "helm get values my-release -n prod", wantVerb: "get", wantSubverb: "values", wantRelease: "my-release", wantNamespace: "prod"},
		{name: "template values and set", raw: "helm template my-release ./chart -f values.yaml --set image.tag=abc", wantVerb: "template", wantRelease: "my-release", wantChart: "./chart", wantValuesFile: "values.yaml", wantSetKey: "image.tag"},
		{name: "install wait atomic", raw: "helm install my-release ./chart -n prod --wait --atomic", wantVerb: "install", wantRelease: "my-release", wantChart: "./chart", wantNamespace: "prod", wantWait: true, wantAtomic: true},
		{name: "upgrade install force create namespace", raw: "helm upgrade my-release ./chart --install --force --create-namespace -n prod", wantVerb: "upgrade", wantRelease: "my-release", wantChart: "./chart", wantNamespace: "prod", wantInstall: true, wantForce: true, wantCreateNS: true},
		{name: "uninstall keep history", raw: "helm uninstall my-release -n prod --keep-history", wantVerb: "uninstall", wantRelease: "my-release", wantNamespace: "prod", wantKeepHist: true},
		{name: "rollback", raw: "helm rollback my-release 2 -n prod", wantVerb: "rollback", wantRelease: "my-release", wantNamespace: "prod"},
		{name: "repo add", raw: "helm repo add bitnami https://charts.bitnami.com/bitnami", wantVerb: "repo", wantSubverb: "add", wantRepoName: "bitnami", wantRepoURL: "https://charts.bitnami.com/bitnami"},
		{name: "plugin install", raw: "helm plugin install https://example.com/plugin.git", wantVerb: "plugin", wantSubverb: "install", wantPluginName: "https://example.com/plugin.git"},
		{name: "registry login", raw: "helm registry login ghcr.io", wantVerb: "registry", wantSubverb: "login", wantRegistry: "ghcr.io"},
		{name: "push", raw: "helm push chart.tgz oci://registry.example.com/charts", wantVerb: "push", wantChart: "chart.tgz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := Parse(tt.raw)
			if len(plan.Commands) != 1 {
				t.Fatalf("len(Commands) = %d, want 1", len(plan.Commands))
			}
			cmd := plan.Commands[0]
			if cmd.Parser != "helm" || cmd.SemanticParser != "helm" || cmd.Helm == nil {
				t.Fatalf("parser state = (%q, %q, %v), want helm semantic", cmd.Parser, cmd.SemanticParser, cmd.Helm)
			}
			got := cmd.Helm
			if got.Verb != tt.wantVerb || got.Subverb != tt.wantSubverb || got.Release != tt.wantRelease ||
				got.Chart != tt.wantChart || got.Namespace != tt.wantNamespace || got.RepoName != tt.wantRepoName ||
				got.RepoURL != tt.wantRepoURL || got.Registry != tt.wantRegistry || got.PluginName != tt.wantPluginName ||
				got.Wait != tt.wantWait || got.Atomic != tt.wantAtomic || got.Install != tt.wantInstall ||
				got.Force != tt.wantForce || got.CreateNamespace != tt.wantCreateNS || got.KeepHistory != tt.wantKeepHist {
				t.Fatalf("Helm = %+v", got)
			}
			if tt.wantValuesFile != "" && !containsTestString(got.ValuesFiles, tt.wantValuesFile) {
				t.Fatalf("ValuesFiles=%#v, want %q", got.ValuesFiles, tt.wantValuesFile)
			}
			if tt.wantSetKey != "" && !containsTestString(got.SetKeys, tt.wantSetKey) {
				t.Fatalf("SetKeys=%#v, want %q", got.SetKeys, tt.wantSetKey)
			}
		})
	}
}
