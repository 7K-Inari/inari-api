package trustedsource_test

import (
	"strings"
	"testing"
	"time"

	trustedsource "github.com/7K-Inari/inari-api/contract/trustedsource/v1"
)

const (
	testSHA      = "0123456789abcdef0123456789abcdef01234567"
	testOtherSHA = "abcdef0123456789abcdef0123456789abcdef01"
	testDigest   = "sha256:52ccaee6d4dd272e54bfccda77738b42e1edf0e4a20c27e23f0b6c15d01aef79"
)

func validSource() trustedsource.Source {
	return trustedsource.Source{
		ID: "community", IndexRef: "ghcr.io/7k-hiroba/argocd/templates/index:latest",
		TrustTier:        trustedsource.TrustReviewed,
		ArtifactPrefixes: []string{"ghcr.io/7k-hiroba/argocd/templates/"},
		Publisher: trustedsource.Publisher{
			Issuer: "https://token.actions.githubusercontent.com", Repository: "7k-hiroba/workflows",
			Workflows: trustedsource.DefaultArtifactWorkflows(), WorkflowSHA: testSHA,
			SourceRepository: "7k-hiroba/argocd", SourceWorkflow: ".github/workflows/release.yml",
			SourceRef: "refs/heads/main", Event: "push", ReleaseEnvironment: "release",
		},
	}
}

func validFile() *trustedsource.File {
	catalog := trustedsource.SourceGroup{SyncInterval: "5m", Sources: []trustedsource.Source{validSource()}}
	templates := trustedsource.SourceGroup{SyncInterval: "5m", Sources: nil}
	return &trustedsource.File{Environment: trustedsource.EnvironmentProduction, Catalog: &catalog, Template: &templates}
}

func validLane() trustedsource.Lane {
	return trustedsource.Lane{Name: "test-packages", Required: true, Status: "passed", RunID: "123", JobID: "456", SourceSHA: testOtherSHA}
}

func validIndex() *trustedsource.Index {
	return &trustedsource.Index{
		APIVersion: trustedsource.APIVersion, Kind: trustedsource.CatalogIndexKind, MediaType: trustedsource.CatalogIndexMediaType,
		SourceSHA: testOtherSHA, SourceRef: "refs/heads/main",
		Producer: trustedsource.Producer{WorkflowPath: ".github/workflows/release.yml", WorkflowSHA: testSHA, RunID: "123", RunAttempt: 1, RunURI: "https://github.com/7k-hiroba/argocd/actions/runs/123"},
		Packages: []trustedsource.IndexEntry{{
			Name: "web-service", Version: "1.2.0", PackageVersion: "1.2.0", Channel: "stable", Type: "kro-rgd",
			Description: "Web service", Category: "compute", OCIRef: "ghcr.io/7k-hiroba/argocd/catalog/web-service:1.2.0",
			Digest: testDigest, SourceSHA: testOtherSHA, ReleaseTag: "web-service-v1.2.0", MediaType: trustedsource.CatalogPackageMediaType,
			Lanes: []trustedsource.Lane{validLane()},
		}},
	}
}

func validEvidence() *trustedsource.TestEvidence {
	return &trustedsource.TestEvidence{
		Schema: trustedsource.TestEvidenceSchema, App: "argocd", Component: "web-service", TrustTier: trustedsource.TrustReviewed,
		SourceRepository: "7k-hiroba/argocd", SourceSHA: testOtherSHA, ReleaseTag: "web-service-v1.2.0",
		Artifact: trustedsource.ArtifactEvidence{Repository: "ghcr.io/7k-hiroba/argocd/catalog/web-service", Digest: testDigest, MediaType: trustedsource.CatalogPackageMediaType},
		Producer: trustedsource.Producer{WorkflowPath: ".github/workflows/release-oci.yml", WorkflowSHA: testSHA, RunID: "123", RunAttempt: 1, RunURI: "https://github.com/7k-hiroba/argocd/actions/runs/123"},
		Lanes:    []trustedsource.Lane{validLane()}, FinishedAt: "2026-11-13T12:00:00Z", EvidenceDigest: testDigest,
	}
}

