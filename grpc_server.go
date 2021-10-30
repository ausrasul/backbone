package grpc_server

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
