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
	ip                string
	port              int
	onConnect         func(string)
	onDisconnect      func(string)
	cmd_handlers      map[string]map[string]func(string, string)
	cmd_handlers_lock chan int
	comm.UnimplementedCommServer
}

func New(Ip string, port int) Server {
	s := Server{
		ip:                Ip,
		port:              port,
		cmd_handlers:      make(map[string]map[string]func(string, string)),
		cmd_handlers_lock: make(chan int, 1),
	}
	s.cmd_handlers_lock <- 1
	return s
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
	client_id := ""
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if in.Name == "id" {
			client_id = in.Arg
			s.onConnect(in.Arg)
		}
		handler, ok := s.getClientCmdHandler(client_id, in.Name)
		if !ok {
			continue
		}
		handler(client_id, in.Arg)
	}
}

func (s *Server) SetCommandHandler(client_id string, cmd_name string, handler func(string, string)) {
	<-s.cmd_handlers_lock
	_, ok := s.cmd_handlers[client_id]
	if !ok {
		s.cmd_handlers[client_id] = make(map[string]func(string, string))
	}
	s.cmd_handlers[client_id][cmd_name] = handler
	s.cmd_handlers_lock <- 1
}

func (s *Server) getClientCmdHandler(client_id string, cmd_name string) (func(string, string), bool) {
	<-s.cmd_handlers_lock
	handlers, ok := s.cmd_handlers[client_id]
	if !ok {
		s.cmd_handlers_lock <- 1
		return nil, false
	}
	handler, ok := handlers[cmd_name]
	s.cmd_handlers_lock <- 1
	return handler, ok
}
