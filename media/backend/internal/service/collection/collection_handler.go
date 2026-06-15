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
	ctx       context.Context   // ctx carries the host startup context for project logging.
	discovery *discoveryRuntime // discovery handles optional Nacos-backed service discovery.
	reports   reportWriter      // reports handles dashboard read-model writes for metric packets.
}

// newEventHandler creates one collection server event handler.
func newEventHandler(ctx context.Context, discovery *discoveryRuntime, reports reportWriter) network.EventHandler {
	return &eventHandler{ctx: ctx, discovery: discovery, reports: reports}
}

// OnConnect records accepted TCP clients.
func (h *eventHandler) OnConnect(conn network.TCPConn) error {
	logger.Infof(h.ctx, "media collection client connected remote=%s", conn.RemoteAddr().String())
	return nil
}

// OnClose records closed TCP clients.
func (h *eventHandler) OnClose(conn network.TCPConn) {
	logger.Infof(h.ctx, "media collection client closed remote=%s", conn.RemoteAddr().String())
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
			h.ctx,
			"media collection discovery register instanceName=%s node=%d privateIp=%s privatePort=%d",
			pkt.GetInstanceName(),
			pkt.GetNode(),
			pkt.GetPrivateIp(),
			pkt.GetPrivatePort(),
		)
		return h.discovery.Register(pkt)
	case *gen.Deregister:
		logger.Infof(
			h.ctx,
			"media collection discovery deregister instanceName=%s node=%d ip=%s port=%d",
			pkt.GetInstanceName(),
			pkt.GetNode(),
			pkt.GetIp(),
			pkt.GetPort(),
		)
		return h.discovery.Deregister(pkt)
	case *gen.Lookup:
		logger.Infof(
			h.ctx,
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
	switch pkt := pkt.(type) {
	case *gen.MachineMetric:
		logger.Infof(
			h.ctx,
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
			return h.reports.HandleMachineMetric(h.ctx, pkt)
		}
	case *gen.NetworkMetric:
		logger.Infof(
			h.ctx,
			"media collection network metric machineId=%s sourceIp=%s destinationIp=%s rtt=%d throughput=%d timestamp=%d",
			pkt.GetMachineId(),
			pkt.GetSourceIp(),
			pkt.GetDestinationIp(),
			pkt.GetRtt(),
			pkt.GetThroughput(),
			pkt.GetTimestamp(),
		)
		if h.reports != nil {
			return h.reports.HandleNetworkMetric(h.ctx, pkt)
		}
	case *gen.StreamMetric:
		logger.Infof(
			h.ctx,
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
			return h.reports.HandleStreamMetric(h.ctx, subcmd, pkt)
		}
	case *gen.SessionMetric:
		logger.Infof(
			h.ctx,
			"media collection session metric subcmd=%d sessionId=%s streamId=%s tenantId=%s protocol=%s timestamp=%d",
			subcmd,
			pkt.GetSessionId(),
			pkt.GetStreamId(),
			pkt.GetTenantId(),
			pkt.GetProtocol().String(),
			pkt.GetTimestamp(),
		)
		if h.reports != nil {
			return h.reports.HandleSessionMetric(h.ctx, subcmd, pkt)
		}
	default:
		logger.Infof(h.ctx, "media collection data report ignored packet=%T", pkt)
	}
	return nil
}

// OnCmdConfig accepts config notifications without applying local state changes.
func (h *eventHandler) OnCmdConfig(_ network.TCPConn, pkt proto.Message) error {
	logger.Infof(h.ctx, "media collection config command ignored packet=%T", pkt)
	return nil
}

// OnCmdEvent accepts event notifications without applying local state changes.
func (h *eventHandler) OnCmdEvent(_ network.TCPConn, pkt proto.Message) error {
	logger.Infof(h.ctx, "media collection event command ignored packet=%T", pkt)
	return nil
}

// OnCmdControl accepts control commands without applying local state changes.
func (h *eventHandler) OnCmdControl(_ network.TCPConn, pkt proto.Message) error {
	logger.Infof(h.ctx, "media collection control command ignored packet=%T", pkt)
	return nil
}
