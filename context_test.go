package rpc

import (
	"testing"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/rpc/encoding"
)

type Test struct {
	Message string `json:"message"`
}

func TestContextCommsFuncs(t *testing.T) {
	conn, _ := nuts.Connect()

	rpc := NewRpc(conn, encoding.Json())
	rpc.HandleRPC("context.test1", "", func(ctx *RpcContext) {
		test := &Test{}
		err := ctx.Bind(test)

		if err != nil {
			t.Errorf("Binding message thew error: %v\n", err)
			t.Fail()
			return
		}

		if test.Message != "Hello world!" {
			t.Fatalf("Message was not matching the expected one")
			return
		}

		if !ctx.IsRPC() {
			t.Fatalf("The message was not marked as replyable")
			return
		}

		ctx.Reply(test)
	})

	body := &Test{"Hello world!"}
	msg := &Test{}
	err := rpc.Request("context.test1", body, msg, 5)

	if err != nil {
		t.Fatalf("RPC request threw error: %v\n", err)
		return
	}

	if msg.Message != "Hello world!" {
		t.Fatalf("Message in reply was not matching the expected one")
		return
	}
}
