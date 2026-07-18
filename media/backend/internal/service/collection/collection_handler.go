// This file handles net-flux TCP collection protocol events.

package collection

import (
	"context"
	"fmt"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/dellinger2023/net-flux/pkg/network"
	"google.golang.org/protobuf/proto"

	"lina-core/pkg/logger"
)

// eventHandler accepts media collection packets from net-flux clients.
type eventHandler struct {
	contextProvider func() context.Context // contextProvider returns the host startup context for logs and writes.
	discovery       *discoveryRuntime      // discovery handles optional Nacos-backed service discovery.
	reports         reportWriter           // reports handles dashboard read-model writes for metric packets.
}

// newEventHandler creates one collection server event handler.
func newEventHandler(ctx context.Context, discovery *discoveryRuntime, reports reportWriter) network.EventHandler {
	if ctx == nil {
		ctx = context.Background()
	}
	return &eventHandler{contextProvider: func() context.Context { return ctx }, discovery: discovery, reports: reports}
}

// context returns the current callback context used for logs and report writes.
func (h *eventHandler) context() context.Context {
	if h == nil || h.contextProvider == nil {
		return context.Background()
	}
	return h.contextProvider()
}

// OnConnect records accepted TCP clients.
func (h *eventHandler) OnConnect(conn network.TCPConn) error {
	logger.Infof(h.context(), "media collection client connected remote=%s", conn.RemoteAddr().String())
	return nil
}

// OnClose records closed TCP clients.
func (h *eventHandler) OnClose(conn network.TCPConn) {
	logger.Infof(h.context(), "media collection client closed remote=%s", conn.RemoteAddr().String())
}

// OnCmdSystem handles basic system commands such as Ping.
func (h *eventHandler) OnCmdSystem(conn network.TCPConn, pkt proto.Message) error {
	switch pkt := pkt.(type) {
	case *gen.Ping:
		return conn.WritePacket(
			uint8(gen.CMD_SYSTEM),
			uint8(gen.SCMDSystem_PONG),
			&gen.Pong{Timestamp: pkt.Timestamp},
		)
	default:
		return fmt.Errorf("media collection unknown system command: %T", pkt)
	}
}

// OnCmdDiscovery handles optional Nacos-backed discovery commands.
func (h *eventHandler) OnCmdDiscovery(conn network.TCPConn, pkt proto.Message) error {
	switch pkt := pkt.(type) {
	case *gen.Instance:
		logger.Infof(
			h.context(),
			"media collection discovery register instanceName=%s node=%d privateIp=%s privatePort=%d",
			pkt.GetInstanceName(),
			pkt.GetNode(),
			pkt.GetPrivateIp(),
			pkt.GetPrivatePort(),
		)
		return h.discovery.Register(pkt)
	case *gen.Deregister:
		logger.Infof(
			h.context(),
			"media collection discovery deregister instanceName=%s node=%d ip=%s port=%d",
			pkt.GetInstanceName(),
			pkt.GetNode(),
			pkt.GetIp(),
			pkt.GetPort(),
		)
		return h.discovery.Deregister(pkt)
	case *gen.Lookup:
		logger.Infof(
			h.context(),
			"media collection discovery lookup serviceName=%s node=%d healthy=%v",
			pkt.GetServiceName(),
			pkt.GetNode(),
			pkt.GetHealthy(),
		)
		ack, err := h.discovery.Lookup(pkt)
		if err != nil {
			return err
		}
		return conn.WritePacket(
			uint8(gen.CMD_DISCOVERY),
			uint8(gen.SCMDDisco_LOOKUP_ACK),
			ack,
		)
	default:
		return fmt.Errorf("media collection unknown discovery command: %T", pkt)
	}
}

// OnCmdDataReport accepts metric reports from net-flux clients.
func (h *eventHandler) OnCmdDataReport(_ network.TCPConn, subcmd uint8, pkt proto.Message) error {
	ctx := h.context()
	switch pkt := pkt.(type) {
	case *gen.MachineMetric:
		logger.Infof(
			ctx,
			"media collection instance metric instanceId=%s machineId=%s status=%s cpuUsage=%f memUsed=%d memTotal=%d timestamp=%d",
			pkt.GetInstanceId(),
			pkt.GetMachineId(),
			pkt.GetStatus(),
			pkt.GetCpuUsage(),
			pkt.GetMemUsed(),
			pkt.GetMemTotal(),
			pkt.GetTimestamp(),
		)
		if h.reports != nil {
			return h.reports.HandleMachineMetric(ctx, pkt)
		}
	case *gen.NetworkMetric:
		logger.Infof(
			ctx,
			"media collection network metric nodeId=%s machineId=%s sourceIp=%s destinationIp=%s rtt=%d jitter=%d packetLoss=%.4f statusCode=%d timestamp=%d",
			pkt.GetNodeId(),
			pkt.GetMachineId(),
			pkt.GetSourceIp(),
			pkt.GetDestinationIp(),
			pkt.GetRtt(),
			pkt.GetJitter(),
			pkt.GetPacketLoss(),
			pkt.GetStatusCode(),
			pkt.GetTimestamp(),
		)
		if h.reports != nil {
			return h.reports.HandleNetworkMetric(ctx, pkt)
		}
	case *gen.StreamMetric:
		logger.Infof(
			ctx,
			"media collection stream metric subcmd=%d machineId=%s streamId=%s status=%s protocol=%s bitrate=%d width=%d height=%d timestamp=%d",
			subcmd,
			pkt.GetMachineId(),
			pkt.GetStreamId(),
			pkt.GetStatus().String(),
			pkt.GetProtocol().String(),
			pkt.GetBitrate(),
			pkt.GetWidth(),
			pkt.GetHeight(),
			pkt.GetTimestamp(),
		)
		if h.reports != nil {
			return h.reports.HandleStreamMetric(ctx, subcmd, pkt)
		}
	case *gen.SessionMetric:
		logger.Infof(
			ctx,
			"media collection session metric subcmd=%d sessionId=%s streamId=%s tenantId=%s protocol=%s timestamp=%d",
			subcmd,
			pkt.GetSessionId(),
			pkt.GetStreamId(),
			pkt.GetTenantId(),
			pkt.GetProtocol().String(),
			pkt.GetTimestamp(),
		)
		if h.reports != nil {
			return h.reports.HandleSessionMetric(ctx, subcmd, pkt)
		}
	default:
		logger.Infof(ctx, "media collection data report ignored packet=%T", pkt)
	}
	return nil
}

// OnCmdConfig accepts config notifications without applying local state changes.
func (h *eventHandler) OnCmdConfig(_ network.TCPConn, pkt proto.Message) error {
	logger.Infof(h.context(), "media collection config command ignored packet=%T", pkt)
	return nil
}

// OnCmdEvent accepts event notifications without applying local state changes.
func (h *eventHandler) OnCmdEvent(_ network.TCPConn, pkt proto.Message) error {
	logger.Infof(h.context(), "media collection event command ignored packet=%T", pkt)
	return nil
}

// OnCmdControl accepts control commands without applying local state changes.
func (h *eventHandler) OnCmdControl(_ network.TCPConn, pkt proto.Message) error {
	logger.Infof(h.context(), "media collection control command ignored packet=%T", pkt)
	return nil
}
