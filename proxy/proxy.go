package proxy

import (
	"log"
	"net/http"
	"strings"

	"github.com/Meduzz/rpc"
	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/proxy/encoding"
	"github.com/Meduzz/rpc/proxy/hub"
	"github.com/julienschmidt/httprouter"
)

type (
	Proxy struct {
		hosts    map[string]*httprouter.Router
		encoding encoding.Codec
		rpc      *rpc.RPC
	}
)

const defaultHost = "default"

// NewProxy - create a new proxy object.
// If the codec is nil, the default codec will be used. The codec set here will be the fallback codec if the Hub does not have one.
// If the client is nil, the code panics.
func NewProxy(codec encoding.Codec, rpc *rpc.RPC) *Proxy {
	hosts := make(map[string]*httprouter.Router, 0)

	if codec == nil {
		codec = encoding.NewDefaultCodec()
	}

	if rpc == nil {
		panic("No client present.")
	}

	p := &Proxy{
		hosts:    hosts,
		encoding: codec,
		rpc:      rpc,
	}

	return p
}

func (p *Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// clean up host
	// look for a matching host
	// if match, use it
	// if no match, use default host

	host := req.Host

	// TODO this will break with ipv6.
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	var handler *httprouter.Router
	if handler = p.hosts[host]; handler == nil {
		handler = p.hosts[defaultHost]
		log.Printf("[%s] %s %s\n", defaultHost, req.Method, req.RequestURI)
	} else {
		log.Printf("[%s] %s %s\n", host, req.Method, req.RequestURI)
	}

	handler.ServeHTTP(res, req)
}

// Start - bind the web server to bind.
func (p *Proxy) Start(bind string) error {
	log.Printf("Server listening on: %s.\n", bind)
	return http.ListenAndServe(bind, p)
}

// Add the method path combo to the host, or the default host.
// No guarantees are made to make sure this combo is not already set.
// In that case the router will still use the old combo, but the proxy
// return you a new hub. That's likely to change in the future though.
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

	log.Printf("Added handler for %s %s requests to namespace: [%s].", method, path, ns)

	return hub
}

func (p *Proxy) handleRequest(bridge *hub.Hub) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
		var err error
		req, err = bridge.Filter(req)

		if err != nil {
			status := bridge.Status(err)
			res.WriteHeader(status)
			res.Write([]byte(err.Error()))
			return
		}

		ps := make(map[string]string, 0)

		for _, param := range params {
			ps[param.Key] = param.Value
		}

		route := bridge.Route(req, ps)

		if !route.Healthy && route.Topic != "" {
			res.WriteHeader(503)
			return
		}

		if route.Topic == "" {
			res.WriteHeader(404)
			return
		}

		var msg *api.Message
		if enc := bridge.Encoder(); enc != nil {
			msg = enc(req, ps)
		} else {
			msg = p.encoding.FromRequest(req, ps)
		}

		if route.RPC {
			reply, err := p.rpc.Request(route.Topic, msg, route.Timeout)

			if err != nil {
				log.Printf("Request to [%s] threw error: %s.\n", route.Topic, err)

				if err == rpc.ErrTimeout {
					res.WriteHeader(503)
				} else {
					res.WriteHeader(500)
				}

				return
			}

			if dec := bridge.Decoder(); dec != nil {
				dec(reply, res)
			} else {
				p.encoding.ToResponse(reply, res)
			}
		} else {
			p.rpc.Trigger(route.Topic, msg)
			res.WriteHeader(200)
		}
	}
}
