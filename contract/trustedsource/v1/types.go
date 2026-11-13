// Package trustedsource defines the versioned producer/consumer contract for
// trusted catalog and template sources. It deliberately contains no registry,
// sigstore, database, or HTTP dependencies.
package trustedsource

const (
	APIVersion = "inari.dev/v1"

	CatalogIndexKind  = "CatalogIndex"
	TemplateIndexKind = "TemplateIndex"

	CatalogIndexMediaType   = "application/vnd.inari.catalog.index.v1+yaml"
	CatalogPackageMediaType = "application/vnd.inari.catalog.package.v1+yaml"
	TemplateIndexMediaType  = "application/vnd.inari.template.index.v1+yaml"
	TemplateMediaType       = "application/vnd.inari.template.v1"
	SmokeSuiteMediaType     = "application/vnd.inari.smoke.suite.v1+yaml"
	HelmChartMediaType      = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"

	SLSAProvenancePredicateType = "https://slsa.dev/provenance/v1"
	TestEvidencePredicateType   = "https://inari.dev/attestations/test-evidence/v1"
	TestEvidenceSchema          = "https://inari.dev/schemas/test-evidence-v1.schema.json"
)

type Environment string

const (
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentDevelopment Environment = "development"
)

type TrustTier string

const (
	TrustOfficial     TrustTier = "official"
	TrustReviewed     TrustTier = "reviewed"
	TrustExperimental TrustTier = "experimental"
)

type AllowUnsigned string

const AllowUnsignedDevelopmentOnly AllowUnsigned = "development-only"

type ArtifactRole string

const (
	ArtifactIndex          ArtifactRole = "index"
	ArtifactCatalogPackage ArtifactRole = "catalogPackage"
	ArtifactTemplate       ArtifactRole = "template"
	ArtifactHelmChart      ArtifactRole = "helmChart"
	ArtifactSmokeSuite     ArtifactRole = "smokeSuite"
)

type ArtifactWorkflows struct {
	Index          string `json:"index" yaml:"index"`
	CatalogPackage string `json:"catalogPackage" yaml:"catalogPackage"`
	Template       string `json:"template" yaml:"template"`
	HelmChart      string `json:"helmChart" yaml:"helmChart"`
	SmokeSuite     string `json:"smokeSuite" yaml:"smokeSuite"`
}

func DefaultArtifactWorkflows() ArtifactWorkflows {
	return ArtifactWorkflows{
		Index:          ".github/workflows/release.yml",
		CatalogPackage: ".github/workflows/release-oci.yml",
		Template:       ".github/workflows/release-oci.yml",
		HelmChart:      ".github/workflows/release-chart.yml",
		SmokeSuite:     ".github/workflows/release-smoke-suite.yml",
	}
}

func (w ArtifactWorkflows) For(role ArtifactRole) string {
	switch role {
	case ArtifactIndex:
		return w.Index
	case ArtifactCatalogPackage:
		return w.CatalogPackage
	case ArtifactTemplate:
		return w.Template
	case ArtifactHelmChart:
		return w.HelmChart
	case ArtifactSmokeSuite:
		return w.SmokeSuite
	default:
		return ""
	}
}

type Publisher struct {
	Issuer             string            `json:"issuer" yaml:"issuer"`
	Repository         string            `json:"repository" yaml:"repository"`
	Workflows          ArtifactWorkflows `json:"workflows" yaml:"workflows"`
	WorkflowSHA        string            `json:"workflowSha" yaml:"workflowSha"`
	SourceRepository   string            `json:"sourceRepository" yaml:"sourceRepository"`
	SourceWorkflow     string            `json:"sourceWorkflow" yaml:"sourceWorkflow"`
	SourceRef          string            `json:"sourceRef,omitempty" yaml:"sourceRef,omitempty"`
	Event              string            `json:"event,omitempty" yaml:"event,omitempty"`
	ReleaseEnvironment string            `json:"releaseEnvironment,omitempty" yaml:"releaseEnvironment,omitempty"`
}

type Source struct {
	ID                        string        `json:"id" yaml:"id"`
	IndexRef                  string        `json:"indexRef" yaml:"indexRef"`
	TrustTier                 TrustTier     `json:"trustTier" yaml:"trustTier"`
	ArtifactPrefixes          []string      `json:"artifactPrefixes" yaml:"artifactPrefixes"`
	Publisher                 Publisher     `json:"publisher" yaml:"publisher"`
	AllowUnsigned             AllowUnsigned `json:"allowUnsigned,omitempty" yaml:"allowUnsigned,omitempty"`
	PluginCompatibilityMaxAge string        `json:"pluginCompatibilityMaxAge,omitempty" yaml:"pluginCompatibilityMaxAge,omitempty"`
}

type SourceGroup struct {
	SyncInterval string   `json:"syncInterval,omitempty" yaml:"syncInterval,omitempty"`
	Sources      []Source `json:"sources" yaml:"sources"`
}

type File struct {
	Environment Environment  `json:"environment,omitempty" yaml:"environment,omitempty"`
	Catalog     *SourceGroup `json:"catalog,omitempty" yaml:"catalog,omitempty"`
	Template    *SourceGroup `json:"template,omitempty" yaml:"template,omitempty"`
}

type Producer struct {
	WorkflowPath string `json:"workflowPath" yaml:"workflowPath"`
	WorkflowSHA  string `json:"workflowSha" yaml:"workflowSha"`
	RunID        string `json:"runId" yaml:"runId"`
	RunAttempt   int    `json:"runAttempt" yaml:"runAttempt"`
	RunURI       string `json:"runUri" yaml:"runUri"`
}

