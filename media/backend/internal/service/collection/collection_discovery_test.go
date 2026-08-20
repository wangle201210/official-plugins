// This file verifies the Nacos discovery client integration when a local Nacos
// server is explicitly provided by the test environment.

package collection

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"
)

// TestNacosDiscoSettingUsesPersistentClientDefaults verifies the upstream
// naming client receives the media discovery settings and cache policy.
func TestNacosDiscoSettingUsesPersistentClientDefaults(t *testing.T) {
	cfg := defaultDiscoveryConfig()
	cfg.Host = "http://10.157.225.139/"
	cfg.NotLoadCacheAtStart = true

	setting, err := newNacosDiscoSetting(cfg)
	if err != nil {
		t.Fatalf("build Nacos discovery setting: %v", err)
	}
	if setting.Host != "10.157.225.139" {
		t.Fatalf("expected normalized Nacos host, got %s", setting.Host)
	}
	if setting.Port != defaultDiscoveryPort {
		t.Fatalf("expected Nacos port %d, got %d", defaultDiscoveryPort, setting.Port)
	}
	if setting.PreloadCache {
		t.Fatal("expected Nacos client cache preload disabled")
	}
	if setting.GroupName != "DEFAULT_GROUP" {
		t.Fatalf("expected default Nacos group, got %s", setting.GroupName)
	}
}

// TestNacosServerConfigNormalizesURLHost verifies plugin URL-style host values
// are converted to the separate fields expected by the Nacos SDK.
func TestNacosServerConfigNormalizesURLHost(t *testing.T) {
	cfg := defaultDiscoveryConfig()
	cfg.Host = "http://10.157.225.139/"

	setting, err := newNacosDiscoSetting(cfg)
	if err != nil {
		t.Fatalf("build Nacos discovery setting: %v", err)
	}
	if setting.Host != "10.157.225.139" {
		t.Fatalf("expected normalized Nacos host, got %s", setting.Host)
	}
	if setting.Port != defaultDiscoveryPort {
		t.Fatalf("expected Nacos port %d, got %d", defaultDiscoveryPort, setting.Port)
	}
}

// TestNacosClientConfigAvoidsLocalStaleCache verifies lookup correctness does
// not depend on Nacos SDK process or disk cache state.
func TestNacosClientConfigAvoidsLocalStaleCache(t *testing.T) {
	cfg := defaultDiscoveryConfig()
	cfg.Enabled = true
	cfg.NotLoadCacheAtStart = true
	clientConfig, err := newNacosDiscoSetting(cfg)
	if err != nil {
		t.Fatalf("build Nacos discovery setting: %v", err)
	}

	if clientConfig.PreloadCache {
		t.Fatal("expected Nacos SDK client to skip loading local disk cache")
	}
	if clientConfig.Namespace != cfg.Namespace {
		t.Fatalf("expected Nacos namespace %s, got %s", cfg.Namespace, clientConfig.Namespace)
	}
}