func validSLSA() *trustedsource.Statement {
	return &trustedsource.Statement{
		Type: "https://in-toto.io/Statement/v1", PredicateType: trustedsource.SLSAProvenancePredicateType,
		Subject: []trustedsource.Subject{{Name: "ghcr.io/7k-hiroba/argocd/catalog/web-service", Digest: map[string]string{"sha256": strings.TrimPrefix(testDigest, "sha256:")}}},
		Predicate: trustedsource.Provenance{
			BuildDefinition: trustedsource.BuildDefinition{
				BuildType:            "https://actions.github.io/buildtypes/workflow/v1",
				ExternalParameters:   trustedsource.ExternalParameters{Repository: "7k-hiroba/argocd", SourceSHA: testOtherSHA, ReleaseTag: "web-service-v1.2.0", PackageVersion: "1.2.0", Event: "push", SourceRef: "refs/heads/main", Environment: "release", RunID: "123", RunAttempt: 1},
				ResolvedDependencies: []trustedsource.Dependency{{URI: "https://github.com/7k-hiroba/workflows/.github/workflows/release-oci.yml", Digest: map[string]string{"sha": testSHA}}},
			},
			RunDetails: trustedsource.RunDetails{Builder: trustedsource.Builder{ID: "https://github.com/7k-hiroba/workflows/.github/workflows/release-oci.yml@" + testSHA}, Metadata: trustedsource.Metadata{InvocationID: "https://github.com/7k-hiroba/argocd/actions/runs/123"}},
		},
	}
}

func TestSourceFileValidateValid(t *testing.T) {
	if err := validFile().Validate(); err != nil {
		t.Fatalf("valid source file rejected: %v", err)
	}
	age, err := validSource().PluginCompatibilityAge()
	if err != nil || age != 7*24*time.Hour {
		t.Fatalf("default compatibility age = %s, %v", age, err)
	}
}

func TestSourceFileValidateNegatives(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*trustedsource.File)
		want   string
	}{
		{"environment", func(f *trustedsource.File) { f.Environment = "qa" }, "environment"},
		{"source id", func(f *trustedsource.File) { f.Catalog.Sources[0].ID = "Bad/ID" }, "catalog.sources[0].id"},
		{"index ref", func(f *trustedsource.File) { f.Catalog.Sources[0].IndexRef = "file:///tmp/index" }, "indexRef"},
		{"trust tier", func(f *trustedsource.File) { f.Catalog.Sources[0].TrustTier = "trusted" }, "trustTier"},
		{"unsigned official", func(f *trustedsource.File) {
			f.Catalog.Sources[0].TrustTier = trustedsource.TrustOfficial
			f.Catalog.Sources[0].AllowUnsigned = trustedsource.AllowUnsignedDevelopmentOnly
		}, "allowUnsigned"},
		{"publisher workflow", func(f *trustedsource.File) {
			f.Catalog.Sources[0].Publisher.Workflows.Index = ".github/workflows/other.yml"
		}, "workflows.index"},
		{"workflow sha", func(f *trustedsource.File) { f.Catalog.Sources[0].Publisher.WorkflowSHA = "main" }, "workflowSha"},
		{"artifact prefix", func(f *trustedsource.File) {
			f.Catalog.Sources[0].ArtifactPrefixes = []string{"https://ghcr.io/7k-hiroba/argocd/"}
		}, "artifactPrefixes"},
		{"sync interval", func(f *trustedsource.File) { f.Catalog.SyncInterval = "often" }, "syncInterval"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := validFile()
			tc.mutate(f)
			if err := f.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want field %q", err, tc.want)
			}
		})
	}
}

func TestSourceFileRejectsDuplicateIDsAcrossKinds(t *testing.T) {
	f := validFile()
	second := validSource()
	f.Template.Sources = []trustedsource.Source{second}
	if err := f.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v, want duplicate source ID", err)
	}
}

func TestIndexValidateValid(t *testing.T) {
	if err := validIndex().Validate(); err != nil {
		t.Fatalf("valid catalog index rejected: %v", err)
	}
	tpl := validIndex()
	tpl.Kind = trustedsource.TemplateIndexKind
	tpl.MediaType = trustedsource.TemplateIndexMediaType
	tpl.Templates = tpl.Packages
	tpl.Packages = nil
	tpl.Templates[0].MediaType = trustedsource.TemplateMediaType
	tpl.Templates[0].Type = "template"
	if err := tpl.Validate(); err != nil {
		t.Fatalf("valid template index rejected: %v", err)
	}
}

