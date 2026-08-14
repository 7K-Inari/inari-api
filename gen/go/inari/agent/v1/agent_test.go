package agentv1_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/7K-Inari/inari-api/gen/go/inari/agent/v1"
)

func roundTripEvent(t *testing.T, eventType string, payload proto.Message) *agentv1.Event {
	t.Helper()
	any, err := anypb.New(payload)
	if err != nil {
		t.Fatalf("anypb.New(%T): %v", payload, err)
	}
	in := &agentv1.Event{
		EventId:    "evt-1",
		ResourceId: "res-1",
		Type:       eventType,
		Payload:    any,
		Time:       timestamppb.New(time.Unix(1, 0).UTC()),
		Sequence:   7,
	}
	raw, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := &agentv1.Event{}
	if err := proto.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.GetEventId() != in.GetEventId() || out.GetType() != in.GetType() || out.GetSequence() != in.GetSequence() {
		t.Fatalf("round trip mismatch: got %v want %v", out, in)
	}
	return out
}

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

func TestCapabilityUpdatePayloadRoundTrip(t *testing.T) {
	schema, err := structpb.NewStruct(map[string]any{"type": "object"})
	if err != nil {
		t.Fatalf("structpb.NewStruct: %v", err)
	}
	payload := &agentv1.CapabilityUpdate{
		FullSync:      true,
		StateChecksum: "sha256:abc",
		Capabilities: []*agentv1.Capability{{
			Kind:           agentv1.CapabilityKind_CAPABILITY_KIND_CRD,
			Name:           "widgets.example.com",
			Group:          "example.com",
			Version:        "v1alpha1",
			Schema:         schema,
			ManagementMode: agentv1.ManagementMode_MANAGEMENT_MODE_OBSERVE_ONLY,
			Action:         agentv1.CapabilityAction_CAPABILITY_ACTION_UPSERT,
		}},
	}
	out := roundTripEvent(t, "inari.agent.capability-update.v1", payload)
	decoded := &agentv1.CapabilityUpdate{}
	if err := out.GetPayload().UnmarshalTo(decoded); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if !decoded.GetFullSync() || decoded.GetStateChecksum() != "sha256:abc" {
		t.Fatalf("capability update mismatch: %v", decoded)
	}
	cap := decoded.GetCapabilities()[0]
	if cap.GetKind() != agentv1.CapabilityKind_CAPABILITY_KIND_CRD ||
		cap.GetManagementMode() != agentv1.ManagementMode_MANAGEMENT_MODE_OBSERVE_ONLY {
		t.Fatalf("capability mismatch: %v", cap)
	}
}

func TestStatusUpdatePayloadRoundTrip(t *testing.T) {
	payload := &agentv1.StatusUpdate{
		Resource: &agentv1.ResourceRef{
			Kind: "Application", Name: "web", Namespace: "argocd", Uid: "uid-1",
		},
		Health:        agentv1.HealthStatus_HEALTH_STATUS_HEALTHY,
		Sync:          agentv1.SyncState_SYNC_STATE_SYNCED,
		Message:       "ok",
		ObservedAt:    timestamppb.New(time.Unix(2, 0).UTC()),
		StateChecksum: "sha256:def",
	}
	out := roundTripEvent(t, "inari.agent.status-update.v1", payload)
	decoded := &agentv1.StatusUpdate{}
	if err := out.GetPayload().UnmarshalTo(decoded); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if decoded.GetHealth() != agentv1.HealthStatus_HEALTH_STATUS_HEALTHY ||
		decoded.GetResource().GetName() != "web" {
		t.Fatalf("status update mismatch: %v", decoded)
	}
}

func TestHeartbeatAndResyncRoundTrip(t *testing.T) {
	ping := &agentv1.Ping{Time: timestamppb.New(time.Unix(3, 0).UTC()), Sequence: 9}
	out := roundTripEvent(t, "inari.agent.ping.v1", ping)
	decodedPing := &agentv1.Ping{}
	if err := out.GetPayload().UnmarshalTo(decodedPing); err != nil {
		t.Fatalf("ping unmarshal: %v", err)
	}
	if decodedPing.GetSequence() != 9 {
		t.Fatalf("ping sequence mismatch: %d", decodedPing.GetSequence())
	}

	resync := &agentv1.ResyncResponse{StateChecksum: "sha256:123", AppliedThroughSequence: 42}
	out = roundTripEvent(t, "inari.agent.resync-response.v1", resync)
	decodedResync := &agentv1.ResyncResponse{}
	if err := out.GetPayload().UnmarshalTo(decodedResync); err != nil {
		t.Fatalf("resync unmarshal: %v", err)
	}
	if decodedResync.GetAppliedThroughSequence() != 42 {
		t.Fatalf("resync mismatch: %v", decodedResync)
	}
}

