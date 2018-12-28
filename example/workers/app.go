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
	msg, err := ctx.BodyAsMessage()

	if err != nil {
		errMsg := api.NewErrorMessage(err.Error())
		ctx.ReplyMessage(errMsg)
		return
	}

	msg.Metadata["result"] = "success"

	ctx.ReplyMessage(msg)
}

func errorHandler(msg *api.Message) *api.Message {
	return api.NewErrorMessage("A very generic error :(")
}
