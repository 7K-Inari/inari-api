import { ConnectRequest, Event } from "./proto/inari/agent/v1/agent_pb";
import { getHealth } from "./rest/client";

const event = new Event({
  eventId: "evt-1",
  resourceId: "res-1",
  type: "inari.agent.heartbeat.v1",
});
const bytes = event.toBinary();
const decoded = Event.fromBinary(bytes);
if (decoded.eventId !== "evt-1") throw new Error("round trip failed");

const req = new ConnectRequest({ event });
if (req.event?.type !== event.type) throw new Error("wrapper failed");

void getHealth;
console.log("smoke ok");
