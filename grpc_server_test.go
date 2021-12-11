/*
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
	"log"
	"testing"

	"context"
	"net"

	pb "github.com/ausrasul/backbone/comm"
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

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	pb.RegisterCommServer(server, &server{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestDepositServer_Deposit(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		commandArg  string
		res         *pb.Command
		errMsg      string
	}{
		{
			"command with response",
			"test one",
			"test one arg",
			&pb.Command{
				name: "test one resp",
				arg:  "test one resp",
			},
			"got wrong resp",
		},
		/*{
			"valid request with non negative amount",
			0.00,
			&pb.DepositResponse{Ok: true},
			codes.OK,
			"",
		},*/
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewCommClient(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &pb.Command{
				name: tt.commandName,
				arg:  tt.commandArg,
			}
			response, err := client.OpenComm(ctx, request)

			if response != nil {
				if response.name != tt.res.name || response.arg != tt.res.arg {
					t.Error("response: expected", tt.res, "received", response)
				}
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