type Lane struct {
	Name      string `json:"name" yaml:"name"`
	Required  bool   `json:"required" yaml:"required"`
	Status    string `json:"status" yaml:"status"`
	RunID     string `json:"runId" yaml:"runId"`
	JobID     string `json:"jobId" yaml:"jobId"`
	SourceSHA string `json:"sourceSha" yaml:"sourceSha"`
}

type Compatibility struct {
	Matrix      map[string]string `json:"matrix" yaml:"matrix"`
	Result      string            `json:"result" yaml:"result"`
	CompletedAt string            `json:"completedAt" yaml:"completedAt"`
	Browser     string            `json:"browser" yaml:"browser"`
	Slots       []string          `json:"slots" yaml:"slots"`
}

type RuntimeEvidence struct {
	Target      string `json:"target" yaml:"target"`
	SuiteDigest string `json:"suiteDigest" yaml:"suiteDigest"`
	Profile     string `json:"profile" yaml:"profile"`
	Result      string `json:"result" yaml:"result"`
}

type IndexEntry struct {
	Name           string           `json:"name" yaml:"name"`
	Version        string           `json:"version" yaml:"version"`
	PackageVersion string           `json:"packageVersion" yaml:"packageVersion"`
	Channel        string           `json:"channel" yaml:"channel"`
	Type           string           `json:"type" yaml:"type"`
	Description    string           `json:"description" yaml:"description"`
	Category       string           `json:"category,omitempty" yaml:"category,omitempty"`
	OCIRef         string           `json:"ociRef" yaml:"ociRef"`
	Digest         string           `json:"digest" yaml:"digest"`
	SourceSHA      string           `json:"sourceSha" yaml:"sourceSha"`
	ReleaseTag     string           `json:"releaseTag" yaml:"releaseTag"`
	MediaType      string           `json:"mediaType" yaml:"mediaType"`
	ChartRef       string           `json:"chartRef,omitempty" yaml:"chartRef,omitempty"`
	ChartDigest    string           `json:"chartDigest,omitempty" yaml:"chartDigest,omitempty"`
	Lanes          []Lane           `json:"lanes" yaml:"lanes"`
	Compatibility  *Compatibility   `json:"compatibility,omitempty" yaml:"compatibility,omitempty"`
	Runtime        *RuntimeEvidence `json:"runtime,omitempty" yaml:"runtime,omitempty"`
}

type Index struct {
	APIVersion string       `json:"apiVersion" yaml:"apiVersion"`
	Kind       string       `json:"kind" yaml:"kind"`
	MediaType  string       `json:"mediaType" yaml:"mediaType"`
	SourceSHA  string       `json:"sourceSha" yaml:"sourceSha"`
	SourceRef  string       `json:"sourceRef" yaml:"sourceRef"`
	Producer   Producer     `json:"producer" yaml:"producer"`
	Packages   []IndexEntry `json:"packages,omitempty" yaml:"packages,omitempty"`
	Templates  []IndexEntry `json:"templates,omitempty" yaml:"templates,omitempty"`
}

type Subject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

type Statement struct {
	Type          string     `json:"_type"`
	Subject       []Subject  `json:"subject"`
	PredicateType string     `json:"predicateType"`
	Predicate     Provenance `json:"predicate"`
}

type Provenance struct {
	BuildDefinition BuildDefinition `json:"buildDefinition"`
	RunDetails      RunDetails      `json:"runDetails"`
}

type BuildDefinition struct {
	BuildType            string             `json:"buildType"`
	ExternalParameters   ExternalParameters `json:"externalParameters"`
	ResolvedDependencies []Dependency       `json:"resolvedDependencies"`
}

type ExternalParameters struct {
	Repository     string `json:"repository"`
	SourceSHA      string `json:"sourceSha"`
	ReleaseTag     string `json:"releaseTag"`
	PackageVersion string `json:"packageVersion"`
	Event          string `json:"event,omitempty"`
	SourceRef      string `json:"sourceRef,omitempty"`
	Environment    string `json:"environment,omitempty"`
	RunID          string `json:"runId"`
	RunAttempt     int    `json:"runAttempt"`
}

type Dependency struct {
	URI    string            `json:"uri"`
	Digest map[string]string `json:"digest"`
}

type RunDetails struct {
	Builder  Builder  `json:"builder"`
	Metadata Metadata `json:"metadata"`
}

type Builder struct {
	ID string `json:"id"`
}

type Metadata struct {
	InvocationID string `json:"invocationId"`
}

type ArtifactEvidence struct {
	Repository string `json:"repository" yaml:"repository"`
	Digest     string `json:"digest" yaml:"digest"`
	MediaType  string `json:"mediaType" yaml:"mediaType"`
}

type TestEvidence struct {
	Schema           string           `json:"schema" yaml:"schema"`
	App              string           `json:"app" yaml:"app"`
	Component        string           `json:"component" yaml:"component"`
	TrustTier        TrustTier        `json:"trustTier" yaml:"trustTier"`
	SourceRepository string           `json:"sourceRepository" yaml:"sourceRepository"`
	SourceSHA        string           `json:"sourceSha" yaml:"sourceSha"`
	ReleaseTag       string           `json:"releaseTag" yaml:"releaseTag"`
	Artifact         ArtifactEvidence `json:"artifact" yaml:"artifact"`
	Producer         Producer         `json:"producer" yaml:"producer"`
	Lanes            []Lane           `json:"lanes" yaml:"lanes"`
	Compatibility    *Compatibility   `json:"compatibility,omitempty" yaml:"compatibility,omitempty"`
	Runtime          *RuntimeEvidence `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	FinishedAt       string           `json:"finishedAt" yaml:"finishedAt"`
	EvidenceDigest   string           `json:"evidenceDigest" yaml:"evidenceDigest"`
}
