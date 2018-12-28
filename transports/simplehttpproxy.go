package transports

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nats-io/go-nats"

	"github.com/Meduzz/rpc/api"
	"github.com/gin-gonic/gin"
)

type (
	SimpleHttpProxy struct {
		context  string
		address  string
		server   *gin.Engine
		delegate api.RpcClient
	}
)

// TODO turn this into a lib

const MAX_BODY = 512 * 1024
const REQRES = "request"
const EVENT = "event"

// This replaces the old proxy.go
func NewSimpleHttpProxy(context, address string, transport api.RpcClient) api.RpcServer {
	if context == "" {
		context = "/rpc"
	}

	server := gin.Default()

	return &SimpleHttpProxy{context, address, server, transport}
}

func (s *SimpleHttpProxy) RegisterWorker(function string, handler api.Worker) {}

func (s *SimpleHttpProxy) RegisterEventer(function string, handler api.Eventer) {}

func (s *SimpleHttpProxy) RegisterHandler(function string, handler api.Handler) {}

func (s *SimpleHttpProxy) Start() {
	s.server.POST(s.context, func(ctx *gin.Context) {
		meta := make(map[string]string, 0)
		meta["Content-Type"] = ctx.ContentType()
		meta["X-Client-Ip"] = ctx.ClientIP()
		meta["X-Request-Id"] = generate()

		for k := range ctx.Request.Header {
			if k != "Content-Type" {
				meta[k] = ctx.Request.Header.Get(k)
			}
		}

		action := ctx.Query("action")

		if action == "" {
			ctx.AbortWithError(400, errors.New("no action specified"))
			return
		}

		rpc := ctx.DefaultQuery("rpc", REQRES)
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

		if rpc == REQRES {
			// RPC mode, expect an answer.
			msg, err := s.delegate.Request(action, req)

			if err != nil {
				if err == nats.ErrTimeout {
					ctx.AbortWithError(404, fmt.Errorf("worker not found: %s", action))
					return
				} else {
					ctx.AbortWithError(500, err)
					return
				}
			}

			contentType := "application/json"

			for k, v := range msg.Metadata {
				if k == "Content-Type" {
					contentType = v
				} else {
					ctx.Header(k, v)
				}
			}

			switch msg.Metadata["result"] {
			case "success":
				ctx.Data(200, contentType, []byte(msg.Body))
			case "error":
				ctx.Data(500, contentType, []byte(msg.Body))
			default:
				dto := &api.ErrorDTO{"no result header present in metadata"}
				ctx.AbortWithStatusJSON(500, dto)
			}
		} else if rpc == EVENT {
			// EVENT mode, fire and forget.
			err = s.delegate.Trigger(action, req)

			if err != nil {
				ctx.AbortWithError(500, err)
				return
			} else {
				ctx.Status(200)
			}
		} else {
			ctx.AbortWithError(400, errors.New("unsupported rpc type defined"))
		}
	})

	s.server.Run(s.address)
}

func generate() string {
	bs := make([]byte, 100)
	rand.Read(bs)

	hasher := sha1.New()
	hasher.Write(bs)
	return hex.EncodeToString(hasher.Sum(nil))
}
