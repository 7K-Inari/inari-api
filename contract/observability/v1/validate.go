package observability

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

var (
	appPattern          = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	digestPattern       = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	datasourcePattern   = regexp.MustCompile(`^\$\{DS_[A-Z0-9_]+\}$`)
	linkVariablePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	slugPattern         = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

var supportedLinkVariables = map[string]bool{
	"namespace":  true,
	"instance":   true,
	"app":        true,
	"org":        true,
	"cluster":    true,
	"datasource": true,
}

func field(path, format string, args ...any) error {
	return fmt.Errorf("observability: %s: %s", path, fmt.Sprintf(format, args...))
}

func validAssetPath(value string) bool {
	return value != "" && value != path.Clean(value) == false && !strings.HasPrefix(value, "/") && !strings.Contains(value, "\\") && path.Clean(value) != ".." && !strings.HasPrefix(path.Clean(value), "../")
}

func parseRange(value, fieldName string) (*semver.Constraints, error) {
	if value == "" || strings.EqualFold(strings.TrimSpace(value), "latest") {
		return nil, field(fieldName, "an explicit SemVer range is required; latest is forbidden")
	}
	r, err := semver.NewConstraint(value)
	if err != nil {
		return nil, field(fieldName, "invalid SemVer range %q", value)
	}
	return r, nil
}

func (c *Capability) Validate() error {
	if c == nil {
		return field("capability", "is required")
	}
	if c.APIVersion != APIVersion {
		return field("apiVersion", "must be %q", APIVersion)
	}
	if c.Kind != CapabilityKind {
		return field("kind", "must be %q", CapabilityKind)
	}
	if !appPattern.MatchString(c.App) {
		return field("app", "must be a lowercase application ID")
	}
	if !validAssetPath(c.Folder) || strings.HasSuffix(c.Folder, "/") {
		return field("folder", "must be a safe relative path")
	}
	// The merged Grafana resolver and generated pack v1/v2 selection currently
	// provide one primary datasource. Do not silently provision a partial map.
	if len(c.Datasources) != 1 {
		return field("datasources", "exactly one primary datasource mapping is supported by the current render contract")
	}
	for name, variable := range c.Datasources {
		if !appPattern.MatchString(name) || !datasourcePattern.MatchString(variable) {
			return field("datasources."+name, "must contain a ${DS_*} datasource variable")
		}
	}
	seenVariables := map[string]bool{}
	for i, name := range c.LinkVariables {
		if !linkVariablePattern.MatchString(name) || !supportedLinkVariables[name] {
			return field(fmt.Sprintf("linkVariables[%d]", i), "unsupported link variable %q", name)
		}
		if seenVariables[name] {
			return field(fmt.Sprintf("linkVariables[%d]", i), "duplicate link variable %q", name)
		}
		seenVariables[name] = true
	}
	if len(c.Dashboards) == 0 {
		return field("dashboards", "at least one dashboard is required")
	}
	seenUIDs := map[string]bool{}
	for i, dashboard := range c.Dashboards {
		prefix := fmt.Sprintf("dashboards[%d]", i)
		if strings.TrimSpace(dashboard.Title) == "" || !strings.HasPrefix(dashboard.UID, c.App+"-") {
			return field(prefix+".uid", "dashboard title and app-prefixed uid are required")
		}
		if seenUIDs[dashboard.UID] {
			return field(prefix+".uid", "duplicate dashboard uid %q", dashboard.UID)
		}
		seenUIDs[dashboard.UID] = true
		if !validAssetPath(dashboard.File) {
			return field(prefix+".file", "unsafe asset path")
		}
	}
	if len(c.Rules) == 0 {
		return field("rules", "at least one rule file is required")
	}
	for i, rule := range c.Rules {
		if !validAssetPath(rule.File) {
			return field(fmt.Sprintf("rules[%d].file", i), "unsafe asset path")
		}
		if rule.Evaluation != EvaluationCentral && rule.Evaluation != EvaluationLocal {
			return field(fmt.Sprintf("rules[%d].evaluation", i), "must be central or local")
		}
	}
	if len(c.Targets) == 0 {
		return field("targets", "at least one target evaluation is required")
	}
	seenTargets := map[Evaluation]bool{}
	for _, target := range c.Targets {
		if target != EvaluationCentral && target != EvaluationLocal {
			return field("targets", "unsupported target %q", target)
		}
		if seenTargets[target] {
			return field("targets", "duplicate target %q", target)
		}
		seenTargets[target] = true
	}
	if len(c.SeverityRouting) == 0 {
		return field("severityRouting", "at least one severity mapping is required")
	}
	if _, err := parseRange(c.MinimumAppVersion, "minimumAppVersion"); err != nil {
		return err
	}
	if !appPattern.MatchString(c.ObservabilityPack.Package) || strings.Contains(c.ObservabilityPack.Package, "/") {
		return field("observabilityPack.package", "must be a package name within the application source")
	}
	if _, err := parseRange(c.ObservabilityPack.VersionRange, "observabilityPack.versionRange"); err != nil {
		return err
	}
	return nil
}

func (c *Capability) ValidateAppVersion(version string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	rng, _ := parseRange(c.MinimumAppVersion, "minimumAppVersion")
	v, err := semver.NewVersion(version)
	if err != nil {
		return field("app.version", "invalid SemVer %q", version)
	}
	if !rng.Check(v) {
		return field("app.version", "version %s is outside required range %q", version, c.MinimumAppVersion)
	}
	return nil
}

func (c *Capability) ValidatePackVersion(version string) error {
	if err := c.Validate(); err != nil {
		return err
	}
	rng, _ := parseRange(c.ObservabilityPack.VersionRange, "observabilityPack.versionRange")
	v, err := semver.NewVersion(version)
	if err != nil {
		return field("pack.version", "invalid SemVer %q", version)
	}
	if !rng.Check(v) {
		return field("pack.version", "version %s is outside required range %q", version, c.ObservabilityPack.VersionRange)
	}
	return nil
}

// Digest binds downstream render/readiness state to the normalized signed
// capability. encoding/json orders map keys, making the digest deterministic.
func (c *Capability) Digest() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("observability: canonical capability: %w", err)
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(raw)), nil
}

