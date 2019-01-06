package proxy

import (
	"log"
	"net/http"

	"github.com/nats-io/go-nats"

	"./encoding"
	"./hub"
	"github.com/Meduzz/rpc/api"
	"github.com/julienschmidt/httprouter"
)

type (
	Proxy struct {
		hosts    map[string]*httprouter.Router
		encoding encoding.Codec
		client   api.RpcClient
	}
)

const defaultHost = "default"

// NewProxy - create a new proxy object.
// If the codec is nil, the default codec will be used.
// If the client is nil, the code panics.
func NewProxy(codec encoding.Codec, client api.RpcClient) *Proxy {
	hosts := make(map[string]*httprouter.Router, 0)

	if codec == nil {
		codec = encoding.NewDefaultCodec()
	}

	if client == nil {
		panic("No client present.")
	}

	p := &Proxy{
		hosts:    hosts,
		encoding: codec,
		client:   client,
	}

	return p
}

func (p *Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// look for a matching host
	// if match, use it
	// if no match, use default host

	var handler *httprouter.Router
	if handler = p.hosts[req.Host]; handler == nil {
		handler = p.hosts[defaultHost]
	}

	handler.ServeHTTP(res, req)
}

func (p *Proxy) Start(bind string) error {
	log.Printf("Server listening on: %s.\n", bind)
	return http.ListenAndServe(bind, p)
}

// Add the method path combo to the host, or the default host.
// No guarantees are made to make sure this combo is not already set.
// In that case the router will still use the old combo, but the proxy
// return you a new hub.
func (p *Proxy) Add(host *string, method, path string) *hub.Hub {
	// if host is set
	// lookup if the host is already registered
	// else register the host
	// add method and path binding to host
	// if host is not set
	// add method and path to default host

	var handler *httprouter.Router
	var ns string

	if host != nil {
		ns = *host
	} else {
		ns = defaultHost
	}

	if handler = p.hosts[ns]; handler == nil {
		handler = &httprouter.Router{}
		p.hosts[ns] = handler
	}

	hub := hub.NewHub()
	handler.Handle(method, path, p.handleRequest(hub))

	log.Printf("Added handler for %s requests on (%s) to %s.", method, ns, path)

	return hub
}

func (p *Proxy) handleRequest(bridge *hub.Hub) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
		living := bridge.Liveliness()

		if living == hub.DEAD {
			res.WriteHeader(503)
			return
		}

		var err error
		req, err = bridge.Filter(req)

		if err != nil {
			log.Printf("Filter threw error: %s.", err.Error())
			status := bridge.Status(err)
			res.WriteHeader(status)
			res.Write([]byte(err.Error()))
			return
		}

		ps := make(map[string]string, 0)

		for _, param := range params {
			ps[param.Key] = param.Value
		}

		topic, rpc := bridge.Route(req, ps)

		if topic == "" {
			res.WriteHeader(404)
			return
		}

		var msg *api.Message
		if enc := bridge.Encoder(); enc != nil {
			msg = enc(req, ps)
		} else {
			msg = p.encoding.FromRequest(req, ps)
		}

		if rpc {
			reply, err := p.client.Request(topic, msg)

			if err != nil {
				if err == nats.ErrTimeout {
					if living == hub.UNKNOWN {
						res.WriteHeader(503)
						log.Printf("%s did not return in time: %s.", topic, err.Error())
					} else {
						res.WriteHeader(500)
						log.Printf("%s caused error: %s.", topic, err.Error())
					}
				} else {
					res.WriteHeader(500)
					log.Printf("%s caused error: %s.", topic, err.Error())
				}
			}

			if dec := bridge.Decoder(); dec != nil {
				dec(reply, res)
			} else {
				p.encoding.ToResponse(reply, res)
			}
		} else {
			p.client.Trigger(topic, msg)
			res.WriteHeader(200)
		}
	}
}
