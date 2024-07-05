package rpc

import (
	"context"
	"encoding/json"

	"github.com/Meduzz/rpc/encoding"
	nats "github.com/nats-io/nats.go"
)

type (
	RpcContext struct {
		conn  *nats.Conn
		msg   *nats.Msg
		codec encoding.Codec
	}
)

func NewRpcContext(conn *nats.Conn, msg *nats.Msg, codec encoding.Codec) *RpcContext {
	return &RpcContext{
		conn:  conn,
		msg:   msg,
		codec: codec,
	}
}

func (c *RpcContext) Bind(to any) error {
	return c.codec.Unmarshal(c.msg.Data, to)
}

func (c *RpcContext) Raw() []byte {
	return c.msg.Data
}

func (c *RpcContext) Msg() *nats.Msg {
	return c.msg
}

func (c *RpcContext) Reply(msg any) error {
	if c.IsRPC() {
		bs, err := json.Marshal(msg)

		if err != nil {
			return err
		}

		return c.conn.Publish(c.msg.Reply, bs)
	}

	return ErrUnexpectedReply
}

func (c *RpcContext) Trigger(topic string, event any) error {
	return Trigger(c.conn, c.codec, topic, event)
}

func (c *RpcContext) Request(topic string, msg, response any, timeout int) error {
	return Request(c.conn, c.codec, topic, msg, timeout, response)
}

func (c *RpcContext) RequestContext(ctx context.Context, topic string, msg, response any) error {
	return RequestContext(ctx, c.conn, c.codec, topic, msg, response)
}

func (c *RpcContext) Forward(topic string, msg any) error {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := c.codec.Marshal(msg)

		if err != nil {
			return err
		}

		if c.IsRPC() {
			return c.conn.PublishRequest(topic, c.msg.Reply, bs)
		} else {
			return c.conn.Publish(topic, bs)
		}
	} else {
		if topic != "" {
			natsMsg.Subject = topic
		}

		return c.conn.PublishMsg(natsMsg)
	}
}

func (c *RpcContext) IsRPC() bool {
	return c.msg.Reply != ""
}

func (c *RpcContext) Header(name string) string {
	return c.msg.Header.Get(name)
}
