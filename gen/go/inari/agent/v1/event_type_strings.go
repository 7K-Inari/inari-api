package agentv1

var eventTypeStrings = map[EventType]string{
	EventType_EVENT_TYPE_CAPABILITY_UPDATE:   "inari.agent.capability-update.v1",
	EventType_EVENT_TYPE_STATUS_UPDATE:       "inari.agent.status-update.v1",
	EventType_EVENT_TYPE_PING:                "inari.agent.ping.v1",
	EventType_EVENT_TYPE_PONG:                "inari.agent.pong.v1",
	EventType_EVENT_TYPE_RESYNC_REQUEST:      "inari.agent.resync-request.v1",
	EventType_EVENT_TYPE_RESYNC_RESPONSE:     "inari.agent.resync-response.v1",
	EventType_EVENT_TYPE_APPLY_BUNDLE:        "inari.agent.apply-bundle.v1",
	EventType_EVENT_TYPE_REGISTER_ARGOCD_APP: "inari.agent.register-argocd-app.v1",
	EventType_EVENT_TYPE_INVOKE_ACTION:       "inari.agent.invoke-action.v1",
	EventType_EVENT_TYPE_RENDER_RGD_INSTANCE: "inari.agent.render-rgd-instance.v1",
	EventType_EVENT_TYPE_COMMAND_ACK:         "inari.agent.command-ack.v1",
	EventType_EVENT_TYPE_COMMAND_NACK:        "inari.agent.command-nack.v1",
}

var eventTypeByString = func() map[string]EventType {
	m := make(map[string]EventType, len(eventTypeStrings))
	for k, v := range eventTypeStrings {
		m[v] = k
	}
	return m
}()

func EventTypeString(t EventType) string {
	return eventTypeStrings[t]
}

func EventTypeFromString(s string) EventType {
	if t, ok := eventTypeByString[s]; ok {
		return t
	}
	return EventType_EVENT_TYPE_UNSPECIFIED
}
