package backbone

import (
	"context"
	"log"
	"net"
	"testing"

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
	when client call connect, it should be able to send commands to server
	and it should receive commands from server.
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
		assert.Equal(t, client.server_addr, data.addr, "Host not set")
		assert.Equal(t, client.id, data.id, "Id not set")
	}
}

func TestOnConnectCallback(t *testing.T) {
	callback_called := false
	testData := []struct {
		function    func()
		expect_call bool
		msg         string
	}{
		{
			function:    func() { callback_called = true },
			expect_call: true,
			msg:         "Call back should be called",
		},
		{
			function:    func() { callback_called = true },
			expect_call: false,
			msg:         "Call back should not be called",
		},
	}
	_ = callback_called
	client := New("123", ":1234")
	for _, data := range testData {
		client.SetOnConnect(data.function)
		if data.expect_call {
			client.onConnect()
		}
		assert.Equal(t, data.expect_call, callback_called, data.msg)
		callback_called = false
	}
}

func TestSetOnDisconnectCallback(t *testing.T) {
	callback_called := false
	testData := []struct {
		function    func()
		expect_call bool
		msg         string
	}{
		{
			function:    func() { callback_called = true },
			expect_call: true,
			msg:         "Call back should be called",
		},
		{
			function:    func() { callback_called = true },
			expect_call: false,
			msg:         "Call back should not be called",
		},
	}
	_ = callback_called
	client := New("123", "localhost:1234")
	for _, data := range testData {
		client.SetOnDisconnect(data.function)
		if data.expect_call {
			client.onDisconnect()
		}
		assert.Equal(t, data.expect_call, callback_called, data.msg)
		callback_called = false
	}
}
func TestClientSendCommand(t *testing.T) {
	tests := []struct {
		name       string
		cmdName    string
		cmdArg     string
		expectCall bool
		errMsg     string
	}{
		{
			"send command",
			"test_cmd",
			"1234",
			true,
			"command should be sent",
		},
		/*		{
					"Valid id",
					"id",
					"12345",
					"12345",
					"Id was not received",
				},
				{
					"Duplicate id",
					"id",
					"12345",
					"",
					"Duplicate Id, should not call on connect",
				},
				{
					"Not id",
					"not_id",
					"123",
					"",
					"Should not call on connect",
				},*/
	}
	c := New("123123", ":1234")
	s, conn := start_grpc_server()
	_, _ = s, conn

	//stream := connect_grpc(s, conn, "clientA")
	/*wait_connect := make(chan struct{})
	s.SetOnConnect(func(client_id string) {
		close(wait_connect)
	})
	client := comm.NewCommClient(conn)
	stream, _ := client.OpenComm(context.Background())
	stream.Send(&comm.Command{Name: "id", Arg: id})
	<-wait_connect
	return stream
	*/

	recvChan := make(chan string)
	c.SetOnConnect(func() { recvChan <- "connected" })
	_ = tests
	_ = c

}

func TestClientConnect(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		cmdArg   string
		expected string
		errMsg   string
	}{
		{
			"Valid id",
			"id",
			"1234",
			"1234",
			"Id was not received",
		},
		/*		{
					"Valid id",
					"id",
					"12345",
					"12345",
					"Id was not received",
				},
				{
					"Duplicate id",
					"id",
					"12345",
					"",
					"Duplicate Id, should not call on connect",
				},
				{
					"Not id",
					"not_id",
					"123",
					"",
					"Should not call on connect",
				},*/
	}
	c := New("123123", ":1234")
	s, conn := start_grpc_server()
	_, _ = s, conn

	//stream := connect_grpc(s, conn, "clientA")
	/*wait_connect := make(chan struct{})
	s.SetOnConnect(func(client_id string) {
		close(wait_connect)
	})
	client := comm.NewCommClient(conn)
	stream, _ := client.OpenComm(context.Background())
	stream.Send(&comm.Command{Name: "id", Arg: id})
	<-wait_connect
	return stream
	*/

	recvChan := make(chan string)
	c.SetOnConnect(func() { recvChan <- "connected" })
	_ = tests
	_ = c

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

func start_grpc_server() (*server.Server, *grpc.ClientConn) {
	ctx := context.Background()
	s := server.New("127.0.0.1:1234")
	dialer_ := dialer(s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	return s, conn
}
