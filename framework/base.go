package framework

import (
	"encoding/base64"
	"encoding/json"
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
		conn     *nats.Conn
		handlers map[string]interface{}
		specs    []*TopicSpec
	}

	TopicSpec struct {
		Topic string
		Group string
	}

	Worker     func(*api.Req) (*api.Res, error)
	Eventer    func(*api.Req) error
	RawWorker  func([]byte) ([]byte, error)
	RawEventer func([]byte) error
)

func NewBuilder() *Builder {
	return &Builder{
		handlers: make(map[string]interface{}),
		specs:    make([]*TopicSpec, 0),
	}
}

func (b *Builder) Nats(connectUrl string, options ...nats.Option) {
	conn, err := nats.Connect(connectUrl, options...)
	b.conn = conn

	if err != nil {
		panic(err)
	}
}

func (b *Builder) WorkerGroup(topic, group string, worker Worker) {
	b.addSpecc(&TopicSpec{topic, group})
	b.handlers[topic] = worker
}

func (b *Builder) Worker(topic string, worker Worker) {
	b.addSpecc(&TopicSpec{topic, ""})
	b.handlers[topic] = worker
}

func (b *Builder) EventGroup(topic, group string, eventer Eventer) {
	b.addSpecc(&TopicSpec{topic, group})
	b.handlers[topic] = eventer
}

func (b *Builder) Event(topic string, eventer Eventer) {
	b.addSpecc(&TopicSpec{topic, ""})
	b.handlers[topic] = eventer
}

func (b *Builder) RawGroupEvent(topic, group string, handler RawEventer) {
	b.addSpecc(&TopicSpec{topic, group})
	b.handlers[topic] = handler
}

func (b *Builder) RawEvent(topic string, handler RawEventer) {
	b.addSpecc(&TopicSpec{topic, ""})
	b.handlers[topic] = handler
}

func (b *Builder) RawWorkerGroup(topic, group string, handler RawWorker) {
	b.addSpecc(&TopicSpec{topic, group})
	b.handlers[topic] = handler
}

func (b *Builder) RawWorker(topic string, handler RawWorker) {
	b.addSpecc(&TopicSpec{topic, ""})
	b.handlers[topic] = handler
}

func (b *Builder) Connection() *nats.Conn {
	return b.conn
}

func (b *Builder) addSpecc(spec *TopicSpec) {
	b.specs = append(b.specs, spec)
}

func (b *Builder) Build() (*Service, error) {
	in := make(chan *nats.Msg)
	s := &Service{b.conn, make(map[string]*nats.Subscription), b.handlers, in}

	for _, specc := range b.specs {
		if specc.Group != "" {
			sub, err := b.conn.ChanQueueSubscribe(specc.Topic, specc.Group, in)
			s.subscriptions[specc.Topic] = sub

			if err != nil {
				println(err.Error())
			}
		} else {
			sub, err := b.conn.ChanSubscribe(specc.Topic, in)
			s.subscriptions[specc.Topic] = sub

			if err != nil {
				println(err.Error())
			}
		}
	}

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
			case RawWorker:
				bs, err := h(msg.Data)
				if err != nil {
					dto := errToDTO(err)
					dtoBs, _ := json.Marshal(dto)
					s.conn.Publish(msg.Reply, dtoBs)
				} else {
					s.conn.Publish(msg.Reply, bs)
				}
			case RawEventer:
				h(msg.Data)
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
	headers["Content-Type"] = "application/json"

	dto := errToDTO(err)
	bs, _ := json.Marshal(dto)

	res.Code = 500
	res.Metadata = headers
	res.Body = base64.StdEncoding.EncodeToString(bs)

	return res
}

func errToDTO(err error) *api.ErrorDTO {
	dto := &api.ErrorDTO{}

	dto.Message = err.Error()

	return dto
}

func handleResponse(res *api.Res, conn *nats.Conn, reply string) error {
	bs, _ := json.Marshal(res)

	return conn.Publish(reply, bs)
}
