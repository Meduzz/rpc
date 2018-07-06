package framework

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

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
		natsUrl  string
		natsOpts []nats.Option
		workers  map[channel]Worker
		event    map[channel]Eventer
	}

	Worker  func(*api.Req) (*api.Res, error)
	Eventer func(*api.Req) error
)

func NewBuilder() Builder {
	return Builder{
		workers: make(map[channel]Worker),
		event:   make(map[channel]Eventer),
	}
}

func (b Builder) Nats(connectUrl string, options ...nats.Option) Builder {
	b.natsUrl = connectUrl
	b.natsOpts = options

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

func (b Builder) Build() (*Service, error) {
	conn, err := nats.Connect(b.natsUrl, b.natsOpts...)

	if err != nil {
		return nil, err
	}

	s := &Service{conn}
	s.startWorkers(b.workers)
	s.startEventers(b.event)

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

func handleError(err error, conn *nats.Conn, reply string) {
	res := &api.Res{}

	res.Code = 500
	res.ContentType = "text/html"
	res.Body = hex.EncodeToString([]byte(fmt.Sprintf("%i", err)))

	bs, _ := json.Marshal(res)

	conn.Publish(reply, bs)
}

func handleResponse(res *api.Res, conn *nats.Conn, reply string) error {
	bs, _ := json.Marshal(res)

	return conn.Publish(reply, bs)
}
