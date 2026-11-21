package agentv1_test

import (
	"testing"
	"time"

	agentv1 "github.com/7K-Inari/inari-api/gen/go/inari/agent/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestVerificationEventTypeIsNextContractValue(t *testing.T) {
	const eventType = agentv1.EventType(16)
	if got := agentv1.EventTypeString(eventType); got != "inari.agent.verification-report.v1" {
		t.Fatalf("EventTypeString(16) = %q, want verification report", got)
	}
	if got := agentv1.EventTypeFromString("inari.agent.verification-report.v1"); got != eventType {
		t.Fatalf("EventTypeFromString verification = %v, want %v", got, eventType)
	}
}

func TestVerificationReportPayloadRoundTrip(t *testing.T) {
	payload := &agentv1.VerificationReport{
		RunId: "vr-zlve7jj4scmgxyozd6fyisjqrlurdvp7", Attempt: 2,
		TriggerFingerprint: "sha256:trigger", SuiteDigest: "sha256:suite", Profile: "prod",
		RenderFingerprint: "sha256:render", ResourceVersion: "42", Deleted: false,
		Workflow:   &agentv1.ResourceRef{Kind: "Workflow", Name: "vr-run-a2", Namespace: "verification", Uid: "workflow-uid"},
		ObservedAt: timestamppb.New(time.Unix(10, 0).UTC()), StartedAt: timestamppb.New(time.Unix(5, 0).UTC()),
		FinishedAt: timestamppb.New(time.Unix(9, 0).UTC()), ExpiresAt: timestamppb.New(time.Unix(100, 0).UTC()),
		Result:      agentv1.VerificationResult_VERIFICATION_RESULT_PASSED,
		SubjectKind: "observability.companion", InstanceUid: "instance-uid", AppInstanceId: "app-1",
		DeploymentPlanId: "dp-1", AppRevision: "sha256:revision", AppDigest: "sha256:app",
		TargetId: "rt-mon", TargetRevision: 4, CompanionId: "oc-1", CompanionGeneration: 3,
		PackDigest: "sha256:pack", CapabilityDigest: "sha256:capability",
		OutcomeJson: `{"result":"passed"}`, OutcomeDigest: "sha256:outcome", StateChecksum: "sha256:state",
		PolicyRevision: 7, GrantRevision: 3,
	}
	out := roundTripEvent(t, "inari.agent.verification-report.v1", payload)
	decoded := &agentv1.VerificationReport{}
	if err := out.GetPayload().UnmarshalTo(decoded); err != nil {
		t.Fatalf("unmarshal verification report: %v", err)
	}
	if decoded.GetRunId() != payload.GetRunId() || decoded.GetAttempt() != 2 || decoded.GetCompanionGeneration() != 3 ||
		decoded.GetWorkflow().GetUid() != "workflow-uid" || decoded.GetOutcomeJson() != payload.GetOutcomeJson() ||
		decoded.GetResult() != agentv1.VerificationResult_VERIFICATION_RESULT_PASSED {
		t.Fatalf("verification report mismatch: %v", decoded)
	}
}

func TestStatusUpdateIncludesObservedRevisionIdentity(t *testing.T) {
	payload := &agentv1.StatusUpdate{
		Resource:         &agentv1.ResourceRef{Kind: "Application", Name: "web"},
		ObservedRevision: "0123456789abcdef0123456789abcdef01234567",
		SyncOperationId:  "op-7", SyncOperationPhase: "Succeeded",
		OperationFinishedAt: timestamppb.New(time.Unix(12, 0).UTC()),
	}
	out := roundTripEvent(t, "inari.agent.status-update.v1", payload)
	decoded := &agentv1.StatusUpdate{}
	if err := out.GetPayload().UnmarshalTo(decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.GetObservedRevision() != payload.GetObservedRevision() || decoded.GetSyncOperationId() != "op-7" ||
		decoded.GetSyncOperationPhase() != "Succeeded" || decoded.GetOperationFinishedAt().GetSeconds() != 12 {
		t.Fatalf("observed revision identity mismatch: %v", decoded)
	}
}
