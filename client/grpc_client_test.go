package backbone

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/ausrasul/backbone/comm"
	"github.com/ausrasul/backbone/server"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

/*
	DONE - instantiated with server address and port
	DONE - setOnConnect and onDisconnect
	after instantiation, we call client connect,
	that will do openComm and start bidirectional communication and identify itself.
	but this is implementation.
	the behavior is:
	when client call connect,
	1. it should be able to send commands to server (handled by server handler)
	2. and it should receive commands from server (handled by client handler)

	- connects to server by calling openComm
	- sends commands to server
	- add commands handlers
	- receives commands from server.

	NOTE: look at server test see how client should connect to it
*/

func TestInstantiateClient(t *testing.T) {
	testData := []struct {
		addr string
		id   string
	}{
		{addr: ":12345", id: "123"},
		{addr: "localhost1:123456", id: "1234"},
	}
	for _, data := range testData {
		client := New(data.id, data.addr)
		assert.IsType(t, &Client{}, client, "New client should be of type client")
		assert.Equal(t, client.serverAddr, data.addr, "Host not set")
		assert.Equal(t, client.id, data.id, "Id not set")
	}
}

func TestOnConnectCallback(t *testing.T) {
	callbackCalled := false
	testData := []struct {
		function   func()
		expectCall bool
		msg        string
	}{
		{
			function:   func() { callbackCalled = true },
			expectCall: true,
			msg:        "Call back should be called",
		},
		{
			function:   func() { callbackCalled = true },
			expectCall: false,
			msg:        "Call back should not be called",
		},
	}
	_ = callbackCalled
	client := New("123", ":1234")
	for _, data := range testData {
		client.SetOnConnect(data.function)
		if data.expectCall {
			client.onConnect()
		}
		assert.Equal(t, data.expectCall, callbackCalled, data.msg)
		callbackCalled = false
	}
}

func TestSetOnDisconnectCallback(t *testing.T) {
	callbackCalled := false
	testData := []struct {
		function   func()
		expectCall bool
		msg        string
	}{
		{
			function:   func() { callbackCalled = true },
			expectCall: true,
			msg:        "Call back should be called",
		},
		{
			function:   func() { callbackCalled = true },
			expectCall: false,
			msg:        "Call back should not be called",
		},
	}
	_ = callbackCalled
	client := New("123", "localhost:1234")
	for _, data := range testData {
		client.SetOnDisconnect(data.function)
		if data.expectCall {
			client.onDisconnect()
		}
		assert.Equal(t, data.expectCall, callbackCalled, data.msg)
		callbackCalled = false
	}
}

func TestClientSendCommand(t *testing.T) {
	tests := []struct {
		name     string
		clientId string
		cmdName  string
		cmdArg   string
		errMsg   string
	}{
		{
			name:     "send a command",
			clientId: "client1",
			cmdName:  "test_cmd",
			cmdArg:   "1234",
			errMsg:   "command should be handled by server",
		},
		{
			name:     "send a command",
			clientId: "client2",
			cmdName:  "test_command_2",
			cmdArg:   "12345",
			errMsg:   "command should be handled by server",
		},
	}
	for _, test := range tests {
		// prepare server
		s, conn := startGrpcServer()
		_, _ = s, conn
		cmdRecieved := make(chan string, 10)
		serverHandler := func(clientId string, arg string) {
			cmdRecieved <- arg
		}
		s.SetCommandHandler(test.clientId, test.cmdName, serverHandler)
		s.SetOnConnect(func(any string) {})
		// test client
		c := New(test.clientId, ":1234")
		c.connect(conn)
		assert.Nil(t, c.Send(test.cmdName, test.cmdArg))
		select {
		case <-cmdRecieved:
		case <-time.After(1 * time.Second):
			t.FailNow()
		}
	}
}

func TestClientReceiveCommand(t *testing.T) {
	tests := []struct {
		name            string
		clientId        string
		cmdName         string
		cmdArg          string
		cmdNameByServer string
		cmdArgByServer  string
		errMsg          string
	}{
		{
			name:            "send a command",
			clientId:        "client1",
			cmdName:         "test_cmd1",
			cmdArg:          "1234",
			cmdNameByServer: "test_cmd1",
			cmdArgByServer:  "1234",
			errMsg:          "command should be handled by server",
		},
		{
			name:            "send a command",
			clientId:        "client2",
			cmdName:         "test_command_21",
			cmdArg:          "12345",
			cmdNameByServer: "test_command_21",
			cmdArgByServer:  "12345",
			errMsg:          "command should be handled by server",
		},
	}
	for _, test := range tests {
		// prepare server
		s, conn := startGrpcServer()
		_, _ = s, conn
		s.SetCommandHandler(test.clientId, test.cmdName, func(any string, any_ string) {})
		s.SetOnConnect(func(any string) {})

		// test client
		cmdRecieved := make(chan string, 10)
		c := New(test.clientId, ":1234")
		c.SetCommandHandler(test.cmdName, func(args string) { cmdRecieved <- args })
		c.connect(conn)

		s.Send(test.clientId, test.cmdNameByServer, test.cmdArgByServer)
		select {
		case <-cmdRecieved:
		case <-time.After(5 * time.Second):
			t.FailNow()
		}
	}
}

/*
	dialer_ := dialer(s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, tt := range tests {
		client := comm.NewCommClient(conn)
		stream, _ := client.OpenComm(ctx)
		stream.Send(&comm.Command{Name: tt.cmdName, Arg: tt.cmdArg})
		res := ""
		select {
		case res = <-recvChan:
		case <-time.After(time.Millisecond * 10):
		}
		assert.Equal(t, tt.expected, res, tt.errMsg)

	}*/

func dialer(s *server.Server) func(context.Context, string) (net.Conn, error) {
	lis := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()

	comm.RegisterCommServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func startGrpcServer() (*server.Server, *grpc.ClientConn) {
	ctx := context.Background()
	s := server.New("127.0.0.1:1234")
	dialer_ := dialer(s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	return s, conn
}
