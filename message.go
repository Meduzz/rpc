package rpc

import (
	"github.com/Meduzz/rpc/encoding"
	"github.com/nats-io/nats.go"
)

// NewMessage - create a new message from the provided topic and body encoded with the provided codec.
func NewMessage(codec encoding.Codec, topic string, body any) (*nats.Msg, error) {
	msg := nats.NewMsg(topic)

	bs, err := codec.Marshal(body)

	if err != nil {
		return nil, err
	}

	msg.Data = bs
	return msg, nil
}
