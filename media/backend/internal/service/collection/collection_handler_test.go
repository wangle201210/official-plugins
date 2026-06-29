// This file verifies media collection TCP protocol handlers.

package collection

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/dellinger2023/net-flux/pkg/network"
	"google.golang.org/protobuf/proto"
)

// TestEventHandlerRespondsToPing verifies system ping packets receive pong responses.
func TestEventHandlerRespondsToPing(t *testing.T) {
	conn := &recordingConn{}
	handler := newEventHandler(context.Background(), newDiscoveryRuntime(defaultDiscoveryConfig()), nil)

	err := handler.OnCmdSystem(conn, &gen.Ping{Timestamp: 12345})
	if err != nil {
		t.Fatalf("handle ping: %v", err)
	}
	if conn.cmd != uint8(gen.CMD_SYSTEM) {
		t.Fatalf("expected system cmd, got %d", conn.cmd)
	}
	if conn.subcmd != uint8(gen.SCMDSystem_PONG) {
		t.Fatalf("expected pong subcmd, got %d", conn.subcmd)
	}
	pong, ok := conn.pkt.(*gen.Pong)
	if !ok {
		t.Fatalf("expected pong packet, got %T", conn.pkt)
	}
	if pong.Timestamp != 12345 {
		t.Fatalf("expected pong timestamp 12345, got %d", pong.Timestamp)
	}
}

// TestEventHandlerAcceptsMetrics verifies known metric packets do not fail.
func TestEventHandlerAcceptsMetrics(t *testing.T) {
	reports := &fakeReportWriter{}
	handler := newEventHandler(context.Background(), newDiscoveryRuntime(defaultDiscoveryConfig()), reports)
	packets := []proto.Message{
		&gen.MachineMetric{InstanceId: "instance-a"},
		&gen.NetworkMetric{MachineId: "machine-a", SourceIp: "127.0.0.1"},
		&gen.StreamMetric{MachineId: "machine-a", StreamId: "stream-a"},
		&gen.SessionMetric{SessionId: "session-a"},
	}
	for _, packet := range packets {
		if err := handler.OnCmdDataReport(&recordingConn{}, uint8(gen.SCMDDataReport_STREAM_STATUS), packet); err != nil {
			t.Fatalf("handle metric %T: %v", packet, err)
		}
	}
	if reports.machineCount != 1 {
		t.Fatalf("expected one machine metric write, got %d", reports.machineCount)
	}
	if reports.networkCount != 1 {
		t.Fatalf("expected one network metric write, got %d", reports.networkCount)
	}
	if reports.streamCount != 1 {
		t.Fatalf("expected one stream metric write, got %d", reports.streamCount)
	}
	if reports.sessionCount != 1 {
		t.Fatalf("expected one session metric write, got %d", reports.sessionCount)
	}
}

// TestEventHandlerRejectsDiscoveryWhenDisabled verifies discovery commands need explicit config.
func TestEventHandlerRejectsDiscoveryWhenDisabled(t *testing.T) {
	handler := newEventHandler(context.Background(), newDiscoveryRuntime(defaultDiscoveryConfig()), nil)

	err := handler.OnCmdDiscovery(&recordingConn{}, &gen.Lookup{ServiceName: "media-node", Node: 1})
	if err == nil {
		t.Fatal("expected disabled discovery error")
	}
}

// TestEventHandlerRegistersDiscoveryInstance verifies Instance packets are registered.
func TestEventHandlerRegistersDiscoveryInstance(t *testing.T) {
	client := &fakeDiscoveryClient{}
	handler := newTestDiscoveryHandler(t, client)

	err := handler.OnCmdDiscovery(&recordingConn{}, &gen.Instance{
		InstanceName: "media-node",
		PrivateIp:    "10.0.0.12",
		PrivatePort:  8080,
		Node:         2,
	})
	if err != nil {
		t.Fatalf("register discovery instance: %v", err)
	}
	if client.registered == nil {
		t.Fatal("expected registered instance")
	}
	if client.registered.GetInstanceName() != "media-node" {
		t.Fatalf("expected registered service name, got %s", client.registered.GetInstanceName())
	}
	if client.registered.GetNode() != 2 {
		t.Fatalf("expected registered node 2, got %d", client.registered.GetNode())
	}
	if !client.closed {
		t.Fatal("expected discovery client closed after register")
	}
}

