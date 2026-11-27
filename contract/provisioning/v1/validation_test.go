package provisioning

import (
	"strings"
	"testing"
)

func TestValidateIdempotencyKey(t *testing.T) {
	for _, key := range []string{"", "has space", "has\nnewline", strings.Repeat("x", 256)} {
		if err := ValidateIdempotencyKey(key); err == nil {
			t.Errorf("ValidateIdempotencyKey(%q) accepted invalid key", key)
		}
	}
	if err := ValidateIdempotencyKey("evt_01J:replay-1"); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
}

func TestValidateSlug(t *testing.T) {
	for _, slug := range []string{"", "a", "Acme", "acme_", "-acme", "acme-", strings.Repeat("a", 64)} {
		if err := ValidateSlug(slug); err == nil {
			t.Errorf("ValidateSlug(%q) accepted invalid slug", slug)
		}
	}
	for _, slug := range []string{"acme", "acme-2", "a1"} {
		if err := ValidateSlug(slug); err != nil {
			t.Errorf("ValidateSlug(%q): %v", slug, err)
		}
	}
}

func TestValidateAdminEmail(t *testing.T) {
	if err := ValidateAdminEmail("not-an-email"); err == nil {
		t.Fatal("accepted malformed email")
	}
	if err := ValidateAdminEmail("Admin <admin@example.com>"); err == nil {
		t.Fatal("accepted display-name address")
	}
	if err := ValidateAdminEmail("Admin@Example.com"); err != nil {
		t.Fatalf("valid email rejected: %v", err)
	}
	if got := NormalizeEmail(" Admin@Example.COM "); got != "admin@example.com" {
		t.Fatalf("NormalizeEmail = %q", got)
	}
}

func TestValidateLimitsDistinguishesZeroAndUnlimited(t *testing.T) {
	zero := int64(0)
	limits := &Limits{MaxClusters: &zero, FeatureGates: map[string]bool{"policy-packs": true}}
	if err := ValidateLimits(limits); err != nil {
		t.Fatalf("explicit zero must be valid: %v", err)
	}
	if limits.MaxClusters == nil || *limits.MaxClusters != 0 {
		t.Fatal("explicit zero became unlimited")
	}
	if err := ValidateLimits(&Limits{}); err != nil {
		t.Fatalf("absent numeric limits mean unlimited and must be valid: %v", err)
	}
	negative := int64(-1)
	if err := ValidateLimits(&Limits{MaxDeveloperSeats: &negative}); err == nil {
		t.Fatal("accepted negative seat limit")
	}
	if err := ValidateLimits(&Limits{FeatureGates: map[string]bool{"bad gate": true}}); err == nil {
		t.Fatal("accepted malformed feature gate")
	}
	if err := ValidateLimits(nil); err == nil {
		t.Fatal("accepted nil limits document")
	}
}

func TestValidatePlan(t *testing.T) {
	if err := ValidatePlan(""); err == nil {
		t.Fatal("accepted empty plan")
	}
	if err := ValidatePlan("test-plan:v1"); err != nil {
		t.Fatalf("opaque plan rejected: %v", err)
	}
}
