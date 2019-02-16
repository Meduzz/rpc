package transports

import (
	"fmt"
	"time"

	"github.com/Meduzz/rpc/api"
)

type (
	LocalRpcServer struct {
		name     string
		workers  map[string]api.Worker
		eventers map[string]api.Eventer
		handlers map[string]api.Handler
	}

	localContext struct {
		retChan chan *api.Message
		msg     *api.Message
		client  api.RpcClient
	}
)

func NewLocalRpcServer(name string) (api.RpcServer, error) {
	workers := make(map[string]api.Worker, 0)
	eventers := make(map[string]api.Eventer, 0)
	handlers := make(map[string]api.Handler, 0)
	return &LocalRpcServer{name, workers, eventers, handlers}, nil
}

func NewLocalRpcClient(server *LocalRpcServer) api.RpcClient {
	return api.RpcClient(server)
}

func (l *LocalRpcServer) Request(function string, body *api.Message) (*api.Message, error) {
	for k, v := range l.workers {
		if k == function {
			ret := v(body)

			return ret, nil
		}
	}

	for k, v := range l.handlers {
		if k == function {
			retChan := make(chan *api.Message, 1)
			v(newLocalContext(retChan, body, l))

			var ret *api.Message
			var err error
			ticker := time.Tick(3 * time.Second)

			select {
			case ret = <-retChan:
				break
			case <-ticker:
				err = ErrTimeout
				break
			}

			return ret, err
		}
	}

	return api.NewErrorMessage(fmt.Sprintf("Nothing bound to: %s", function)), nil
}

func (l *LocalRpcServer) Trigger(function string, body *api.Message) error {
	for k, v := range l.eventers {
		if k == function {
			v(body)
			return nil
		}
	}

	for k, v := range l.handlers {
		if k == function {
			v(newLocalContext(nil, body, l))
			return nil
		}
	}

	return nil
}

func (l *LocalRpcServer) RegisterWorker(function string, handler api.Worker) {
	l.workers[function] = handler
}

func (l *LocalRpcServer) RegisterEventer(function string, handler api.Eventer) {
	l.eventers[function] = handler
}

func (l *LocalRpcServer) RegisterHandler(function string, handler api.Handler) {
	l.handlers[function] = handler
}

func (l *LocalRpcServer) Start(block bool) {
}

func (l *LocalRpcServer) Remove(function string) {
	_, ok := l.eventers[function]

	if ok {
		delete(l.eventers, function)
	}

	_, ok = l.workers[function]

	if ok {
		delete(l.workers, function)
	}

	_, ok = l.handlers[function]

	if ok {
		delete(l.handlers, function)
	}
}

func newLocalContext(retChan chan *api.Message, msg *api.Message, client api.RpcClient) *localContext {
	return &localContext{retChan, msg, client}
}

func (c *localContext) Body() (*api.Message, error) {
	return c.msg, nil
}

func (c *localContext) End() {
	if c.retChan != nil {
		close(c.retChan)
	}
}

func (c *localContext) Reply(msg *api.Message) error {
	if c.retChan != nil {
		c.retChan <- msg
		close(c.retChan)
	}

	return nil
}

func (c *localContext) Trigger(topic string, event *api.Message) error {
	return c.client.Trigger(topic, event)
}

func (c *localContext) Request(topic string, event *api.Message) (*api.Message, error) {
	return c.client.Request(topic, event)
}

func (c *localContext) Forward(topic string, message *api.Message) error {
	// This will have to be good enough for local? Reply will not be possible though.
	return c.client.Trigger(topic, message)
}
