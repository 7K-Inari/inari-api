package extensionv1_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	extensionv1 "github.com/7K-Inari/inari-api/contract/extension/v1"
)

func TestNormalizeSupportSet(t *testing.T) {
	got, err := extensionv1.NormalizeSupportSet([]string{" inari.plugin.v2 ", "inari.plugin.v1"})
	if err != nil || !reflect.DeepEqual(got, []string{"inari.plugin.v1", "inari.plugin.v2"}) {
		t.Fatalf("NormalizeSupportSet = %v, %v", got, err)
	}
	for _, ids := range [][]string{nil, {}, {""}, {"inari.plugin.v1", "inari.plugin.v1"}, {"not-a-contract"}} {
		if _, err := extensionv1.NormalizeSupportSet(ids); err == nil {
			t.Fatalf("NormalizeSupportSet(%v) accepted invalid set", ids)
		}
	}
}

func TestIntersects(t *testing.T) {
	if !extensionv1.Intersects([]string{"inari.plugin.v1", "inari.plugin.v2"}, []string{"inari.plugin.v2"}) {
		t.Fatal("expected v2 intersection")
	}
	if extensionv1.Intersects([]string{"inari.plugin.v1"}, []string{"inari.plugin.v2"}) {
		t.Fatal("unexpected cross-major intersection")
	}
	if extensionv1.Intersects(nil, []string{"inari.plugin.v1"}) {
		t.Fatal("empty set must not intersect")
	}
}

func TestContractsValidate(t *testing.T) {
	for _, sdkRange := range []string{">=0.1.8 <2.0.0", "^0.1.8", "~0.1.8", "^0.1.8 || ^1.0.0"} {
		t.Run("range="+sdkRange, func(t *testing.T) {
			valid := &extensionv1.Contracts{Backend: []string{extensionv1.BackendContractV1}, UI: &extensionv1.UIContract{SDK: sdkRange, HostAPI: extensionv1.UIHostAPIV1}}
			if err := valid.Validate(); err != nil {
				t.Fatalf("valid declaration rejected: %v", err)
			}
		})
	}
	tests := []struct {
		name      string
		contracts *extensionv1.Contracts
		want      string
	}{
		{"nil", nil, "required"},
		{"empty", &extensionv1.Contracts{}, "at least one"},
		{"empty backend", &extensionv1.Contracts{Backend: []string{}}, "backend"},
		{"duplicate backend", &extensionv1.Contracts{Backend: []string{extensionv1.BackendContractV1, extensionv1.BackendContractV1}}, "duplicate"},
		{"wrong backend family", &extensionv1.Contracts{Backend: []string{extensionv1.UIHostAPIV1}}, "inari.plugin"},
		{"missing host", &extensionv1.Contracts{UI: &extensionv1.UIContract{SDK: "^0.1.8"}}, "hostApi"},
		{"wrong host family", &extensionv1.Contracts{UI: &extensionv1.UIContract{SDK: "^0.1.8", HostAPI: extensionv1.BackendContractV1}}, "inari.ui"},
		{"missing sdk", &extensionv1.Contracts{UI: &extensionv1.UIContract{HostAPI: extensionv1.UIHostAPIV1}}, "sdk"},
		{"invalid sdk", &extensionv1.Contracts{UI: &extensionv1.UIContract{SDK: "not-a-range", HostAPI: extensionv1.UIHostAPIV1}}, "sdk"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.contracts.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestValidateDeclarationPresenceAndSkew(t *testing.T) {
	if err := extensionv1.ValidateDeclaration(nil, true, false, false); err != nil {
		t.Fatalf("legacy optional declaration rejected: %v", err)
	}
	err := extensionv1.ValidateDeclaration(nil, true, false, true)
	var contractErr *extensionv1.ContractError
	if !errors.As(err, &contractErr) || contractErr.Code != extensionv1.CodeContractRequired {
		t.Fatalf("required error = %#v", err)
	}
	if err := extensionv1.ValidateDeclaration(&extensionv1.Contracts{Backend: []string{"inari.plugin.v1", "inari.plugin.v2"}}, true, false, true); err != nil {
		t.Fatalf("future v1+v2 declaration rejected: %v", err)
	}
	if err := extensionv1.ValidateDeclaration(&extensionv1.Contracts{Backend: []string{extensionv1.BackendContractV1}}, false, true, true); err == nil {
		t.Fatal("missing UI declaration accepted when UI is required")
	}
}
