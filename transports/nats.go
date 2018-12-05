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
