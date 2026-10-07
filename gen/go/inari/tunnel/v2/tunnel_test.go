package tunnelv2_test

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

	tunnelv2 "github.com/7K-Inari/inari-api/gen/go/inari/tunnel/v2"
	"github.com/7K-Inari/inari-api/gen/go/inari/tunnel/v2/tunnelv2connect"
)

// TestTunnelMessagePayloads round-trips every oneof payload variant through
// proto marshal/unmarshal to pin the wire contract shape.
func TestTunnelMessagePayloads(t *testing.T) {
	msgs := []*tunnelv2.TunnelMessage{
		{ConnectionId: "conn-1", Payload: &tunnelv2.TunnelMessage_Open{Open: &tunnelv2.TunnelOpen{
			Method: "GET",
			Path:   "/api/v1/namespaces?limit=10",
			Headers: map[string]*tunnelv2.StringList{
				"Impersonate-User":  {Values: []string{"alice"}},
				"Impersonate-Group": {Values: []string{"/tenant-acme/devs", "/tenant-acme/ops"}},
			},
			UpgradeExpected: true,
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv2.TunnelMessage_OpenResult{OpenResult: &tunnelv2.TunnelOpenResult{
			Status:  101,
			Headers: map[string]*tunnelv2.StringList{"Upgrade": {Values: []string{"websocket"}}},
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv2.TunnelMessage_OpenResult{OpenResult: &tunnelv2.TunnelOpenResult{
			Error: "apiserver unreachable",
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv2.TunnelMessage_Frame{Frame: &tunnelv2.TunnelFrame{
			Data:      []byte("hello"),
			HalfClose: true,
		}}},
		{ConnectionId: "conn-1", Payload: &tunnelv2.TunnelMessage_Close{Close: &tunnelv2.TunnelClose{Reason: "done"}}},
		{ConnectionId: "conn-1", Payload: &tunnelv2.TunnelMessage_Ping{Ping: &tunnelv2.TunnelPing{
			Time: timestamppb.New(time.Unix(1000, 0).UTC()),
		}}},
	}
	for _, msg := range msgs {
		data, err := proto.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal %T: %v", msg.GetPayload(), err)
		}
		var got tunnelv2.TunnelMessage
		if err := proto.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %T: %v", msg.GetPayload(), err)
		}
		if !proto.Equal(msg, &got) {
			t.Fatalf("round-trip mismatch for %T: want %v, got %v", msg.GetPayload(), msg, &got)
		}
	}
}

// TestMultiValuedHeadersPreserved pins the v2 headline change: repeated
// header values (e.g. Set-Cookie) survive the wire un-joined.
func TestMultiValuedHeadersPreserved(t *testing.T) {
	msg := &tunnelv2.TunnelMessage{
		ConnectionId: "conn-1",
		Payload: &tunnelv2.TunnelMessage_OpenResult{OpenResult: &tunnelv2.TunnelOpenResult{
			Status: 200,
			Headers: map[string]*tunnelv2.StringList{
				"Set-Cookie": {Values: []string{"a=1; Path=/", "b=2; Path=/api"}},
			},
		}},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got tunnelv2.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	values := got.GetOpenResult().GetHeaders()["Set-Cookie"].GetValues()
	if len(values) != 2 || values[0] != "a=1; Path=/" || values[1] != "b=2; Path=/api" {
		t.Fatalf("multi-valued header not preserved: %v", values)
	}
}

// TestConnectStream wires a generated Connect client to a generated handler
// over httptest and echoes one TunnelOpen/TunnelFrame pair, proving the bidi
// stream contract compiles and round-trips end to end.
func TestConnectStream(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(tunnelv2connect.NewTunnelServiceHandler(echoTunnelHandler{}))
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
	client := tunnelv2connect.NewTunnelServiceClient(h2client, srv.URL)
	stream := client.Connect(context.Background())
	defer func() { _ = stream.CloseRequest() }()

	open := &tunnelv2.TunnelMessage{
		ConnectionId: "conn-1",
		Payload: &tunnelv2.TunnelMessage_Open{Open: &tunnelv2.TunnelOpen{
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
	stream *connect.BidiStream[tunnelv2.TunnelMessage, tunnelv2.TunnelMessage],
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