func TestCommandPayloadsRoundTrip(t *testing.T) {
	params, err := structpb.NewStruct(map[string]any{"replicas": 3})
	if err != nil {
		t.Fatalf("structpb.NewStruct: %v", err)
	}
	commands := []struct {
		eventType string
		payload   proto.Message
	}{
		{"inari.agent.apply-bundle.v1", &agentv1.ApplyBundle{
			CommandId: "cmd-1",
			Source:    &agentv1.ApplyBundle_OciRef{OciRef: "oci://registry/bundle:v1"},
			Target:    &agentv1.GitTarget{Repo: "acme-inari-state", Path: "bundles/web"},
			Policy:    agentv1.CommitPolicy_COMMIT_POLICY_PULL_REQUEST,
			Checksum:  "sha256:aaa",
		}},
		{"inari.agent.register-argocd-app.v1", &agentv1.RegisterArgoCDApp{
			CommandId: "cmd-2",
			Name:      "web",
			Project:   "default",
			Source: &agentv1.ApplicationSource{
				RepoUrl: "https://git.example/repo", Path: "app", TargetRevision: "main",
			},
			DestinationServer:    "https://kubernetes.default.svc",
			DestinationNamespace: "web",
			SyncPolicy:           &agentv1.SyncPolicy{Automated: true, SelfHeal: true},
		}},
		{"inari.agent.invoke-action.v1", &agentv1.InvokeAction{
			CommandId:  "cmd-3",
			Action:     "restart",
			Resource:   &agentv1.ResourceRef{Kind: "Deployment", Name: "web", Namespace: "web"},
			Parameters: params,
			Timeout:    durationpb.New(30 * time.Second),
		}},
		{"inari.agent.render-rgd-instance.v1", &agentv1.RenderRgdInstance{
			CommandId:       "cmd-4",
			RgdRef:          "oci://registry/rgd/postgres:v1",
			InstanceName:    "db",
			TargetNamespace: "data",
			Parameters:      params,
			Target:          &agentv1.GitTarget{Repo: "acme-inari-state", Path: "instances/db"},
			Policy:          agentv1.CommitPolicy_COMMIT_POLICY_DIRECT_COMMIT,
		}},
		{"inari.agent.command-ack.v1", &agentv1.CommandAck{
			CommandId: "cmd-1",
			Result:    agentv1.CommandResult_COMMAND_RESULT_APPLIED,
		}},
	}
	for _, tc := range commands {
		t.Run(tc.eventType, func(t *testing.T) {
			out := roundTripEvent(t, tc.eventType, tc.payload)
			raw, err := proto.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload: %v", err)
			}
			if err := proto.Unmarshal(raw, tc.payload); err != nil {
				t.Fatalf("payload self round trip: %v", err)
			}
			if !out.GetPayload().MessageIs(tc.payload) {
				t.Fatalf("payload type mismatch for %s", tc.eventType)
			}
		})
	}
}

