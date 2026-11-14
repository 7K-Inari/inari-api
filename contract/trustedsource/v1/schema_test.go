package trustedsource_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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
