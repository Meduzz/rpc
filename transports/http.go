package transports

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Meduzz/rpc/api"
	"github.com/gin-gonic/gin"
)

type (
	// Bind your rpc methods onto a http url.
	HttpRpcServer struct {
		address string
		server  *gin.Engine
	}

	HttpRpcClient struct {
		host string
	}
)

// Creates a gin server, and binds it to address on Start()
func NewHttpTransport(address string) api.RpcServer {
	return &HttpRpcServer{address, gin.Default()}
}

// Embedds into a gin server you created, does nothing on Start()
func NewEmbeddedHttpTransport(engine *gin.Engine) api.RpcServer {
	return &HttpRpcServer{"", engine}
}

func NewHttpRpcClient(host string) api.RpcClient {
	return &HttpRpcClient{host}
}

func (h *HttpRpcServer) RegisterWorker(function string, handler api.Worker) {
	h.server.POST(function, h.workerWrapper(handler))
}

func (h *HttpRpcServer) RegisterEventer(function string, handler api.Eventer) {
	h.server.POST(function, h.eventerWrapper(handler))
}

func (h *HttpRpcServer) Start() {
	if h.address != "" {
		h.server.Run(h.address)
	}
}

func (h *HttpRpcServer) workerWrapper(handler api.Worker) func(*gin.Context) {
	return func(ctx *gin.Context) {
		meta := make(map[string]string, 0)

		for k, v := range ctx.Request.Header {
			meta[k] = strings.Join(v, " ")
		}

		meta["Content-Type"] = ctx.ContentType()
		meta["X-Client-Ip"] = ctx.ClientIP()
		meta["X-Request-Id"] = generate()
		data, err := ctx.GetRawData()

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		if len(data) > MAX_BODY {
			ctx.AbortWithError(400, errors.New("body is too large"))
			return
		}

		req := api.NewEmptyMessage()
		req.Body = json.RawMessage(data)
		req.Metadata = meta

		res := handler(req)

		contentType := "application/json"

		for k, v := range res.Metadata {
			if k == "Content-Type" {
				contentType = v
			} else {
				ctx.Header(k, v)
			}
		}

		switch res.Metadata["result"] {
		case "success":
			ctx.Data(200, contentType, []byte(res.Body))
		case "error":
			ctx.Data(500, contentType, []byte(res.Body))
		default:
			dto := &api.ErrorDTO{"no result header present in metadata"}
			ctx.AbortWithStatusJSON(500, dto)
		}
	}
}

func (h *HttpRpcServer) eventerWrapper(handler api.Eventer) func(*gin.Context) {
	return func(ctx *gin.Context) {
		meta := make(map[string]string, 0)
		meta["Content-Type"] = ctx.ContentType()
		meta["X-Client-Ip"] = ctx.ClientIP()
		meta["X-Request-Id"] = generate()
		data, err := ctx.GetRawData()

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		if len(data) > MAX_BODY {
			ctx.AbortWithError(400, errors.New("body is too large"))
			return
		}

		req := api.NewEmptyMessage()
		req.Body = json.RawMessage(data)
		req.Metadata = meta

		handler(req)

		ctx.Status(200)
	}
}

func (h *HttpRpcClient) Request(function string, body *api.Message) (*api.Message, error) {
	req, err := http.NewRequest("POST", h.url(function), bytes.NewReader([]byte(body.Body)))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	for k, v := range body.Metadata {
		req.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	retBs, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	ret := api.NewEmptyMessage()

	if res.StatusCode == 200 {
		ret.Metadata["result"] = "success"
	} else {
		ret.Metadata["result"] = "error"
	}
	ret.Body = json.RawMessage(retBs)

	for k, v := range res.Header {
		ret.Metadata[k] = strings.Join(v, " ")
	}

	return ret, nil
}

func (h *HttpRpcClient) Trigger(function string, body *api.Message) error {
	req, err := http.NewRequest("POST", h.url(function), bytes.NewReader([]byte(body.Body)))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	for k, v := range body.Metadata {
		req.Header.Add(k, v)
	}

	_, err = http.DefaultClient.Do(req)

	if err != nil {
		return err
	} else {
		return nil
	}
}

func (h *HttpRpcClient) url(ctx string) string {
	return fmt.Sprintf("%s%s", h.host, ctx)
}
