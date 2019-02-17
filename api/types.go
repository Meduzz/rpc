package api

import (
	"encoding/json"
)

type (
	// Handler a function that takes a Context.
	Handler func(Context)

	// Message contains what's needed for rpc.
	Message struct {
		Metadata map[string]string `json:"metadata,omitempty"`
		Body     json.RawMessage   `json:"body,omitempty"`
	}

	MessageBuilder struct {
		message *Message
	}

	// ErrorDTO the body that's used when ever there's a server error.
	ErrorDTO struct {
		Message string `json:"message"`
	}

	// Context a new take to simplify things.
	Context interface {
		// Body fetch the body and bind it to a Message
		Body() (*Message, error)
		// Reply reply with a message.
		Reply(*Message) error
		// Trigger an event.
		Trigger(string, *Message) error
		// Request a response.
		Request(string, *Message, int) (*Message, error)
		// Forward a message
		Forward(string, *Message) error
	}
)