// TestEventHandlerDeregistersDiscoveryInstance verifies Deregister packets remove instances.
func TestEventHandlerDeregistersDiscoveryInstance(t *testing.T) {
	client := &fakeDiscoveryClient{}
	handler := newTestDiscoveryHandler(t, client)

	err := handler.OnCmdDiscovery(&recordingConn{}, &gen.Deregister{
		InstanceName: "media-node",
		Ip:           "10.0.0.12",
		Port:         8080,
		Node:         3,
	})
	if err != nil {
		t.Fatalf("deregister discovery instance: %v", err)
	}
	expected := fakeDeregisterCall{
		serviceName: "media-node",
		groupName:   "3",
		ip:          "10.0.0.12",
		port:        8080,
	}
	if client.deregistered != expected {
		t.Fatalf("unexpected deregister call: %#v", client.deregistered)
	}
	if !client.closed {
		t.Fatal("expected discovery client closed after deregister")
	}
}

// TestEventHandlerLooksUpDiscoveryInstance verifies Lookup packets write LookupAck responses.
func TestEventHandlerLooksUpDiscoveryInstance(t *testing.T) {
	client := &fakeDiscoveryClient{
		lookupResult: &gen.Instance{
			InstanceName: "4@@media-node",
			PrivateIp:    "10.0.0.12",
			PrivatePort:  8080,
			Node:         4,
		},
	}
	handler := newTestDiscoveryHandler(t, client)
	conn := &recordingConn{}

	err := handler.OnCmdDiscovery(conn, &gen.Lookup{
		ServiceName: "media-node",
		Node:        4,
		Healthy:     true,
	})
	if err != nil {
		t.Fatalf("lookup discovery instance: %v", err)
	}
	if client.lookupServiceName != "media-node" || client.lookupGroupName != "4" {
		t.Fatalf("unexpected lookup service=%s group=%s", client.lookupServiceName, client.lookupGroupName)
	}
	if conn.cmd != uint8(gen.CMD_DISCOVERY) {
		t.Fatalf("expected discovery cmd, got %d", conn.cmd)
	}
	if conn.subcmd != uint8(gen.SCMDDisco_LOOKUP_ACK) {
		t.Fatalf("expected lookup ack subcmd, got %d", conn.subcmd)
	}
	ack, ok := conn.pkt.(*gen.LookupAck)
	if !ok {
		t.Fatalf("expected lookup ack packet, got %T", conn.pkt)
	}
	if len(ack.GetServices()) != 1 {
		t.Fatalf("expected one service, got %d", len(ack.GetServices()))
	}
	service := ack.GetServices()[0]
	if service.GetName() != "media-node" {
		t.Fatalf("expected service name media-node, got %s", service.GetName())
	}
	if service.GetGroupName() != "4" {
		t.Fatalf("expected service group 4, got %s", service.GetGroupName())
	}
	if !service.GetValid() {
		t.Fatal("expected service valid flag true")
	}
	instance := service.GetInstances()[0]
	if instance.GetInstanceName() != "media-node" {
		t.Fatalf("expected clean lookup instance name media-node, got %s", instance.GetInstanceName())
	}
	if instance.GetPrivateIp() != client.lookupResult.GetPrivateIp() ||
		instance.GetPrivatePort() != client.lookupResult.GetPrivatePort() {
		t.Fatalf("unexpected lookup instance endpoint: %#v", service.GetInstances())
	}
	if !client.closed {
		t.Fatal("expected discovery client closed after lookup")
	}
}

