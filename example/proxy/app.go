package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Meduzz/rpc/api"
	"github.com/Meduzz/rpc/proxy"
	"github.com/Meduzz/rpc/proxy/util"
	"github.com/Meduzz/rpc/transports"
)

const MAX_BODY = 512 * 1024

func main() {
	client, _ := transports.NewNatsRpcClient("nats://localhost:4222", nil)
	server := proxy.NewProxy(nil, client)
	hub := server.Add(nil, "POST", "/rpc")

	hub.SetFilter(filter, nil)
	hub.SetRoute(routing)
	hub.SetEncoder(encoder)

	server.Start(":8080")
}

func filter(req *http.Request) (*http.Request, error) {
	query := req.URL.Query()

	rpc := query.Get("rpc")
	if rpc == "" {
		return nil, fmt.Errorf("No rpc type set, must be either (event or request)")
	}

	action := query.Get("action")
	if action == "" {
		return nil, fmt.Errorf("No action set, dont know how to route this request")
	}

	if req.ContentLength > MAX_BODY {
		return nil, fmt.Errorf("Body is too large")
	}

	return req, nil
}

func encoder(req *http.Request, params map[string]string) *api.Message {
	query := req.URL.Query()
	operation := query.Get("operation")

	if operation != "" {
		params["operation"] = operation
	}

	util.Headers(req, params, "Content-Type")
	util.RemoteAddr(req, "Remote-Addr", params)

	params["Request-Id"] = generate()

	user, pass, ok := req.BasicAuth()

	if ok {
		params["Username"] = user
		params["Password"] = pass
	}

	bs, _ := ioutil.ReadAll(req.Body)

	msg := api.NewBytesMessage(bs)

	msg.Metadata = params

	return msg
}

func routing(req *http.Request, params map[string]string) (string, bool) {
	query := req.URL.Query()

	rpc := query.Get("rpc")
	action := query.Get("action")

	if rpc == "event" {
		return action, false
	}

	return action, true
}

func generate() string {
	bs := make([]byte, 100)
	rand.Read(bs)

	hasher := sha1.New()
	hasher.Write(bs)
	return hex.EncodeToString(hasher.Sum(nil))
}
