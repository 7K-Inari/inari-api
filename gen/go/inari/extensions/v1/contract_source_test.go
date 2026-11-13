package extensionsv1

import (
	"os"
	"strings"
	"testing"
)

func readContractFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("required extension gateway contract file %s: %v", path, err)
	}
	return string(data)
}

func TestAgentGatewaySourceContract(t *testing.T) {
	proto := readContractFile(t, "../../../../../proto/inari/extensions/v1/agent_gateway.proto")
	for _, want := range []string{
		"package inari.extensions.v1;",
		"service AgentGateway",
		"rpc InvokeAction(inari.agent.v1.InvokeAction)",
		"returns (inari.agent.v1.CommandAck)",
	} {
		if !strings.Contains(proto, want) {
			t.Errorf("agent_gateway.proto missing %q", want)
		}
	}

	metadata := readContractFile(t, "metadata.go")
	for _, want := range []string{
		`MetadataTenant = "x-inari-tenant"`,
		`MetadataCluster = "x-inari-cluster"`,
		`MetadataUserCredential = "x-inari-user-credential"`,
		`MetadataUserCredentialRef = "x-inari-user-credential-ref"`,
	} {
		if !strings.Contains(metadata, want) {
			t.Errorf("metadata.go missing canonical constant %q", want)
		}
	}

	generated := readContractFile(t, "extensionsv1connect/agent_gateway.connect.go")
	if !strings.Contains(generated, `AgentGatewayInvokeActionProcedure = "/inari.extensions.v1.AgentGateway/InvokeAction"`) {
		t.Error("generated Connect client does not preserve the existing InvokeAction procedure")
	}
}
