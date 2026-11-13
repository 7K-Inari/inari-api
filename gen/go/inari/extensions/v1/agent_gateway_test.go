package extensionsv1_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectinprocess"
	agentv1 "github.com/7K-Inari/inari-api/gen/go/inari/agent/v1"
	"github.com/7K-Inari/inari-api/gen/go/inari/extensions/v1/extensionsv1connect"
)

type stubAgentGateway struct {
	extensionsv1connect.UnimplementedAgentGatewayHandler
	last *agentv1.InvokeAction
}

func (s *stubAgentGateway) InvokeAction(_ context.Context, req *agentv1.InvokeAction) (*agentv1.CommandAck, error) {
	s.last = req
	return &agentv1.CommandAck{
		CommandId: req.GetCommandId(),
		Result:    agentv1.CommandResult_COMMAND_RESULT_APPLIED,
		Message:   "applied",
	}, nil
}

func TestAgentGatewayProcedureAndRoundTrip(t *testing.T) {
	if got, want := extensionsv1connect.AgentGatewayInvokeActionProcedure, "/inari.extensions.v1.AgentGateway/InvokeAction"; got != want {
		t.Fatalf("procedure = %q, want %q", got, want)
	}
	if got, want := extensionsv1connect.AgentGatewayName, "inari.extensions.v1.AgentGateway"; got != want {
		t.Fatalf("service = %q, want %q", got, want)
	}

	handler := &stubAgentGateway{}
	server := connect.NewServer()
	extensionsv1connect.RegisterAgentGatewayHandler(server, handler)
	client := extensionsv1connect.NewAgentGatewayClient(connect.NewClient(connectinprocess.New(server)))

	ack, err := client.InvokeAction(context.Background(), &agentv1.InvokeAction{
		CommandId: "command-1",
		Action:    "sync",
	})
	if err != nil {
		t.Fatalf("InvokeAction: %v", err)
	}
	if ack.GetCommandId() != "command-1" || ack.GetResult() != agentv1.CommandResult_COMMAND_RESULT_APPLIED {
		t.Fatalf("ack = %+v", ack)
	}
	if handler.last.GetAction() != "sync" {
		t.Fatalf("handler command = %+v", handler.last)
	}
}

func TestAgentGatewayUnimplemented(t *testing.T) {
	server := connect.NewServer()
	extensionsv1connect.RegisterAgentGatewayHandler(server, extensionsv1connect.UnimplementedAgentGatewayHandler{})
	client := extensionsv1connect.NewAgentGatewayClient(connect.NewClient(connectinprocess.New(server)))

	_, err := client.InvokeAction(context.Background(), &agentv1.InvokeAction{})
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) || connectErr.Code() != connect.CodeUnimplemented {
		t.Fatalf("err = %v, want UNIMPLEMENTED", err)
	}
}
