package main

import (
	"github.com/Meduzz/rpc/transports"
)

func main() {
	nats, _ := transports.NewNatsRpcClient("nats://localhost:4222", nil)
	proxy := transports.NewSimpleHttpProxy("/rpc", ":8080", nats)

	proxy.Start()
}
