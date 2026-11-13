package trustedsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

var (
	sourceIDPattern   = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	shaPattern        = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digestPattern     = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	workflowPattern   = regexp.MustCompile(`^\.github/workflows/[A-Za-z0-9_.-]+\.yml$`)
)

func field(path, format string, args ...any) error {
	return fmt.Errorf("trustedsource: %s: %s", path, fmt.Sprintf(format, args...))
}

// DecodeFile strictly decodes one YAML source document. Unknown fields,
// duplicate mapping keys, multiple documents, and null/empty documents fail.
func DecodeFile(data []byte) (*File, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, field("file", "document is empty")
	}
	var root yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		return nil, field("file", "invalid YAML: %v", err)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, field("file", "expected one mapping document")
	}
	if err := rejectDuplicateYAMLKeys(root.Content[0], "file"); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, field("file", "multiple documents are not allowed")
		}
		return nil, field("file", "invalid YAML: %v", err)
	}
	dec = yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f File
	if err := dec.Decode(&f); err != nil {
		return nil, field("file", "invalid source configuration: %v", err)
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return &f, nil
}

func rejectDuplicateYAMLKeys(n *yaml.Node, path string) error {
	if n.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			key := n.Content[i].Value
			if seen[key] {
				return field(path, "duplicate key %q", key)
			}
			seen[key] = true
			if err := rejectDuplicateYAMLKeys(n.Content[i+1], path+"."+key); err != nil {
				return err
			}
		}
	}
	if n.Kind == yaml.SequenceNode || n.Kind == yaml.DocumentNode {
		for i, child := range n.Content {
			if err := rejectDuplicateYAMLKeys(child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// DecodeIndex strictly decodes and validates one CatalogIndex or TemplateIndex YAML document.
func DecodeIndex(data []byte) (*Index, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, field("index", "document is empty")
	}
	var root yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		return nil, field("index", "invalid YAML: %v", err)
	}
	if len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, field("index", "expected one mapping document")
	}
	if err := rejectDuplicateYAMLKeys(root.Content[0], "index"); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, field("index", "multiple documents are not allowed")
		}
		return nil, field("index", "invalid YAML: %v", err)
	}
	dec = yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var idx Index
	if err := dec.Decode(&idx); err != nil {
		return nil, field("index", "invalid index: %v", err)
	}
	if err := idx.Validate(); err != nil {
		return nil, err
	}
	return &idx, nil
}

// DecodeTestEvidence strictly decodes a test-evidence predicate. Freshness
// and compatibility policy are checked separately by ValidateAt.
func DecodeTestEvidence(data []byte) (*TestEvidence, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var evidence TestEvidence
	if err := dec.Decode(&evidence); err != nil {
		return nil, field("evidence", "invalid JSON: %v", err)
	}
	if dec.More() {
		return nil, field("evidence", "multiple JSON values are not allowed")
	}
	return &evidence, nil
}

// DecodeStatement strictly decodes and validates an in-toto SLSA statement.
func DecodeStatement(data []byte) (*Statement, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var statement Statement
	if err := dec.Decode(&statement); err != nil {
		return nil, field("statement", "invalid JSON: %v", err)
	}
	if dec.More() {
		return nil, field("statement", "multiple JSON values are not allowed")
	}
	if err := statement.ValidateSLSA(); err != nil {
		return nil, err
	}
	return &statement, nil
}

func (f *File) Validate() error {
	if f == nil {
		return field("file", "is required")
	}
	switch f.Environment {
	case "", EnvironmentProduction, EnvironmentStaging, EnvironmentDevelopment:
	default:
		return field("environment", "must be production, staging, or development")
	}
	ids := map[string]string{}
	if err := validateGroup(f.Catalog, "catalog", ids); err != nil {
		return err
	}
	return validateGroup(f.Template, "template", ids)
}

