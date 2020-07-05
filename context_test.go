package rpc

import (
	"testing"

	"github.com/Meduzz/rpc/api"
	nats "github.com/nats-io/nats.go"

	"github.com/Meduzz/helper/nuts"
)

func TestContextCommsFuncs(t *testing.T) {
	conn, err := nuts.Connect()

	rpc := NewRpc(conn)
	rpc.Handler("test", "", func(ctx api.Context) {
		msg, _ := ctx.Body()
		ctx.Reply(msg)
	})

	msg := &nats.Msg{
		Data:    []byte("Hello world!"),
		Reply:   "test",
		Subject: "test",
	}

	ctx := newNatsContext(conn, msg)
	err = ctx.Reply(api.NewEmptyMessage())

	if err != nil {
		t.Errorf("Did not expect an error: %s", err.Error())
	}

	err = ctx.Forward("test", api.NewEmptyMessage())

	if err != nil {
		t.Errorf("There was an error: %s", err.Error())
	}
}
