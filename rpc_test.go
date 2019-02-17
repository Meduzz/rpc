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
	sub.Handler("test", "", testHandler)

	err = sub.Trigger("test", api.NewEmptyMessage())

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
	sub.Handler("test", "asdf", testHandler)

	msg, err := sub.Request("test", api.NewEmptyMessage())

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}

	if msg == nil {
		t.Error("Expected a message in the response")
	}
}

func testHandler(ctx api.Context) {
	msg, _ := ctx.Body()

	ctx.Reply(msg)
}
