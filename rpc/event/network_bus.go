package event

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

// NetworkBus - object capable of subscribing to remote event buses in addition
// to remote event bus subscribing to it's local event bus. a server and client
type NetworkBus struct {
	*Client
	*Server
	service   *NetworkBusService
	sharedBus IBus
	address   string
	path      string
}

// NewNetworkBus - returns a new network bus object at the server address and path
func NewNetworkBus(address, path string) *NetworkBus {
	bus := new(NetworkBus)
	bus.sharedBus = New()
	bus.Server = NewServer(address, path, bus.sharedBus)
	bus.Client = NewClient(address, path, bus.sharedBus)
	bus.service = &NetworkBusService{&sync.WaitGroup{}, false}
	bus.address = address
	bus.path = path
	return bus
}

// EventBus - returns wrapped event bus
func (networkBus *NetworkBus) EventBus() IBus {
	return networkBus.sharedBus
}

// NetworkBusService - object capable of serving the network bus
type NetworkBusService struct {
	wg      *sync.WaitGroup
	started bool
}

// Start - helper method to serve a network bus service
func (networkBus *NetworkBus) Start() error {
	var err error
	service := networkBus.service
	clientService := networkBus.Client.service
	serverService := networkBus.Server.service
	if !service.started {
		server := rpc.NewServer()
		_ = server.RegisterName("ServerService", serverService)
		_ = server.RegisterName("ClientService", clientService)
		server.HandleHTTP(networkBus.path, "/debug"+networkBus.path)
		l, e := net.Listen("tcp", networkBus.address)
		if e != nil {
			err = fmt.Errorf("listen error: %v", e)
		}
		service.wg.Add(1)
		go func() { _ = http.Serve(l, nil) }()
	} else {
		err = errors.New("server bus already started")
	}
	return err
}

// Stop - signal for the service to stop serving
func (networkBus *NetworkBus) Stop() {
	service := networkBus.service
	if service.started {
		service.wg.Done()
		service.started = false
	}
}
