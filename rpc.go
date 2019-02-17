package rpc

import (
	"os"
	"os/signal"

	"github.com/Meduzz/rpc/api"
	nats "github.com/nats-io/go-nats"
)

type (
	RPC struct {
		conn *nats.Conn
		subz map[string]*nats.Subscription
	}

	natsContext struct {
		conn *nats.Conn
		msg  *nats.Msg
	}
)

func NewRpc(conn *nats.Conn) *RPC {
	subz := make(map[string]*nats.Subscription)
	return &RPC{conn, subz}
}

func (r *RPC) Handler(topic, queue string, handler api.Handler) error {
	if queue != "" {
		sub, err := r.conn.QueueSubscribe(topic, queue, r.handlerWrapper(handler))

		if err != nil {
			return err
		}

		r.subz[topic] = sub
	} else {
		sub, err := r.conn.Subscribe(topic, r.handlerWrapper(handler))

		if err != nil {
			return err
		}

		r.subz[topic] = sub
	}

	return nil
}

func (r *RPC) Remove(topic string) {
	sub, ok := r.subz[topic]

	if ok {
		sub.Drain()
		delete(r.subz, topic)
	}
}

func (r *RPC) Trigger(topic string, message *api.Message) error {
	return trigger(r.conn, topic, message)
}

func (r *RPC) Request(topic string, message *api.Message, timeout int) (*api.Message, error) {
	return request(r.conn, topic, message, timeout)
}

// Run - Helper to block waiting for Interrupt then cleanup helper.
// But not really needed otherwise.
func (t *RPC) Run() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	t.conn.Close()
}

func (t *RPC) handlerWrapper(handler api.Handler) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		ctx := newNatsContext(t.conn, msg)
		handler(ctx)
	}
}
