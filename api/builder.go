package api

import "encoding/json"

func Builder() *MessageBuilder {
	return &MessageBuilder{NewEmptyMessage()}
}

func (m *MessageBuilder) Message() *Message {
	return m.message
}

func (m *MessageBuilder) Header(key, value string) {
	m.message.Metadata[key] = value
}

func (m *MessageBuilder) Text(text string) {
	m.Bytes([]byte(text))
}

func (m *MessageBuilder) Bytes(bs []byte) {
	m.message.Body = json.RawMessage(bs)
}

func (m *MessageBuilder) Json(any interface{}) error {
	bs, err := json.Marshal(any)

	if err != nil {
		return err
	}

	m.Bytes(bs)

	return nil
}
