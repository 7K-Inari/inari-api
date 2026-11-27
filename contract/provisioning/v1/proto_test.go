package provisioning

import (
	"testing"

	planlimitsv1 "github.com/7K-Inari/inari-api/gen/go/inari/planlimits/v1"
	"github.com/7K-Inari/inari-api/gen/go/inari/provisioning/v1/provisioningv1connect"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestLimitsProtoRoundTripPreservesZeroAndUnlimited(t *testing.T) {
	zero := int64(0)
	input := &Limits{MaxClusters: &zero, FeatureGates: map[string]bool{"policy-packs": true}}
	wire, err := LimitsToProto(input)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := protojson.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded planlimitsv1.PlanLimits
	if err := protojson.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.MaxClusters == nil || *decoded.MaxClusters != 0 {
		t.Fatalf("explicit zero lost in %s", raw)
	}
	if decoded.MaxDeveloperSeats != nil {
		t.Fatal("absent seat limit became present")
	}
	output, err := LimitsFromProto(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	*output.MaxClusters = 99
	output.FeatureGates["policy-packs"] = false
	if *input.MaxClusters != 0 || !input.FeatureGates["policy-packs"] {
		t.Fatal("conversion aliased mutable input")
	}
}

func TestConnectServiceContract(t *testing.T) {
	if got := provisioningv1connect.ProvisioningServiceName; got != "inari.provisioning.v1.ProvisioningService" {
		t.Fatalf("service name = %q", got)
	}
	if got := provisioningv1connect.ProvisioningServiceCreateTenantProcedure; got != "/inari.provisioning.v1.ProvisioningService/CreateTenant" {
		t.Fatalf("create procedure = %q", got)
	}
	if IdempotencyKeyHeader != "Idempotency-Key" || ProvisioningScope != "provisioning" {
		t.Fatal("unexpected wire metadata constants")
	}
}
