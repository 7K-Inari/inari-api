package observability_test

import (
	"strings"
	"testing"

	observability "github.com/7K-Inari/inari-api/contract/observability/v1"
)

const validCapabilityYAML = `apiVersion: inari.dev/v1alpha1
kind: Observability
app: payments-api
folder: payments-api
datasources:
  prometheus: ${DS_PROMETHEUS}
linkVariables: [namespace]
dashboards:
  - title: Service overview
    uid: payments-api-overview
    file: dashboards/overview.json
rules:
  - file: rules/availability.yaml
    evaluation: central
targets: [central, local]
severityRouting:
  critical: platform-security
minimumAppVersion: ">=1.3.0 <2.0.0"
observabilityPack:
  package: observability-pack
  versionRange: ">=0.2.0 <0.3.0"
`

func TestDecodeCapabilityStrictYAML(t *testing.T) {
	c, err := observability.DecodeCapability([]byte(validCapabilityYAML))
	if err != nil {
		t.Fatalf("DecodeCapability: %v", err)
	}
	if c.App != "payments-api" || c.ObservabilityPack.Package != "observability-pack" {
		t.Fatalf("decoded capability = %+v", c)
	}
	for _, tc := range []struct{ name, input, want string }{
		{"unknown", validCapabilityYAML + "unknown: true\n", "unknown"},
		{"duplicate", strings.Replace(validCapabilityYAML, "kind: Observability", "kind: Observability\nkind: Observability", 1), "duplicate"},
		{"multiple", validCapabilityYAML + "---\n{}\n", "multiple"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := observability.DecodeCapability([]byte(tc.input)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}
