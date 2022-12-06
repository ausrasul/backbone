// Example usage code.

package main

import (
	"log"
	"time"

	client "github.com/ausrasul/backbone/client"
	server "github.com/ausrasul/backbone/server"
)

func main() {
	go start_server()
	c := client.New("client1", ":1234")
	c.SetOnConnect(func(c *client.Client) {
		log.Println("client doing stuff when connected")
		c.SetCommandHandler("response", func(c *client.Client, arg string) {
			time.Sleep(time.Millisecond * 500)
			c.Send("test_command", "more stuff")
			log.Println("received response , ", arg)
		})
	})
	c.SetOnDisconnect(func(*client.Client) { log.Println("client disconnected!") })
	log.Println("client connecting...")
	c.Start()
	log.Println("client connected...")
	c.Send("test_command", "client is sending this command")
	<-time.After(time.Second * 5)

}

func start_server() {
	log.Println("starting server")
	s := server.New("localhost:1234")
	inChan := make(chan string, 5)
	s.SetOnConnect(func(str string) {
		log.Println("server: a client connected")
		inChan <- "something connected " + str
	})
	s.SetOnDisconnect(func(str string) {
		log.Println("server: client disconnected")
		inChan <- "something disconnect " + str
	})
	handler := func(client_id string, arg string) {
		log.Println("server handling cmd")
		inChan <- "command: " + client_id + " -- " + arg
		s.Send(client_id, "response", "cmdArg")
		log.Println("server handled cmd")
	}
	s.SetCommandHandler("client1", "test_command", handler)
	go s.Start()
	log.Println("started.")
	for in := range inChan {
		log.Println("svr got ", in)
	}
	log.Println("end.")

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
