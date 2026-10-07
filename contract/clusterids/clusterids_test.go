package clusterids_test

import (
	"regexp"
	"testing"

	"github.com/7K-Inari/inari-api/contract/clusterids"
)

// charsetRe is the reference charset the testdata pins: start with an
// alphanumeric, then alphanumerics plus . _ : - (safe unquoted in YAML
// names and URL paths). Consumers' own validators must accept/reject
// exactly the testdata; this regex only checks the testdata against the
// documented reference.
var charsetRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)

// serverIssuedRe pins the canonical shape inari-server's clusterregistry
// issues ("cluster:" + uuid).
var serverIssuedRe = regexp.MustCompile(`^cluster:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestValidIDsMatchReferenceCharset(t *testing.T) {
	for _, id := range clusterids.ValidIDs {
		if !charsetRe.MatchString(id) {
			t.Errorf("ValidIDs entry %q violates the reference charset", id)
		}
	}
}

func TestInvalidIDsViolateReferenceCharset(t *testing.T) {
	for _, id := range clusterids.InvalidIDs {
		if charsetRe.MatchString(id) {
			t.Errorf("InvalidIDs entry %q matches the reference charset", id)
		}
	}
}

// TestCanonicalShapePresent guarantees the testdata always carries the
// server's issued shape — the regression this contract exists for.
func TestCanonicalShapePresent(t *testing.T) {
	found := false
	for _, id := range clusterids.ValidIDs {
		if serverIssuedRe.MatchString(id) {
			found = true
		}
	}
	if !found {
		t.Error("ValidIDs carries no canonical cluster:<uuid> entry")
	}
}
