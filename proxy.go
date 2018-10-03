package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/Meduzz/rpc/api"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/go-nats"
)

const MAX_BODY = 512 * 1024
const REQRES = "REQRES"
const EVENT = "EVENT"

func main() {
	natsURL := os.Getenv("NATS_URL")
	natsUser := os.Getenv("NATS_USER")
	natsPass := os.Getenv("NATS_PASS")
	natsOptions := make([]nats.Option, 0)

	if natsUser != "" {
		natsOptions = append(natsOptions, nats.UserInfo(natsUser, natsPass))
	}

	conn, err := nats.Connect(natsURL, natsOptions...)

	if err != nil {
		panic(err)
	}

	server := gin.Default()

	server.POST("/rpc", func(ctx *gin.Context) {
		meta := make(map[string]string, 0)
		meta["Content-Type"] = ctx.ContentType()
		meta["X-Client-Ip"] = ctx.ClientIP()
		meta["X-Request-Id"] = generate()

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
			ctx.AbortWithStatus(400)
			return
		}

		body := encode(data)
		req := &api.Req{meta, body}
		jsonBody, err := json.Marshal(req)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		if rpc == REQRES {
			// RPC mode, expect an answer.
			msg, err := conn.Request(action, jsonBody, 3*time.Second)

			if err != nil {
				ctx.AbortWithError(500, err)
				return
			}

			res := &api.Res{}
			err = json.Unmarshal(msg.Data, res)

			if err != nil {
				ctx.AbortWithError(500, err)
				return
			}

			bodyBytes, err := decode(res.Body)

			if err != nil {
				ctx.AbortWithError(500, err)
				return
			}

			contentType := "application/json"

			for k, v := range res.Metadata {
				if k == "Content-Type" {
					contentType = v
				} else {
					ctx.Header(k, v)
				}
			}

			ctx.Data(res.Code, contentType, bodyBytes)
		} else if rpc == EVENT {
			// EVENT mode, fire and forget.
			err = conn.Publish(action, jsonBody)

			if err != nil {
				ctx.AbortWithError(500, err)
				return
			} else {
				ctx.Status(200)
			}
		} else {
			ctx.AbortWithStatus(400)
		}
	})

	server.Run(":8081")
}

func generate() string {
	bs := make([]byte, 100)
	rand.Read(bs)

	hasher := sha1.New()
	hasher.Write(bs)
	return hex.EncodeToString(hasher.Sum(nil))
}

func encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
