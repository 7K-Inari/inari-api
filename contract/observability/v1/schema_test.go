package observability_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

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
