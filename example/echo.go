package main

import (
	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/framework"
)

func main() {
	b := framework.NewBuilder()

	service, err := b.Nats("nats://localhost:4222").WorkerGroup("echo", "1", handler).Build()

	if err != nil {
		panic(err)
	}

	println("Started.")

	service.Start()
}

func handler(req *api.Req) (*api.Res, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	return &api.Res{200, headers, req.Body}, nil
}
