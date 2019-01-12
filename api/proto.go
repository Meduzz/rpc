package api

import (
	"encoding/json"
)

type (
	// Worker a function that takes a Message and returns a Message.
	Worker func(*Message) *Message
	// Eventer a function that takes a Message.
	Eventer func(*Message)
	// Handler a function that takes a Context.
	Handler func(Context)

	// RpcClient base interface for all rpc clients.
	RpcClient interface {
		// Request - send a request to function and expect a reply.
		// Implementations are expected to handle timeouts and return
		// a ErrTimeout over a "protocol"-specific error, in those cases.
		Request(function string, message *Message) (*Message, error)
		// Trigger - send an event to function.
		Trigger(function string, message *Message) error
	}

	// RpcServer base interface for all rpc servers.
	RpcServer interface {
		// RegisterWorker lets you bind a Worker to function.
		RegisterWorker(function string, handler Worker)
		// RegisterEventer lets you bind an Eventer to a function.
		RegisterEventer(function string, handler Eventer)
		// RegisterHandler lets you bind a Handler to a function.
		RegisterHandler(function string, handler Handler)
		// Start depends on the impl, but usually this will block.
		Start()
	}

	// Message contains what's needed for rpc.
	Message struct {
		Metadata map[string]string `json:"metadata,omitempty"`
		Body     json.RawMessage   `json:"body,omitempty"`
	}

	// ErrorDTO the body that's used when ever there's a server error.
	ErrorDTO struct {
		Message string `json:"message"`
	}

	// Context a new take to simplify things.
	Context interface {
		// Body fetch the body and bind it to a Message
		Body() (*Message, error)
		// End let the context clean up after itself.
		End()
		// Reply reply with a message.
		Reply(*Message) error
		// Trigger an event.
		Trigger(string, *Message) error
		// Request a response.
		Request(string, *Message) (*Message, error)
		// Forward a message
		Forward(string, *Message) error
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
