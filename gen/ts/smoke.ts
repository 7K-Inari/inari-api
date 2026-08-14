import { ConnectRequest, Event, EventType } from "./proto/inari/agent/v1/agent_pb";
import {
  CapabilityAction,
  CapabilityKind,
  CapabilityUpdate,
  ManagementMode,
  Ping,
  ResyncResponse,
  StatusUpdate,
} from "./proto/inari/agent/v1/events_pb";
import {
  ApplyBundle,
  CommandAck,
  CommandResult,
  CommitPolicy,
  InvokeAction,
  RegisterArgoCDApp,
  RenderRgdInstance,
} from "./proto/inari/agent/v1/commands_pb";
import {
  RegisterClusterRequest,
  RegisterClusterResponse,
} from "./proto/inari/agent/v1/registration_pb";
import { RegistrationService } from "./proto/inari/agent/v1/registration_connect";
import { EventStreamService } from "./proto/inari/agent/v1/agent_connect";
import { getHealth } from "./rest/client";

const event = new Event({
  eventId: "evt-1",
  resourceId: "res-1",
  type: "inari.agent.heartbeat.v1",
  sequence: 7n,
});
const bytes = event.toBinary();
const decoded = Event.fromBinary(bytes);
if (decoded.eventId !== "evt-1") throw new Error("round trip failed");
if (decoded.sequence !== 7n) throw new Error("sequence round trip failed");

const req = new ConnectRequest({ event });
if (req.event?.type !== event.type) throw new Error("wrapper failed");

const cap = new CapabilityUpdate({
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

const status = new StatusUpdate({ message: "ok" });
if (status.message !== "ok") throw new Error("status failed");

const ping = new Ping({ sequence: 1n });
if (ping.sequence !== 1n) throw new Error("ping failed");

const resync = new ResyncResponse({ appliedThroughSequence: 42n });
if (resync.appliedThroughSequence !== 42n) throw new Error("resync failed");

const bundle = new ApplyBundle({
  commandId: "cmd-1",
  source: { case: "ociRef", value: "oci://registry/bundle:v1" },
  policy: CommitPolicy.PULL_REQUEST,
});
if (bundle.source.case !== "ociRef") throw new Error("apply-bundle failed");

const app = new RegisterArgoCDApp({ commandId: "cmd-2", name: "web" });
const action = new InvokeAction({ commandId: "cmd-3", action: "restart" });
const rgd = new RenderRgdInstance({ commandId: "cmd-4", rgdRef: "oci://registry/rgd:v1" });
const ack = new CommandAck({ commandId: "cmd-1", result: CommandResult.APPLIED });
void [app, action, rgd, ack];

const regReq = new RegisterClusterRequest({ registrationToken: "tok-1" });
const regResp = new RegisterClusterResponse({ clientId: "cluster-1" });
if (regReq.registrationToken !== "tok-1" || regResp.clientId !== "cluster-1") {
  throw new Error("registration failed");
}

if (EventType.CAPABILITY_UPDATE !== 1) throw new Error("enum failed");

void RegistrationService;
void EventStreamService;
void getHealth;
console.log("smoke ok");
