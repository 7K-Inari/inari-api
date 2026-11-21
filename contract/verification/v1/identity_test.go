package verification_test

import (
	"strings"
	"testing"
	"time"

	verification "github.com/7K-Inari/inari-api/contract/verification/v1"
)

const (
	testInstanceUID    = "2b6e4424-87a4-4f7c-b6d8-dfe2f6a894f8"
	testArtifactDigest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	testSuiteDigest    = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func validAppRevisionInput() verification.AppRevisionInput {
	return verification.AppRevisionInput{
		InstanceUID:    testInstanceUID,
		ItemID:         "hiroba/example",
		Version:        "1.4.0",
		ArtifactDigest: testArtifactDigest,
		SpecDigest:     "sha256:" + strings.Repeat("b", 64),
		Generation:     4,
		DesiredCommit:  "0123456789abcdef0123456789abcdef01234567",
	}
}

func TestComputeAppRevisionGolden(t *testing.T) {
	got, err := verification.ComputeAppRevision(validAppRevisionInput())
	if err != nil {
		t.Fatalf("ComputeAppRevision: %v", err)
	}
	const want = "sha256:99595485ac9f27595465be7873826681a9a19732e39a8588ffcf827c536308f9"
	if got != want {
		t.Fatalf("app revision = %q, want %q", got, want)
	}
}

func TestComputeRunIdentityGolden(t *testing.T) {
	const appRevision = "sha256:99595485ac9f27595465be7873826681a9a19732e39a8588ffcf827c536308f9"
	trigger := verification.Trigger{Kind: verification.TriggerOnRevisionChange, Profile: "prod"}
	got, err := verification.ComputeRunIdentity(testInstanceUID, appRevision, testSuiteDigest, trigger)
	if err != nil {
		t.Fatalf("ComputeRunIdentity: %v", err)
	}
	if got.AppRevision != appRevision {
		t.Fatalf("app revision = %q", got.AppRevision)
	}
	if got.RunKey != "sha256:caea4fa53c90986be1d91f8b8449308ae911d5ff007de162428e790532edc843" {
		t.Fatalf("run key = %q", got.RunKey)
	}
	if got.RunID != "vr-zlve7jj4scmgxyozd6fyisjqrlurdvp7" {
		t.Fatalf("run ID = %q", got.RunID)
	}
	if got.TriggerFingerprint != "sha256:88bb58e1ffa1f709a5bb6890463177024fd282454038679fc23f3af61a537d7d" {
		t.Fatalf("trigger fingerprint = %q", got.TriggerFingerprint)
	}
	again, err := verification.ComputeRunIdentity(testInstanceUID, appRevision, testSuiteDigest, trigger)
	if err != nil || again != got {
		t.Fatalf("identity is not deterministic: %#v, %v", again, err)
	}
}

func TestRunIdentityChangesWithEveryKeyDimension(t *testing.T) {
	base := validAppRevisionInput()
	appRevision, err := verification.ComputeAppRevision(base)
	if err != nil {
		t.Fatal(err)
	}
	trigger := verification.Trigger{Kind: verification.TriggerManual, Profile: "stage", RequestKey: "req-1"}
	original, err := verification.ComputeRunIdentity(testInstanceUID, appRevision, testSuiteDigest, trigger)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name                 string
		uid, revision, suite string
		trigger              verification.Trigger
	}{
		{name: "instance", uid: "3c7f5535-98b5-508d-a709-eef3a7b90509", revision: appRevision, suite: testSuiteDigest, trigger: trigger},
		{name: "revision", uid: testInstanceUID, revision: "sha256:" + strings.Repeat("d", 64), suite: testSuiteDigest, trigger: trigger},
		{name: "suite", uid: testInstanceUID, revision: appRevision, suite: "sha256:" + strings.Repeat("e", 64), trigger: trigger},
		{name: "trigger", uid: testInstanceUID, revision: appRevision, suite: testSuiteDigest, trigger: verification.Trigger{Kind: verification.TriggerManual, Profile: "stage", RequestKey: "req-2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := verification.ComputeRunIdentity(tt.uid, tt.revision, tt.suite, tt.trigger)
			if err != nil {
				t.Fatal(err)
			}
			if got.RunKey == original.RunKey || got.RunID == original.RunID {
				t.Fatalf("identity did not change: %#v", got)
			}
		})
	}
}

func TestCronTriggerFromChildName(t *testing.T) {
	got, err := verification.CronTriggerFromChildName("vc-nightly", "vc-nightly-1767223200", "prod", "nightly")
	if err != nil {
		t.Fatalf("CronTriggerFromChildName: %v", err)
	}
	wantTime := time.Unix(1767223200, 0).UTC().Format(time.RFC3339)
	if got.Kind != verification.TriggerCron || got.Profile != "prod" || got.ScheduleSetID != "nightly" || got.ScheduledAt != wantTime {
		t.Fatalf("unexpected trigger: %#v", got)
	}
	for _, child := range []string{"other-1767223200", "vc-nightly-not-a-time", "vc-nightly-1767223200-extra", "vc-nightly-0"} {
		if _, err := verification.CronTriggerFromChildName("vc-nightly", child, "prod", "nightly"); err == nil {
			t.Errorf("child %q was accepted", child)
		}
	}
}

func TestDigestJSONCanonical(t *testing.T) {
	a, err := verification.DigestJSON([]byte(`{"b":2,"a":{"z":1,"y":[3,2]}}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := verification.DigestJSON([]byte(`{"a":{"y":[3,2],"z":1},"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if a == "" || a != b {
		t.Fatalf("canonical digests differ: %q, %q", a, b)
	}
	if _, err := verification.DigestJSON([]byte(`{"a":`)); err == nil {
		t.Fatal("malformed JSON was accepted")
	}
}
