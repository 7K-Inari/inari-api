package observability_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	observability "github.com/7K-Inari/inari-api/contract/observability/v1"
)

const testDigest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func validCapability() *observability.Capability {
	return &observability.Capability{
		APIVersion:        observability.APIVersion,
		Kind:              observability.CapabilityKind,
		App:               "payments-api",
		Folder:            "payments-api",
		Datasources:       map[string]string{"prometheus": "${DS_PROMETHEUS}"},
		LinkVariables:     []string{"namespace"},
		Dashboards:        []observability.Dashboard{{Title: "Service overview", UID: "payments-api-overview", File: "dashboards/overview.json"}},
		Rules:             []observability.Rule{{File: "rules/availability.yaml", Evaluation: observability.EvaluationCentral}},
		Targets:           []observability.Evaluation{observability.EvaluationCentral, observability.EvaluationLocal},
		SeverityRouting:   map[string]string{"critical": "platform-security"},
		MinimumAppVersion: ">=1.3.0 <2.0.0",
		ObservabilityPack: observability.PackReference{Package: "observability-pack", VersionRange: ">=0.2.0 <0.3.0"},
	}
}

func TestCapabilityValidAndDeterministicDigest(t *testing.T) {
	c := validCapability()
	if err := c.Validate(); err != nil {
		t.Fatalf("valid capability rejected: %v", err)
	}
	first, err := c.Digest()
	if err != nil {
		t.Fatal(err)
	}
	c.Datasources["prometheus"] = "${DS_PROMETHEUS}"
	second, err := c.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if first != second || !strings.HasPrefix(first, "sha256:") {
		t.Fatalf("digest is not deterministic: %q / %q", first, second)
	}
}