// TestEventHandlerWritesEmptyLookupAckWhenDiscoveryHasNoInstance verifies empty Nacos lookups still reply.
func TestEventHandlerWritesEmptyLookupAckWhenDiscoveryHasNoInstance(t *testing.T) {
	client := &fakeDiscoveryClient{
		lookupErr: errors.New("instance list is empty!"),
	}
	handler := newTestDiscoveryHandler(t, client)
	conn := &recordingConn{}

	err := handler.OnCmdDiscovery(conn, &gen.Lookup{
		ServiceName: "media-node",
		Node:        5,
		Healthy:     true,
	})
	if err != nil {
		t.Fatalf("lookup empty discovery instance: %v", err)
	}
	if conn.cmd != uint8(gen.CMD_DISCOVERY) || conn.subcmd != uint8(gen.SCMDDisco_LOOKUP_ACK) {
		t.Fatalf("expected lookup ack response, got cmd=%d subcmd=%d", conn.cmd, conn.subcmd)
	}
	ack, ok := conn.pkt.(*gen.LookupAck)
	if !ok {
		t.Fatalf("expected lookup ack packet, got %T", conn.pkt)
	}
	if len(ack.GetServices()) != 0 {
		t.Fatalf("expected empty services after no instance, got %#v", ack.GetServices())
	}
	if !client.closed {
		t.Fatal("expected discovery client closed after empty lookup")
	}
}

// TestEventHandlerCreatesFreshDiscoveryClientPerLookup verifies lookups do not reuse stale Nacos caches.
func TestEventHandlerCreatesFreshDiscoveryClientPerLookup(t *testing.T) {
	firstClient := &fakeDiscoveryClient{
		lookupResult: &gen.Instance{
			InstanceName: "6@@media-node",
			PrivateIp:    "10.0.0.12",
			PrivatePort:  8080,
			Node:         6,
		},
	}
	secondClient := &fakeDiscoveryClient{
		lookupErr: errors.New("instance list is empty!"),
	}
	handler := newTestDiscoveryHandlerWithFactory(t, func(DiscoveryConfig) (discoveryClient, error) {
		if !firstClient.closed {
			return firstClient, nil
		}
		return secondClient, nil
	})

	firstConn := &recordingConn{}
	if err := handler.OnCmdDiscovery(firstConn, &gen.Lookup{ServiceName: "media-node", Node: 6, Healthy: true}); err != nil {
		t.Fatalf("first lookup discovery instance: %v", err)
	}
	firstAck, ok := firstConn.pkt.(*gen.LookupAck)
	if !ok || len(firstAck.GetServices()) != 1 {
		t.Fatalf("expected first lookup to return one service, got %#v", firstConn.pkt)
	}
	if !firstClient.closed {
		t.Fatal("expected first discovery client closed after lookup")
	}

	secondConn := &recordingConn{}
	if err := handler.OnCmdDiscovery(secondConn, &gen.Lookup{ServiceName: "media-node", Node: 6, Healthy: true}); err != nil {
		t.Fatalf("second lookup discovery instance: %v", err)
	}
	secondAck, ok := secondConn.pkt.(*gen.LookupAck)
	if !ok {
		t.Fatalf("expected second lookup ack packet, got %T", secondConn.pkt)
	}
	if len(secondAck.GetServices()) != 0 {
		t.Fatalf("expected second lookup to observe empty discovery state, got %#v", secondAck.GetServices())
	}
	if !secondClient.closed {
		t.Fatal("expected second discovery client closed after lookup")
	}
}

// newTestDiscoveryHandler creates a handler wired to a fake discovery client.
func newTestDiscoveryHandler(t *testing.T, client *fakeDiscoveryClient) network.EventHandler {
	t.Helper()

	return newTestDiscoveryHandlerWithFactory(t, func(DiscoveryConfig) (discoveryClient, error) {
		return client, nil
	})
}

// newTestDiscoveryHandlerWithFactory creates a handler wired to a fake discovery factory.
func newTestDiscoveryHandlerWithFactory(
	t *testing.T,
	factory discoveryClientFactory,
) network.EventHandler {
	t.Helper()

	cfg := defaultDiscoveryConfig()
	cfg.Enabled = true
	runtime := newDiscoveryRuntime(cfg)
	runtime.factory = factory
	return newEventHandler(context.Background(), runtime, nil)
}

// fakeReportWriter records data report business handling.
type fakeReportWriter struct {
	machineCount int
	networkCount int
	streamCount  int
	sessionCount int
}

// HandleMachineMetric records one instance metric carried by MachineMetric.
func (w *fakeReportWriter) HandleMachineMetric(context.Context, *gen.MachineMetric) error {
	w.machineCount++
	return nil
}

// HandleNetworkMetric records one network metric.
func (w *fakeReportWriter) HandleNetworkMetric(context.Context, *gen.NetworkMetric) error {
	w.networkCount++
	return nil
}

