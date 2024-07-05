package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/rpc/encoding"
)

func TestRPC(t *testing.T) {
	data := &Test{"Hello?"}
	dataMsg, err := NewMessage(encoding.Json(), "rpc.rpc", data)

	if err != nil {
		t.Fatal("could not create dataMsg")
	}

	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn, encoding.Json())

	t.Run("rpc", func(t *testing.T) {
		sub.HandleRPC("rpc.rpc", "", testHandler)
		sub.HandleRPC("rpc.forward", "", testForwardHandler)

		// happyCase
		t.Run("happy case request", func(t *testing.T) {
			response := &Test{}
			err = sub.Request("rpc.rpc", data, response, 3)

			if err != nil {
				t.Errorf("request threw error %s", err.Error())
			}

			if data.Message != response.Message {
				t.Errorf("request did not match response req: %s res: %s", data.Message, response.Message)
			}
		})

		// happyCaseMsg
		t.Run("happy case request msg", func(t *testing.T) {
			response := &Test{}
			err = sub.Request("rpc.rpc", dataMsg, response, 3)

			if err != nil {
				t.Errorf("request threw error %s", err.Error())
			}

			if data.Message != response.Message {
				t.Errorf("request did not match response req: %s res: %s", data.Message, response.Message)
			}
		})

		// forward
		t.Run("forward request", func(t *testing.T) {
			response := &Test{}
			err = sub.Request("rpc.forward", data, response, 3)

			if err != nil {
				t.Errorf("request threw error %s", err.Error())
			}

			if data.Message != response.Message {
				t.Errorf("request did not match response req: %s res: %s", data.Message, response.Message)
			}
		})

		// withContext
		t.Run("request with context", func(t *testing.T) {
			ctx := context.Background()
			ctx, cncl := context.WithDeadline(ctx, time.Now().Add(3*time.Second))

			response := &Test{}
			err = sub.RequestContext(ctx, "rpc.forward", data, response)

			if err != nil {
				t.Errorf("request threw error %s", err.Error())
			}

			if data.Message != response.Message {
				t.Errorf("request did not match response req: %s res: %s", data.Message, response.Message)
			}

			defer cncl()
		})

		sub.Remove("rpc.forward")
		sub.Remove("rpc.rpc")
	})

	t.Run("event", func(t *testing.T) {
		feedback := make(chan *Test)
		sub.HandleEvent("rpc.event", "", testEventHandler(feedback))

		dataMsg.Subject = "rpc.event"

		// happyCase
		t.Run("happy case", func(t *testing.T) {
			err = sub.Trigger("rpc.event", data)

			if err != nil {
				t.Errorf("trigger event threw error %s", err.Error())
			}

			res := <-feedback

			if res.Message != data.Message {
				t.Errorf("event body did not match request body req: %s evt: %s", data.Message, res.Message)
			}
		})

		// hapyCaseMsg
		t.Run("happy case msg", func(t *testing.T) {
			err = sub.Trigger("", dataMsg)

			if err != nil {
				t.Errorf("trigger event threw error %s", err.Error())
			}

			res := <-feedback

			if res.Message != data.Message {
				t.Errorf("event body did not match request body req: %s evt: %s", data.Message, res.Message)
			}
		})
	})

	conn.Drain()
}

func testHandler(ctx *RpcContext) {
	event := &Test{}
	ctx.Bind(event)

	if ctx.IsRPC() {
		ctx.Reply(event)
	}
}

func testForwardHandler(ctx *RpcContext) {
	msg := ctx.Msg()

	ctx.Forward("rpc.rpc", msg)
}

func testEventHandler(feedback chan *Test) func(*EventContext) {
	return func(ctx *EventContext) {
		event := &Test{}
		ctx.Bind(event)

		feedback <- event
	}
}
