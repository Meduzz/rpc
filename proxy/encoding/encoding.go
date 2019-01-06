package encoding

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Meduzz/rpc/proxy/util"

	"github.com/Meduzz/rpc/api"
)

type (
	Codec interface {
		// FromRequest craft a Message from a http request and it's path params.
		FromRequest(*http.Request, map[string]string) *api.Message
		// FromResponse write a message to a http response.
		ToResponse(*api.Message, http.ResponseWriter)
	}

	defaultCodec struct{}
)

func NewDefaultCodec() Codec {
	return &defaultCodec{}
}

func (c *defaultCodec) FromRequest(req *http.Request, params map[string]string) *api.Message {
	util.Headers(req, params, "Content-Type")
	util.RemoteAddr(req, "Remote-Addr", params)

	params["Request-Id"] = generate()

	user, pass, ok := req.BasicAuth()

	if ok {
		params["Username"] = user
		params["Password"] = pass
	}

	body, _ := ioutil.ReadAll(req.Body)

	msg := api.NewBytesMessage(body)
	msg.Metadata = params

	return msg
}

func (c *defaultCodec) ToResponse(msg *api.Message, res http.ResponseWriter) {
	switch msg.Metadata["result"] {
	case "success":
		for k, v := range msg.Metadata {
			res.Header().Set(k, v)
		}
		res.Write(msg.Body)
	case "error":
		for k, v := range msg.Metadata {
			res.Header().Set(k, v)
		}
		res.WriteHeader(500)
		res.Write(msg.Body)
	default:
		dto := &api.ErrorDTO{"no result header present in metadata"}
		bs, _ := json.Marshal(dto)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(500)
		res.Write(bs)
	}
}

func generate() string {
	bs := make([]byte, 100)
	rand.Read(bs)

	hasher := sha1.New()
	hasher.Write(bs)
	return hex.EncodeToString(hasher.Sum(nil))
}