func TestIndexValidateNegatives(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*trustedsource.Index)
		want   string
	}{
		{"api version", func(i *trustedsource.Index) { i.APIVersion = "v2" }, "apiVersion"},
		{"kind", func(i *trustedsource.Index) { i.Kind = "Other" }, "kind"},
		{"top source sha", func(i *trustedsource.Index) { i.SourceSHA = "main" }, "sourceSha"},
		{"producer workflow", func(i *trustedsource.Index) { i.Producer.WorkflowPath = ".github/workflows/release-oci.yml" }, "producer.workflowPath"},
		{"missing digest", func(i *trustedsource.Index) { i.Packages[0].Digest = "" }, "digest"},
		{"mutable ref", func(i *trustedsource.Index) {
			i.Packages[0].OCIRef = "ghcr.io/7k-hiroba/argocd/catalog/web-service:latest"
		}, "ociRef"},
		{"duplicate version", func(i *trustedsource.Index) { i.Packages = append(i.Packages, i.Packages[0]) }, "duplicate"},
		{"wrong media", func(i *trustedsource.Index) { i.Packages[0].MediaType = trustedsource.TemplateMediaType }, "mediaType"},
		{"platform chart required", func(i *trustedsource.Index) { i.Packages[0].Type = "platform-app" }, "chartRef"},
		{"failed lane", func(i *trustedsource.Index) { i.Packages[0].Lanes[0].Status = "failed" }, "lanes[0].status"},
		{"lane source", func(i *trustedsource.Index) { i.Packages[0].Lanes[0].SourceSHA = testSHA }, "lanes[0].sourceSha"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := validIndex()
			tc.mutate(i)
			if err := i.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want field %q", err, tc.want)
			}
		})
	}
}

func TestSLSAValidate(t *testing.T) {
	if err := validSLSA().ValidateSLSA(); err != nil {
		t.Fatalf("valid SLSA rejected: %v", err)
	}
	bad := validSLSA()
	bad.Predicate.RunDetails.Builder.ID = "https://github.com/untrusted/workflow@main"
	if err := bad.ValidateSLSA(); err == nil || !strings.Contains(err.Error(), "builder.id") {
		t.Fatalf("error = %v, want builder.id", err)
	}
}

func TestEvidenceValidateAndFreshness(t *testing.T) {
	now := time.Date(2026, 11, 13, 12, 30, 0, 0, time.UTC)
	e := validEvidence()
	e.Compatibility = &trustedsource.Compatibility{Matrix: map[string]string{"stable": "passed", "previous-stable": "passed"}, Result: "passed", CompletedAt: "2026-11-13T12:00:00Z", Browser: "passed", Slots: []string{"cluster-tab"}}
	if err := e.ValidateAt(now, 7*24*time.Hour, true); err != nil {
		t.Fatalf("valid evidence rejected: %v", err)
	}
	e.Compatibility.CompletedAt = "2026-11-06T11:59:59Z"
	if err := e.ValidateAt(now, 7*24*time.Hour, true); err == nil || !strings.Contains(err.Error(), "completedAt") {
		t.Fatalf("error = %v, want stale compatibility", err)
	}
}

func TestEvidenceNegatives(t *testing.T) {
	now := time.Date(2026, 11, 13, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*trustedsource.TestEvidence)
		want   string
	}{
		{"schema", func(e *trustedsource.TestEvidence) { e.Schema = "other" }, "schema"},
		{"tier", func(e *trustedsource.TestEvidence) { e.TrustTier = trustedsource.TrustOfficial }, "trustTier"},
		{"digest", func(e *trustedsource.TestEvidence) { e.Artifact.Digest = "sha256:bad" }, "artifact.digest"},
		{"workflow", func(e *trustedsource.TestEvidence) { e.Producer.WorkflowPath = ".github/workflows/release.yml" }, "producer.workflowPath"},
		{"lane", func(e *trustedsource.TestEvidence) { e.Lanes[0].Status = "skipped" }, "lanes[0].status"},
		{"finished", func(e *trustedsource.TestEvidence) { e.FinishedAt = "2026-11-14T12:00:00Z" }, "finishedAt"},
		{"compat required", func(e *trustedsource.TestEvidence) { e.Compatibility = nil }, "compatibility"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := validEvidence()
			e.Compatibility = &trustedsource.Compatibility{Matrix: map[string]string{"stable": "passed", "previous-stable": "passed"}, Result: "passed", CompletedAt: "2026-11-13T12:00:00Z", Browser: "passed", Slots: []string{"cluster-tab"}}
			tc.mutate(e)
			if err := e.ValidateAt(now, 7*24*time.Hour, true); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want field %q", err, tc.want)
			}
		})
	}
}

func TestSourceFingerprintStable(t *testing.T) {
	a, err := validFile().Fingerprint(validSource())
	if err != nil || len(a) != len("sha256:")+64 {
		t.Fatalf("fingerprint = %q, %v", a, err)
	}
	b, _ := validFile().Fingerprint(validSource())
	if a != b {
		t.Fatalf("fingerprint is not deterministic: %q != %q", a, b)
	}
}
