package trustedsource_test

import (
	"strings"
	"testing"
	"time"

	trustedsource "github.com/7K-Inari/inari-api/contract/trustedsource/v1"
)

const validSourceYAML = `environment: production
catalog:
  syncInterval: 5m
  sources:
    - id: community
      indexRef: ghcr.io/7k-hiroba/argocd/templates/index:latest
      trustTier: reviewed
      artifactPrefixes:
        - ghcr.io/7k-hiroba/argocd/templates/
      publisher:
        issuer: https://token.actions.githubusercontent.com
        repository: 7k-hiroba/workflows
        workflows:
          index: .github/workflows/release.yml
          catalogPackage: .github/workflows/release-oci.yml
          template: .github/workflows/release-oci.yml
          helmChart: .github/workflows/release-chart.yml
          smokeSuite: .github/workflows/release-smoke-suite.yml
        workflowSha: 0123456789abcdef0123456789abcdef01234567
        sourceRepository: 7k-hiroba/argocd
        sourceWorkflow: .github/workflows/release.yml
        sourceRef: refs/heads/main
        event: push
        releaseEnvironment: release
template:
  sources: []
`

func TestDecodeFile(t *testing.T) {
	f, err := trustedsource.DecodeFile([]byte(validSourceYAML))
	if err != nil {
		t.Fatalf("DecodeFile: %v", err)
	}
	if len(f.Catalog.Sources) != 1 || f.Catalog.Sources[0].ID != "community" {
		t.Fatalf("decoded file = %+v", f)
	}
}

func TestDecodeFileRejectsUnknownAndDuplicateKeys(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"empty", "", "empty"},
		{"unknown", validSourceYAML + "unknown: true\n", "unknown"},
		{"duplicate", strings.Replace(validSourceYAML, "event: push", "event: push\n        event: workflow_dispatch", 1), "duplicate"},
		{"multiple documents", validSourceYAML + "---\n{}\n", "multiple"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := trustedsource.DecodeFile([]byte(tc.input)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestEvidenceValidatesChartAndSmokeLanes(t *testing.T) {
	now := mustTime(t, "2026-11-13T12:30:00Z")
	for _, tc := range []struct{ name, mediaType, workflow string }{
		{"chart", "application/vnd.cncf.helm.chart.content.v1.tar+gzip", ".github/workflows/release-chart.yml"},
		{"smoke", trustedsource.SmokeSuiteMediaType, ".github/workflows/release-smoke-suite.yml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := validEvidence()
			e.Artifact.MediaType = tc.mediaType
			e.Producer.WorkflowPath = tc.workflow
			if err := e.ValidateAt(now, 0, false); err != nil {
				t.Fatalf("valid %s evidence rejected: %v", tc.name, err)
			}
		})
	}
}

func mustTime(t *testing.T, value string) (out time.Time) {
	t.Helper()
	out, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
