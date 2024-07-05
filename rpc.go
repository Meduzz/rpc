package rpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Meduzz/rpc/encoding"
	nats "github.com/nats-io/nats.go"
)

type (
	// RpcHandler a function that takes a RpcContext.
	RpcHandler func(*RpcContext)

	// EventHandler a funciton that takes an EventContext
	EventHandler func(*EventContext)
)

func rpcWrapper(conn *nats.Conn, codec encoding.Codec, handler RpcHandler) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		ctx := NewRpcContext(conn, msg, codec)
		handler(ctx)
	}
}

func eventWrapper(conn *nats.Conn, codec encoding.Codec, handler EventHandler) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		ctx := NewEventContext(conn, msg, codec)
		handler(ctx)
	}
}

func Trigger(conn *nats.Conn, codec encoding.Codec, topic string, msg any) error {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := codec.Marshal(msg)

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

func Request(conn *nats.Conn, codec encoding.Codec, topic string, msg any, timeout int, response any) error {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := codec.Marshal(msg)

		if err != nil {
			return err
		}

		reply, err := conn.Request(topic, bs, time.Duration(timeout)*time.Second)

		if err != nil {
			return err
		}

		return json.Unmarshal(reply.Data, response)
	} else {
		if topic != "" && natsMsg.Subject == "" {
			natsMsg.Subject = topic
		}

		reply, err := conn.RequestMsg(natsMsg, time.Duration(timeout*int(time.Second)))

		if err != nil {
			return err
		}

		return json.Unmarshal(reply.Data, response)
	}
}

func RequestContext(ctx context.Context, conn *nats.Conn, codec encoding.Codec, topic string, msg, response any) error {
	natsMsg, ok := msg.(*nats.Msg)

	if !ok {
		bs, err := codec.Marshal(msg)

		if err != nil {
			return err
		}

		reply, err := conn.RequestWithContext(ctx, topic, bs)

		if err != nil {
			return err
		}

		return codec.Unmarshal(reply.Data, response)
	} else {
		if topic != "" && natsMsg.Subject == "" {
			natsMsg.Subject = topic
		}

		reply, err := conn.RequestMsgWithContext(ctx, natsMsg)

		if err != nil {
			return err
		}

		return codec.Unmarshal(reply.Data, response)
	}
}

func HandleRPC(conn *nats.Conn, codec encoding.Codec, topic, queue string, handler RpcHandler) (*nats.Subscription, error) {
	if queue != "" {
		sub, err := conn.QueueSubscribe(topic, queue, rpcWrapper(conn, codec, handler))

		if err != nil {
			return nil, err
		}

		return sub, nil
	} else {
		sub, err := conn.Subscribe(topic, rpcWrapper(conn, codec, handler))

		if err != nil {
			return nil, err
		}

		return sub, nil
	}
}

func HandleEvent(conn *nats.Conn, codec encoding.Codec, topic, queue string, handler EventHandler) (*nats.Subscription, error) {
	if queue != "" {
		sub, err := conn.QueueSubscribe(topic, queue, eventWrapper(conn, codec, handler))

		if err != nil {
			return nil, err
		}

		return sub, nil
	} else {
		sub, err := conn.Subscribe(topic, eventWrapper(conn, codec, handler))

		if err != nil {
			return nil, err
		}

		return sub, nil
	}
}
