package extensionv1_test

import (
	"errors"
	"strings"
	"testing"

	extensionv1 "github.com/7K-Inari/inari-api/contract/extension/v1"
)

func TestDecodeSpecValidFullManifestAndAdditiveFields(t *testing.T) {
	doc := []byte(`apiVersion: inari.io/v1alpha1
kind: Extension
spec:
  kinds: [backend, ui]
  contracts:
    backend: [inari.plugin.v1, inari.plugin.v2]
    ui:
      sdk: ">=0.1.8 <2.0.0"
      hostApi: inari.ui.v1
    futureOptional: preserved-by-forward-compatible-readers
  auth:
    methods: []
`)
	contracts, err := extensionv1.DecodeSpec(doc)
	if err != nil {
		t.Fatalf("DecodeSpec valid document: %v", err)
	}
	if contracts == nil || len(contracts.Backend) != 2 || contracts.Backend[1] != "inari.plugin.v2" {
		t.Fatalf("backend = %#v", contracts)
	}
	if contracts.UI == nil || contracts.UI.HostAPI != extensionv1.UIHostAPIV1 || contracts.UI.SDK != ">=0.1.8 <2.0.0" {
		t.Fatalf("ui = %#v", contracts.UI)
	}
}

func TestDecodeSpecRESTSpecAndMissing(t *testing.T) {
	contracts, err := extensionv1.DecodeSpec([]byte(`{"contracts":{"backend":["inari.plugin.v1"]}}`))
	if err != nil || contracts == nil || len(contracts.Backend) != 1 {
		t.Fatalf("REST spec decode = %#v, %v", contracts, err)
	}
	contracts, err = extensionv1.DecodeSpec([]byte(`{"auth":{"methods":[]}}`))
	if err != nil || contracts != nil {
		t.Fatalf("missing contracts = %#v, %v; want nil, nil", contracts, err)
	}
}

func TestDecodeSpecRejectsMalformedDocuments(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want string
	}{
		{"empty", "", "empty"},
		{"multiple documents", "spec: {}\n---\nspec: {}\n", "multiple"},
		{"duplicate keys", "spec:\n  contracts:\n    backend: [inari.plugin.v1]\n    backend: [inari.plugin.v1]\n", "duplicate"},
		{"null contracts", "spec:\n  contracts: null\n", "mapping"},
		{"empty declaration", "spec:\n  contracts: {}\n", "at least one"},
		{"empty backend", "spec:\n  contracts:\n    backend: []\n", "backend"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := extensionv1.DecodeSpec([]byte(tc.doc))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestDecodeSpecErrorIsTyped(t *testing.T) {
	_, err := extensionv1.DecodeSpec([]byte("spec:\n  contracts:\n    backend: []\n"))
	var contractErr *extensionv1.ContractError
	if !errors.As(err, &contractErr) || contractErr.Code != extensionv1.CodeContractMalformed {
		t.Fatalf("error = %#v, want contract_malformed", err)
	}
}
