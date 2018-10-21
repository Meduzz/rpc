package api

import (
	"encoding/json"
)

type (
	Worker  func(*Message) *Message
	Eventer func(*Message)

	RpcClient interface {
		Request(function string, message *Message) (*Message, error)
		Trigger(function string, message *Message) error
	}

	RpcServer interface {
		RegisterWorker(function string, handler Worker)
		RegisterEventer(function string, handler Eventer)
		Start()
	}

	Message struct {
		Metadata map[string]string `json:"metadata,omitempty"`
		Body     json.RawMessage   `json:"body,omitempty"`
	}

	ErrorDTO struct {
		Message string `json:"message"`
	}
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
