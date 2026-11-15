package trustedsource_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	trustedsource "github.com/7K-Inari/inari-api/contract/trustedsource/v1"
)

func TestSourceFileSchemaRejectsUnsignedOutsideDevelopment(t *testing.T) {
	raw, err := os.ReadFile("schema/source-file-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("source-file-v1.schema.json", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("source-file-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		environment trustedsource.Environment
		wantValid   bool
	}{
		{trustedsource.EnvironmentDevelopment, true},
		{"", false},
		{trustedsource.EnvironmentProduction, false},
		{trustedsource.EnvironmentStaging, false},
	} {
		t.Run("environment="+string(tc.environment), func(t *testing.T) {
			f := validFile()
			f.Environment = tc.environment
			f.Template.Sources = []trustedsource.Source{}
			f.Catalog.Sources[0].TrustTier = trustedsource.TrustExperimental
			f.Catalog.Sources[0].AllowUnsigned = trustedsource.AllowUnsignedDevelopmentOnly
			raw, err := json.Marshal(f)
			if err != nil {
				t.Fatal(err)
			}
			var document any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			err = schema.Validate(document)
			if tc.wantValid && err != nil {
				t.Fatalf("development config rejected: %v", err)
			}
			if !tc.wantValid && err == nil {
				t.Fatal("non-development unsigned config unexpectedly accepted")
			}
		})
	}
}

func TestJSONSchemasAreValidAndVersioned(t *testing.T) {
	for _, name := range []string{"source-file-v1.schema.json", "catalog-index-v1.schema.json", "template-index-v1.schema.json", "test-evidence-v1.schema.json"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("schema", name))
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
		})
	}
}
