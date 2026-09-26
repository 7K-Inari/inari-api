package agentv1_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
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
	req *connect.Request[agentv1.RedeemUserCredentialRequest],
) (*connect.Response[agentv1.RedeemUserCredentialResponse], error) {
	h.lastRef = req.Msg.GetCredentialRef()
	if h.lastRef == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("credential_ref is required"))
	}
	return connect.NewResponse(&agentv1.RedeemUserCredentialResponse{
		BearerToken: "tok-live",
		ExpiresAt:   timestamppb.New(time.Unix(3600, 0).UTC()),
	}), nil
}

func newCredentialsClient(t *testing.T, h agentv1connect.AgentCredentialsServiceHandler) agentv1connect.AgentCredentialsServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(agentv1connect.NewAgentCredentialsServiceHandler(h))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return agentv1connect.NewAgentCredentialsServiceClient(srv.Client(), srv.URL)
}

func TestRedeemUserCredentialOverTheWire(t *testing.T) {
	h := &stubCredentialsHandler{}
	client := newCredentialsClient(t, h)

	resp, err := client.RedeemUserCredential(context.Background(),
		connect.NewRequest(&agentv1.RedeemUserCredentialRequest{CredentialRef: "ucr-abc123"}))
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if resp.Msg.GetBearerToken() != "tok-live" {
		t.Fatalf("bearer token mismatch: %q", resp.Msg.GetBearerToken())
	}
	if resp.Msg.GetExpiresAt().GetSeconds() != 3600 {
		t.Fatalf("expires_at mismatch: %v", resp.Msg.GetExpiresAt())
	}
	if h.lastRef != "ucr-abc123" {
		t.Fatalf("handler saw ref %q, want ucr-abc123", h.lastRef)
	}
}

func TestRedeemUserCredentialEmptyRefRejected(t *testing.T) {
	client := newCredentialsClient(t, &stubCredentialsHandler{})

	_, err := client.RedeemUserCredential(context.Background(),
		connect.NewRequest(&agentv1.RedeemUserCredentialRequest{}))
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %v", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want invalid_argument", connectErr.Code())
	}
}

func TestRedeemUserCredentialUnimplemented(t *testing.T) {
	client := newCredentialsClient(t, agentv1connect.UnimplementedAgentCredentialsServiceHandler{})

	_, err := client.RedeemUserCredential(context.Background(),
		connect.NewRequest(&agentv1.RedeemUserCredentialRequest{CredentialRef: "ucr-1"}))
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeUnimplemented {
		t.Fatalf("expected unimplemented, got %v", err)
	}
}
