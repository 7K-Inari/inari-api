package extensionv1_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestExtensionContractSchema(t *testing.T) {
	raw, err := os.ReadFile("schema/extension-contracts-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("extension-contracts-v1.schema.json", document); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("extension-contracts-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		doc   any
		valid bool
	}{
		{"v1 additive", map[string]any{"backend": []any{"inari.plugin.v1"}, "future": true}, true},
		{"future v1 v2", map[string]any{"backend": []any{"inari.plugin.v1", "inari.plugin.v2"}}, true},
		{"ui", map[string]any{"ui": map[string]any{"sdk": ">=0.1.8 <2.0.0", "hostApi": "inari.ui.v1"}}, true},
		{"empty", map[string]any{}, false},
		{"empty backend", map[string]any{"backend": []any{}}, false},
		{"duplicate backend", map[string]any{"backend": []any{"inari.plugin.v1", "inari.plugin.v1"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.Validate(tc.doc)
			if tc.valid && err != nil {
				t.Fatalf("valid document rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("invalid document accepted")
			}
		})
	}
}
