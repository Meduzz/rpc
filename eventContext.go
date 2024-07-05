package rpc

import (
	"context"

	"github.com/Meduzz/rpc/encoding"
	nats "github.com/nats-io/nats.go"
)

type (
	EventContext struct {
		conn  *nats.Conn
		msg   *nats.Msg
		codec encoding.Codec
	}
)

func NewEventContext(conn *nats.Conn, msg *nats.Msg, codec encoding.Codec) *EventContext {
	return &EventContext{
		conn:  conn,
		msg:   msg,
		codec: codec,
	}
}

func (c *EventContext) Bind(to any) error {
	return c.codec.Unmarshal(c.msg.Data, to)
}

func (c *EventContext) Raw() []byte {
	return c.msg.Data
}

func (c *EventContext) Msg() *nats.Msg {
	return c.msg
}

func (c *EventContext) Trigger(topic string, event any) error {
	return Trigger(c.conn, c.codec, topic, event)
}

func (c *EventContext) Request(topic string, msg, response any, timeout int) error {
	return Request(c.conn, c.codec, topic, msg, timeout, response)
}

func (c *EventContext) RequestContext(ctx context.Context, topic string, msg, response any) error {
	return RequestContext(ctx, c.conn, c.codec, topic, msg, response)
}
