package extensionsv1

import "testing"

func TestAgentGatewayMetadataConstants(t *testing.T) {
	t.Parallel()
	got := map[string]string{
		"tenant":         MetadataTenant,
		"cluster":        MetadataCluster,
		"credential":     MetadataUserCredential,
		"credential_ref": MetadataUserCredentialRef,
	}
	want := map[string]string{
		"tenant":         "x-inari-tenant",
		"cluster":        "x-inari-cluster",
		"credential":     "x-inari-user-credential",
		"credential_ref": "x-inari-user-credential-ref",
	}
	for name, value := range want {
		if got[name] != value {
			t.Errorf("%s metadata = %q, want %q", name, got[name], value)
		}
	}
}
