package backbone

import (
	"log"

	"github.com/ausrasul/backbone/comm"
	"google.golang.org/grpc"
)

type Client struct {
	server_addr  string
	id           string
	onConnect    func()
	onDisconnect func()
}

func New(id string, addr string) *Client {
	return &Client{
		server_addr: addr,
		id:          id,
	}
}

func (c *Client) SetOnConnect(callback func()) {
	c.onConnect = callback
}

func (c *Client) SetOnDisconnect(callback func()) {
	c.onDisconnect = callback
}

func (c *Client) Start() {
	conn := getConnection(c.server_addr)
	defer conn.Close()
	//c.connect(conn)
}

func (c *Client) connect(conn *grpc.ClientConn) {
	client := comm.NewCommClient(conn)
	_ = client
	/*
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
}
func getConnection(addr string) *grpc.ClientConn {
	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return conn
}

/*func New(Ip string, port int) *Server {
	s := Server{
		ip:                Ip,
		port:              port,
		cmd_handlers:      make(map[string]map[string]func(string, string)),
		cmd_handlers_lock: make(chan int, 1),
		clientsOutbox:     make(map[string]chan *comm.Command),
		onDisconnect:      func(s string) {},
		//clientsOutboxMgr:  make(chan map[string]interface{}),
	}
	s.cmd_handlers_lock <- 1
	return &s
}
*/
