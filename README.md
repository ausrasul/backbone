# Backbone
A Golang library for server client communication, uses gRPC under the hood.
### Documentation
[Documentation](https://pkg.go.dev/github.com/ausrasul/backbone).

## Features:
- Dynamically add/change RPC methods (SetCommandHandler)
- User defined authentication and authorization (using OnConnect event)
- Detect and handle disconnects (OnDisconnect event)
- Transmitts and receives strings/serialized data only.
- Uses bidirectional gRPC stream under the hood.

## Example usage:

#### Server:
```golang
func main() {
    // Instantiate the server
    s := server.New("localhost:1234")

    // Define event handlers
    eventChan := make(chan string, 5)
    s.SetOnConnect(func(clientId string) {
        eventChan <- "a client disconnected " + clientId
    })
    s.SetOnDisconnect(func(clientId string) {
        eventChan <- "a client disconnected " + clientId
    })
    
    // Define command handlers
    s.SetCommandHandler("client id 123", "command_1", func(clientId string, arg string) {
        eventChan <- "command received: " + clientId + " -- " + arg
        // handle the command..
        // ..
        // Can respond to client:
        s.Send(clientId, "response_command_known_at_client", "cmdArg")
    }
    
    // Start server
    go s.Start()

    for in := range eventChan {
        log.Println("server received: ", in)
    }
}
```

#### Client:
```golang
func main() {
    // Instantiate client
    c := client.New("client id 123", ":1234")
    // Set event handlers
    c.SetOnConnect(func() { log.Println("connected!") })
    c.SetOnDisconnect(func() { log.Println("disconnected!") })
    // Set command handlers
    c.SetCommandHandler("response", func(arg string) {
        log.Println("received command from server , ", arg)
        // handle command
        // ...
        // Can respond to server
        c.Send("command_1", "arguments")
    })
    // connect to server
    c.Start()
    c.Send("test_command", "client is sending this command")
    <-time.After(time.Second * 5)
}
```

## License

MIT
