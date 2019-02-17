package api

import (
	"encoding/json"
)

func NewMessage(body interface{}) (*Message, error) {
	bs, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	headers := make(map[string]string, 0)
	headers["result"] = "success"
	return &Message{headers, json.RawMessage(bs)}, nil
}

func NewBytesMessage(bs []byte) *Message {
	headers := make(map[string]string, 0)
	headers["result"] = "success"

	msg := &Message{}
	msg.Metadata = headers
	msg.Body = json.RawMessage(bs)

	return msg
}

func NewEmptyMessage() *Message {
	headers := make(map[string]string, 0)

	msg := &Message{}
	msg.Metadata = headers

	return msg
}

func NewErrorMessage(errorMessage string) *Message {
	errorDto := &ErrorDTO{}
	errorDto.Message = errorMessage

	bs, err := json.Marshal(errorDto)

	if err != nil {
		// TODO now what?
		return nil
	}

	headers := make(map[string]string, 0)
	headers["result"] = "error"

	return &Message{headers, json.RawMessage(bs)}
}

func (m *Message) Json(into interface{}) error {
	return json.Unmarshal(m.Body, into)
}

func (m *Message) String() (string, error) {
	str := ""
	err := m.Json(&str)

	if err != nil {
		return "", err
	}

	return str, nil
}
