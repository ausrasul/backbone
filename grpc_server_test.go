/*
	Background:
	DONE I need a package that establishes grpc server,
	DONE we should be able to:
	DONE API- instanciate it with ip and port
	DONE API- assign connect/disconnect handlers
	DONE API- tell it to start.

	DONE the reason we have to tell it to start, is to give us time to load on connect/disconnect handlers.

	DONE after it starts, it:
	DONE - allow grpc clients to automatically connect to it.
	DONE - when client sends id, it calls onconnect with id.
	DONE API- the server allow me to set handlers per client id
	DONE GRPC- when client sends other commands, it calls the appropriate handler.

	DONE - call onDisconnect when client disconnects.
	- delete disconnected clients' handlers
	- remember all connected clients. should I?
	- handle race condition when sending command to a disconnected client.
	API- allow me to query how many clients are connected to it
	API- allow me to send commands to different clients.

	so that clients data can be handled by my code.
	should the server be able to shutdown? no.
	should it handle errors? yes!


	when a client "connects" it actually sends a command, and an input output streams are established.
	I should arrange so that data go in ot the stream while no race condition may happen.

	v the following behaviours are needed:
	v - instantiate with config
	v - assign callbacks
		v - onconnect(id string)
*		- ondisconnect(id string)
		x - onauth(id string) don't implement it.
	- server gives the following methods:
		v - addHandler(id, cmdname, callback)
		v - start()

	v - start starts a grpc server.
	v - on each connection start bidirectional stream.
*	- probably use channels.

*/

package backbone

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/ausrasul/backbone/comm"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestInstantiate(t *testing.T) {
	testData := []struct {
		ip   string
		port int
	}{
		{ip: "localhost", port: 1234},
		{ip: "localhost1", port: 12345},
	}
	for _, data := range testData {
		Server := New(data.ip, data.port)
		assert.Equal(t, Server.ip, data.ip, "Ip not set")
		assert.Equal(t, Server.port, data.port, "Port not set")
	}
}

func TestOnConnectCallback(t *testing.T) {
	varToBeChangedByCallBack := ""
	testData := []struct {
		function func(string)
		input    string
		expected string
	}{
		{
			input:    "monkey",
			function: func(input string) { varToBeChangedByCallBack = "monkey" },
			expected: "monkey",
		},
	}
	Server := New("localhost", 1234)
	for _, data := range testData {
		Server.SetOnConnect(data.function)
		Server.onConnect(data.input)
		assert.Equal(t, varToBeChangedByCallBack, data.expected, "Bad onconnect callback")
	}
}

func TestSetOnDisconnectCallback(t *testing.T) {
	varToBeChangedByCallBack := ""
	testData := []struct {
		function func(string)
		input    string
		expected string
	}{
		{
			input:    "monkey",
			function: func(input string) { varToBeChangedByCallBack = "monkey" },
			expected: "monkey",
		},
	}
	Server := New("localhost", 1234)
	for _, data := range testData {
		Server.SetOnDisconnect(data.function)
		Server.onDisconnect(data.input)
		assert.Equal(t, varToBeChangedByCallBack, data.expected, "Bad onconnect callback")
	}
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
		{
			"Valid id",
			"id",
			"12345",
			"12345",
			"Id was not received",
		},
		{
			"Not id",
			"not_id",
			"123",
			"",
			"Should not call on connect",
		},
	}
	s := New("127.0.0.1", 1234)
	recvChan := make(chan string)
	s.SetOnConnect(func(str string) { recvChan <- str })

	ctx := context.Background()
	dialer_ := dialer(&s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := comm.NewCommClient(conn)
	stream, err := client.OpenComm(ctx)
	if err != nil {
		t.Error("got err", err)
	}
	for _, tt := range tests {
		stream.Send(&comm.Command{Name: tt.cmdName, Arg: tt.cmdArg})
		res := ""
		select {
		case res = <-recvChan:
		case <-time.After(time.Millisecond * 10):
		}
		assert.Equal(t, tt.expected, res, tt.errMsg)

	}

}
func TestOnDisconnect(t *testing.T) {
	s, conn := start_grpc()

	tests := []struct {
		name            string
		client_id       string
		will_disconnect bool
		err_msg         string
	}{
		{
			"client don't disconnects",
			"client1",
			false,
			"onDisconnect should not be called before client disconnect",
		},
		{
			"client disconnects",
			"client1",
			true,
			"onDisconnect should be called.",
		},
		{
			"client disconnects",
			"client2",
			true,
			"onDisconnect should be called.",
		},
	}
	for _, tt := range tests {
		stream := connect_grpc(s, conn, tt.client_id)
		disconnected := make(chan string)
		s.SetOnDisconnect(func(client_id string) { disconnected <- client_id })

		if tt.will_disconnect {
			stream.CloseSend()
		}
		select {
		case res := <-disconnected:
			if !tt.will_disconnect {
				t.Error(tt.name + ", " + tt.err_msg)
			} else {
				assert.Equal(t, res, tt.client_id, tt.name+", "+tt.err_msg)
			}
		case <-time.After(time.Millisecond * 10):
			if tt.will_disconnect {
				t.Error(tt.name + ", " + tt.err_msg)
			}
		}
	}

}

func TestCallCommandHandler(t *testing.T) {
	s, conn := start_grpc()
	wait_for_handler := make(chan map[string]string)

	handler := func(client_id string, arg string) {
		wait_for_handler <- map[string]string{"id": client_id, "arg": arg}
	}
	tests := []struct {
		name         string
		handler_cmd  string
		handler_func func(string, string)
		client_id    string
		cmd_name     string
		cmd_arg      string
	}{
		{
			"Valid handler test 1",
			"cmd1",
			handler,
			"client1",
			"cmd1",
			"ape",
		},
		{
			"Valid handler test 2",
			"cmd1",
			handler,
			"client1",
			"cmd1",
			"dog",
		},
		{
			"Valid handler test 3",
			"cmd2",
			handler,
			"client2",
			"cmd2",
			"dog",
		},
	}

	for _, tt := range tests {
		stream := connect_grpc(s, conn, tt.client_id)
		s.SetCommandHandler(tt.client_id, tt.handler_cmd, tt.handler_func)
		stream.Send(&comm.Command{Name: tt.cmd_name, Arg: tt.cmd_arg})
		res := <-wait_for_handler
		assert.Equal(t, res["id"], tt.client_id, tt.name+", handler should get client id")
		assert.Equal(t, res["arg"], tt.cmd_arg, tt.name+", handler should get cmd arg")
	}
}

func dialer(s *Server) func(context.Context, string) (net.Conn, error) {
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

func start_grpc() (*Server, *grpc.ClientConn) {
	ctx := context.Background()
	s := New("127.0.0.1", 1234)
	dialer_ := dialer(&s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	return &s, conn
}

func connect_grpc(s *Server, conn *grpc.ClientConn, id string) comm.Comm_OpenCommClient {
	wait_connect := make(chan struct{})
	s.SetOnConnect(func(client_id string) {
		close(wait_connect)
	})
	client := comm.NewCommClient(conn)
	stream, _ := client.OpenComm(context.Background())
	stream.Send(&comm.Command{Name: "id", Arg: id})
	<-wait_connect
	return stream
}
