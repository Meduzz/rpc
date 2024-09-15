package rpc

import (
	"testing"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/rpc/encoding"
	"github.com/Meduzz/rpc/messages"
)

type Test struct {
	Message string `json:"message"`
}

func TestContextCommsFuncs(t *testing.T) {
	conn, _ := nuts.Connect()

	rpc := NewRpc(conn, encoding.Json())

	t.Run("RPC", func(t *testing.T) {
		var (
			test1 = "rpc.context.test1"
			test2 = "rpc.context.test2"
		)

		rpc.HandleRPC(test1, "", func(ctx *RpcContext) {
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

		rpc.HandleRPC(test2, "", func(ctx *RpcContext) {
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

			ctx.ReplyBuilder(func(mb *messages.MsgBuilder) {
				mb.WithBody(test)
			})
		})

		t.Run("test1", func(t *testing.T) {
			body := &Test{"Hello world!"}
			msg := &Test{}
			err := rpc.Request(test1, body, msg, 5)

			if err != nil {
				t.Fatalf("RPC request threw error: %v\n", err)
				return
			}

			if msg.Message != "Hello world!" {
				t.Fatalf("Message in reply was not matching the expected one")
				return
			}
		})

		t.Run("test2", func(t *testing.T) {
			body := &Test{"Hello world!"}
			msg := &Test{}
			err := rpc.Request(test2, body, msg, 5)

			if err != nil {
				t.Fatalf("RPC request threw error: %v\n", err)
				return
			}

			if msg.Message != "Hello world!" {
				t.Fatalf("Message in reply was not matching the expected one")
				return
			}
		})
	})

	t.Run("Events", func(t *testing.T) {
		var (
			test1 = "events.context.test1"
		)

		rpc.HandleEvent(test1, "", func(ctx *EventContext) {
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
		})

		t.Run("test1", func(t *testing.T) {
			body := &Test{"Hello world!"}

			err := rpc.Trigger(test1, body)

			if err != nil {
				t.Fatalf("Trigger Event threw error: %v\n", err)
				return
			}
		})
	})
}
