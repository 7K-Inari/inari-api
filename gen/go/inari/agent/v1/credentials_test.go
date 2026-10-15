package agentv1_test

import (
	"context"
	"errors"
	"testing"
	"time"

	connect "connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectinprocess"
	"google.golang.org/protobuf/types/known/timestamppb"

	agentv1 "github.com/7K-Inari/inari-api/gen/go/inari/agent/v1"
	"github.com/7K-Inari/inari-api/gen/go/inari/agent/v1/agentv1connect"
)

type stubCredentialsHandler struct {
	agentv1connect.UnimplementedAgentCredentialsServiceHandler
	lastRef string
}

func (h *stubCredentialsHandler) RedeemUserCredential(
	_ context.Context,
	req *agentv1.RedeemUserCredentialRequest,
) (*agentv1.RedeemUserCredentialResponse, error) {
	h.lastRef = req.GetCredentialRef()
	if h.lastRef == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "credential_ref is required")
	}
	return &agentv1.RedeemUserCredentialResponse{
		BearerToken: "tok-live",
		ExpiresAt:   timestamppb.New(time.Unix(3600, 0).UTC()),
	}, nil
}

func newCredentialsClient(h agentv1connect.AgentCredentialsServiceHandler) agentv1connect.AgentCredentialsServiceClient {
	server := connect.NewServer()
	agentv1connect.RegisterAgentCredentialsServiceHandler(server, h)
	return agentv1connect.NewAgentCredentialsServiceClient(connect.NewClient(connectinprocess.New(server)))
}

func TestRedeemUserCredentialOverTheWire(t *testing.T) {
	h := &stubCredentialsHandler{}
	client := newCredentialsClient(h)

	resp, err := client.RedeemUserCredential(context.Background(),
		&agentv1.RedeemUserCredentialRequest{CredentialRef: "ucr-abc123"})
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if resp.GetBearerToken() != "tok-live" {
		t.Fatalf("bearer token mismatch: %q", resp.GetBearerToken())
	}
	if resp.GetExpiresAt().GetSeconds() != 3600 {
		t.Fatalf("expires_at mismatch: %v", resp.GetExpiresAt())
	}
	if h.lastRef != "ucr-abc123" {
		t.Fatalf("handler saw ref %q, want ucr-abc123", h.lastRef)
	}
}

func TestRedeemUserCredentialEmptyRefRejected(t *testing.T) {
	client := newCredentialsClient(&stubCredentialsHandler{})

	_, err := client.RedeemUserCredential(context.Background(),
		&agentv1.RedeemUserCredentialRequest{})
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %v", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want invalid_argument", connectErr.Code())
	}
}

func TestRedeemUserCredentialUnimplemented(t *testing.T) {
	client := newCredentialsClient(agentv1connect.UnimplementedAgentCredentialsServiceHandler{})

	_, err := client.RedeemUserCredential(context.Background(),
		&agentv1.RedeemUserCredentialRequest{CredentialRef: "ucr-1"})
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeUnimplemented {
		t.Fatalf("expected unimplemented, got %v", err)
	}
}