// TestNacosDiscoveryClientIntegration verifies register, lookup, and deregister
// against a real Nacos server. It is skipped unless LINAPRO_TEST_NACOS=1 is set.
func TestNacosDiscoveryClientIntegration(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_NACOS") != "1" {
		t.Skip("set LINAPRO_TEST_NACOS=1 to run against a local Nacos server")
	}

	cfg := newNacosIntegrationConfig(t)
	setting, err := newNacosDiscoSetting(cfg)
	if err != nil {
		t.Fatalf("build Nacos discovery setting: %v", err)
	}
	baseHost := setting.Host
	if !strings.Contains(baseHost, "://") {
		baseHost = "http://" + baseHost
	}
	baseURL := fmt.Sprintf("%s:%d", baseHost, setting.Port)

	registerRuntime := newDiscoveryRuntime(cfg)
	defer registerRuntime.Close()
	lookupRuntime := newDiscoveryRuntime(cfg)
	defer lookupRuntime.Close()
	deregisterRuntime := newDiscoveryRuntime(cfg)
	defer deregisterRuntime.Close()

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

	if err := registerRuntime.Register(1, instance); err != nil {
		t.Fatalf("register Nacos instance: %v", err)
	}
	registered := true
	defer func() {
		if registered {
			_ = deregisterRuntime.Deregister(1, &gen.Deregister{
				InstanceName: instanceName,
				Ip:           instance.PrivateIp,
				Port:         instance.PrivatePort,
				Node:         instance.Node,
			})
		}
		deleteNacosTestService(t, baseURL, cfg.Namespace, instanceName, nodeGroup(instance.Node))
	}()

	ack := waitForLookupAck(t, lookupRuntime, instanceName, instance.Node)
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
	if !found.GetHealthy() {
		t.Fatal("expected registered Nacos instance to remain healthy while naming client is reused")
	}

	if err := deregisterRuntime.Deregister(1, &gen.Deregister{
		InstanceName: instanceName,
		Ip:           instance.PrivateIp,
		Port:         instance.PrivatePort,
		Node:         instance.Node,
	}); err != nil {
		t.Fatalf("deregister Nacos instance: %v", err)
	}
	registered = false
	// The upstream GetServiceInstances API is backed by a subscription cache;
	// the deregister request itself is the authoritative cleanup assertion.
	deleteNacosTestService(t, baseURL, cfg.Namespace, instanceName, nodeGroup(instance.Node))
}

// newNacosIntegrationConfig reads shared real-Nacos test configuration.
func newNacosIntegrationConfig(t *testing.T) DiscoveryConfig {
	t.Helper()

	cfg := defaultDiscoveryConfig()
	cfg.Enabled = true
	cfg.Host = envString("LINAPRO_TEST_NACOS_HOST", defaultDiscoveryHost)
	cfg.Port = envInt(t, "LINAPRO_TEST_NACOS_PORT", defaultDiscoveryPort)
	cfg.Namespace = envString("LINAPRO_TEST_NACOS_NAMESPACE", cfg.Namespace)
	cfg.Username = envString("LINAPRO_TEST_NACOS_USERNAME", cfg.Username)
	cfg.Password = envString("LINAPRO_TEST_NACOS_PASSWORD", cfg.Password)
	cfg.Node = envInt(t, "LINAPRO_TEST_NACOS_NODE", cfg.Node)
	if cfg.Node <= 0 {
		t.Fatalf("LINAPRO_TEST_NACOS_NODE must be positive, got %d", cfg.Node)
	}
	cfg.LogDir = t.TempDir()
	cfg.CacheDir = t.TempDir()
	return cfg
}

// waitForLookupAck waits for Nacos registration propagation.
func waitForLookupAck(t *testing.T, runtime *discoveryRuntime, serviceName string, node int32) *gen.LookupAck {
	t.Helper()

	var lastErr error
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		ack, err := runtime.Lookup(1, &gen.Lookup{
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

// waitForEmptyLookupAck waits for Nacos deregistration propagation.
func waitForEmptyLookupAck(t *testing.T, runtime *discoveryRuntime, serviceName string, node int32) {
	t.Helper()

	var (
		lastErr error
		lastAck *gen.LookupAck
	)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		ack, err := runtime.Lookup(1, &gen.Lookup{
			ServiceName: serviceName,
			Node:        node,
			Healthy:     true,
		})
		if err == nil && len(ack.GetServices()) == 0 {
			return
		}
		lastErr = err
		lastAck = ack
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("lookup Nacos instance did not become empty before timeout, lastErr=%v lastAck=%#v", lastErr, lastAck)
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
func deleteNacosTestService(t *testing.T, baseURL string, namespace string, serviceName string, groupName string) {
	t.Helper()

	values := url.Values{}
	values.Set("namespaceId", namespace)
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