func TestEventTypeStringParity(t *testing.T) {
	want := map[agentv1.EventType]string{
		agentv1.EventType_EVENT_TYPE_CAPABILITY_UPDATE:   "inari.agent.capability-update.v1",
		agentv1.EventType_EVENT_TYPE_STATUS_UPDATE:       "inari.agent.status-update.v1",
		agentv1.EventType_EVENT_TYPE_PING:                "inari.agent.ping.v1",
		agentv1.EventType_EVENT_TYPE_PONG:                "inari.agent.pong.v1",
		agentv1.EventType_EVENT_TYPE_RESYNC_REQUEST:      "inari.agent.resync-request.v1",
		agentv1.EventType_EVENT_TYPE_RESYNC_RESPONSE:     "inari.agent.resync-response.v1",
		agentv1.EventType_EVENT_TYPE_APPLY_BUNDLE:        "inari.agent.apply-bundle.v1",
		agentv1.EventType_EVENT_TYPE_REGISTER_ARGOCD_APP: "inari.agent.register-argocd-app.v1",
		agentv1.EventType_EVENT_TYPE_INVOKE_ACTION:       "inari.agent.invoke-action.v1",
		agentv1.EventType_EVENT_TYPE_RENDER_RGD_INSTANCE: "inari.agent.render-rgd-instance.v1",
		agentv1.EventType_EVENT_TYPE_COMMAND_ACK:         "inari.agent.command-ack.v1",
		agentv1.EventType_EVENT_TYPE_COMMAND_NACK:        "inari.agent.command-nack.v1",
	}
	for enumVal, typeString := range want {
		if got := agentv1.EventTypeString(enumVal); got != typeString {
			t.Fatalf("EventTypeString(%v) = %q, want %q", enumVal, got, typeString)
		}
		if got := agentv1.EventTypeFromString(typeString); got != enumVal {
			t.Fatalf("EventTypeFromString(%q) = %v, want %v", typeString, got, enumVal)
		}
	}
	if got := agentv1.EventTypeFromString("inari.agent.does-not-exist.v1"); got != agentv1.EventType_EVENT_TYPE_UNSPECIFIED {
		t.Fatalf("unknown type string = %v, want UNSPECIFIED", got)
	}
}

func TestUnknownPayloadTolerance(t *testing.T) {
	// An N-1 agent must tolerate unknown Any payloads: unmarshal succeeds and
	// the payload is dropped-and-logged, never fatal (compatibility contract).
	in := &agentv1.Event{
		EventId: "evt-x",
		Type:    "inari.agent.future-thing.v9",
		Payload: &anypb.Any{TypeUrl: "type.googleapis.com/inari.agent.v1.FutureThing", Value: []byte{0x0a, 0x01, 0x78}},
	}
	raw, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := &agentv1.Event{}
	if err := proto.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal with unknown payload must not fail: %v", err)
	}
	if out.GetPayload().MessageIs(&agentv1.Ping{}) {
		t.Fatal("unknown payload must not match known types")
	}
}

func TestRegistrationMessages(t *testing.T) {
	req := &agentv1.RegisterClusterRequest{
		RegistrationToken: "tok-1",
		AgentVersion:      "0.2.0",
		ContractVersion:   "inari.agent.v1",
		ClusterLabels:     map[string]string{"env": "prod"},
		KubernetesVersion: "v1.34.0",
	}
	resp := &agentv1.RegisterClusterResponse{
		ClusterId:     "cluster-1",
		OidcIssuerUrl: "https://keycloak.example/realms/inari",
		ClientId:      "cluster-cluster-1",
		ClientSecretDelivery: &agentv1.SecretDeliveryReference{
			EsoSecretStore:  "inari-platform",
			SecretName:      "inari-agent-oidc",
			SecretNamespace: "inari-system",
			SecretKey:       "client-secret",
		},
	}
	for _, m := range []proto.Message{req, resp} {
		raw, err := proto.Marshal(m)
		if err != nil {
			t.Fatalf("marshal %T: %v", m, err)
		}
		if err := proto.Unmarshal(raw, m); err != nil {
			t.Fatalf("unmarshal %T: %v", m, err)
		}
	}
	if resp.GetClientSecretDelivery().GetSecretName() != "inari-agent-oidc" {
		t.Fatalf("delivery reference mismatch: %v", resp.GetClientSecretDelivery())
	}
}

func TestHandshakeVersionNegotiation(t *testing.T) {
	req := &agentv1.HandshakeRequest{
		AgentVersion:          "0.2.0",
		TenantId:              "t-1",
		ContractVersion:       "inari.agent.v1",
		LastSeenStateChecksum: "sha256:old",
	}
	resp := &agentv1.HandshakeResponse{
		SessionId:              "s-1",
		ServerContractVersions: "inari.agent.v1",
		ResyncRequired:         true,
	}
	if req.GetContractVersion() != "inari.agent.v1" || !resp.GetResyncRequired() {
		t.Fatalf("handshake negotiation fields missing: %v %v", req, resp)
	}
}
