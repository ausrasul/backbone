package backbone

//"google.golang.org/grpc"

//"github.com/golang/protobuf/proto"

//"github.com/ausrasul/backbone/comm"

import (
	"log"

	pb "github.com/ausrasul/backbone/comm"
)

type Server struct {
	ip           string
	port         int
	onConnect    func(string)
	onDisconnect func(string)
}

func New(Ip string, port int) Server {
	return Server{ip: Ip, port: port}
}

func (s *Server) SetOnConnect(onConnectCallback func(string)) {
	s.onConnect = onConnectCallback
}

func (s *Server) SetOnDisconnect(onDisconnectCallback func(string)) {
	s.onDisconnect = onDisconnectCallback
}

func (s *Server) OpenComm(c *pb.Command) error {
	log.Println("hhhhhhhhhh")
	return nil
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
