package main

import (
	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/transports"
)

func main() {
	nats, err := transports.NewNatsRpcServer("example", "nats://localhost:4222", nil, true)

	if err != nil {
		panic(err)
	}

	nats.RegisterWorker("echo", echoHandler)
	nats.RegisterWorker("error", errorHandler)
	nats.Start()
}

func echoHandler(msg *api.Message) *api.Message {
	msg.Metadata["result"] = "success"
	return msg
}

func errorHandler(msg *api.Message) *api.Message {
	return api.NewErrorMessage("A very generic error :(")
}
