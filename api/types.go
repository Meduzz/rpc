package api

type (
	// Handler a function that takes a Context.
	Handler func(Context)

	// Context a new take to simplify things.
	Context interface {
		// Bind body of message to param
		Json(interface{}) error
		// Return the raw body
		Raw() []byte
		// Return the body as a string
		Text() string
		// Reply reply with something we can turn into json.
		Reply(interface{}) error
		// Trigger an event (where event is something we can turn into json).
		Trigger(string, interface{}) error
		// Request a response (where request is something we can turn into json).
		Request(string, interface{}, int) (Context, error)
		// Forward a message (where the message is something we can turn into json).
		Forward(string, interface{}) error
		// CanReply lets us know if the message had reply topic set
		CanReply() bool
	}
)
