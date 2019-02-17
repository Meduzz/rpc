package rpc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Meduzz/rpc/api"
	nats "github.com/nats-io/go-nats"
)

func newNatsContext(conn *nats.Conn, msg *nats.Msg) *natsContext {
	return &natsContext{
		conn: conn,
		msg:  msg,
	}
}

func (c *natsContext) Body() (*api.Message, error) {
	msg := &api.Message{}
	err := json.Unmarshal(c.msg.Data, msg)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (c *natsContext) Reply(msg *api.Message) error {
	if c.msg.Reply != "" {
		bs, err := json.Marshal(msg)

		if err != nil {
			return err
		}

		return c.conn.Publish(c.msg.Reply, bs)
	}

	return fmt.Errorf("Message did not expect reply")
}

func (c *natsContext) Trigger(topic string, event *api.Message) error {
	return trigger(c.conn, topic, event)
}

func (c *natsContext) Request(topic string, msg *api.Message) (*api.Message, error) {
	return request(c.conn, topic, msg)
}

func (c *natsContext) Forward(topic string, msg *api.Message) error {
	bs, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	if c.msg.Reply != "" {
		return c.conn.PublishRequest(topic, c.msg.Reply, bs)
	} else {
		return c.conn.Publish(topic, bs)
	}
}

func trigger(conn *nats.Conn, topic string, msg *api.Message) error {
	bs, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	return conn.Publish(topic, bs)
}

func request(conn *nats.Conn, topic string, msg *api.Message) (*api.Message, error) {
	bs, err := json.Marshal(msg)

	if err != nil {
		return nil, err
	}

	reply, err := conn.Request(topic, bs, 3*time.Second)

	if err != nil {
		if err == nats.ErrTimeout {
			return nil, ErrTimeout
		}

		return nil, err
	}

	ret := &api.Message{}
	err = json.Unmarshal(reply.Data, ret)

	if err != nil {
		return nil, err
	}

	return ret, nil
}