package discovery

import (
	"testing"

	"github.com/Meduzz/helper/nuts"
	"github.com/Meduzz/rpc"

	"github.com/Meduzz/rpc/api"
)

func TestCreateDiscovery_Defaults(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatalf("Could not connect to nats: %s.", err)
	}

	rpc := rpc.NewRpc(conn)

	a := NewDiscovery(rpc, &Settings{})

	if a.settings.Namespace != "*" {
		fel("Expected namespace to be wildcard.")
		t.Fail()
	}

	if a.settings.DiscoveryTopic != "discovery" {
		fel("Expected discovery topic to be discovery, but was: %s", a.settings.DiscoveryTopic)
		t.Fail()
	}

	if a.settings.Interests == nil {
		fel("Expected interests to be set")
		t.Fail()
	}

	if a.settings.Version != "DEVELOPMENT" {
		fel("Expected version to be DEVELOPMENT but was: %s", a.settings.Version)
	}
}

func TestValidateTriggerAndRequest(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatalf("Could not connect to nats: %s.", err)
	}

	rpc := rpc.NewRpc(conn)

	sub := NewDiscovery(rpc, &Settings{})

	_, a := sub.Request("", "", api.NewEmptyMessage(), 3)

	if a == nil {
		fel("(a) Expected an error")
		t.Fail()
	}

	b := sub.Trigger("", "", api.NewEmptyMessage())

	if b == nil {
		fel("(b) Expected an error")
		t.Fail()
	}
}

func TestRemove(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatalf("Could not connect to nats: %s.", err)
	}

	rpc := rpc.NewRpc(conn)

	sub := NewDiscovery(rpc, &Settings{})

	sub.RegisterHandler("test", "", handler(), "fqn")
	sub.Remove("test")

	_, err = sub.find("fqn", "*")

	if err == nil {
		fel("Expected an error")
		t.Fail()
	}
}

func TestDiscoveryUpdated(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatalf("Could not connect to nats: %s.", err)
	}

	rpc := rpc.NewRpc(conn)

	sub := NewDiscovery(rpc, &Settings{})

	sub.Start(false)

	addr := newAddress("test", "", "", "")
	sub.registry.Update(addr)

	a, err := sub.find("test", "")

	if err != nil {
		fel("Did not expect an error: %s", err.Error())
		t.FailNow()
	}

	if a == nil {
		fel("Expected a match on semi wildcard")
		t.Fail()
	}
}

func TestHandler(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatalf("Could not connect to nats: %s.", err)
	}

	rpc := rpc.NewRpc(conn)

	sub := NewDiscovery(rpc, &Settings{})

	sub.RegisterHandler("handler", "", handler(), "test.handler")
	sub.Start(false)

	addr := newAddress("test.handler", "", "handler", "")
	sub.registry.Update(addr)

	msg, _ := sub.Request("test.handler", "", api.NewEmptyMessage(), 3)

	if msg == nil {
		fel("Did not expect response to be nil")
		t.FailNow()
	}

	if len(msg.Body) > 0 {
		fel("Expected an empty body")
		t.Fail()
	}
}

func TestTriggerGlobalNamespaceLookup(t *testing.T) {
	conn, err := nuts.Connect()

	if err != nil {
		t.Fatalf("Could not connect to nats: %s.", err)
	}

	rpc := rpc.NewRpc(conn)

	sub := NewDiscovery(rpc, &Settings{
		Namespace: "asdf",
	})

	sub.RegisterHandler("handler", "", handler(), "test.handler")
	sub.Start(false)

	addr := newAddress("test.handler", "", "handler", "")
	sub.registry.Update(addr)

	msg, _ := sub.Request("test.handler", "", api.NewEmptyMessage(), 3)

	if msg == nil {
		fel("Did not expect response to be nil")
		t.FailNow()
	}

	if len(msg.Body) > 0 {
		fel("Expected an empty body")
		t.Fail()
	}
}

func eventer(out chan *api.Message) func(*api.Message) {
	return func(msg *api.Message) {
		out <- msg
	}
}

func worker() func(*api.Message) *api.Message {
	return func(msg *api.Message) *api.Message {
		return msg
	}
}

func handler() func(api.Context) {
	return func(ctx api.Context) {
		msg, _ := ctx.Body()
		ctx.Reply(msg)
	}
}
