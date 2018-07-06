package main

import (
	"encoding/json"

	"github.com/Meduzz/rpc/api"

	"github.com/nats-io/go-nats"
)

func main() {
	conn, err := nats.Connect("nats://localhost:4222")

	if err != nil {
		panic(err)
	}

	println("Started.")

	conn.QueueSubscribe("echo", "test", func(msg *nats.Msg) {
		println("Echoing a message.")

		req := &api.Req{}
		json.Unmarshal(msg.Data, req)

		res := &api.Res{200, "text/html", req.Body}
		bodyBytes, _ := json.Marshal(res)

		conn.Publish(msg.Reply, bodyBytes)
	})

	for {
	}
}
