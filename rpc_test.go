package rpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Meduzz/rpc/api"
	"github.com/nats-io/nats.go"

	"github.com/Meduzz/helper/nuts"
)

var data = &Test{"Hello?"}

func TestSubscribeAndTrigger(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test1", "", testHandler)

	err = sub.Trigger("rpc.test1", data)

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

	msg, err := sub.Request("rpc.test2", data, 3)

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}

	if msg == nil {
		t.Error("Expected a message in the response")
	}
}

func TestRequestAMsg(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test3", "", testHandler)

	bs, err := json.Marshal(data)

	if err != nil {
		t.Fatal("cant serialize data")
	}

	msg := &nats.Msg{}
	msg.Header = make(nats.Header)

	msg.Data = bs
	msg.Header.Add("Hello", "world!")

	ret, err := sub.Request("rpc.test3", msg, 3)

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}

	if ret == nil {
		t.Error("Expected a message in the response")
	}
}

func TestTriggerMsg(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test4", "", testHandler)

	bs, err := json.Marshal(data)

	if err != nil {
		t.Fatal("cant serialize data")
	}

	msg := &nats.Msg{}
	msg.Header = make(nats.Header)

	msg.Data = bs
	msg.Header.Add("Hello", "world!")

	err = sub.Trigger("rpc.test4", msg)

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}
}

func TestForwardMsg(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test5", "", func(ctx api.Context) {
		msg := ctx.Msg()

		ctx.Forward("rpc.test5.1", msg)
	})
	sub.Handler("rpc.test5.1", "", testHandler)

	bs, err := json.Marshal(data)

	if err != nil {
		t.Fatal("cant serialize data")
	}

	msg := &nats.Msg{}
	msg.Header = make(nats.Header)

	msg.Data = bs
	msg.Header.Add("Hello", "world!")

	err = sub.Trigger("rpc.test5", msg)

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}
}

func TestRequestContextAMsg(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatal("cant connect to nats.")
	}

	sub := NewRpc(conn)
	sub.Handler("rpc.test6", "", testHandler)

	bs, err := json.Marshal(data)

	if err != nil {
		t.Fatal("cant serialize data")
	}

	msg := &nats.Msg{}
	msg.Header = make(nats.Header)

	msg.Data = bs
	msg.Header.Add("Hello", "world!")

	ctx := context.Background()
	ctx, cnl := context.WithDeadline(ctx, time.Now().Add(3*time.Second))
	ret, err := sub.RequestContext(ctx, "rpc.test6", msg)

	defer cnl()

	if err != nil {
		t.Errorf("Did not expect any errors when trigger message: %s", err.Error())
	}

	if ret == nil {
		t.Error("Expected a message in the response")
	}
}

func testHandler(ctx api.Context) {
	event := &Test{}
	ctx.Bind(event)

	if ctx.IsRPC() {
		ctx.Reply(event)
	}
}
