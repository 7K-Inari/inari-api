package trustedsource_test

import (
	"os"
	"testing"
	"time"

	trustedsource "github.com/7K-Inari/inari-api/contract/trustedsource/v1"
)

func TestCommittedContractFixtures(t *testing.T) {
	for _, name := range []string{"source-file.yaml", "catalog-index.yaml", "template-index.yaml"} {
		raw, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		switch name {
		case "source-file.yaml":
			if _, err := trustedsource.DecodeFile(raw); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		default:
			if _, err := trustedsource.DecodeIndex(raw); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		}
	}
	for _, name := range []string{"test-evidence.json", "slsa-provenance.json"} {
		raw, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if name == "test-evidence.json" {
			e, err := trustedsource.DecodeTestEvidence(raw)
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 11, 13, 12, 30, 0, 0, time.UTC)
			if err := e.ValidateAt(now, 7*24*time.Hour, false); err != nil {
				t.Fatal(err)
			}
		} else if _, err := trustedsource.DecodeStatement(raw); err != nil {
			t.Fatal(err)
		}
	}
}
