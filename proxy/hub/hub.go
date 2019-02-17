package hub

import (
	"net/http"

	"github.com/Meduzz/rpc/api"
)

type (
	// Hub will act as a bridge into the server.
	// With the hub you can to manipulate presence (200->503->404),
	// you can control routing. You can even abort requests.
	Hub struct {
		routing RouteFunc
		filter  FilterFunc
		errors  ErrorFunc
		encoder EncodeFunc
		decoder DecodeFunc
	}

	// Route a small struct that carries what the proxy
	// needs to make the descition to route and to where
	// and if to expect a response.
	Route struct {
		Healthy bool
		Topic   string
		RPC     bool
		Timeout int
	}

	RouteFunc  func(*http.Request, map[string]string) *Route
	FilterFunc func(*http.Request) (*http.Request, error)
	ErrorFunc  func(error) int
	EncodeFunc func(*http.Request, map[string]string) *api.Message
	DecodeFunc func(*api.Message, http.ResponseWriter)
)

func NewHub() *Hub {
	return &Hub{}
}

/*
	TODO
	There are 2 sides of Hub, what the proxy sees and what the impl sees. Atm both see everything, that could change to slim down the api a bit.
*/

// Route - The proxy will call this to get a nats topic to
// send the request to. The bool tells the proxy whether its
// a rpc or an event call. This method will delegate to
// your RouterFunc.
// Returning empty string makes the proxy return 404.
func (h *Hub) Route(req *http.Request, params map[string]string) *Route {
	if h.routing != nil {
		return h.routing(req, params)
	}

	return &Route{false, "", false, 0}
}

// SetRoute - sets the routing to use for this hub.
func (h *Hub) SetRoute(routing RouteFunc) {
	h.routing = routing
}

// SetFilter  - allows you to execute arbetrary checks
// on the request, change it, or abort the request with an error.
// Second param, errors lets you map your errors to status codes,
// later used by the proxy. Also note, that errors wont be called
// unless there's a filter.
func (h *Hub) SetFilter(filter FilterFunc, errors ErrorFunc) {
	h.filter = filter
	h.errors = errors
}

// Filter - used by the proxy to execute any FilterFuncs.
func (h *Hub) Filter(req *http.Request) (*http.Request, error) {
	if h.filter != nil {
		return h.filter(req)
	}

	return req, nil
}

// Status - used by the proxy to map any errors generated
// by the previous filter. Defaults to 500 if not set.
func (h *Hub) Status(err error) int {
	if h.errors != nil {
		return h.errors(err)
	}

	return 500
}

func (h *Hub) SetEncoder(enc EncodeFunc) {
	h.encoder = enc
}

func (h *Hub) SetDecoder(dec DecodeFunc) {
	h.decoder = dec
}

func (h *Hub) Encoder() EncodeFunc {
	return h.encoder
}

func (h *Hub) Decoder() DecodeFunc {
	return h.decoder
}
