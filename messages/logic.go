package messages

import (
	"fmt"

	"github.com/Meduzz/rpc/encoding"
	"github.com/nats-io/nats.go"
)

const (
	CodeHeader    = "code"
	MessageHeader = "message"
	ContentHeader = "contentType"
)

func CreateMessage(topic string, codec encoding.Codec, building func(*MsgBuilder)) (*nats.Msg, error) {
	msg := nats.NewMsg(topic)
	builder := &MsgBuilder{}
	building(builder)

	if builder.code != 0 {
		msg.Header.Add(CodeHeader, fmt.Sprintf("%d", builder.code))
	}

	if builder.msg != "" {
		msg.Header.Add(MessageHeader, builder.msg)
	}

	if builder.body != nil {
		bs, err := codec.Marshal(builder.body)

		if err != nil {
			return nil, err
		}

		msg.Data = bs
		msg.Header.Add(ContentHeader, codec.Mime())
	}

	return msg, nil
}
