//go:build manual_nacos

// This file provides an explicitly tagged Nacos registration test that keeps
// its persistent external state for manual inspection.

package collection

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"
)

// TestNacosDiscoveryClientRegisterWithoutCleanup intentionally leaves one
// persistent Nacos instance for manual inspection. The manual_nacos build tag
// excludes it from normal test suites, while LINAPRO_TEST_NACOS_KEEP=1 provides
// a second explicit confirmation that the caller accepts the external residue.
func TestNacosDiscoveryClientRegisterWithoutCleanup(t *testing.T) {
	if os.Getenv("LINAPRO_TEST_NACOS_KEEP") != "1" {
		t.Skip("set LINAPRO_TEST_NACOS_KEEP=1 to register and retain a Nacos instance")
	}

	cfg := newNacosIntegrationConfig(t)
	serviceName := envString(
		"LINAPRO_TEST_NACOS_KEEP_SERVICE",
		fmt.Sprintf("linapro-media-kept-test-%d", time.Now().UnixNano()),
	)
	instanceIP := envString("LINAPRO_TEST_NACOS_KEEP_IP", "127.0.0.1")
	instancePort := envInt(t, "LINAPRO_TEST_NACOS_KEEP_PORT", 19091)
	if instancePort <= 0 || instancePort > 65535 {
		t.Fatalf("LINAPRO_TEST_NACOS_KEEP_PORT must be between 1 and 65535, got %d", instancePort)
	}

	runtime := newDiscoveryRuntime(cfg)
	defer runtime.Close()
	instance := &gen.Instance{
		InstanceId:   serviceName,
		InstanceName: serviceName,
		PrivateIp:    instanceIP,
		PrivatePort:  int32(instancePort),
		InnerIp:      instanceIP,
		InnerPort:    int32(instancePort),
		PublicIp:     instanceIP,
		PublicPort:   int32(instancePort),
		Node:         int32(cfg.Node),
	}
	if err := runtime.Register(instance); err != nil {
		t.Fatalf("register retained Nacos instance: %v", err)
	}

	t.Logf(
		"registered retained Nacos instance namespace=%s group=%s service=%s endpoint=%s:%d",
		cfg.Namespace,
		nodeGroup(instance.Node),
		serviceName,
		instanceIP,
		instancePort,
	)
}
