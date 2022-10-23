package backbone

import (
	"context"
	"log"

	"github.com/ausrasul/backbone/comm"
	"google.golang.org/grpc"
)

type Client struct {
	serverAddr     string
	id             string
	onConnect      func()
	onDisconnect   func()
	commandHanlder map[string]func(string)
	stream         comm.Comm_OpenCommClient
}

func New(id string, addr string) *Client {
	return &Client{
		serverAddr:     addr,
		id:             id,
		commandHanlder: map[string]func(string){},
	}
}

func (c *Client) SetOnConnect(callback func()) {
	c.onConnect = callback
}

func (c *Client) SetOnDisconnect(callback func()) {
	c.onDisconnect = callback
}

func (c *Client) SetCommandHandler(commandName string, callback func(string)) {
	c.commandHanlder[commandName] = callback
}
func (c *Client) Start() {
	conn := getConnection(c.serverAddr)
	defer conn.Close()
	c.connect(conn)
}

func (c *Client) connect(conn *grpc.ClientConn) {
	client := comm.NewCommClient(conn)
	stream, err := client.OpenComm(context.Background())
	if err != nil {
		log.Fatal("Unable to establish bidirectional stream")
	}
	c.stream = stream
	go func() {
		for {
			in, _ := c.stream.Recv()

			if handler, found := c.commandHanlder[in.Name]; found {
				handler(in.Arg)
			}
		}
	}()
	if err := c.Send("id", c.id); err != nil {
		log.Fatal("Server not authorizing client ", err)
	}
}

func (c *Client) Send(cmdName string, cmdArg string) error {
	return c.stream.Send(&comm.Command{Name: cmdName, Arg: cmdArg})
}

func getConnection(addr string) *grpc.ClientConn {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}
