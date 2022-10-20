package main

import (
	"io"
	"time"

	"github.com/ausrasul/backbone/comm"
	server "github.com/ausrasul/backbone/server"
	"github.com/stretchr/testify/assert"
)

func main() {
	go start_server()
	time.Sleep(time.Millisecond * 100)

	s, conn := start_grpc()
	stream := connect_grpc(s, conn, "clientA")
	s.Send("clientA", &comm.Command{Name: "testCommand", Arg: "cmdArg"})
	in, err := stream.Recv()
	assert.Equal(t, err, nil, "Should not receive err ")
	assert.Equal(t, in.Name, "testCommand", "Invalid command name received")
	assert.Equal(t, in.Arg, "cmdArg", "Invalid command arg received")
	// should send to different clients.
	stream2 := connect_grpc(s, conn, "clientB")
	s.Send("clientB", &comm.Command{Name: "testCommand2", Arg: "cmdArg2"})
	in, err = stream2.Recv()
	assert.Equal(t, err, nil, "Should not receive err ")
	assert.Equal(t, in.Name, "testCommand2", "Invalid command name received")
	assert.Equal(t, in.Arg, "cmdArg2", "Invalid command arg received")

	// send doesn't work wehn client disconnects.
	disconnected := make(chan int)
	s.SetOnDisconnect(func(client_id string) {
		disconnected <- 1
	})
	stream.CloseSend()
	<-disconnected
	s.Send("clientA", &comm.Command{Name: "testCommand", Arg: "cmdArg"})
	_, err = stream.Recv()
	assert.Equal(t, err, io.EOF, "client should only receive EOF on closure.")

	/*inChan := make(chan string, 10)
	c := client.New("client1", ":1234")
	c.SetOnConnect(func() { inChan <- "connecting" })
	c.SetOnDisconnect(func() { inChan <- "disconnected" })
	c.Start()
	*/
}

func start_server() {
	s := server.New("localhost:1234")
	inChan := make(chan string, 10)
	//s.SetOnConnect(func(str string) { inChan <- "connect" + str })
	//s.SetOnDisconnect(func(str string) { inChan <- "disconnect" + str })
	handler := func(client_id string, arg string) {
		inChan <- "command: " + client_id + " -- " + arg
		s.Send("client1", &comm.Command{Name: "testCommand", Arg: "cmdArg"})
	}
	s.SetCommandHandler("client1", "test_command", handler)
	select {
	case _ = <-inChan:
	case <-time.After(time.Millisecond * 100000):
	}
	/*for data := range inChan {
		log.Println(data)
	}*/

}
