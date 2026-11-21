package trustedsource_test

import (
	"errors"
	"testing"
	"time"

	trustedsource "github.com/7K-Inari/inari-api/contract/trustedsource/v1"
)

func boolPtr(v bool) *bool { return &v }

func makeApplicationEligible(e *trustedsource.TestEvidence) {
	e.TargetAdapter = "manifests"
	e.FixtureOnly = boolPtr(false)
	e.ExecutableProfiles = []string{"ci", "stage", "prod"}
	e.EvidenceScope = trustedsource.EvidenceScopeApplication
	e.AppEvidenceEligible = boolPtr(true)
}

func eligibilityCode(err error) string {
	var typed *trustedsource.EligibilityError
	if errors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

func TestRealEnvironmentEligibility(t *testing.T) {
	tests := []struct {
		name            string
		mutate          func(*trustedsource.TestEvidence)
		profile         string
		environment     trustedsource.Environment
		legacyException bool
		wantCode        string
	}{
		{name: "application production", mutate: makeApplicationEligible, profile: "prod", environment: trustedsource.EnvironmentProduction},
		{name: "legacy denied", mutate: func(*trustedsource.TestEvidence) {}, profile: "stage", environment: trustedsource.EnvironmentStaging, wantCode: trustedsource.CodeSuiteEligibilityUnknown},
		{name: "legacy production exception denied", mutate: func(*trustedsource.TestEvidence) {}, profile: "prod", environment: trustedsource.EnvironmentProduction, legacyException: true, wantCode: trustedsource.CodeSuiteEligibilityUnknown},
		{name: "legacy development exception", mutate: func(*trustedsource.TestEvidence) {}, profile: "stage", environment: trustedsource.EnvironmentDevelopment, legacyException: true},
		{name: "explicitly ineligible", mutate: func(e *trustedsource.TestEvidence) {
			makeApplicationEligible(e)
			e.AppEvidenceEligible = boolPtr(false)
		}, profile: "prod", environment: trustedsource.EnvironmentProduction, wantCode: trustedsource.CodeSuiteCIOnly},
		{name: "fixture only", mutate: func(e *trustedsource.TestEvidence) {
			makeApplicationEligible(e)
			e.FixtureOnly = boolPtr(true)
			e.EvidenceScope = trustedsource.EvidenceScopeFixture
			e.AppEvidenceEligible = boolPtr(false)
			e.ExecutableProfiles = []string{"ci"}
		}, profile: "ci", environment: trustedsource.EnvironmentDevelopment, wantCode: trustedsource.CodeSuiteCIOnly},
		{name: "profile not executable", mutate: makeApplicationEligible, profile: "other", environment: trustedsource.EnvironmentProduction, wantCode: trustedsource.CodeSuiteProfileNotExecutable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEvidence()
			tt.mutate(e)
			err := e.ValidateRealEnvironmentEligibility(tt.profile, tt.environment, tt.legacyException)
			if tt.wantCode == "" && err != nil {
				t.Fatalf("unexpected denial: %v", err)
			}
			if tt.wantCode != "" && eligibilityCode(err) != tt.wantCode {
				t.Fatalf("error = %v, want code %q", err, tt.wantCode)
			}
		})
	}
}

func TestEligibilityMetadataMustBeComplete(t *testing.T) {
	e := validEvidence()
	e.TargetAdapter = "manifests"
	e.FixtureOnly = boolPtr(false)
	if err := e.ValidateRealEnvironmentEligibility("prod", trustedsource.EnvironmentProduction, false); eligibilityCode(err) != trustedsource.CodeSuiteEligibilityUnknown {
		t.Fatalf("partial metadata error = %v", err)
	}
	if err := e.ValidateAt(time.Now(), 7*24*time.Hour, false); err == nil {
		t.Fatal("partial eligibility metadata passed general evidence validation")
	}
}