func validateGroup(g *SourceGroup, path string, ids map[string]string) error {
	if g == nil {
		return nil
	}
	if g.SyncInterval != "" {
		d, err := time.ParseDuration(g.SyncInterval)
		if err != nil || d <= 0 {
			return field(path+".syncInterval", "must be a positive duration")
		}
	}
	prefixOwners := map[string]string{}
	for i := range g.Sources {
		s := &g.Sources[i]
		sp := fmt.Sprintf("%s.sources[%d]", path, i)
		if err := s.validate(sp); err != nil {
			return err
		}
		if previous, found := ids[s.ID]; found {
			return field(sp+".id", "duplicate source ID %q (already used by %s)", s.ID, previous)
		}
		ids[s.ID] = sp
		for _, prefix := range s.ArtifactPrefixes {
			if previous, found := prefixOwners[prefix]; found {
				return field(sp+".artifactPrefixes", "prefix %q is already owned by %s", prefix, previous)
			}
			prefixOwners[prefix] = s.ID
		}
	}
	return nil
}

func (s *Source) validate(path string) error {
	if len(s.ID) == 0 || len(s.ID) > 40 || !sourceIDPattern.MatchString(s.ID) {
		return field(path+".id", "must be a lowercase RFC1123 label of at most 40 characters")
	}
	if err := validateOCIRef(s.IndexRef, true); err != nil {
		return field(path+".indexRef", "%v", err)
	}
	switch s.TrustTier {
	case TrustOfficial, TrustReviewed, TrustExperimental:
	default:
		return field(path+".trustTier", "must be official, reviewed, or experimental")
	}
	if s.AllowUnsigned != "" {
		if s.AllowUnsigned != AllowUnsignedDevelopmentOnly {
			return field(path+".allowUnsigned", "must be development-only when set")
		}
		if s.TrustTier != TrustExperimental {
			return field(path+".allowUnsigned", "is permitted only for experimental sources")
		}
	}
	if len(s.ArtifactPrefixes) == 0 {
		return field(path+".artifactPrefixes", "at least one prefix is required")
	}
	for i, prefix := range s.ArtifactPrefixes {
		if !strings.HasSuffix(prefix, "/") || strings.Contains(prefix, "://") || strings.Contains(prefix, "..") || strings.ContainsAny(prefix, " \t\r\n") {
			return field(fmt.Sprintf("%s.artifactPrefixes[%d]", path, i), "must be a normalized OCI repository prefix ending in /")
		}
	}
	if err := s.Publisher.validate(path + ".publisher"); err != nil {
		return err
	}
	if _, err := s.PluginCompatibilityAge(); err != nil {
		return field(path+".pluginCompatibilityMaxAge", "%v", err)
	}
	return nil
}

func (p *Publisher) validate(path string) error {
	if p.Issuer != "https://token.actions.githubusercontent.com" {
		return field(path+".issuer", "must be GitHub Actions OIDC issuer")
	}
	if !repositoryPattern.MatchString(p.Repository) {
		return field(path+".repository", "must be owner/repository")
	}
	if !repositoryPattern.MatchString(p.SourceRepository) {
		return field(path+".sourceRepository", "must be owner/repository")
	}
	if !workflowPattern.MatchString(p.SourceWorkflow) {
		return field(path+".sourceWorkflow", "must be a .github/workflows/*.yml path")
	}
	if !shaPattern.MatchString(p.WorkflowSHA) {
		return field(path+".workflowSha", "must be a full lowercase 40-hex commit SHA")
	}
	expected := DefaultArtifactWorkflows()
	actual := []struct {
		name string
		got  string
		want string
	}{
		{"index", p.Workflows.Index, expected.Index},
		{"catalogPackage", p.Workflows.CatalogPackage, expected.CatalogPackage},
		{"template", p.Workflows.Template, expected.Template},
		{"helmChart", p.Workflows.HelmChart, expected.HelmChart},
		{"smokeSuite", p.Workflows.SmokeSuite, expected.SmokeSuite},
	}
	for _, w := range actual {
		if w.got != w.want {
			return field(path+".workflows."+w.name, "must be %s", w.want)
		}
	}
	for name, value := range map[string]string{"sourceRef": p.SourceRef, "event": p.Event, "releaseEnvironment": p.ReleaseEnvironment} {
		if strings.ContainsAny(value, " \t\r\n") {
			return field(path+"."+name, "must not contain whitespace")
		}
	}
	return nil
}

