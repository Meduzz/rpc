package discovery

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Meduzz/rpc/api"
)

type (
	Discovery struct {
		client   api.RpcClient
		server   api.RpcServer
		registry *Registry
		funcs    []*address
		settings *Settings
	}

	Settings struct {
		// an optional namespace to limit discoveries to
		Namespace string
		// an optional version, defaults to DEVELOPMENT
		Version string
		// The topic where discoveries will be sent to and received from
		DiscoveryTopic string
		// any particular fqn:s these rpc functions are interested in (lets us ignore the rest)
		Interests []string
		// how old (in seconds) can a last seen address be before we're not interested of it (missed hello)
		MaxAge int
		// send discoveries with this interval (seconds)
		DiscoveryInterval int
	}

	address struct {
		FQN       string `json:"fqn"`
		Topic     string `json:"topic"`
		Version   string `json:"version"`
		Namespace string `json:"namespace"`
	}

	Addresses []*address
)

func NewDiscovery(client api.RpcClient, server api.RpcServer, settings *Settings) *Discovery {
	if settings.Namespace == "" {
		settings.Namespace = "*"
	}

	if settings.DiscoveryTopic == "" {
		settings.DiscoveryTopic = "discovery"
	}

	if settings.Interests == nil {
		settings.Interests = make([]string, 0)
	}

	if settings.Version == "" {
		settings.Version = "DEVELOPMENT"
	}

	if settings.MaxAge == 0 {
		settings.MaxAge = 30
	}

	if settings.DiscoveryInterval == 0 {
		settings.DiscoveryInterval = 15
	}

	registry := NewRegistry(settings.Interests...)
	funcs := make([]*address, 0)

	return &Discovery{client, server, registry, funcs, settings}
}

func (d *Discovery) RegisterEventer(topic string, eventer api.Eventer, fqn string) {
	if fqn == "" {
		panic("fqn cant be empty")
	}

	d.server.RegisterEventer(topic, eventer)
	d.funcs = append(d.funcs, &address{fqn, topic, d.settings.Version, d.settings.Namespace})
}

func (d *Discovery) RegisterWorker(topic string, worker api.Worker, fqn string) {
	if fqn == "" {
		panic("fqn cant be empty")
	}

	d.server.RegisterWorker(topic, worker)
	d.funcs = append(d.funcs, &address{fqn, topic, d.settings.Version, d.settings.Namespace})
}

func (d *Discovery) RegisterHandler(topic string, handler api.Handler, fqn string) {
	if fqn == "" {
		panic("fqn cant be empty")
	}

	d.server.RegisterHandler(topic, handler)
	d.funcs = append(d.funcs, &address{fqn, topic, d.settings.Version, d.settings.Namespace})
}

func (d *Discovery) Start(block bool) {
	d.hello(true)
	d.enableDiscovery()

	go d.scheduledHello()

	d.server.Start(block)
}

func (d *Discovery) Remove(topic string) {
	d.server.Remove(topic)

	keepers := make([]*address, 0)

	for _, self := range d.funcs {
		if self.Topic != topic {
			keepers = append(keepers, self)
		}
	}

	d.funcs = keepers
}

func (d *Discovery) Trigger(fqn, version string, message *api.Message) error {
	if fqn == "" {
		return fmt.Errorf("fqn must be set")
	}

	addr, err := d.find(fqn, version)

	if err != nil {
		return err
	}

	return d.client.Trigger(addr.Topic, message)
}

func (d *Discovery) Request(fqn, version string, message *api.Message) (*api.Message, error) {
	if fqn == "" {
		return nil, fmt.Errorf("fqn must be set")
	}

	addr, err := d.find(fqn, version)

	if err != nil {
		return nil, err
	}

	return d.client.Request(addr.Topic, message)
}

func (d *Discovery) find(fqn, version string) (*address, error) {
	if version == "" {
		version = "*"
	}

	addr := d.registry.Find(fqn, d.settings.Namespace, version, d.settings.MaxAge)

	if addr == nil {
		if d.settings.Namespace != "*" && version == "*" {
			addr = d.registry.Find(fqn, "*", version, d.settings.MaxAge)

			if addr == nil {
				return nil, fmt.Errorf("Did not find any rpc functions for fqn:%s in neither local (%s) or global (*) namespace, matching version: %s", fqn, d.settings.Namespace, version)
			}

			return addr, nil
		}

		return nil, fmt.Errorf("Did not find any rpc functions for fqn:%s in global namespace matching version: %s", fqn, version)
	}

	return addr, nil
}

func (d *Discovery) scheduledHello() {
	tick := time.Tick(time.Duration(d.settings.DiscoveryInterval) * time.Second)

	for {
		select {
		case <-tick:
			d.hello(false)
			break
		}
	}
}

func (d *Discovery) hello(first bool) {
	d.triggerHello(d.funcs, first)
}

func (d *Discovery) triggerHello(self []*address, first bool) {
	hello, _ := api.NewMessage(self)

	if first {
		hello.Metadata["Register"] = "true"
	}

	d.client.Trigger(d.settings.DiscoveryTopic, hello)
}

func (d *Discovery) enableDiscovery() {
	d.server.RegisterEventer(d.settings.DiscoveryTopic, func(msg *api.Message) {
		if msg.Metadata["Register"] == "true" {
			// TODO atm we might respond to our own register...
			d.hello(false)
		}

		// TODO atm we'll add our own funcs to registry.
		addr := Addresses{}
		json.Unmarshal(msg.Body, &addr)

		for _, a := range addr {
			d.registry.Update(a)
		}
	})
}
