package rpc

import (
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

func (c *natsContext) Json(to interface{}) error {
	return json.Unmarshal(c.msg.Data, to)
}

func (c *natsContext) Raw() []byte {
	return c.msg.Data
}

func (c *natsContext) Reply(msg interface{}) error {
	if c.CanReply() {
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

func (c *natsContext) Forward(topic string, msg interface{}) error {
	bs, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	if c.CanReply() {
		return c.conn.PublishRequest(topic, c.msg.Reply, bs)
	} else {
		return c.conn.Publish(topic, bs)
	}
}

func (c *natsContext) CanReply() bool {
	return c.msg.Reply != ""
}

func trigger(conn *nats.Conn, topic string, msg interface{}) error {
	bs, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	return conn.Publish(topic, bs)
}

func request(conn *nats.Conn, topic string, msg interface{}, timeout int) (api.Deserializer, error) {
	bs, err := json.Marshal(msg)

	if err != nil {
		return nil, err
	}

	reply, err := conn.Request(topic, bs, time.Duration(timeout)*time.Second)

	if err != nil {
		return nil, err
	}

	return newNatsContext(conn, reply), nil
}
