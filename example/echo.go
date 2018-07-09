package main

import (
	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/framework"
)

func main() {
	b := framework.NewBuilder()

	_, err := b.Nats("nats://localhost:4222").WorkerGroup("echo", "1", handler).Build()

	if err != nil {
		panic(err)
	}

	println("Started.")

	for {
	}
}

func handler(req *api.Req) (*api.Res, error) {
	return &api.Res{200, "text/html", req.Body}, nil
}
