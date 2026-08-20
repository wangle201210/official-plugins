// This file verifies the connection/service-scoped naming client factory.

package collection

import (
	"testing"

	"github.com/dellinger2023/net-flux/pkg/naming"
)

// TestDiscoClientFactoryReusesClientPerConnectionAndService verifies the
// upstream naming client is created once for one connection/service pair.
func TestDiscoClientFactoryReusesClientPerConnectionAndService(t *testing.T) {
	client := &fakeDiscoveryClient{}
	created := 0
	factory := newDiscoClientFactory(naming.DiscoSetting{})
	factory.newClient = func(naming.DiscoSetting) (discoveryClient, error) {
		created++
		return client, nil
	}

	first, err := factory.GetDiscoClient(7, "media-node")
	if err != nil {
		t.Fatalf("get first discovery client: %v", err)
	}
	second, err := factory.GetDiscoClient(7, "media-node")
	if err != nil {
		t.Fatalf("get second discovery client: %v", err)
	}
	if first != second {
		t.Fatal("expected one discovery client per connection and service")
	}
	if created != 1 {
		t.Fatalf("expected one client creation, got %d", created)
	}

	factory.RemoveAll(7)
	if !client.closed {
		t.Fatal("expected connection cleanup to close the discovery client")
	}
}
