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
	channel struct {
		queue string
		group string
	}

	Service struct {
		conn *nats.Conn
	}

	Builder struct {
		conn    *nats.Conn
		workers map[channel]Worker
		event   map[channel]Eventer
		raw     map[channel]nats.MsgHandler
	}

	Worker  func(*api.Req) (*api.Res, error)
	Eventer func(*api.Req) error
)

func NewBuilder() Builder {
	return Builder{
		workers: make(map[channel]Worker),
		event:   make(map[channel]Eventer),
		raw:     make(map[channel]nats.MsgHandler),
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
	c := channel{topic, group}
	b.workers[c] = worker

	return b
}

func (b Builder) Worker(topic string, worker Worker) Builder {
	return b.WorkerGroup(topic, "", worker)
}

func (b Builder) EventGroup(topic, group string, eventer Eventer) Builder {
	c := channel{topic, group}
	b.event[c] = eventer

	return b
}

func (b Builder) Event(topic string, eventer Eventer) Builder {
	return b.EventGroup(topic, "", eventer)
}

func (b Builder) RawGroup(topic, group string, handler nats.MsgHandler) Builder {
	c := channel{topic, group}
	b.raw[c] = handler

	return b
}

func (b Builder) Raw(topic string, handler nats.MsgHandler) Builder {
	return b.RawGroup(topic, "", handler)
}

func (b Builder) Connection() *nats.Conn {
	return b.conn
}

func (b Builder) Build() (*Service, error) {
	s := &Service{b.conn}
	s.startWorkers(b.workers)
	s.startEventers(b.event)
	s.startRaw(b.raw)

	return s, nil
}

func (s *Service) startWorkers(workers map[channel]Worker) {
	for c, w := range workers {
		if c.group != "" {
			s.conn.QueueSubscribe(c.queue, c.group, s.worker(w))
		} else {
			s.conn.Subscribe(c.queue, s.worker(w))
		}
	}
}

func (s *Service) startEventers(eventers map[channel]Eventer) {
	for c, e := range eventers {
		if c.group != "" {
			s.conn.QueueSubscribe(c.queue, c.group, s.eventer(e))
		} else {
			s.conn.Subscribe(c.queue, s.eventer(e))
		}
	}
}

func (s *Service) startRaw(raw map[channel]nats.MsgHandler) {
	for c, h := range raw {
		if c.group != "" {
			s.conn.QueueSubscribe(c.queue, c.group, h)
		} else {
			s.conn.Subscribe(c.queue, h)
		}
	}
}

func (s *Service) worker(w Worker) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		req := &api.Req{}
		err := json.Unmarshal(msg.Data, req)

		if err != nil {
			handleError(err, s.conn, msg.Reply)
		}

		res, err := w(req)

		if err != nil {
			handleError(err, s.conn, msg.Reply)
		} else {
			handleResponse(res, s.conn, msg.Reply)
		}
	}
}

func (s *Service) eventer(e Eventer) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		req := &api.Req{}
		err := json.Unmarshal(msg.Data, req)

		if err != nil {
			handleError(err, s.conn, msg.Reply)
		}

		err = e(req)

		if err != nil {
			handleError(err, s.conn, msg.Reply)
		}
	}
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
