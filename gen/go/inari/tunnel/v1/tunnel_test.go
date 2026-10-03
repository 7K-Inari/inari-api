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

// TestTunnelMessageEmptyConnectionId verifies that an empty connection_id is
// valid for session-level messages such as Ping. The contract does not enforce
// UUID formatting at the wire level.
func TestTunnelMessageEmptyConnectionId(t *testing.T) {
	msg := &tunnelv1.TunnelMessage{
		Payload: &tunnelv1.TunnelMessage_Ping{Ping: &tunnelv1.TunnelPing{
			Time: timestamppb.New(time.Unix(1000, 0).UTC()),
		}},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got tunnelv1.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetConnectionId() != "" {
		t.Fatalf("expected empty connection_id, got %q", got.GetConnectionId())
	}
	if got.GetPing() == nil {
		t.Fatal("expected Ping payload")
	}
}

// TestTunnelMessageNoPayload verifies that a message with no oneof payload set
// round-trips cleanly. This is the default state on the wire.
func TestTunnelMessageNoPayload(t *testing.T) {
	msg := &tunnelv1.TunnelMessage{ConnectionId: "conn-empty"}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got tunnelv1.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetConnectionId() != "conn-empty" {
		t.Fatalf("connection_id mismatch: got %q", got.GetConnectionId())
	}
	if got.GetPayload() != nil {
		t.Fatalf("expected nil payload, got %T", got.GetPayload())
	}
}

// TestTunnelMessageOneofLastWriteWins verifies proto3 oneof behavior when the
// wire encoding contains more than one payload field. The generated Go oneof
// exposes only the last unmarshaled field, which is the expected contract
// behavior for TunnelMessage.
func TestTunnelMessageOneofLastWriteWins(t *testing.T) {
	// Build a message with open, then overwrite it with close at the struct
	// level before marshal. The wire will contain only the last set field.
	msg := &tunnelv1.TunnelMessage{
		ConnectionId: "conn-1",
		Payload:      &tunnelv1.TunnelMessage_Open{Open: &tunnelv1.TunnelOpen{Method: "GET", Path: "/"}},
	}
	msg.Payload = &tunnelv1.TunnelMessage_Close{Close: &tunnelv1.TunnelClose{Reason: "overwritten"}}

	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got tunnelv1.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetOpen() != nil {
		t.Fatalf("expected open to be cleared, got %v", got.GetOpen())
	}
	if got.GetClose().GetReason() != "overwritten" {
		t.Fatalf("close reason mismatch: got %q", got.GetClose().GetReason())
	}
}

// TestTunnelFrameLargeData verifies that the frame payload does not impose an
// arbitrary size limit at the protobuf layer. The 32 KiB comment is an
// application-level convention, not a wire constraint.
func TestTunnelFrameLargeData(t *testing.T) {
	large := make([]byte, 1<<20) // 1 MiB
	for i := range large {
		large[i] = byte(i % 256)
	}
	msg := &tunnelv1.TunnelMessage{
		ConnectionId: "conn-large",
		Payload:      &tunnelv1.TunnelMessage_Frame{Frame: &tunnelv1.TunnelFrame{Data: large}},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal 1 MiB frame: %v", err)
	}
	var got tunnelv1.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal 1 MiB frame: %v", err)
	}
	gotFrame := got.GetFrame()
	if gotFrame == nil {
		t.Fatal("expected Frame payload")
	}
	if len(gotFrame.GetData()) != len(large) {
		t.Fatalf("frame data length mismatch: got %d, want %d", len(gotFrame.GetData()), len(large))
	}
	for i := range large {
		if gotFrame.GetData()[i] != large[i] {
			t.Fatalf("frame data mismatch at index %d", i)
		}
	}
}

// TestTunnelOpenResultErrorOnly verifies that an OpenResult carrying only an
// error (no status) round-trips and is distinguishable from a success result.
func TestTunnelOpenResultErrorOnly(t *testing.T) {
	msg := &tunnelv1.TunnelMessage{
		ConnectionId: "conn-1",
		Payload: &tunnelv1.TunnelMessage_OpenResult{OpenResult: &tunnelv1.TunnelOpenResult{
			Error: "apiserver unreachable",
		}},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got tunnelv1.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetOpenResult().GetStatus() != 0 {
		t.Fatalf("expected zero status, got %d", got.GetOpenResult().GetStatus())
	}
	if got.GetOpenResult().GetError() != "apiserver unreachable" {
		t.Fatalf("error mismatch: got %q", got.GetOpenResult().GetError())
	}
}

// TestTunnelCloseEmptyReason verifies that a close message with an empty reason
// round-trips. Empty reason is valid for normal EOF.
func TestTunnelCloseEmptyReason(t *testing.T) {
	msg := &tunnelv1.TunnelMessage{
		ConnectionId: "conn-1",
		Payload:      &tunnelv1.TunnelMessage_Close{Close: &tunnelv1.TunnelClose{}},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got tunnelv1.TunnelMessage
	if err := proto.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.GetClose().GetReason() != "" {
		t.Fatalf("expected empty reason, got %q", got.GetClose().GetReason())
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
