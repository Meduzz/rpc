package rpc

import (
	"testing"

	"github.com/Meduzz/rpc/api"

	"github.com/Meduzz/helper/nuts"
)

type Test struct {
	Message string `json:"message"`
}

func TestContextCommsFuncs(t *testing.T) {
	conn, err := nuts.Connect()

	rpc := NewRpc(conn)
	rpc.Handler("test", "", func(ctx api.Context) {
		msg, _ := ctx.Body()

		test := &Test{}
		err := ctx.Bind(test)

		if err != nil {
			t.Fatalf("Binding message thew error: %v\n", err)
			return
		}

		if test.Message != "Hello world!" {
			t.Fatalf("Message was not matching the expected one")
			return
		}

		ctx.Reply(msg)
	})

	body, _ := api.NewMessage(Test{"Hello world!"})
	msg, err := rpc.Request("test", body, 5)

	if err != nil {
		t.Fatalf("RPC request threw error: %v\n", err)
		return
	}

	subject := &Test{}
	msg.Json(subject)

	if subject.Message != "Hello world!" {
		t.Fatalf("Message in reply was not matching the expected one")
		return
	}
}
