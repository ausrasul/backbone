package server

//"google.golang.org/grpc"

//"github.com/golang/protobuf/proto"

//"github.com/ausrasul/backbone/comm"

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/ausrasul/backbone/comm"
	"google.golang.org/grpc"
)

type Server struct {
	addr                 string
	onConnect            func(string)
	onDisconnect         func(string)
	cmd_handlers         map[string]map[string]func(string, string)
	cmd_handlers_mutex   sync.RWMutex
	clientsOutbox        map[string]chan *comm.Command
	clientsOutboxRWMutex sync.RWMutex
	//clientsOutboxMgr  chan map[string]interface{}
	comm.UnimplementedCommServer
}

func New(addr string) *Server {
	s := Server{
		addr:          addr,
		cmd_handlers:  make(map[string]map[string]func(string, string)),
		clientsOutbox: make(map[string]chan *comm.Command),
		onDisconnect:  func(s string) {},
		//clientsOutboxMgr:  make(chan map[string]interface{}),
	}
	return &s
}

func (s *Server) SetOnConnect(onConnectCallback func(string)) {
	s.onConnect = onConnectCallback
}

func (s *Server) SetOnDisconnect(onDisconnectCallback func(string)) {
	s.onDisconnect = onDisconnectCallback
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
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

func (s *Server) Send(client_id string, command *comm.Command) error {
	s.clientsOutboxRWMutex.RLock()
	outbox, ok := s.clientsOutbox[client_id]
	s.clientsOutboxRWMutex.RUnlock()

	if ok {
		outbox <- command
	}
	return nil
}

func (s *Server) OpenComm(stream comm.Comm_OpenCommServer) error {
	client_id := ""
	done := make(chan int)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			done <- 1
			s.clientsOutboxRWMutex.Lock()
			delete(s.clientsOutbox, client_id)
			s.clientsOutboxRWMutex.Unlock()

			s.deleteClientCmdHandlers(client_id)
			go s.onDisconnect(client_id)
			return nil
		}
		if err != nil {
			return err
		}
		if in.Name == "id" && client_id == "" {
			client_id = in.Arg
			s.clientsOutboxRWMutex.RLock()
			_, ok := s.clientsOutbox[client_id]
			s.clientsOutboxRWMutex.RUnlock()

			if ok {
				return errors.New("client id already exist")
			}
			s.clientsOutboxRWMutex.Lock()
			s.clientsOutbox[client_id] = make(chan *comm.Command)
			outboxCh := s.clientsOutbox[client_id]
			s.clientsOutboxRWMutex.Unlock()

			go func() {

				for {
					select {
					case <-done:
						return
					case command := <-outboxCh:
						stream.Send(command)
					}
				}
			}()
			s.onConnect(in.Arg)
		}
		handler, ok := s.getClientCmdHandler(client_id, in.Name)
		if !ok {
			continue
		}
		go handler(client_id, in.Arg)
	}
}

func (s *Server) deleteClientCmdHandlers(client_id string) {
	s.cmd_handlers_mutex.Lock()
	delete(s.cmd_handlers, client_id)
	s.cmd_handlers_mutex.Unlock()
}
func (s *Server) SetCommandHandler(client_id string, cmd_name string, handler func(string, string)) {
	s.cmd_handlers_mutex.Lock()
	_, ok := s.cmd_handlers[client_id]
	if !ok {
		s.cmd_handlers[client_id] = make(map[string]func(string, string))
	}
	s.cmd_handlers[client_id][cmd_name] = handler
	s.cmd_handlers_mutex.Unlock()
}

func (s *Server) getClientCmdHandler(client_id string, cmd_name string) (func(string, string), bool) {
	s.cmd_handlers_mutex.RLock()
	handlers, ok := s.cmd_handlers[client_id]
	if !ok {
		s.cmd_handlers_mutex.RUnlock()
		return nil, false
	}
	handler, ok := handlers[cmd_name]
	s.cmd_handlers_mutex.RUnlock()
	return handler, ok
}
