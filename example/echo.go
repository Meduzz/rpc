package main

import (
	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/framework"

	"errors"
)

func main() {
	b := framework.NewBuilder()
	b.Nats("nats://localhost:4222")
	b.WorkerGroup("echo", "1", echoHandler)
	b.WorkerGroup("error", "1", errorThrower)

	service, err := b.Build()

	if err != nil {
		panic(err)
	}

	println("Started.")

	service.Start()
}

func echoHandler(req *api.Req) (*api.Res, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	return &api.Res{200, headers, req.Body}, nil
}
func echoHandler(req *api.Req) (*api.Res, error) {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	return &api.Res{200, headers, req.Body}, nil
}

func errorThrower(req *api.Req) (*api.Res, error) {
	return nil, errors.New("A random error.")
}
