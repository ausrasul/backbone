/*
	Background:
	we need a package that establishes grpc server,
	we should be able to:
	API- instanciate it with ip and port
	API- assign connect/disconnect handlers
	API- tell it to start.

	the reason we have to tell it to start, is to give us time to
	load on connect/disconnect handlers.

	after it starts, it:
	- allow grpc clients to automatically connect to it.
	API- allow me to query how many clients are connected to it
	API- allow me to assign command handlers to individual clients.

	so that clients data can be handled by my code.
	should the server be able to shutdown? no.
	should it handle errors? yes!


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
	"fmt"
	"log"
	"net"
	"testing"

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

/*func TestStart(t *testing.T) {
	server := New("localhost", 1234)
	assert.Equal(t, server.Start(), nil, "Grpc didn't start")

}
*/

func dialer(s *Server) func(context.Context, string) (net.Conn, error) {
	lis := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()

	comm.RegisterCommServer(grpcServer, s)
	//pb.RegisterDepositServiceServer(server, &DepositServer{})
	//log.Print("asdf")

	//grpcServer.Serve(lis)
	log.Print("asdf")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Print("asdf")

	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestCommServer_Deposit(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		cmdArg  string
		res     *comm.Command
		errMsg  string
	}{
		{
			"Command",
			"testCmd",
			"testArg",
			&comm.Command{},
			fmt.Sprintf("cannot deposit %v", -1.11),
		},
	}
	s := New("127.0.0.1", 1234)
	ctx := context.Background()
	log.Print("asdf1")
	dialer_ := dialer(&s)
	log.Print("asdf1")

	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	log.Print("asdf2")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := comm.NewCommClient(conn)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &comm.Command{Name: tt.cmdName, Arg: tt.cmdArg}

			response, err := client.OpenComm(ctx, request)

			if response != nil {
				if response.Name != tt.res.Name {
					t.Error("response: expected", tt.res.Name, "received", response.Name)
				}
			}

			if err != nil {
				t.Error("unknown err", err)
			}
		})
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
