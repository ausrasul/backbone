package backbone

//"google.golang.org/grpc"

//"github.com/golang/protobuf/proto"

//"github.com/ausrasul/backbone/comm"

import (
	"fmt"
	"io"
	"net"

	"github.com/ausrasul/backbone/comm"
	"google.golang.org/grpc"
)

type Server struct {
	ip           string
	port         int
	onConnect    func(string)
	onDisconnect func(string)
	handler      func(string, string)
	comm.UnimplementedCommServer
}

func New(Ip string, port int) Server {
	return Server{ip: Ip, port: port, handler: func(a string, b string) {}}
}

func (s *Server) SetOnConnect(onConnectCallback func(string)) {
	s.onConnect = onConnectCallback
}

func (s *Server) SetOnDisconnect(onDisconnectCallback func(string)) {
	s.onDisconnect = onDisconnectCallback
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+fmt.Sprint(s.port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	comm.RegisterCommServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (s *Server) OpenComm(stream comm.Comm_OpenCommServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if in.Name == "id" {
			s.onConnect(in.Arg)
		}
		s.handler("123", "abc")
	}
}

func (s *Server) SetCommandHandler(client_id string, handler func(string, string)) {
	s.handler = handler
}

/* everytime we get a new client:
- create a client object
    - the id is random string
    - create input channel
        - every time input channel got entry:
            - send command to client.
- run onConnect (hopefully someone will add handlers)
- add handlers (if called) to teh client
- when receive command from client parse and call handler.
*/
