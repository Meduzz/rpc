package rpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Meduzz/rpc/api"
	nats "github.com/nats-io/nats.go"
)

func newNatsContext(conn *nats.Conn, msg *nats.Msg) *natsContext {
	return &natsContext{
		conn: conn,
		msg:  msg,
	}
}

func (c *natsContext) Bind(to interface{}) error {
	return json.Unmarshal(c.msg.Data, to)
}

func (c *natsContext) Raw() []byte {
	return c.msg.Data
}

func (c *natsContext) Msg() *nats.Msg {
	return c.msg
}

func (c *natsContext) Reply(msg interface{}) error {
	if c.IsRPC() {
		bs, err := json.Marshal(msg)

		if err != nil {
			return err
		}

		return c.conn.Publish(c.msg.Reply, bs)
	}

	return ErrUnexpectedReply
}

func (c *natsContext) Trigger(topic string, event interface{}) error {
	return trigger(c.conn, topic, event)
}

func (c *natsContext) Request(topic string, msg interface{}, timeout int) (api.Deserializer, error) {
	return request(c.conn, topic, msg, timeout)
}

func (c *natsContext) RequestContext(ctx context.Context, topic string, msg interface{}) (api.Deserializer, error) {
	return requestContext(ctx, c.conn, topic, msg)
}

func (c *natsContext) Forward(topic string, msg interface{}) error {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := json.Marshal(msg)

		if err != nil {
			return err
		}

		if c.IsRPC() {
			return c.conn.PublishRequest(topic, c.msg.Reply, bs)
		} else {
			return c.conn.Publish(topic, bs)
		}
	} else {
		if topic != "" && natsMsg.Subject == "" {
			natsMsg.Subject = topic
		}

		return c.conn.PublishMsg(natsMsg)
	}
}

func (c *natsContext) IsRPC() bool {
	return c.msg.Reply != ""
}

func trigger(conn *nats.Conn, topic string, msg interface{}) error {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := json.Marshal(msg)

		if err != nil {
			return err
		}

		return conn.Publish(topic, bs)
	} else {
		if topic != "" && natsMsg.Subject == "" {
			natsMsg.Subject = topic
		}

		return conn.PublishMsg(natsMsg)
	}
}

func request(conn *nats.Conn, topic string, msg interface{}, timeout int) (api.Deserializer, error) {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := json.Marshal(msg)

		if err != nil {
			return nil, err
		}

		reply, err := conn.Request(topic, bs, time.Duration(timeout)*time.Second)

		if err != nil {
			return nil, err
		}

		return newNatsContext(conn, reply), nil
	} else {
		if topic != "" && natsMsg.Subject == "" {
			natsMsg.Subject = topic
		}

		reply, err := conn.RequestMsg(natsMsg, time.Duration(timeout*int(time.Second)))

		if err != nil {
			return nil, err
		}

		return newNatsContext(conn, reply), nil
	}
}

func requestContext(ctx context.Context, conn *nats.Conn, topic string, msg interface{}) (api.Deserializer, error) {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := json.Marshal(msg)

		if err != nil {
			return nil, err
		}

		reply, err := conn.RequestWithContext(ctx, topic, bs)

		if err != nil {
			return nil, err
		}

		return newNatsContext(conn, reply), nil
	} else {
		if topic != "" && natsMsg.Subject == "" {
			natsMsg.Subject = topic
		}

		reply, err := conn.RequestMsgWithContext(ctx, natsMsg)

		if err != nil {
			return nil, err
		}

		return newNatsContext(conn, reply), nil
	}
}
