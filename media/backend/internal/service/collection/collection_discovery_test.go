// This file verifies the Nacos discovery client integration when a local Nacos
// server is explicitly provided by the test environment.

package collection

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"
)

// TestNewDeregisterInstanceParamUsesEphemeral verifies deregistration targets
// the same temporary instance type used by registration.
func TestNewDeregisterInstanceParamUsesEphemeral(t *testing.T) {
	param := newDeregisterInstanceParam("media-node", "901", "127.0.0.1", 19091)

	if !param.Ephemeral {
		t.Fatal("expected deregister request to target ephemeral Nacos instances")
	}
	if param.ServiceName != "media-node" || param.GroupName != "901" ||
		param.Ip != "127.0.0.1" || param.Port != 19091 {
		t.Fatalf("unexpected deregister request: %#v", param)
	}
}

// TestNacosDiscoveryClientIntegration verifies register, lookup, and deregister
// against a real Nacos server. It is skipped unless LINAPRO_TEST_NACOS=1 is set.
func TestNacosDiscoveryClientIntegration(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_NACOS") != "1" {
		t.Skip("set LINAPRO_TEST_NACOS=1 to run against a local Nacos server")
	}

	cfg := defaultDiscoveryConfig()
	cfg.Enabled = true
	cfg.Host = envString("LINAPRO_TEST_NACOS_HOST", defaultDiscoveryHost)
	cfg.Port = envInt(t, "LINAPRO_TEST_NACOS_PORT", defaultDiscoveryPort)
	cfg.LogDir = t.TempDir()
	cfg.CacheDir = t.TempDir()
	baseURL := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)

	runtime := newDiscoveryRuntime(cfg)
	defer runtime.Close()

	instanceName := fmt.Sprintf("linapro-media-collection-test-%d", time.Now().UnixNano())
	instance := &gen.Instance{
		InstanceName: instanceName,
		PrivateIp:    "127.0.0.1",
		PrivatePort:  19091,
		InnerIp:      "127.0.0.1",
		InnerPort:    19092,
		PublicIp:     "127.0.0.1",
		PublicPort:   19093,
		Node:         901,
	}

	if err := runtime.Register(instance); err != nil {
		t.Fatalf("register Nacos instance: %v", err)
	}
	registered := true
	defer func() {
		if registered {
			_ = runtime.Deregister(&gen.Deregister{
				InstanceName: instanceName,
				Ip:           instance.PrivateIp,
				Port:         instance.PrivatePort,
				Node:         instance.Node,
			})
		}
		deleteNacosTestService(t, baseURL, instanceName, nodeGroup(instance.Node))
	}()

	ack := waitForLookupAck(t, runtime, instanceName, instance.Node)
	service := ack.GetServices()[0]
	if service.GetName() != instanceName {
		t.Fatalf("expected service name %s, got %s", instanceName, service.GetName())
	}
	if service.GetGroupName() != nodeGroup(instance.Node) {
		t.Fatalf("expected group %s, got %s", nodeGroup(instance.Node), service.GetGroupName())
	}
	found := service.GetInstances()[0]
	if found.GetPrivateIp() != instance.PrivateIp || found.GetPrivatePort() != instance.PrivatePort {
		t.Fatalf("unexpected lookup instance private endpoint: %s:%d", found.GetPrivateIp(), found.GetPrivatePort())
	}

	if err := runtime.Deregister(&gen.Deregister{
		InstanceName: instanceName,
		Ip:           instance.PrivateIp,
		Port:         instance.PrivatePort,
		Node:         instance.Node,
	}); err != nil {
		t.Fatalf("deregister Nacos instance: %v", err)
	}
	registered = false
	deleteNacosTestService(t, baseURL, instanceName, nodeGroup(instance.Node))
}

// waitForLookupAck waits for Nacos registration propagation.
func waitForLookupAck(t *testing.T, runtime *discoveryRuntime, serviceName string, node int32) *gen.LookupAck {
	t.Helper()

	var lastErr error
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		ack, err := runtime.Lookup(&gen.Lookup{
			ServiceName: serviceName,
			Node:        node,
			Healthy:     true,
		})
		if err == nil && len(ack.GetServices()) > 0 && len(ack.GetServices()[0].GetInstances()) > 0 {
			return ack
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("lookup Nacos instance did not return a service before timeout, lastErr=%v", lastErr)
	return nil
}

// envString returns a string environment value or fallback.
func envString(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// envInt returns an integer environment value or fallback.
func envInt(t *testing.T, key string, fallback int) int {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("parse %s: %v", key, err)
	}
	return parsed
}

// deleteNacosTestService removes the empty service left after deregistration.
func deleteNacosTestService(t *testing.T, baseURL string, serviceName string, groupName string) {
	t.Helper()

	values := url.Values{}
	values.Set("serviceName", serviceName)
	values.Set("groupName", groupName)
	request, err := http.NewRequest(http.MethodDelete, baseURL+"/nacos/v1/ns/service?"+values.Encode(), nil)
	if err != nil {
		t.Fatalf("create Nacos service delete request: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Logf("delete Nacos test service skipped after request error: %v", err)
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Fatalf("close Nacos delete response body: %v", err)
		}
	}()
	if response.StatusCode >= http.StatusBadRequest {
		t.Logf("delete Nacos test service returned status %s", response.Status)
	}
}
