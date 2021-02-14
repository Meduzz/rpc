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
	rpc.Handler("context.test1", "", func(ctx api.Context) {
		msg := ctx.Body()

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

		ctx.Reply(msg)
	})

	body, _ := api.NewMessage(Test{"Hello world!"})
	msg, err := rpc.Request("context.test1", body, 5)

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
