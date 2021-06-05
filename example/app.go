package main

import (
	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/rpc"
	"github.com/Meduzz/rpc/api"
)

func main() {
	conn, _ := nuts.Connect()
	rpc := rpc.NewRpc(conn)

	rpc.Handler("echo", "a", echoHandler)

	rpc.Run()
}

func echoHandler(ctx api.Context) {
	ctx.Reply(ctx.Text())
}
