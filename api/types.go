package api

import "context"

type (
	// Handler a function that takes a Context.
	Handler func(Context)

	Deserializer interface {
		// Bind body of message to param
		Bind(interface{}) error
		// Return the raw body
		Raw() []byte
	}

	// Context a new take to simplify things.
	Context interface {
		Deserializer
		// Reply reply with something we can turn into json.
		Reply(interface{}) error
		// Trigger an event (where event is something we can turn into json).
		Trigger(string, interface{}) error
		// Request a response (where request is something we can turn into json).
		Request(string, interface{}, int) (Deserializer, error)
		// RequestContext request a response using a context instead of a timeout
		RequestContext(context.Context, string, interface{}) (Deserializer, error)
		// Forward a message (where the message is something we can turn into json).
		Forward(string, interface{}) error
		// IsRPC lets us know if the message had reply topic set
		IsRPC() bool
	}
)
