// Package provisioning defines dependency-light validation and wire metadata
// shared by producers and consumers of the generic provisioning contract.
package provisioning

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

const (
	IdempotencyKeyHeader    = "Idempotency-Key"
	ProvisioningScope       = "provisioning"
	MaxIdempotencyKeyLength = 255
)

var (
	slugPattern        = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	planPattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)
	featureGatePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]{0,127}$`)
)

// Limits is the dependency-light form of the generic plan-limits document.
// A nil numeric pointer means unlimited; a pointer to zero allows none.
type Limits struct {
	MaxClusters       *int64
	MaxDeveloperSeats *int64
	AllFeatures       bool
	FeatureGates      map[string]bool
}

func NormalizeEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func ValidateIdempotencyKey(value string) error {
	if value == "" || len(value) > MaxIdempotencyKeyLength {
		return fmt.Errorf("provisioning: idempotency key length must be 1-%d bytes", MaxIdempotencyKeyLength)
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 0x21 || value[i] > 0x7e {
			return fmt.Errorf("provisioning: idempotency key must contain visible ASCII only")
		}
	}
	return nil
}

func ValidateSlug(value string) error {
	if len(value) < 2 || len(value) > 63 || !slugPattern.MatchString(value) {
		return fmt.Errorf("provisioning: slug must be a 2-63 character lowercase DNS label")
	}
	return nil
}

func ValidateAdminEmail(value string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("provisioning: admin email must be a bare address")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || strings.Count(value, "@") != 1 {
		return fmt.Errorf("provisioning: admin email must be a bare valid address")
	}
	return nil
}

func ValidatePlan(value string) error {
	if !planPattern.MatchString(value) {
		return fmt.Errorf("provisioning: plan must be a non-empty opaque identifier of at most 128 characters")
	}
	return nil
}

func ValidateLimits(limits *Limits) error {
	if limits == nil {
		return fmt.Errorf("provisioning: limits are required")
	}
	for name, value := range map[string]*int64{
		"max clusters":        limits.MaxClusters,
		"max developer seats": limits.MaxDeveloperSeats,
	} {
		if value != nil && *value < 0 {
			return fmt.Errorf("provisioning: %s cannot be negative", name)
		}
	}
	if len(limits.FeatureGates) > 128 {
		return fmt.Errorf("provisioning: at most 128 feature gates are allowed")
	}
	for name := range limits.FeatureGates {
		if !featureGatePattern.MatchString(name) {
			return fmt.Errorf("provisioning: invalid feature gate name %q", name)
		}
	}
	return nil
}