func BindLinkVariables(declared []string, values map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(declared))
	seen := map[string]bool{}
	for _, name := range declared {
		if !linkVariablePattern.MatchString(name) || !supportedLinkVariables[name] || seen[name] {
			return nil, field("linkVariables", "unsupported or duplicate variable %q", name)
		}
		seen[name] = true
		value := strings.TrimSpace(values[name])
		if value == "" {
			return nil, field("linkVariables."+name, "no server-owned binding is available")
		}
		out[name] = value
	}
	return out, nil
}

func validVerificationProfile(profile string) bool {
	return profile == ProfileCI || profile == ProfileStage || profile == ProfileProd
}

func validateIdentity(r *VerificationResult) error {
	if r == nil {
		return field("verification", "is required")
	}
	if r.APIVersion != APIVersion || r.Kind != VerificationKind {
		return field("verification", "unexpected apiVersion/kind")
	}
	if r.Companion.ID == "" || r.Companion.Generation <= 0 || r.Target.ID == "" || r.Target.Revision <= 0 || r.Target.ClusterID == "" {
		return field("verification.identity", "companion generation and target identity are required")
	}
	if r.App.PlanID == "" || r.App.Revision == "" || (r.App.InstanceID == "" && r.App.ScaffoldRunID == "") {
		return field("app", "plan, revision, and instance or scaffold identity are required")
	}
	for name, digest := range map[string]string{"app.sourceDigest": r.App.SourceDigest, "pack.digest": r.Pack.Digest, "capabilityDigest": r.CapabilityDigest, "suite.digest": r.Suite.Digest} {
		if !digestPattern.MatchString(digest) {
			return field(name, "an immutable sha256 digest is required")
		}
	}
	if r.Pack.ItemID == "" || r.Pack.Version == "" || !validVerificationProfile(r.Suite.Profile) || r.Run.ID == "" || r.Run.Attempt <= 0 {
		return field("verification.identity", "pack, generic ci/stage/prod suite profile, and run identity are required")
	}
	if r.Result != "passed" && r.Result != "failed" {
		return field("result", "must be passed or failed")
	}
	return nil
}

func (r *VerificationResult) ValidateAt(now time.Time) error {
	if err := validateIdentity(r); err != nil {
		return err
	}
	if now.IsZero() || r.Run.FinishedAt.IsZero() || r.Run.FinishedAt.After(now.Add(5*time.Minute)) {
		return field("run.finishedAt", "invalid or future completion time")
	}
	if !r.ExpiresAt.After(r.Run.FinishedAt) {
		return field("expiresAt", "must be after the run completion")
	}
	if r.Result == "failed" {
		return nil
	}
	if !now.Before(r.ExpiresAt) {
		return field("expiresAt", "passing evidence is expired")
	}
	if r.Checks.Folder.State != "passed" || r.Checks.Folder.Name == "" || r.Checks.Folder.UID == "" {
		return field("checks.folder", "folder evidence must pass")
	}
	if len(r.Checks.Dashboards) == 0 || len(r.Checks.Datasources) == 0 || len(r.Checks.Rules) == 0 {
		return field("checks", "dashboard, datasource, and rule evidence is required")
	}
	for _, dashboard := range r.Checks.Dashboards {
		if dashboard.State != "passed" || dashboard.UID == "" || dashboard.Title == "" || !slugPattern.MatchString(dashboard.Slug) {
			return field("checks.dashboards", "every dashboard must pass with a safe UID/title/slug")
		}
	}
	for _, datasource := range r.Checks.Datasources {
		if datasource.State != "passed" || datasource.Logical == "" || datasource.Name == "" || datasource.Type == "" {
			return field("checks.datasources", "every datasource must pass with logical/name/type evidence")
		}
	}
	for _, rule := range r.Checks.Rules {
		if rule.State != "passed" || !validAssetPath(rule.File) || (rule.Evaluation != EvaluationCentral && rule.Evaluation != EvaluationLocal) {
			return field("checks.rules", "every selected rule must pass with file/evaluation evidence")
		}
	}
	return nil
}
