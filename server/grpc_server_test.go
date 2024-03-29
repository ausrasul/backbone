package server

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/ausrasul/backbone/comm"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestInstantiate(t *testing.T) {
	testData := []struct {
		addr string
	}{
		{addr: "localhost:1234"},
		{addr: "localhost1:12345"},
	}
	for _, data := range testData {
		Server := New(data.addr)
		assert.Equal(t, Server.addr, data.addr, "Host not set")
	}
}

func TestOnConnectCallback(t *testing.T) {
	varToBeChangedByCallBack := ""
	testData := []struct {
		function func(string)
		input    string
		expected string
	}{
		{
			input:    "monkey",
			function: func(input string) { varToBeChangedByCallBack = "monkey" },
			expected: "monkey",
		},
	}
	s := New("localhost:1234")
	for _, data := range testData {
		s.SetOnConnect(data.function)
		s.onConnect(data.input)
		assert.Equal(t, varToBeChangedByCallBack, data.expected, "Bad onconnect callback")
	}
}

func TestSetOnDisconnectCallback(t *testing.T) {
	varToBeChangedByCallBack := ""
	testData := []struct {
		function func(string)
		input    string
		expected string
	}{
		{
			input:    "monkey",
			function: func(input string) { varToBeChangedByCallBack = "monkey" },
			expected: "monkey",
		},
	}
	s := New("localhost:1234")
	for _, data := range testData {
		s.SetOnDisconnect(data.function)
		s.onDisconnect(data.input)
		assert.Equal(t, varToBeChangedByCallBack, data.expected, "Bad onconnect callback")
	}
}

