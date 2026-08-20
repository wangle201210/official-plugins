// This file manages one reusable net-flux naming client per TCP connection
// and service, matching the upstream discovery server lifecycle.

package collection

import (
	"sync"

	"github.com/dellinger2023/net-flux/pkg/naming"
	"github.com/gogf/gf/v2/errors/gerror"
)

// discoveryClient aliases the upstream naming client contract used by media.
type discoveryClient = naming.DiscoClient

// discoClientFactory stores one naming client per TCP connection and service.
type discoClientFactory struct {
	sync.RWMutex
	settings       naming.DiscoSetting
	serviceStorage map[uint32]map[string]naming.DiscoClient
	newClient      func(naming.DiscoSetting) (discoveryClient, error)
}

// newDiscoClientFactory creates a production discovery client factory.
func newDiscoClientFactory(settings naming.DiscoSetting) *discoClientFactory {
	return &discoClientFactory{
		settings:       settings,
		serviceStorage: make(map[uint32]map[string]naming.DiscoClient),
		newClient:      naming.NewNacosDiscoverClient,
	}
}

// GetDiscoClient returns the reusable client for one connection and service.
func (f *discoClientFactory) GetDiscoClient(connID uint32, serviceName string) (discoveryClient, error) {
	if f == nil {
		return nil, gerror.New("media discovery client factory cannot be nil")
	}
	if serviceName == "" {
		serviceName = "web"
	}

	f.RLock()
	if clients, ok := f.serviceStorage[connID]; ok {
		if client, ok := clients[serviceName]; ok {
			f.RUnlock()
			return client, nil
		}
	}
	f.RUnlock()

	if f.newClient == nil {
		return nil, gerror.New("media discovery client creator cannot be nil")
	}
	client, err := f.newClient(f.settings)
	if err != nil {
		return nil, gerror.Wrap(err, "create media discovery client failed")
	}

	f.Lock()
	defer f.Unlock()
	clients := f.serviceStorage[connID]
	if clients == nil {
		clients = make(map[string]naming.DiscoClient)
		f.serviceStorage[connID] = clients
	}
	if existing, ok := clients[serviceName]; ok {
		client.Close()
		return existing, nil
	}
	clients[serviceName] = client
	return client, nil
}

// RemoveDiscoClient closes and removes one connection-scoped service client.
func (f *discoClientFactory) RemoveDiscoClient(connID uint32, serviceName string) {
	if f == nil {
		return
	}
	if serviceName == "" {
		serviceName = "web"
	}
	f.Lock()
	defer f.Unlock()
	if clients, ok := f.serviceStorage[connID]; ok {
		if client, ok := clients[serviceName]; ok {
			client.Close()
			delete(clients, serviceName)
		}
		if len(clients) == 0 {
			delete(f.serviceStorage, connID)
		}
	}
}

// RemoveAll closes and removes every client belonging to one TCP connection.
func (f *discoClientFactory) RemoveAll(connID uint32) {
	if f == nil {
		return
	}
	f.Lock()
	defer f.Unlock()
	if clients, ok := f.serviceStorage[connID]; ok {
		for _, client := range clients {
			client.Close()
		}
		delete(f.serviceStorage, connID)
	}
}

// Close releases every naming client owned by the factory.
func (f *discoClientFactory) Close() {
	if f == nil {
		return
	}
	f.Lock()
	defer f.Unlock()
	for connID, clients := range f.serviceStorage {
		for _, client := range clients {
			client.Close()
		}
		delete(f.serviceStorage, connID)
	}
}