// HandleStreamMetric records one stream metric.
func (w *fakeReportWriter) HandleStreamMetric(context.Context, uint8, *gen.StreamMetric) error {
	w.streamCount++
	return nil
}

// HandleSessionMetric records one session metric.
func (w *fakeReportWriter) HandleSessionMetric(context.Context, uint8, *gen.SessionMetric) error {
	w.sessionCount++
	return nil
}

// fakeDiscoveryClient records discovery operations.
type fakeDiscoveryClient struct {
	registered        *gen.Instance
	deregistered      fakeDeregisterCall
	lookupServiceName string
	lookupGroupName   string
	lookupResult      *gen.Instance
	lookupErr         error
	closed            bool
}

// RegisterInstance records one registered instance.
func (c *fakeDiscoveryClient) RegisterInstance(instance *gen.Instance) error {
	c.registered = instance
	return nil
}

// DeregisterInstance records one deregister call.
func (c *fakeDiscoveryClient) DeregisterInstance(serviceName, groupName, ip string, port uint64) error {
	c.deregistered = fakeDeregisterCall{
		serviceName: serviceName,
		groupName:   groupName,
		ip:          ip,
		port:        port,
	}
	return nil
}

// GetServiceInstanceByGroup records one lookup call.
func (c *fakeDiscoveryClient) GetServiceInstanceByGroup(serviceName, groupName string) (*gen.Instance, error) {
	c.lookupServiceName = serviceName
	c.lookupGroupName = groupName
	return c.lookupResult, c.lookupErr
}

// Close records that the fake client was closed.
func (c *fakeDiscoveryClient) Close() {
	c.closed = true
}

// fakeDeregisterCall records one deregister call.
type fakeDeregisterCall struct {
	serviceName string
	groupName   string
	ip          string
	port        uint64
}

// recordingConn records packets written by the collection event handler.
type recordingConn struct {
	cmd    uint8
	subcmd uint8
	pkt    proto.Message
}

// ID returns a stable test connection ID.
func (c *recordingConn) ID() uint32 { return 1 }

// RemoteAddr returns a stable test remote address.
func (c *recordingConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1911}
}

// LocalAddr returns a stable test local address.
func (c *recordingConn) LocalAddr() net.Addr { return &net.TCPAddr{IP: net.IPv4zero, Port: 1911} }

// SetReadDeadline records no state for tests.
func (c *recordingConn) SetReadDeadline(time.Time) error { return nil }

// SetWriteDeadline records no state for tests.
func (c *recordingConn) SetWriteDeadline(time.Time) error { return nil }

// SetHeartbeatInterval records no state for tests.
func (c *recordingConn) SetHeartbeatInterval(time.Duration) error { return nil }

// SetNoNagle records no state for tests.
func (c *recordingConn) SetNoNagle(bool) error { return nil }

// SetKeepAlive records no state for tests.
func (c *recordingConn) SetKeepAlive(bool) error { return nil }

// SetKeepAlivePeriod records no state for tests.
func (c *recordingConn) SetKeepAlivePeriod(time.Duration) error { return nil }

// SetReuseAddr records no state for tests.
func (c *recordingConn) SetReuseAddr(bool) error { return nil }

// SetReusePort records no state for tests.
func (c *recordingConn) SetReusePort(bool) error { return nil }

// RawConn returns no raw connection for tests.
func (c *recordingConn) RawConn() net.Conn { return nil }

// Write records no raw bytes for tests.
func (c *recordingConn) Write(p []byte) (int, error) { return len(p), nil }

// Read returns EOF-like empty data for tests.
func (c *recordingConn) Read(_ []byte) (int, error) { return 0, nil }

// Connect records no state for tests.
func (c *recordingConn) Connect(string) error { return nil }

// ConnectWithTLS records no state for tests.
func (c *recordingConn) ConnectWithTLS(string, *tls.Config) error { return nil }

// Close records no state for tests.
func (c *recordingConn) Close() error { return nil }

// IsClosed reports the test connection as open.
func (c *recordingConn) IsClosed() bool { return false }

// WritePacket records one protobuf packet for assertions.
func (c *recordingConn) WritePacket(cmd, subcmd uint8, pkt proto.Message) error {
	c.cmd = cmd
	c.subcmd = subcmd
	c.pkt = pkt
	return nil
}
