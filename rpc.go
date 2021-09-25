package rpc

import (
	"context"
	"os"
	"os/signal"

	"github.com/Meduzz/rpc/api"
	nats "github.com/nats-io/nats.go"
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

func (r *RPC) Trigger(topic string, message interface{}) error {
	return trigger(r.conn, topic, message)
}

func (r *RPC) Request(topic string, message interface{}, timeout int) (api.Deserializer, error) {
	return request(r.conn, topic, message, timeout)
}

func (r *RPC) RequestContext(ctx context.Context, topic string, message interface{}) (api.Deserializer, error) {
	return requestContext(ctx, r.conn, topic, message)
}

// Run - Helper to block waiting for Interrupt then cleanup helper.
// But not really needed otherwise.
func (t *RPC) Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	t.conn.Drain()
}

func (t *RPC) handlerWrapper(handler api.Handler) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		ctx := newNatsContext(t.conn, msg)
		handler(ctx)
	}
}
