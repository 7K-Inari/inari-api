package observability_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	observability "github.com/7K-Inari/inari-api/contract/observability/v1"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestJSONSchemasAreValidAndVersioned(t *testing.T) {
	for _, name := range []string{"observability-v1.schema.json", "verification-result-v1.schema.json"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile("schema/" + name)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("invalid JSON Schema: %v", err)
			}
			if doc["$schema"] == nil || doc["$id"] == nil || doc["type"] != "object" {
				t.Fatalf("incomplete schema: %+v", doc)
			}
			compiler := jsonschema.NewCompiler()
			if err := compiler.AddResource(name, doc); err != nil {
				t.Fatal(err)
			}
			if _, err := compiler.Compile(name); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func verificationSchemaError(t *testing.T, result *observability.VerificationResult) error {
	t.Helper()
	raw, err := os.ReadFile("schema/verification-result-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("verification-result-v1.schema.json", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("verification-result-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	return schema.Validate(document)
}

func TestVerificationSchemaMatchesGoContract(t *testing.T) {
	now := time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*observability.VerificationResult)
	}{
		{name: "valid", mutate: func(*observability.VerificationResult) {}},
		{name: "failed empty checks", mutate: func(r *observability.VerificationResult) {
			r.Result, r.Checks = "failed", observability.VerificationChecks{}
		}},
		{name: "missing app execution identity", mutate: func(r *observability.VerificationResult) { r.App.InstanceID, r.App.ScaffoldRunID = "", "" }},
		{name: "empty folder uid", mutate: func(r *observability.VerificationResult) { r.Checks.Folder.UID = "" }},
		{name: "unsafe dashboard slug", mutate: func(r *observability.VerificationResult) { r.Checks.Dashboards[0].Slug = "../other" }},
		{name: "empty datasource type", mutate: func(r *observability.VerificationResult) { r.Checks.Datasources[0].Type = "" }},
		{name: "retired synthetic profile", mutate: func(r *observability.VerificationResult) { r.Suite.Profile = "observability-readonly-v1" }},
		{name: "missing rules", mutate: func(r *observability.VerificationResult) { r.Checks.Rules = nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := validVerification()
			tc.mutate(result)
			goErr := result.ValidateAt(now)
			schemaErr := verificationSchemaError(t, result)
			if (goErr == nil) != (schemaErr == nil) {
				t.Fatalf("Go validation = %v, JSON Schema validation = %v", goErr, schemaErr)
			}
		})
	}
}

func TestCapabilitySchemaMatchesCanonicalFixture(t *testing.T) {
	raw, err := os.ReadFile("schema/observability-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("observability-v1.schema.json", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("observability-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(validCapability())
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(document); err != nil {
		t.Fatalf("canonical fixture rejected by schema: %v", err)
	}
}
