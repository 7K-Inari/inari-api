package agentv1_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/7K-Inari/inari-api/gen/go/inari/agent/v1"
)

func TestEventRoundTrip(t *testing.T) {
	hs := &agentv1.HandshakeRequest{AgentVersion: "0.1.0", TenantId: "t-1"}
	payload, err := anypb.New(hs)
	if err != nil {
		t.Fatalf("anypb.New: %v", err)
	}
	in := &agentv1.Event{
		EventId:    "evt-1",
		ResourceId: "res-1",
		Type:       "inari.agent.heartbeat.v1",
		Payload:    payload,
		Time:       timestamppb.New(time.Unix(1, 0).UTC()),
	}
	raw, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := &agentv1.Event{}
	if err := proto.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.GetEventId() != in.GetEventId() || out.GetType() != in.GetType() {
		t.Fatalf("round trip mismatch: got %v want %v", out, in)
	}
	decoded := &agentv1.HandshakeRequest{}
	if err := out.GetPayload().UnmarshalTo(decoded); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if decoded.GetTenantId() != "t-1" {
		t.Fatalf("payload tenant mismatch: %q", decoded.GetTenantId())
	}
}
