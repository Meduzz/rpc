package discovery

import (
	"testing"

	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/transports"
)

func TestCreateDiscovery_Defaults(t *testing.T) {
	server, client := pair()
	a := NewDiscovery(client, server, &Settings{})

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
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{})

	_, a := sub.Request("", "", api.NewEmptyMessage())

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
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{})

	sub.RegisterEventer("test", eventer(nil), "fqn")
	sub.Remove("test")

	_, err := sub.find("fqn", "*")

	if err == nil {
		fel("Expected an error")
		t.Fail()
	}
}

func TestDiscoveryUpdated(t *testing.T) {
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{})

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

func TestEventer(t *testing.T) {
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{})
	evtChan := make(chan *api.Message)
	sub.RegisterEventer("eventer", eventer(evtChan), "test.eventer")
	sub.Start(false)

	addr := newAddress("test.eventer", "", "eventer", "")
	sub.registry.Update(addr)

	go sub.Trigger("test.eventer", "", api.NewEmptyMessage())

	msg := <-evtChan

	if len(msg.Body) > 0 {
		fel("Expected an empty body")
		t.Fail()
	}
}

func TestWorker(t *testing.T) {
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{})
	sub.RegisterWorker("worker", worker(), "test.worker")
	sub.Start(false)

	addr := newAddress("test.worker", "", "worker", "")
	sub.registry.Update(addr)

	msg, _ := sub.Request("test.worker", "", api.NewEmptyMessage())

	if msg == nil {
		fel("Did not expect response to be nil")
		t.FailNow()
	}

	if len(msg.Body) > 0 {
		fel("Expected an empty body")
		t.Fail()
	}
}

func TestHandler(t *testing.T) {
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{})
	sub.RegisterHandler("handler", handler(), "test.handler")
	sub.Start(false)

	addr := newAddress("test.handler", "", "handler", "")
	sub.registry.Update(addr)

	msg, _ := sub.Request("test.handler", "", api.NewEmptyMessage())

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
	server, client := pair()
	sub := NewDiscovery(client, server, &Settings{
		Namespace: "asdf",
	})
	sub.RegisterHandler("handler", handler(), "test.handler")
	sub.Start(false)

	addr := newAddress("test.handler", "", "handler", "")
	sub.registry.Update(addr)

	msg, _ := sub.Request("test.handler", "", api.NewEmptyMessage())

	if msg == nil {
		fel("Did not expect response to be nil")
		t.FailNow()
	}

	if len(msg.Body) > 0 {
		fel("Expected an empty body")
		t.Fail()
	}
}

func pair() (api.RpcServer, api.RpcClient) {
	server, _ := transports.NewLocalRpcServer("test")
	return server, transports.NewLocalRpcClient(server.(*transports.LocalRpcServer))
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
