package pluginv1_test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	pluginv1 "github.com/7K-Inari/inari-api/gen/go/inari/plugin/v1"
)

func TestPluginInfoRoundTrip(t *testing.T) {
	in := &pluginv1.GetInfoResponse{
		Info: &pluginv1.PluginInfo{
			Name:       "inari-plugin-example",
			Version:    "0.1.0",
			ApiVersion: "inari.plugin.v1",
		},
	}
	raw, err := proto.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := &pluginv1.GetInfoResponse{}
	if err := proto.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.GetInfo().GetApiVersion() != in.GetInfo().GetApiVersion() {
		t.Fatalf("round trip mismatch: got %v want %v", out, in)
	}
}

func TestHandshakeConfigFields(t *testing.T) {
	cfg := &pluginv1.HandshakeConfig{
		ProtocolVersion:  "1",
		MagicCookieKey:   "INARI_PLUGIN",
		MagicCookieValue: "inari",
	}
	if cfg.GetMagicCookieKey() == "" || cfg.GetProtocolVersion() == "" {
		t.Fatalf("handshake config fields not set: %v", cfg)
	}
}
