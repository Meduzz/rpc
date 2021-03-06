package rpc

import (
	"testing"

	"github.com/Meduzz/rpc/api"

	"github.com/Meduzz/helper/nuts"
)

var visited = false

func TestSubscribeAndTrigger(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test1", "", testHandler)

	err = sub.Trigger("rpc.test1", api.NewEmptyMessage())

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}
}

func TestSubscribeAndRequest(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test2", "asdf", testHandler)

	msg, err := sub.Request("rpc.test2", api.NewEmptyMessage(), 3)

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}

	if msg == nil {
		t.Error("Expected a message in the response")
	}
}

func testHandler(ctx api.Context) {
	msg := ctx.Body()

	ctx.Reply(msg)
}
