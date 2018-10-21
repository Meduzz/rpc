package main

import (
	"encoding/json"
	"fmt"

	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/transports"
)

func main() {
	server, _ := transports.NewLocalRpcServer("example")
	server.RegisterEventer("print", printHandler)

	client := transports.NewLocalRpcClient(server.(*transports.LocalRpcServer))

	headers := make(map[string]string)
	headers["hello"] = "world"

	msg := &api.Message{}
	msg.Metadata = headers
	msg.Body = json.RawMessage([]byte("Hello %s!\n"))

	client.Trigger("print", msg)
}

func printHandler(msg *api.Message) {
	fmt.Printf(string(msg.Body), msg.Metadata["hello"])
}
