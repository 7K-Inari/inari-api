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

func TestAuthMethodsRoundTrip(t *testing.T) {
	in := &pluginv1.GetInfoResponse{
		Info: &pluginv1.PluginInfo{Name: "inari-plugin-git", Version: "0.2.0", ApiVersion: "inari.plugin.v1"},
		AuthMethods: []*pluginv1.AuthMethod{
			{
				Type:      pluginv1.AuthMethod_TYPE_OIDC_USER,
				Audience:  "api://git.example",
				Scopes:    []string{"read", "write"},
				IsDefault: true,
			},
			{Type: pluginv1.AuthMethod_TYPE_SERVICE_ACCOUNT},
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
	methods := out.GetAuthMethods()
	if len(methods) != 2 {
		t.Fatalf("auth methods count: got %d want 2", len(methods))
	}
	if methods[0].GetType() != pluginv1.AuthMethod_TYPE_OIDC_USER ||
		methods[0].GetAudience() != "api://git.example" ||
		len(methods[0].GetScopes()) != 2 ||
		!methods[0].GetIsDefault() {
		t.Fatalf("auth method mismatch: %v", methods[0])
	}
}

func TestAuthMethodEnumValues(t *testing.T) {
	want := map[pluginv1.AuthMethod_Type]string{
		pluginv1.AuthMethod_TYPE_UNSPECIFIED:      "TYPE_UNSPECIFIED",
		pluginv1.AuthMethod_TYPE_OIDC_USER:        "TYPE_OIDC_USER",
		pluginv1.AuthMethod_TYPE_SERVICE_ACCOUNT:  "TYPE_SERVICE_ACCOUNT",
		pluginv1.AuthMethod_TYPE_API_KEY:          "TYPE_API_KEY",
		pluginv1.AuthMethod_TYPE_SHARED_SECRET:    "TYPE_SHARED_SECRET",
		pluginv1.AuthMethod_TYPE_OIDC_SSO_SESSION: "TYPE_OIDC_SSO_SESSION",
	}
	vals := pluginv1.AuthMethod_TYPE_UNSPECIFIED.Descriptor().Values()
	if len(want) != vals.Len() {
		t.Fatalf("enum coverage: map has %d, descriptor has %d", len(want), vals.Len())
	}
	for v, name := range want {
		if v.String() != name {
			t.Fatalf("enum value %d = %q, want %q", int32(v), v.String(), name)
		}
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
