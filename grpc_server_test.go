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
	API- the server allow me to set handlers per client id
	GRPC- when client sends other commands, it calls the appropriate handler.

	- remember all connected clients.
	- delete disconnected clients.
	- handle race condition when sending command to a disconnected client.
	API- allow me to query how many clients are connected to it
	API- allow me to assign command handlers to individual clients.
	API- allow me to send commands to different clients.

	so that clients data can be handled by my code.
	should the server be able to shutdown? no.
	should it handle errors? yes!


	when a client "connects" it actually sends a command, and an input output streams are established.
	I should learn how they work.
	I should arrange so that data go in ot the stream while no race condition may happen.

	the following behaviours are needed:
	- instantiate with config
	- assign callbacks
		- onconnect(id string)
		- ondisconnect(id string)
		- onauth(id string) don't implement it.
	- server gives the following methods:
		- addHandler(id, cmdname, callback)
		- start()

	- start starts a grpc server.
	- on each connection start bidirectional stream.
	- probably use channels.

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

func TestOnDisconnectCallback(t *testing.T) {
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

func TestClientHandlers(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		cmdArg   string
		expected string
		errMsg   string
	}{
		{
			"Existing handler",
			"handle_this",
			"1234",
			"",
			"Existing handler was not called",
		},
		{
			"Existing handler 2",
			"handle_this_too",
			"12345",
			"",
			"Existing handler was not called",
		},
		{
			"Non existing handler",
			"unkown_command",
			"123",
			"",
			"No handler is called",
		},
	}
	s := New("127.0.0.1", 1234)
	recvChan := make(chan string)
	// From here on, must be rewritten.
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

// Test that it start grpc server.

// test that it creates a client one a client connects?
// test that it creates dual channel?
// not decided how this should work...

/*func TestAddHandler(t *testing.T){
	//varToBeChangedByCallBack := ""
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
		Server.AddHandler(data.clientId, data.command, data.callback)
		// what to expect?
		// every time a client receives a command that matches this one,
		// call this callback.
		// 1 - fake connect a client.
		// 2- fake receive a command.
		Server.onDisconnect(data.input)
		assert.Equal(t, varToBeChangedByCallBack, data.expected, "Bad onconnect callback")
	}

}*/
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
