package backbone

import (
	"context"
	"errors"
	"log"

	"github.com/ausrasul/backbone/comm"
	"google.golang.org/grpc"
)

type Client struct {
	serverAddr     string
	id             string
	onConnect      func(*Client)
	onDisconnect   func(*Client)
	commandHanlder map[string]func(*Client, string)
	stream         comm.Comm_OpenCommClient
}

func New(id string, addr string) *Client {
	return &Client{
		serverAddr:     addr,
		id:             id,
		commandHanlder: map[string]func(*Client, string){},
		onConnect:      func(c *Client) {},
		onDisconnect:   func(c *Client) {},
		stream:         nil,
	}
}

func (c *Client) SetOnConnect(callback func(*Client)) {
	c.onConnect = callback
}

func (c *Client) SetOnDisconnect(callback func(*Client)) {
	c.onDisconnect = callback
}

func (c *Client) SetCommandHandler(commandName string, callback func(*Client, string)) {
	c.commandHanlder[commandName] = callback
}
func (c *Client) Start() error {
	// getConnection is separated so that conn can be mocked during tests.
	conn := getConnection(c.serverAddr)
	return c.connect(conn)
}

func (c *Client) connect(conn *grpc.ClientConn) error {
	client := comm.NewCommClient(conn)
	stream, err := client.OpenComm(context.Background())
	if err != nil {
		return errors.New("cannot establish stream")
	}
	c.stream = stream
	go func() {
		for {
			in, err := c.stream.Recv()
			if err != nil {
				c.onDisconnect(c)
				break
			}
			if handler, found := c.commandHanlder[in.Name]; found {
				handler(c, in.Arg)
			}
		}
	}()
	if err := c.Send("id", c.id); err != nil {
		log.Fatal("Server not registering client ", err)
	}
	c.onConnect(c)
	return nil
}

func (c *Client) Send(cmdName string, cmdArg string) error {
	return c.stream.Send(&comm.Command{Name: cmdName, Arg: cmdArg})
}

func getConnection(addr string) *grpc.ClientConn {
	// Set up a connection to the server. this is mocked during testing
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}
