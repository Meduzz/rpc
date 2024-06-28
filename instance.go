package rpc

import (
	"context"
	"os"
	"os/signal"

	"github.com/Meduzz/rpc/encoding"
	nats "github.com/nats-io/nats.go"
)

type (
	RPC struct {
		conn  *nats.Conn
		subz  map[string]*nats.Subscription
		codec encoding.Codec
	}
)

// NewRpc - encapsulate a nats-conn and a codec into a RPC instance.
func NewRpc(conn *nats.Conn, codec encoding.Codec) *RPC {
	subz := make(map[string]*nats.Subscription)
	return &RPC{conn, subz, codec}
}

func (r *RPC) HandleRPC(topic, queue string, handler RpcHandler) error {
	sub, err := HandleRPC(r.conn, r.codec, topic, queue, handler)

	if err != nil {
		return err
	}

	r.subz[topic] = sub

	return nil
}

func (r *RPC) HandleEvent(topic, queue string, handler EventHandler) error {
	sub, err := HandleEvent(r.conn, r.codec, topic, queue, handler)

	if err != nil {
		return err
	}

	r.subz[topic] = sub

	return nil
}

func (r *RPC) Remove(topic string) {
	sub, ok := r.subz[topic]

	if ok {
		sub.Drain()
		delete(r.subz, topic)
	}
}

func (r *RPC) Trigger(topic string, message any) error {
	return Trigger(r.conn, r.codec, topic, message)
}

func (r *RPC) Request(topic string, message, response any, timeout int) error {
	return Request(r.conn, r.codec, topic, message, timeout, response)
}

func (r *RPC) RequestContext(ctx context.Context, topic string, message, response any) error {
	return RequestContext(ctx, r.conn, r.codec, topic, message, response)
}

// Run - Helper to block waiting for Interrupt then cleanup helper.
// But not really needed otherwise.
func (t *RPC) Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	t.conn.Drain()
}
