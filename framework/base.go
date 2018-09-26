package framework

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Meduzz/rpc/api"
	"github.com/nats-io/go-nats"
)

type (
	Service struct {
		conn          *nats.Conn
		subscriptions map[string]*nats.Subscription
		handlers      map[string]interface{}
		in            chan *nats.Msg
	}

	Builder struct {
		conn          *nats.Conn
		in            chan *nats.Msg
		subscriptions map[string]*nats.Subscription
		handlers      map[string]interface{}
	}

	Worker  func(*api.Req) (*api.Res, error)
	Eventer func(*api.Req) error
)

func NewBuilder() Builder {
	return Builder{
		in:            make(chan *nats.Msg),
		subscriptions: make(map[string]*nats.Subscription),
		handlers:      make(map[string]interface{}),
	}
}

func (b Builder) Nats(connectUrl string, options ...nats.Option) Builder {
	conn, err := nats.Connect(connectUrl, options...)
	b.conn = conn

	if err != nil {
		panic(err)
	}

	return b
}

func (b Builder) WorkerGroup(topic, group string, worker Worker) Builder {
	sub, _ := b.conn.ChanQueueSubscribe(topic, group, b.in)

	b.handlers[topic] = worker
	b.subscriptions[topic] = sub

	return b
}

func (b Builder) Worker(topic string, worker Worker) Builder {
	sub, _ := b.conn.ChanSubscribe(topic, b.in)

	b.handlers[topic] = worker
	b.subscriptions[topic] = sub

	return b
}

func (b Builder) EventGroup(topic, group string, eventer Eventer) Builder {
	sub, _ := b.conn.ChanQueueSubscribe(topic, group, b.in)

	b.handlers[topic] = eventer
	b.subscriptions[topic] = sub

	return b
}

func (b Builder) Event(topic string, eventer Eventer) Builder {
	sub, _ := b.conn.ChanSubscribe(topic, b.in)

	b.handlers[topic] = eventer
	b.subscriptions[topic] = sub

	return b
}

func (b Builder) RawGroup(topic, group string, handler nats.MsgHandler) Builder {
	sub, _ := b.conn.ChanQueueSubscribe(topic, group, b.in)

	b.handlers[topic] = handler
	b.subscriptions[topic] = sub

	return b
}

func (b Builder) Raw(topic string, handler nats.MsgHandler) Builder {
	sub, _ := b.conn.ChanSubscribe(topic, b.in)

	b.handlers[topic] = handler
	b.subscriptions[topic] = sub

	return b
}

func (b Builder) Connection() *nats.Conn {
	return b.conn
}

func (b Builder) Build() (*Service, error) {
	s := &Service{b.conn, b.subscriptions, b.handlers, b.in}

	return s, nil
}

func (s *Service) Start() {
	for {
		select {
		case msg := <-s.in:
			handler := s.handlers[msg.Subject]
			switch h := handler.(type) {
			case Worker:
				res, err := s.worker(h, msg.Data)
				if err != nil {
					handleError(err, s.conn, msg.Reply)
				} else {
					handleResponse(res, s.conn, msg.Reply)
				}
			case Eventer:
				s.eventer(h, msg.Data)
			case nats.MsgHandler:
				h(msg)
			}
		}
	}
}

func (s *Service) Unsubscribe(topic string) error {
	err := s.subscriptions[topic].Drain()

	if err != nil {
		return err
	}

	return s.subscriptions[topic].Unsubscribe()
}

func (s *Service) worker(w Worker, msg []byte) (*api.Res, error) {
	req := &api.Req{}
	err := json.Unmarshal(msg, req)

	if err != nil {
		return nil, err
	}

	return w(req)
}

func (s *Service) eventer(e Eventer, msg []byte) error {
	req := &api.Req{}
	err := json.Unmarshal(msg, req)

	if err != nil {
		return err
	}

	return e(req)
}

func Request(conn *nats.Conn, path string, req *api.Req) *api.Res {
	jsonBytes, err := json.Marshal(req)

	if err != nil {
		return errToRes(err)
	}

	msg, err := conn.Request(path, jsonBytes, 3*time.Second)

	if err != nil {
		return errToRes(err)
	}

	res := &api.Res{}
	err = json.Unmarshal(msg.Data, res)

	if err != nil {
		return errToRes(err)
	}

	return res
}

func Trigger(conn *nats.Conn, path string, event interface{}) {
	jsonBytes, err := json.Marshal(event)

	if err != nil {
		return
	}

	conn.Publish(path, jsonBytes)
}

func DecodeBytes(source, dst []byte) error {
	_, err := base64.StdEncoding.Decode(source, dst)
	return err
}

func DecodeString(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func EncodeBytes(source []byte) string {
	return base64.StdEncoding.EncodeToString(source)
}

func handleError(err error, conn *nats.Conn, reply string) {
	res := errToRes(err)

	bs, _ := json.Marshal(res)

	conn.Publish(reply, bs)
}

func errToRes(err error) *api.Res {
	res := &api.Res{}
	headers := make(map[string]string)
	headers["Content-Type"] = "text/html"

	res.Code = 500
	res.Metadata = headers
	res.Body = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", err)))

	return res
}

func handleResponse(res *api.Res, conn *nats.Conn, reply string) error {
	bs, _ := json.Marshal(res)

	return conn.Publish(reply, bs)
}