func TestCapabilityRejectsInvalidMetadata(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*observability.Capability)
		want   string
	}{
		{"api version", func(c *observability.Capability) { c.APIVersion = "v1" }, "apiVersion"},
		{"app prefix", func(c *observability.Capability) { c.Dashboards[0].UID = "other-overview" }, "uid"},
		{"duplicate uid", func(c *observability.Capability) { c.Dashboards = append(c.Dashboards, c.Dashboards[0]) }, "duplicate"},
		{"unsafe asset", func(c *observability.Capability) { c.Rules[0].File = "../rules.yaml" }, "path"},
		{"unsupported target", func(c *observability.Capability) { c.Targets = []observability.Evaluation{"fleet"} }, "targets"},
		{"latest pack", func(c *observability.Capability) { c.ObservabilityPack.VersionRange = "latest" }, "latest"},
		{"multiple datasource mappings", func(c *observability.Capability) { c.Datasources["logs"] = "${DS_LOGS}" }, "datasource"},
		{"unknown link variable", func(c *observability.Capability) { c.LinkVariables = []string{"pod"} }, "linkVariables"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := validCapability()
			tc.mutate(c)
			if err := c.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCompatibilityRanges(t *testing.T) {
	c := validCapability()
	if err := c.ValidateAppVersion("1.4.0"); err != nil {
		t.Fatalf("compatible app rejected: %v", err)
	}
	if err := c.ValidateAppVersion("1.2.0"); err == nil {
		t.Fatal("incompatible app accepted")
	}
	if err := c.ValidatePackVersion("0.2.1"); err != nil {
		t.Fatalf("compatible pack rejected: %v", err)
	}
	if err := c.ValidatePackVersion("0.3.0"); err == nil {
		t.Fatal("incompatible pack accepted")
	}
}

func TestLinkVariablesUseOnlyDeclaredBindings(t *testing.T) {
	got, err := observability.BindLinkVariables([]string{"namespace"}, map[string]string{"namespace": "payments", "instance": "ignored"})
	if err != nil || got["namespace"] != "payments" || len(got) != 1 {
		t.Fatalf("bindings = %v, %v", got, err)
	}
	if _, err := observability.BindLinkVariables([]string{"namespace"}, nil); err == nil {
		t.Fatal("missing binding accepted")
	}
}

func validVerification() *observability.VerificationResult {
	return &observability.VerificationResult{
		APIVersion:       observability.APIVersion,
		Kind:             observability.VerificationKind,
		Companion:        observability.CompanionIdentity{ID: "oc-1", Generation: 3},
		Target:           observability.TargetIdentity{ID: "rt-1", Revision: 7, ClusterID: "cluster-1"},
		App:              observability.AppIdentity{PlanID: "dp-1", InstanceID: "app-1", SourceDigest: testDigest, Revision: "generation-1"},
		Pack:             observability.PackIdentity{ItemID: "hiroba/observability-pack", Version: "0.2.0", Digest: testDigest},
		CapabilityDigest: testDigest,
		Suite:            observability.SuiteIdentity{Digest: testDigest, Profile: "stage"},
		Run:              observability.RunIdentity{ID: "vr-184", Attempt: 1, FinishedAt: time.Date(2026, 11, 15, 21, 42, 0, 0, time.UTC)},
		Result:           "passed",
		Checks: observability.VerificationChecks{
			Folder:      observability.FolderCheck{Check: observability.Check{State: "passed"}, Name: "acme-inc/payments-api", UID: "folder-1"},
			Dashboards:  []observability.DashboardCheck{{Check: observability.Check{State: "passed"}, UID: "payments-api-overview", Title: "Service overview", Slug: "overview"}},
			Datasources: []observability.DatasourceCheck{{Check: observability.Check{State: "passed"}, Logical: "prometheus", Name: "metrics-central", Type: "prometheus"}},
			Rules:       []observability.RuleCheck{{Check: observability.Check{State: "passed"}, File: "rules/availability.yaml", Evaluation: observability.EvaluationCentral}},
		},
		ExpiresAt: time.Date(2026, 11, 16, 21, 42, 0, 0, time.UTC),
	}
}

func TestVerificationResultValidatesFreshReadOnlyEvidence(t *testing.T) {
	r := validVerification()
	now := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	if err := r.ValidateAt(now); err != nil {
		t.Fatalf("valid verification rejected: %v", err)
	}
	r.Result = "failed"
	if err := r.ValidateAt(now); err != nil {
		t.Fatalf("well-formed failure evidence rejected: %v", err)
	}
}

func TestVerificationUsesGenericProfilesWithoutGrafanaAccess(t *testing.T) {
	now := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	for _, profile := range []string{"ci", "stage", "prod"} {
		t.Run(profile, func(t *testing.T) {
			r := validVerification()
			r.Suite.Profile = profile
			if err := r.ValidateAt(now); err != nil {
				t.Fatalf("valid %s evidence without Grafana access rejected: %v", profile, err)
			}
		})
	}
	r := validVerification()
	r.Suite.Profile = "observability-readonly-v1"
	if err := r.ValidateAt(now); err == nil {
		t.Fatal("retired synthetic profile accepted")
	}
	if _, found := reflect.TypeOf(observability.VerificationChecks{}).FieldByName("Access"); found {
		t.Fatal("VerificationChecks still contains the removed Grafana Access field")
	}
}

func TestVerificationResultRejectsPartialOrExpiredEvidence(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*observability.VerificationResult)
		now    time.Time
	}{
		{"expired", func(r *observability.VerificationResult) {}, time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)},
		{"folder not passed", func(r *observability.VerificationResult) { r.Checks.Folder.State = "failed" }, time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)},
		{"missing dashboard", func(r *observability.VerificationResult) { r.Checks.Dashboards = nil }, time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)},
		{"bad digest", func(r *observability.VerificationResult) { r.CapabilityDigest = "latest" }, time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := validVerification()
			tc.mutate(r)
			if err := r.ValidateAt(tc.now); err == nil {
				t.Fatal("invalid passing evidence accepted")
			}
		})
	}
}
