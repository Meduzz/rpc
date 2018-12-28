package transports

import (
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/Meduzz/rpc/api"
	"github.com/nats-io/go-nats"
)

type (
	NatsRpcServer struct {
		conn   *nats.Conn
		name   string
		queued bool
	}

	NatsRpcClient struct {
		conn *nats.Conn
	}

	natsContext struct {
		conn *nats.Conn
		msg  *nats.Msg
	}
)

func NewNatsRpcServer(serviceName, url string, options []nats.Option, queued bool) (api.RpcServer, error) {
	conn, err := nats.Connect(url, options...)

	if err != nil {
		return nil, err
	}

	return &NatsRpcServer{conn, serviceName, queued}, nil
}

func NewNatsRpcServerConn(serviceName string, conn *nats.Conn, queued bool) api.RpcServer {
	return &NatsRpcServer{conn, serviceName, queued}
}

func NewNatsRpcClient(url string, options []nats.Option) (api.RpcClient, error) {
	conn, err := nats.Connect(url, options...)

	if err != nil {
		return nil, err
	}

	return &NatsRpcClient{conn}, nil
}

func NewNatsRpcClientConn(conn *nats.Conn) api.RpcClient {
	return &NatsRpcClient{conn}
}

func (t *NatsRpcClient) Request(function string, body *api.Message) (*api.Message, error) {
	bs, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	msg, err := t.conn.Request(function, bs, 3*time.Second)

	if err != nil {
		return nil, err
	}

	ret := &api.Message{}
	err = json.Unmarshal(msg.Data, ret)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (t *NatsRpcClient) Trigger(function string, body *api.Message) error {
	bs, err := json.Marshal(body)

	if err != nil {
		return err
	}

	return t.conn.Publish(function, bs)
}

func (t *NatsRpcServer) RegisterWorker(function string, handler api.Worker) {
	if t.queued {
		t.conn.QueueSubscribe(function, t.name, t.workerWrapper(handler))
	} else {
		t.conn.Subscribe(function, t.workerWrapper(handler))
	}
}

func (t *NatsRpcServer) RegisterEventer(function string, handler api.Eventer) {
	if t.queued {
		t.conn.QueueSubscribe(function, t.name, t.eventerWrapper(handler))
	} else {
		t.conn.Subscribe(function, t.eventerWrapper(handler))
	}
}

func (t *NatsRpcServer) RegisterHandler(function string, handler api.Handler) {
	if t.queued {
		t.conn.QueueSubscribe(function, t.name, t.handlerWrapper(handler))
	} else {
		t.conn.Subscribe(function, t.handlerWrapper(handler))
	}
}

func (t *NatsRpcServer) Start() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	t.conn.Close()
}

func (t *NatsRpcServer) workerWrapper(handler api.Worker) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		req := &api.Message{}
		json.Unmarshal(msg.Data, req)

		res := handler(req)

		bs, _ := json.Marshal(res)
		t.conn.Publish(msg.Reply, bs)
	}
}

func (t *NatsRpcServer) eventerWrapper(handler api.Eventer) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		req := &api.Message{}
		json.Unmarshal(msg.Data, req)

		handler(req)
	}
}

func (t *NatsRpcServer) handlerWrapper(handler api.Handler) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		ctx := newNatsContext(t.conn, msg)
		handler(ctx)
	}
}

func newNatsContext(conn *nats.Conn, msg *nats.Msg) *natsContext {
	return &natsContext{
		conn: conn,
		msg:  msg,
	}
}

func (c *natsContext) BodyAsJSON(into interface{}) error {
	return json.Unmarshal(c.msg.Data, into)
}

func (c *natsContext) BodyAsBytes() []byte {
	return c.msg.Data
}

func (c *natsContext) BodyAsMessage() (*api.Message, error) {
	msg := &api.Message{}
	err := json.Unmarshal(c.msg.Data, msg)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (c *natsContext) End() {}
func (c *natsContext) ReplyJSON(data interface{}) error {
	if c.msg.Reply != "" {
		bs, err := json.Marshal(data)

		if err != nil {
			return err
		}

		return c.conn.Publish(c.msg.Reply, bs)
	}

	return nil // TODO potentially we return an error here instead.
}

func (c *natsContext) ReplyBinary(data []byte) error {
	if c.msg.Reply != "" {
		return c.conn.Publish(c.msg.Reply, data)
	}

	return nil // TODO potentially we return an error here instead.
}

func (c *natsContext) ReplyMessage(msg *api.Message) error {
	if c.msg.Reply != "" {
		bs, err := json.Marshal(msg)

		if err != nil {
			return err
		}

		return c.conn.Publish(c.msg.Reply, bs)
	}

	return nil // TODO potentially we return an error here instead.
}

func (c *natsContext) Event(topic string, event []byte) error {
	return c.conn.Publish(topic, event)
}
