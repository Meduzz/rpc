package main

import (
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
	natsUrl := os.Getenv("NATS_URL")
	conn, err := nats.Connect(natsUrl)

	if err != nil {
		panic(err)
	}

	server := gin.Default()

	// TODO timeouts?
	// TODO nats settings?
	// TODO request ids?
	server.POST("/rpc", func(ctx *gin.Context) {
		meta := make(map[string]string, 0)
		meta["contentType"] = ctx.ContentType()
		meta["clientIp"] = ctx.ClientIP()

		action := ctx.Query("action")

		if action == "" {
			ctx.AbortWithError(400, errors.New("No action specified."))
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

		body := hex.EncodeToString(data)
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

			bodyBytes, err := hex.DecodeString(res.Body)

			if err != nil {
				ctx.AbortWithError(500, err)
				return
			}

			ctx.Data(res.Code, res.ContentType, bodyBytes)
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

	server.Run(":8080")
}
