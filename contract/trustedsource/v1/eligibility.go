package trustedsource

import "fmt"

const (
	CodeSuiteCIOnly               = "suite_ci_only"
	CodeSuiteEligibilityUnknown   = "suite_evidence_eligibility_unknown"
	CodeSuiteProfileNotExecutable = "suite_profile_not_executable"
)

// EligibilityError is a stable real-environment evidence denial.
type EligibilityError struct {
	Code   string
	Reason string
}

func (e *EligibilityError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

// ValidateRealEnvironmentEligibility checks signed evidence before a real
// environment run. legacyNonProductionException is deliberately narrow and
// never permits staging or production use.
func (e *TestEvidence) ValidateRealEnvironmentEligibility(profile string, environment Environment, legacyNonProductionException bool) error {
	known, err := e.validateEligibilityMetadata()
	if err != nil {
		if e.eligibilityFieldCount() == 5 && ((e.FixtureOnly != nil && *e.FixtureOnly) || e.EvidenceScope == EvidenceScopeFixture || (e.AppEvidenceEligible != nil && !*e.AppEvidenceEligible)) {
			return &EligibilityError{Code: CodeSuiteCIOnly, Reason: "fixture or application-ineligible evidence cannot authorize a real environment run"}
		}
		return &EligibilityError{Code: CodeSuiteEligibilityUnknown, Reason: err.Error()}
	}
	if !known {
		if legacyNonProductionException && environment == EnvironmentDevelopment {
			return nil
		}
		return &EligibilityError{Code: CodeSuiteEligibilityUnknown, Reason: "signed evidence has no runtime eligibility declaration"}
	}
	if e.FixtureOnly == nil || *e.FixtureOnly || e.EvidenceScope != EvidenceScopeApplication || e.AppEvidenceEligible == nil || !*e.AppEvidenceEligible {
		return &EligibilityError{Code: CodeSuiteCIOnly, Reason: "fixture or application-ineligible evidence cannot authorize a real environment run"}
	}
	for _, candidate := range e.ExecutableProfiles {
		if candidate == profile {
			return nil
		}
	}
	return &EligibilityError{Code: CodeSuiteProfileNotExecutable, Reason: fmt.Sprintf("profile %q is not executable by the signed suite evidence", profile)}
}

func (e *TestEvidence) eligibilityFieldCount() int {
	if e == nil {
		return 0
	}
	present := 0
	if e.TargetAdapter != "" {
		present++
	}
	if e.FixtureOnly != nil {
		present++
	}
	if e.ExecutableProfiles != nil {
		present++
	}
	if e.EvidenceScope != "" {
		present++
	}
	if e.AppEvidenceEligible != nil {
		present++
	}
	return present
}

// validateEligibilityMetadata returns known=false only when all five additive
// fields are absent. Partial declarations are always malformed.
func (e *TestEvidence) validateEligibilityMetadata() (bool, error) {
	if e == nil {
		return false, fmt.Errorf("evidence is required")
	}
	present := e.eligibilityFieldCount()
	if present == 0 {
		return false, nil
	}
	if present != 5 {
		return false, fmt.Errorf("runtime eligibility requires targetAdapter, fixtureOnly, executableProfiles, evidenceScope, and appEvidenceEligible together")
	}
	if e.TargetAdapter != "helm" && e.TargetAdapter != "manifests" {
		return false, fmt.Errorf("targetAdapter must be helm or manifests")
	}
	if len(e.ExecutableProfiles) == 0 {
		return false, fmt.Errorf("executableProfiles must not be empty")
	}
	seen := map[string]bool{}
	for _, profile := range e.ExecutableProfiles {
		if profile != "ci" && profile != "stage" && profile != "prod" {
			return false, fmt.Errorf("invalid executable profile %q", profile)
		}
		if seen[profile] {
			return false, fmt.Errorf("duplicate executable profile %q", profile)
		}
		seen[profile] = true
	}
	if *e.FixtureOnly {
		if e.TargetAdapter != "manifests" || len(e.ExecutableProfiles) != 1 || e.ExecutableProfiles[0] != "ci" || e.EvidenceScope != EvidenceScopeFixture || e.AppEvidenceEligible == nil || *e.AppEvidenceEligible {
			return false, fmt.Errorf("fixtureOnly evidence must be manifests-based, ci-only, fixture-scoped, and application-ineligible")
		}
		return true, nil
	}
	if e.EvidenceScope != EvidenceScopeApplication || e.AppEvidenceEligible == nil || !*e.AppEvidenceEligible {
		return false, fmt.Errorf("artifact-bound evidence must be application-scoped and application-eligible")
	}
	return true, nil
}
