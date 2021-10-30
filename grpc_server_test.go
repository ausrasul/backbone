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

package grpc_server

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
