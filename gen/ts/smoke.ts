import { create, fromBinary, toBinary } from "@bufbuild/protobuf";
import {
  ConnectRequestSchema,
  EventSchema,
  EventType,
} from "./proto/inari/agent/v1/agent_pb";
import {
  CapabilityAction,
  CapabilityKind,
  CapabilityUpdateSchema,
  ManagementMode,
  PingSchema,
  ResyncResponseSchema,
  StatusUpdateSchema,
} from "./proto/inari/agent/v1/events_pb";
import {
  ApplyBundleSchema,
  CommandAckSchema,
  CommandResult,
  CommitPolicy,
  InvokeActionSchema,
  RegisterArgoCDAppSchema,
  RenderRgdInstanceSchema,
} from "./proto/inari/agent/v1/commands_pb";
import {
  RegisterClusterRequestSchema,
  RegisterClusterResponseSchema,
  RegistrationService,
} from "./proto/inari/agent/v1/registration_pb";
import { EventStreamService } from "./proto/inari/agent/v1/agent_pb";
import {
  TunnelCloseSchema,
  TunnelMessageSchema,
  TunnelOpenSchema,
  TunnelService,
} from "./proto/inari/tunnel/v1/tunnel_pb";
import { getHealth, getGetClusterKubeconfigUrl } from "./rest/client";

const event = create(EventSchema, {
  eventId: "evt-1",
  resourceId: "res-1",
  type: "inari.agent.heartbeat.v1",
  sequence: 7n,
});
const bytes = toBinary(EventSchema, event);
const decoded = fromBinary(EventSchema, bytes);
if (decoded.eventId !== "evt-1") throw new Error("round trip failed");
if (decoded.sequence !== 7n) throw new Error("sequence round trip failed");

const req = create(ConnectRequestSchema, { event });
if (req.event?.type !== event.type) throw new Error("wrapper failed");

const cap = create(CapabilityUpdateSchema, {
  fullSync: true,
  stateChecksum: "sha256:abc",
  capabilities: [
    {
      kind: CapabilityKind.CRD,
      name: "widgets.example.com",
      managementMode: ManagementMode.OBSERVE_ONLY,
      action: CapabilityAction.UPSERT,
    },
  ],
});
if (cap.capabilities[0]?.kind !== CapabilityKind.CRD) throw new Error("capability failed");

const status = create(StatusUpdateSchema, { message: "ok" });
if (status.message !== "ok") throw new Error("status failed");

const ping = create(PingSchema, { sequence: 1n });
if (ping.sequence !== 1n) throw new Error("ping failed");

const resync = create(ResyncResponseSchema, { appliedThroughSequence: 42n });
if (resync.appliedThroughSequence !== 42n) throw new Error("resync failed");

const bundle = create(ApplyBundleSchema, {
  commandId: "cmd-1",
  source: { case: "ociRef", value: "oci://registry/bundle:v1" },
  policy: CommitPolicy.PULL_REQUEST,
});
if (bundle.source.case !== "ociRef") throw new Error("apply-bundle failed");

const app = create(RegisterArgoCDAppSchema, { commandId: "cmd-2", name: "web" });
const action = create(InvokeActionSchema, { commandId: "cmd-3", action: "restart" });
const rgd = create(RenderRgdInstanceSchema, { commandId: "cmd-4", rgdRef: "oci://registry/rgd:v1" });
const ack = create(CommandAckSchema, { commandId: "cmd-1", result: CommandResult.APPLIED });
void [app, action, rgd, ack];

const regReq = create(RegisterClusterRequestSchema, { registrationToken: "tok-1" });
const regResp = create(RegisterClusterResponseSchema, { clientId: "cluster-1" });
if (regReq.registrationToken !== "tok-1" || regResp.clientId !== "cluster-1") {
  throw new Error("registration failed");
}

if (EventType.CAPABILITY_UPDATE !== 1) throw new Error("enum failed");

const tunnelOpen = create(TunnelMessageSchema, {
  connectionId: "conn-1",
  payload: { case: "open", value: create(TunnelOpenSchema, { method: "GET", path: "/api/v1/pods" }) },
});
const tunnelBytes = toBinary(TunnelMessageSchema, tunnelOpen);
const tunnelDecoded = fromBinary(TunnelMessageSchema, tunnelBytes);
if (tunnelDecoded.payload.case !== "open" || tunnelDecoded.payload.value.method !== "GET") {
  throw new Error("tunnel open round trip failed");
}

const tunnelClose = create(TunnelMessageSchema, {
  connectionId: "conn-1",
  payload: { case: "close", value: create(TunnelCloseSchema, { reason: "done" }) },
});
if (tunnelClose.payload.case !== "close") throw new Error("tunnel close payload failed");

void TunnelService;
void getGetClusterKubeconfigUrl;

void RegistrationService;
void EventStreamService;
void getHealth;
console.log("smoke ok");