func TestClientConnect(t *testing.T) {
	tests := []struct {
		name     string
		cmdName  string
		cmdArg   string
		expected string
		errMsg   string
	}{
		{
			"Valid id",
			"id",
			"1234",
			"1234",
			"Id was not received",
		},
		{
			"Valid id",
			"id",
			"12345",
			"12345",
			"Id was not received",
		},
		{
			"Duplicate id",
			"id",
			"12345",
			"",
			"Duplicate Id, should not call on connect",
		},
		{
			"Not id",
			"not_id",
			"123",
			"",
			"Should not call on connect",
		},
	}
	s := New("127.0.0.1:1234")
	recvChan := make(chan string)
	s.SetOnConnect(func(str string) { recvChan <- str })

	ctx := context.Background()
	dialer_ := dialer(s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for _, tt := range tests {
		client := comm.NewCommClient(conn)
		stream, _ := client.OpenComm(ctx)
		stream.Send(&comm.Command{Name: tt.cmdName, Arg: tt.cmdArg})
		res := ""
		select {
		case res = <-recvChan:
		case <-time.After(time.Millisecond * 10):
		}
		assert.Equal(t, tt.expected, res, tt.errMsg)

	}

}
func TestOnDisconnect(t *testing.T) {
	s, conn := start_grpc()

	tests := []struct {
		name            string
		client_id       string
		will_disconnect bool
		err_msg         string
	}{
		{
			"client don't disconnects",
			"client1",
			false,
			"onDisconnect should not be called before client disconnect",
		},
		{
			"client disconnects",
			"client2",
			true,
			"onDisconnect should be called.",
		},
		{
			"client disconnects",
			"client2",
			false,
			"onDisconnect should not be called before client disconnect",
		},
	}
	for _, tt := range tests {
		stream := connect_grpc(s, conn, tt.client_id)
		disconnected := make(chan string)
		s.SetOnDisconnect(func(client_id string) { disconnected <- client_id })

		if tt.will_disconnect {
			stream.CloseSend()
		}
		select {
		case res := <-disconnected:
			if !tt.will_disconnect {
				t.Error(tt.name + ", " + tt.err_msg)
			} else {
				assert.Equal(t, res, tt.client_id, tt.name+", "+tt.err_msg)
			}
		case <-time.After(time.Millisecond * 10):
			if tt.will_disconnect {
				t.Error(tt.name + ", " + tt.err_msg)
			}
		}
	}

}

func TestCallCommandHandler(t *testing.T) {
	s, conn := start_grpc()
	wait_for_handler := make(chan map[string]string)

	handler := func(client_id string, arg string) {
		wait_for_handler <- map[string]string{"id": client_id, "arg": arg}
	}
	tests := []struct {
		name         string
		handler_cmd  string
		handler_func func(string, string)
		client_id    string
		cmd_name     string
		cmd_arg      string
	}{
		{
			"Valid handler test 1",
			"cmd1",
			handler,
			"client1",
			"cmd1",
			"ape",
		},
		{
			"Valid handler test 2",
			"cmd1",
			handler,
			"client1",
			"cmd1",
			"dog",
		},
		{
			"Valid handler test 3",
			"cmd2",
			handler,
			"client2",
			"cmd2",
			"dog",
		},
	}
	var stream comm.Comm_OpenCommClient
	prev_client_id := ""
	for _, tt := range tests {
		if prev_client_id != tt.client_id {
			prev_client_id = tt.client_id
			stream = connect_grpc(s, conn, tt.client_id)
		}
		s.SetCommandHandler(tt.client_id, tt.handler_cmd, tt.handler_func)
		stream.Send(&comm.Command{Name: tt.cmd_name, Arg: tt.cmd_arg})
		res := <-wait_for_handler
		assert.Equal(t, res["id"], tt.client_id, tt.name+", handler should get client id")
		assert.Equal(t, res["arg"], tt.cmd_arg, tt.name+", handler should get cmd arg")
	}
}

func TestCallCommandHandlerAsGoRoutine(t *testing.T) {
	s, conn := start_grpc()
	stream := connect_grpc(s, conn, "client1")

	handler_called := make(chan int, 10)

	s.SetCommandHandler("client1", "longProcess", func(string, string) {
		handler_called <- 1
		time.Sleep(time.Second)
	})

	expectedCalledTimes := 5
	for i := 0; i < expectedCalledTimes; i++ {
		stream.Send(&comm.Command{Name: "longProcess", Arg: "cmdArg"})
	}
	handlerCalledTimes := 0
	for {
		select {
		case <-handler_called:
			handlerCalledTimes += 1
		case <-time.After(200 * time.Millisecond):
			t.Error("expected called", expectedCalledTimes, "got", handlerCalledTimes)
			t.FailNow()
		}
		if handlerCalledTimes == expectedCalledTimes {
			break
		}
	}
}

func TestSendCmdToClient(t *testing.T) {
	s, conn := start_grpc()
	stream := connect_grpc(s, conn, "clientA")
	s.Send("clientA", "testCommand", "cmdArg")
	//s.Send("clientA", &comm.Command{Name: "testCommand", Arg: "cmdArg"})
	in, err := stream.Recv()
	assert.Equal(t, err, nil, "Should not receive err ")
	assert.Equal(t, in.Name, "testCommand", "Invalid command name received")
	assert.Equal(t, in.Arg, "cmdArg", "Invalid command arg received")
	// should send to different clients.
	stream2 := connect_grpc(s, conn, "clientB")
	//s.Send("clientB", &comm.Command{Name: "testCommand2", Arg: "cmdArg2"})
	s.Send("clientB", "testCommand2", "cmdArg2")
	in, err = stream2.Recv()
	assert.Equal(t, err, nil, "Should not receive err ")
	assert.Equal(t, in.Name, "testCommand2", "Invalid command name received")
	assert.Equal(t, in.Arg, "cmdArg2", "Invalid command arg received")

	// send doesn't work wehn client disconnects.
	disconnected := make(chan int)
	s.SetOnDisconnect(func(client_id string) {
		disconnected <- 1
	})
	stream.CloseSend()
	<-disconnected
	//s.Send("clientA", &comm.Command{Name: "testCommand", Arg: "cmdArg"})
	s.Send("clientA", "testCommand", "cmdArg")
	_, err = stream.Recv()
	assert.Equal(t, err, io.EOF, "client should only receive EOF on closure.")

	// send introduced a  race condition
}

/*func waitFuncTimeout(f func()) chan int {
	done := make(chan int)
	go func() {
		f()
		done <- 1
	}()
	return done
}*/

func dialer(s *Server) func(context.Context, string) (net.Conn, error) {
	lis := bufconn.Listen(1024 * 1024)

	grpcServer := grpc.NewServer()

	comm.RegisterCommServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func start_grpc() (*Server, *grpc.ClientConn) {
	ctx := context.Background()
	s := New("127.0.0.1:1234")
	dialer_ := dialer(s)
	conn, err := grpc.DialContext(ctx, ":1234", grpc.WithInsecure(), grpc.WithContextDialer(dialer_))
	if err != nil {
		log.Fatal(err)
	}
	return s, conn
}

func connect_grpc(s *Server, conn *grpc.ClientConn, id string) comm.Comm_OpenCommClient {
	wait_connect := make(chan struct{})
	s.SetOnConnect(func(client_id string) {
		close(wait_connect)
	})
	client := comm.NewCommClient(conn)
	stream, _ := client.OpenComm(context.Background())
	stream.Send(&comm.Command{Name: "id", Arg: id})
	<-wait_connect
	return stream
}
