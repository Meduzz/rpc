package transports

import (
	"github.com/Meduzz/rpc/api"
)

type (
	LocalRpcServer struct {
		name     string
		workers  map[string]api.Worker
		eventers map[string]api.Eventer
	}
)

func NewLocalRpcServer(name string) (api.RpcServer, error) {
	workers := make(map[string]api.Worker, 0)
	eventers := make(map[string]api.Eventer, 0)
	return &LocalRpcServer{name, workers, eventers}, nil
}

func NewLocalRpcClient(server *LocalRpcServer) api.RpcClient {
	return api.RpcClient(server)
}

func (l *LocalRpcServer) Request(function string, body *api.Message) (*api.Message, error) {
	ret := l.workers[function](body)
	return ret, nil
}

func (l *LocalRpcServer) Trigger(function string, body *api.Message) error {
	l.eventers[function](body)
	return nil
}

func (l *LocalRpcServer) RegisterWorker(function string, handler api.Worker) {
	l.workers[function] = handler
}

func (l *LocalRpcServer) RegisterEventer(function string, handler api.Eventer) {
	l.eventers[function] = handler
}

func (l *LocalRpcServer) Start() {
}
