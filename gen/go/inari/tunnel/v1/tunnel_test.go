package tunnelv1_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	tunnelv1 "github.com/7K-Inari/inari-api/gen/go/inari/tunnel/v1"
	"github.com/7K-Inari/inari-api/gen/go/inari/tunnel/v1/tunnelv1connect"
)

// TestTunnelMessagePayloads round-trips every oneof payload variant through
// proto marshal/unmarshal to pin the wire contract shape.
func TestTunnelMessagePayloads(t *testing.T) {
	msgs := []*tunnelv1.TunnelMessage{
		{ConnectionId: "conn-1", Payload: &tunnelv1.TunnelMessage_Open{Open: &tunnelv1.TunnelOpen{
			Method:          "GET",
			Path:            "/api/v1/namespaces?limit=10",
			Headers:         map[string]string{"Impersonate-User": "alice"},
			UpgradeExpected: true,
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv1.TunnelMessage_OpenResult{OpenResult: &tunnelv1.TunnelOpenResult{
			Status:  101,
			Headers: map[string]string{"Upgrade": "websocket"},
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv1.TunnelMessage_OpenResult{OpenResult: &tunnelv1.TunnelOpenResult{
			Error: "apiserver unreachable",
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv1.TunnelMessage_Frame{Frame: &tunnelv1.TunnelFrame{
			Data:      []byte("hello"),
			HalfClose: true,
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv1.TunnelMessage_Close{Close: &tunnelv1.TunnelClose{Reason: "done"}}},
		{ConnectionId: "conn-1", Payload: &tunnelv1.TunnelMessage_Ping{Ping: &tunnelv1.TunnelPing{
			Time: timestamppb.New(time.Unix(1000, 0).UTC()),
		}}},
	}
	for _, msg := range msgs {
		data, err := proto.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal %T: %v", msg.GetPayload(), err)
		}
		var got tunnelv1.TunnelMessage
		if err := proto.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %T: %v", msg.GetPayload(), err)
		}
		if !proto.Equal(msg, &got) {
			t.Fatalf("round-trip mismatch for %T: want %v, got %v", msg.GetPayload(), msg, &got)
		}
	}
}

// TestConnectStream wires a generated Connect client to a generated handler
// over httptest and echoes one TunnelOpen/TunnelFrame pair, proving the bidi
// stream contract compiles and round-trips end to end.
func TestConnectStream(t *testing.T) {
	handler := tunnelv1connect.UnimplementedTunnelServiceHandler{}
	_ = handler // keep the unimplemented type referenced like other packages do

	mux := http.NewServeMux()
	mux.Handle(tunnelv1connect.NewTunnelServiceHandler(echoTunnelHandler{}))
	// Bidi streaming requires HTTP/2; serve h2c and dial with an h2 client.
	srv := httptest.NewUnstartedServer(h2c.NewHandler(mux, &http2.Server{}))
	srv.Start()
	t.Cleanup(srv.Close)

	h2client := &http.Client{Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}}
	client := tunnelv1connect.NewTunnelServiceClient(h2client, srv.URL)
	stream := client.Connect(context.Background())
	defer func() { _ = stream.CloseRequest() }()

	open := &tunnelv1.TunnelMessage{
		ConnectionId: "conn-1",
		Payload: &tunnelv1.TunnelMessage_Open{Open: &tunnelv1.TunnelOpen{
			Method: "GET",
			Path:   "/api/v1/pods",
		}},
	}
	if err := stream.Send(open); err != nil {
		t.Fatalf("send open: %v", err)
	}
	if err := stream.CloseRequest(); err != nil {
		t.Fatalf("close request: %v", err)
	}

	got, err := stream.Receive()
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if got.GetConnectionId() != "conn-1" || got.GetOpen().GetPath() != "/api/v1/pods" {
		t.Fatalf("unexpected echo: %v", got)
	}
	if _, err := stream.Receive(); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after echo, got %v", err)
	}
}

type echoTunnelHandler struct{}

func (echoTunnelHandler) Connect(
	_ context.Context,
	stream *connect.BidiStream[tunnelv1.TunnelMessage, tunnelv1.TunnelMessage],
) error {
	for {
		msg, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
}
