package discovery

import (
	"github.com/Meduzz/rpc"
	"github.com/Meduzz/rpc/api"
)

type (
	Discovery struct {
		rpc      *rpc.RPC
		registry Registry
		settings *Settings
	}

	Settings struct {
		// The topic where discoveries will be sent to and received from
		DiscoveryTopic string
	}

	Address struct {
		ID       string            `json:"id"`
		Topic    string            `json:"topic"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}
)

func NewDiscovery(rpc *rpc.RPC, settings *Settings, registry Registry) *Discovery {
	if settings.DiscoveryTopic == "" {
		settings.DiscoveryTopic = "discovery"
	}

	d := &Discovery{rpc, registry, settings}
	d.startDiscovery()

	return d
}

func (d *Discovery) startDiscovery() {
	d.rpc.Handler(d.settings.DiscoveryTopic, "", d.discoveryHandler)
}

func (d *Discovery) discoveryHandler(ctx api.Context) {
	msg, _ := ctx.Body()
	addr := &Address{}
	msg.Json(addr)

	d.registry.Update(addr)
}

func (d *Discovery) SendDiscovery(addr *Address) {
	msg, _ := api.NewMessage(addr)
	d.rpc.Trigger(d.settings.DiscoveryTopic, msg)
}
