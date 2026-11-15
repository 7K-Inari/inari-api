package extensionv1

import (
	"regexp"
	"sort"
	"strings"

	"github.com/blang/semver"
)

var (
	contractIDPattern = regexp.MustCompile(`^inari\.(plugin|ui)\.v[1-9][0-9]*$`)
	backendIDPattern  = regexp.MustCompile(`^inari\.plugin\.v[1-9][0-9]*$`)
	uiHostIDPattern   = regexp.MustCompile(`^inari\.ui\.v[1-9][0-9]*$`)
)

// NormalizeSupportSet validates and deterministically sorts supported IDs.
func NormalizeSupportSet(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, NewContractError(CodeContractMalformed, "supportedContracts", "at least one supported contract is required")
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if !contractIDPattern.MatchString(id) {
			return nil, NewContractError(CodeContractMalformed, "supportedContracts", "invalid contract ID %q", raw)
		}
		if seen[id] {
			return nil, NewContractError(CodeContractMalformed, "supportedContracts", "duplicate contract ID %q", id)
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

// Intersects reports whether two contract ID sets share at least one ID.
func Intersects(a, b []string) bool {
	set := make(map[string]bool, len(a))
	for _, id := range a {
		set[id] = true
	}
	for _, id := range b {
		if set[id] {
			return true
		}
	}
	return false
}

// Validate checks explicitly present declarations.
func (c *Contracts) Validate() error {
	if c == nil {
		return NewContractError(CodeContractRequired, "spec.contracts", "declaration is required")
	}
	if c.Backend == nil && c.UI == nil {
		return NewContractError(CodeContractMalformed, "spec.contracts", "at least one backend or ui declaration is required")
	}
	if c.Backend != nil {
		if len(c.Backend) == 0 {
			return NewContractError(CodeContractMalformed, "spec.contracts.backend", "at least one backend contract is required")
		}
		seen := map[string]bool{}
		for _, raw := range c.Backend {
			id := strings.TrimSpace(raw)
			if !backendIDPattern.MatchString(id) {
				return NewContractError(CodeContractMalformed, "spec.contracts.backend", "invalid inari.plugin contract ID %q", raw)
			}
			if seen[id] {
				return NewContractError(CodeContractMalformed, "spec.contracts.backend", "duplicate contract ID %q", id)
			}
			seen[id] = true
		}
	}
	if c.UI != nil {
		if !uiHostIDPattern.MatchString(strings.TrimSpace(c.UI.HostAPI)) {
			return NewContractError(CodeContractMalformed, "spec.contracts.ui.hostApi", "invalid or missing inari.ui host API %q", c.UI.HostAPI)
		}
		if strings.TrimSpace(c.UI.SDK) == "" {
			return NewContractError(CodeContractMalformed, "spec.contracts.ui.sdk", "a SemVer range is required")
		}
		if _, err := semver.ParseRange(c.UI.SDK); err != nil {
			return NewContractError(CodeContractMalformed, "spec.contracts.ui.sdk", "invalid SemVer range %q: %v", c.UI.SDK, err)
		}
	}
	return nil
}

// ValidateDeclaration applies kind-specific presence and shape rules.
func ValidateDeclaration(c *Contracts, backend, ui, required bool) error {
	if c == nil {
		if required && backend {
			return NewContractError(CodeContractRequired, "spec.contracts.backend", "declaration is required")
		}
		if required && ui {
			return NewContractError(CodeContractRequired, "spec.contracts.ui", "declaration is required")
		}
		return nil
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if required && backend && c.Backend == nil {
		return NewContractError(CodeContractRequired, "spec.contracts.backend", "declaration is required")
	}
	if required && ui && c.UI == nil {
		return NewContractError(CodeContractRequired, "spec.contracts.ui", "declaration is required")
	}
	return nil
}
