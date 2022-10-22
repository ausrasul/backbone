package main

import (
	"context"
	"io"
	"log"

	"github.com/ausrasul/backbone/comm"
	server "github.com/ausrasul/backbone/server"
	"google.golang.org/grpc"
)

func main() {
	go start_server()

	//time.Sleep(time.Millisecond * 10)

	log.Println("client connecting")
	conn, err := grpc.Dial(":1234", grpc.WithInsecure())
	log.Println("client connected")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := comm.NewCommClient(conn)

	stream, err := client.OpenComm(context.Background())
	if err != nil {
		log.Fatalf("did not open comm: %v", err)
	}

	waitc := make(chan struct{})
	stream.Send(&comm.Command{Name: "id", Arg: "client1"})
	stream.Send(&comm.Command{Name: "test_command", Arg: "test args"})

	for {
		log.Println("client recieving cmd")
		in, err := stream.Recv()
		log.Println("client received cmd")
		if err == io.EOF {
			// read done.
			close(waitc)
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		log.Println("client got ", in.Name, in.Arg)
		stream.Send(&comm.Command{Name: "test_command", Arg: "test args"})
	}

	/*inChan := make(chan string, 10)
	c := client.New("client1", ":1234")
	c.SetOnConnect(func() { inChan <- "connecting" })
	c.SetOnDisconnect(func() { inChan <- "disconnected" })
	c.Start()
	*/
}

func start_server() {
	//time.Sleep(time.Millisecond * 1000)
	log.Println("starting server")
	s := server.New("localhost:1234")
	inChan := make(chan string, 5)
	s.SetOnConnect(func(str string) { inChan <- "something connected " + str })
	s.SetOnDisconnect(func(str string) { inChan <- "something disconnect " + str })
	handler := func(client_id string, arg string) {
		log.Println("server handling cmd")
		inChan <- "command: " + client_id + " -- " + arg
		s.Send(client_id, &comm.Command{Name: "response", Arg: "cmdArg"})
		log.Println("server handled cmd")
	}
	s.SetCommandHandler("client1", "test_command", handler)
	go s.Start()
	log.Println("started.")
	<-inChan
	/*for in := range inChan {
		log.Println("svr got ", in)
	}
	log.Println("end.")*/

}

/* this is how client should work
go start_server()
conn, err := grpc.Dial(":1234", grpc.WithInsecure())
client := comm.NewCommClient(conn)
stream, err := client.OpenComm(context.Background())
stream.Send(&comm.Command{Name: "id", Arg: "client1"}) // we have to send this as first command.
for {
	in, err := stream.Recv()
	if err == io.EOF {
		// read done.
		close(waitc)
		return
	}
	stream.Send(&comm.Command{Name: "test_command", Arg: "test args"})
}
*/