func validateOCIRef(ref string, allowLatest bool) error {
	if ref == "" || strings.Contains(ref, "://") || strings.ContainsAny(ref, " \t\r\n") || !strings.Contains(ref, "/") || strings.Contains(ref, "..") {
		return fmt.Errorf("must be a valid OCI reference")
	}
	if strings.HasSuffix(ref, ":latest") && !allowLatest {
		return fmt.Errorf("latest is not a deployable artifact reference")
	}
	if strings.Contains(ref, "@") && !digestPattern.MatchString(ref[strings.LastIndex(ref, "@")+1:]) {
		return fmt.Errorf("contains an invalid digest")
	}
	return nil
}

func (s Source) PluginCompatibilityAge() (time.Duration, error) {
	if s.PluginCompatibilityMaxAge == "" {
		return 7 * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(s.PluginCompatibilityMaxAge)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("must be a positive duration")
	}
	return d, nil
}

func (f *File) Fingerprint(s Source) (string, error) {
	if err := s.validate("source"); err != nil {
		return "", err
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (i *Index) Validate() error {
	if i == nil {
		return field("index", "is required")
	}
	if i.APIVersion != APIVersion {
		return field("apiVersion", "must be %s", APIVersion)
	}
	var entries []IndexEntry
	var entryPath, mediaType string
	switch i.Kind {
	case CatalogIndexKind:
		entries, entryPath, mediaType = i.Packages, "packages", CatalogIndexMediaType
	case TemplateIndexKind:
		entries, entryPath, mediaType = i.Templates, "templates", TemplateIndexMediaType
	default:
		return field("kind", "must be CatalogIndex or TemplateIndex")
	}
	if i.MediaType != mediaType {
		return field("mediaType", "must be %s", mediaType)
	}
	if !shaPattern.MatchString(i.SourceSHA) {
		return field("sourceSha", "must be a full lowercase 40-hex source SHA")
	}
	if i.SourceRef == "" {
		return field("sourceRef", "is required")
	}
	if err := validateProducer(i.Producer, i.Kind, "producer"); err != nil {
		return err
	}
	seen := map[string]bool{}
	for n := range entries {
		e := &entries[n]
		path := fmt.Sprintf("%s[%d]", entryPath, n)
		if err := e.validate(path, i.Kind); err != nil {
			return err
		}
		key := e.Name + "@" + e.Version
		if seen[key] {
			return field(path, "duplicate package version %q", key)
		}
		seen[key] = true
	}
	return nil
}

func validateProducer(p Producer, kind, path string) error {
	expected := DefaultArtifactWorkflows()
	workflow := expected.Index
	switch kind {
	case string(ArtifactCatalogPackage):
		workflow = expected.CatalogPackage
	case string(ArtifactTemplate):
		workflow = expected.Template
	case string(ArtifactHelmChart):
		workflow = expected.HelmChart
	case string(ArtifactSmokeSuite):
		workflow = expected.SmokeSuite
	}
	if p.WorkflowPath != workflow {
		return field(path+".workflowPath", "must be %s", workflow)
	}
	if !shaPattern.MatchString(p.WorkflowSHA) {
		return field(path+".workflowSha", "must be a full lowercase 40-hex workflow SHA")
	}
	if p.RunID == "" {
		return field(path+".runId", "is required")
	}
	if p.RunAttempt < 1 {
		return field(path+".runAttempt", "must be at least 1")
	}
	if !strings.HasPrefix(p.RunURI, "https://github.com/") {
		return field(path+".runUri", "must be a GitHub workflow run URL")
	}
	return nil
}

func (e *IndexEntry) validate(path, kind string) error {
	if !sourceIDPattern.MatchString(e.Name) || len(e.Name) > 63 {
		return field(path+".name", "must be a lowercase RFC1123 package name")
	}
	for fieldName, value := range map[string]string{"version": e.Version, "packageVersion": e.PackageVersion, "releaseTag": e.ReleaseTag} {
		if value == "" || strings.ContainsAny(value, " \t\r\n") {
			return field(path+"."+fieldName, "is required and must not contain whitespace")
		}
	}
	switch e.Channel {
	case "stable", "incubating", "edge":
	default:
		return field(path+".channel", "must be stable, incubating, or edge")
	}
	switch e.Type {
	case "kro-rgd", "platform-app", "policy-pack", "crossplane-claim", "manifests", "kustomize", "template":
	default:
		return field(path+".type", "unsupported package type %q", e.Type)
	}
	if e.Description == "" {
		return field(path+".description", "is required")
	}
	if err := validateOCIRef(e.OCIRef, false); err != nil {
		return field(path+".ociRef", "%v", err)
	}
	if !digestPattern.MatchString(e.Digest) {
		return field(path+".digest", "must be sha256 followed by 64 lowercase hex characters")
	}
	if !shaPattern.MatchString(e.SourceSHA) {
		return field(path+".sourceSha", "must be a full lowercase 40-hex source SHA")
	}
	expectedMedia := CatalogPackageMediaType
	if kind == TemplateIndexKind {
		expectedMedia = TemplateMediaType
	}
	if e.MediaType != expectedMedia {
		return field(path+".mediaType", "must be %s", expectedMedia)
	}
	role := string(ArtifactCatalogPackage)
	if kind == TemplateIndexKind {
		role = string(ArtifactTemplate)
	}
	if err := validateProducer(e.Producer, role, path+".producer"); err != nil {
		return err
	}
	if kind == TemplateIndexKind && e.Type != "template" {
		return field(path+".type", "must be template for a TemplateIndex")
	}
	if kind == CatalogIndexKind && e.Type == "template" {
		return field(path+".type", "template is not valid for a CatalogIndex")
	}
	if e.Type == "platform-app" {
		chartRef := strings.TrimPrefix(e.ChartRef, "oci://")
		if e.ChartRef == "" || chartRef == e.ChartRef {
			return field(path+".chartRef", "platform-app requires an oci:// chart reference")
		}
		if err := validateOCIRef(chartRef, false); err != nil {
			return field(path+".chartRef", "%v", err)
		}
		if !digestPattern.MatchString(e.ChartDigest) {
			return field(path+".chartDigest", "platform-app requires a sha256 chart digest")
		}
		if e.ChartProducer == nil {
			return field(path+".chartProducer", "platform-app requires the chart release producer")
		}
		if err := validateProducer(*e.ChartProducer, string(ArtifactHelmChart), path+".chartProducer"); err != nil {
			return err
		}
	} else if e.ChartRef != "" || e.ChartDigest != "" || e.ChartProducer != nil {
		return field(path, "chartRef/chartDigest are valid only for platform-app entries")
	}
	if len(e.Lanes) == 0 {
		return field(path+".lanes", "at least one lane is required")
	}
	for n := range e.Lanes {
		if err := e.Lanes[n].validate(fmt.Sprintf("%s.lanes[%d]", path, n), e.SourceSHA); err != nil {
			return err
		}
	}
	if e.Compatibility != nil {
		if err := e.Compatibility.validate(path + ".compatibility"); err != nil {
			return err
		}
	}
	if e.Runtime != nil {
		if err := e.Runtime.validate(path + ".runtime"); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lane) validate(path, sourceSHA string) error {
	if l.Name == "" {
		return field(path+".name", "is required")
	}
	switch l.Status {
	case "passed", "failed", "skipped", "pending":
	default:
		return field(path+".status", "must be passed, failed, skipped, or pending")
	}
	if l.Required && l.Status != "passed" {
		return field(path+".status", "required lane must be passed")
	}
	if l.RunID == "" || l.JobID == "" {
		return field(path, "runId and jobId are required")
	}
	if l.SourceSHA != sourceSHA {
		return field(path+".sourceSha", "must match the artifact source SHA")
	}
	return nil
}

func (c *Compatibility) validate(path string) error {
	if c.Result != "passed" || c.Browser != "passed" {
		return field(path, "result and browser must be passed")
	}
	if len(c.Matrix) < 2 || c.Matrix["stable"] != "passed" || c.Matrix["previous-stable"] != "passed" {
		return field(path+".matrix", "stable and previous-stable results must be passed")
	}
	if len(c.Slots) == 0 {
		return field(path+".slots", "at least one declared slot is required")
	}
	if _, err := time.Parse(time.RFC3339, c.CompletedAt); err != nil {
		return field(path+".completedAt", "must be RFC3339")
	}
	return nil
}

func (r *RuntimeEvidence) validate(path string) error {
	if r.Target == "" || r.Profile == "" {
		return field(path, "target and profile are required")
	}
	if !digestPattern.MatchString(r.SuiteDigest) {
		return field(path+".suiteDigest", "must be a sha256 digest")
	}
	if r.Result != "passed" {
		return field(path+".result", "must be passed")
	}
	return nil
}

func (s *Statement) ValidateSLSA() error {
	if s == nil {
		return field("statement", "is required")
	}
	if s.Type != "https://in-toto.io/Statement/v1" {
		return field("statement._type", "must be in-toto Statement v1")
	}
	if s.PredicateType != SLSAProvenancePredicateType {
		return field("statement.predicateType", "must be %s", SLSAProvenancePredicateType)
	}
	if len(s.Subject) == 0 {
		return field("statement.subject", "at least one subject is required")
	}
	for n, subject := range s.Subject {
		if !strings.Contains(subject.Name, "/") || !digestPattern.MatchString("sha256:"+subject.Digest["sha256"]) {
			return field(fmt.Sprintf("statement.subject[%d]", n), "must contain an OCI repository and sha256 digest")
		}
	}
	bd := s.Predicate.BuildDefinition
	if bd.BuildType != "https://actions.github.io/buildtypes/workflow/v1" {
		return field("predicate.buildDefinition.buildType", "unsupported build type")
	}
	ep := bd.ExternalParameters
	if !repositoryPattern.MatchString(ep.Repository) {
		return field("predicate.buildDefinition.externalParameters.repository", "must be owner/repository")
	}
	if !shaPattern.MatchString(ep.SourceSHA) {
		return field("predicate.buildDefinition.externalParameters.sourceSha", "must be a full source SHA")
	}
	if ep.ReleaseTag == "" || ep.PackageVersion == "" || ep.RunID == "" || ep.RunAttempt < 1 {
		return field("predicate.buildDefinition.externalParameters", "releaseTag, packageVersion, runId, and runAttempt are required")
	}
	if len(bd.ResolvedDependencies) == 0 {
		return field("predicate.buildDefinition.resolvedDependencies", "at least one immutable dependency is required")
	}
	for n, dep := range bd.ResolvedDependencies {
		if !strings.HasPrefix(dep.URI, "https://github.com/") || !shaPattern.MatchString(dep.Digest["sha"]) {
			return field(fmt.Sprintf("predicate.buildDefinition.resolvedDependencies[%d]", n), "must be a GitHub URI with a full SHA digest")
		}
	}
	builder := s.Predicate.RunDetails.Builder.ID
	if !strings.HasPrefix(builder, "https://github.com/") || !strings.Contains(builder, "@") || !shaPattern.MatchString(builder[strings.LastIndex(builder, "@")+1:]) {
		return field("predicate.runDetails.builder.id", "must identify a workflow at an immutable SHA")
	}
	if !strings.HasPrefix(s.Predicate.RunDetails.Metadata.InvocationID, "https://github.com/") {
		return field("predicate.runDetails.metadata.invocationId", "must be a GitHub workflow run URL")
	}
	return nil
}

func (e *TestEvidence) ValidateAt(now time.Time, maxCompatibilityAge time.Duration, requireCompatibility bool) error {
	if e == nil {
		return field("evidence", "is required")
	}
	if e.Schema != TestEvidenceSchema {
		return field("schema", "must be %s", TestEvidenceSchema)
	}
	if e.App == "" || e.Component == "" {
		return field("app", "app and component are required")
	}
	switch e.TrustTier {
	case TrustOfficial, TrustReviewed, TrustExperimental:
	default:
		return field("trustTier", "must be official, reviewed, or experimental")
	}
	if requireCompatibility && e.TrustTier != TrustReviewed {
		return field("trustTier", "UI compatibility evidence is versioned for reviewed artifacts")
	}
	if !repositoryPattern.MatchString(e.SourceRepository) {
		return field("sourceRepository", "must be owner/repository")
	}
	if !shaPattern.MatchString(e.SourceSHA) {
		return field("sourceSha", "must be a full lowercase 40-hex source SHA")
	}
	if e.ReleaseTag == "" {
		return field("releaseTag", "is required")
	}
	if !strings.Contains(e.Artifact.Repository, "/") || strings.Contains(e.Artifact.Repository, "://") {
		return field("artifact.repository", "must be an OCI repository")
	}
	if !digestPattern.MatchString(e.Artifact.Digest) {
		return field("artifact.digest", "must be a sha256 digest")
	}
	if e.Artifact.MediaType == "" {
		return field("artifact.mediaType", "is required")
	}
	role := string(ArtifactCatalogPackage)
	switch e.Artifact.MediaType {
	case TemplateMediaType:
		role = string(ArtifactTemplate)
	case HelmChartMediaType:
		role = string(ArtifactHelmChart)
	case SmokeSuiteMediaType:
		role = string(ArtifactSmokeSuite)
	}
	if err := validateProducer(e.Producer, role, "producer"); err != nil {
		return err
	}
	if len(e.Lanes) == 0 {
		return field("lanes", "at least one lane is required")
	}
	for n := range e.Lanes {
		if err := e.Lanes[n].validate(fmt.Sprintf("lanes[%d]", n), e.SourceSHA); err != nil {
			return err
		}
	}
	if requireCompatibility {
		if e.Compatibility == nil {
			return field("compatibility", "is required for reviewed UI evidence")
		}
		if err := e.Compatibility.validate("compatibility"); err != nil {
			return err
		}
		completed, _ := time.Parse(time.RFC3339, e.Compatibility.CompletedAt)
		if completed.After(now.Add(5 * time.Minute)) {
			return field("compatibility.completedAt", "must not be in the future")
		}
		if maxCompatibilityAge <= 0 {
			return field("compatibility.completedAt", "freshness duration must be positive")
		}
		if completed.Before(now.Add(-maxCompatibilityAge)) {
			return field("compatibility.completedAt", "reviewed UI compatibility evidence is stale")
		}
	}
	if e.Runtime != nil {
		if err := e.Runtime.validate("runtime"); err != nil {
			return err
		}
	}
	finished, err := time.Parse(time.RFC3339, e.FinishedAt)
	if err != nil {
		return field("finishedAt", "must be RFC3339")
	}
	if finished.After(now.Add(5 * time.Minute)) {
		return field("finishedAt", "must not be in the future")
	}
	if !digestPattern.MatchString(e.EvidenceDigest) {
		return field("evidenceDigest", "must be a sha256 digest")
	}
	return nil
}

// SortedSourceIDs is a deterministic helper for status and sync responses.
func (f *File) SortedSourceIDs() []string {
	if f == nil {
		return nil
	}
	seen := map[string]bool{}
	for _, group := range []*SourceGroup{f.Catalog, f.Template} {
		if group == nil {
			continue
		}
		for _, source := range group.Sources {
			seen[source.ID] = true
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
