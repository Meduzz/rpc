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

	nats.RegisterHandler("echo", echoHandler)
	nats.RegisterWorker("error", errorHandler)
	nats.Start()
}

func echoHandler(ctx api.Context) {
	msg, err := ctx.Body()

	if err != nil {
		errMsg := api.NewErrorMessage(err.Error())
		ctx.Reply(errMsg)
		return
	}

	msg.Metadata["result"] = "success"

	ctx.Reply(msg)
}

func errorHandler(msg *api.Message) *api.Message {
	return api.NewErrorMessage("A very generic error :(")
}
