// Command collection-client calls the media plugin net-flux TCP collection server.
//
// It is a small cross-platform development tool for verifying the LinaPro media
// TCP discovery and report paths without calling Nacos or writing dashboard rows
// directly. The command sends net-flux packets to collectionServer.addr and
// prints lookup acknowledgements.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/dellinger2023/net-flux/pkg/network"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Supported client actions.
const (
	actionPing           = "ping"
	actionRegister       = "register"
	actionLookup         = "lookup"
	actionDeregister     = "deregister"
	actionRegisterLookup = "register-lookup"
	actionReport         = "report"
	actionReportClose    = "report-close"
	actionReportCycle    = "report-cycle"
	actionSmoke          = "smoke"
)

// clientConfig holds command-line options for one TCP collection call.
type clientConfig struct {
	addr        string        // addr is the LinaPro media collection TCP address.
	action      string        // action selects the discovery packet flow.
	serviceName string        // serviceName is the Nacos service name sent through TCP.
	instanceID  string        // instanceID is optional and defaults to serviceName.
	node        int           // node becomes the net-flux node group.
	privateIP   string        // privateIP is the registered service endpoint IP.
	privatePort int           // privatePort is the registered service endpoint port.
	publicIP    string        // publicIP is stored as Nacos metadata.
	publicPort  int           // publicPort is stored as Nacos metadata.
	innerIP     string        // innerIP is stored as Nacos metadata.
	innerPort   int           // innerPort is stored as Nacos metadata.
	weight      float64       // weight is sent in the net-flux instance packet.
	healthy     bool          // healthy is sent in the net-flux instance packet.
	enable      bool          // enable is sent in the net-flux instance packet.
	ephemeral   bool          // ephemeral is sent in the net-flux instance packet.
	extra       string        // extra is comma-separated key=value metadata.
	tenantID    string        // tenantID is written into stream and session report packets.
	deviceID    string        // deviceID is written into stream and session report metadata.
	nodeID      string        // nodeID is written into report packets.
	nodeName    string        // nodeName is written into report packets.
	region      string        // region is written into machine report packets.
	streamID    string        // streamID is the reported media stream business key.
	streamName  string        // streamName is the reported display name.
	streamPath  string        // streamPath is the reported source URL.
	sessionID   string        // sessionID is the reported playback session business key.
	clientID    string        // clientID identifies the playback client.
	clientIP    string        // clientIP is written into session report packets.
	clientType  string        // clientType is written into session report packets.
	userName    string        // userName is written into session report packets.
	protocol    string        // protocol is mapped to net-flux StreamProtocol.
	status      string        // status is mapped to net-flux StreamStatus.
	timeout     time.Duration // timeout bounds connect and lookup waiting time.
	settle      time.Duration // settle waits after fire-and-forget writes.
	verbose     bool          // verbose prints connection callbacks.
}

// clientHandler receives asynchronous net-flux packets from the TCP client.
type clientHandler struct {
	lookupAckCh chan *gen.LookupAck
	pongCh      chan *gen.Pong
	verbose     bool
}

// main parses flags, runs the requested action, and exits non-zero on failure.
func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "collection client failed: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags builds a client configuration from command-line flags.
func parseFlags() clientConfig {
	cfg := clientConfig{}
	flag.StringVar(&cfg.addr, "addr", "127.0.0.1:1911", "LinaPro media collection TCP address")
	flag.StringVar(
		&cfg.action,
		"action",
		actionRegisterLookup,
		"action: ping, register, lookup, deregister, register-lookup, report, report-close, report-cycle, smoke",
	)
	flag.StringVar(&cfg.serviceName, "service", "linapro-media-client-test", "service name for discovery")
	flag.StringVar(&cfg.instanceID, "instance-id", "", "instance id; defaults to -service")
	flag.IntVar(&cfg.node, "node", 1, "net-flux node id; LinaPro maps it to the Nacos group")
	flag.StringVar(&cfg.privateIP, "private-ip", "127.0.0.1", "registered private endpoint IP")
	flag.IntVar(&cfg.privatePort, "private-port", 19191, "registered private endpoint port")
	flag.StringVar(&cfg.publicIP, "public-ip", "127.0.0.1", "public endpoint IP stored as metadata")
	flag.IntVar(&cfg.publicPort, "public-port", 19193, "public endpoint port stored as metadata")
	flag.StringVar(&cfg.innerIP, "inner-ip", "127.0.0.1", "container endpoint IP stored as metadata")
	flag.IntVar(&cfg.innerPort, "inner-port", 19192, "container endpoint port stored as metadata")
	flag.Float64Var(&cfg.weight, "weight", 1.0, "instance weight")
	flag.BoolVar(&cfg.healthy, "healthy", true, "instance health flag")
	flag.BoolVar(&cfg.enable, "enable", true, "instance enabled flag")
	flag.BoolVar(&cfg.ephemeral, "ephemeral", false, "instance ephemeral flag; server-side Nacos registration stays persistent")
	flag.StringVar(&cfg.extra, "extra", "", "extra metadata, comma-separated key=value pairs")
	flag.StringVar(&cfg.tenantID, "tenant-id", "tenant-demo", "tenant id used by report packets")
	flag.StringVar(&cfg.deviceID, "device-id", "", "device id used by stream and session reports; defaults to device-<instance-id>")
	flag.StringVar(&cfg.nodeID, "node-id", "", "media node id used by report packets; defaults to node-<node>")
	flag.StringVar(&cfg.nodeName, "node-name", "", "media node name used by report packets; defaults to -node-id")
	flag.StringVar(&cfg.region, "region", "local", "media node region used by machine reports")
	flag.StringVar(&cfg.streamID, "stream-id", "", "stream id used by report packets; defaults to stream-<instance-id>")
	flag.StringVar(&cfg.streamName, "stream-name", "", "stream display name used by report packets; defaults to -stream-id")
	flag.StringVar(&cfg.streamPath, "stream-path", "", "stream URL used by report packets; defaults to rtmp://<private-ip>/live/<stream-id>")
	flag.StringVar(&cfg.sessionID, "session-id", "", "session id used by report packets; defaults to session-<instance-id>")
	flag.StringVar(&cfg.clientID, "client-id", "client-demo", "client id used by session reports")
	flag.StringVar(&cfg.clientIP, "client-ip", "192.0.2.10", "client IP used by session reports")
	flag.StringVar(&cfg.clientType, "client-type", "pc", "client type used by session reports: mobile, pc, web")
	flag.StringVar(&cfg.userName, "user-name", "demo-user", "user name used by session reports")
	flag.StringVar(&cfg.protocol, "protocol", "hls", "stream protocol: rtmp, rtsp, hls, http-flv, ws-flv, https-flv, wss-flv, gb28181")
	flag.StringVar(&cfg.status, "status", "running", "stream status: running, inactive, error, closed, timeout, cancelled, failed")
	flag.DurationVar(&cfg.timeout, "timeout", 5*time.Second, "connection and lookup timeout")
	flag.DurationVar(&cfg.settle, "settle", 500*time.Millisecond, "wait after register or deregister before lookup or exit")
	flag.BoolVar(&cfg.verbose, "v", false, "print connection callbacks")
	flag.Parse()
	return cfg
}

// run opens one net-flux TCP client and executes the selected action.
func run(cfg clientConfig) error {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}
	cfg = normalized

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	handler := newClientHandler(cfg.verbose)
	client, err := network.NewTcpClient(cfg.addr, handler, &network.TCPConnOptions{ParentCtx: ctx})
	if err != nil {
		return fmt.Errorf("create tcp client: %w", err)
	}
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect %s: %w", cfg.addr, err)
	}
	defer closeClient(client)

	switch cfg.action {
	case actionPing:
		return sendPingAndPrint(ctx, client, handler)
	case actionRegister:
		if err := sendRegister(client, cfg); err != nil {
			return err
		}
		return waitSettle(ctx, cfg.settle)
	case actionLookup:
		return sendLookupAndPrint(ctx, client, handler, cfg)
	case actionDeregister:
		if err := sendDeregister(client, cfg); err != nil {
			return err
		}
		return waitSettle(ctx, cfg.settle)
	case actionRegisterLookup:
		if err := sendRegister(client, cfg); err != nil {
			return err
		}
		if err := waitSettle(ctx, cfg.settle); err != nil {
			return err
		}
		return sendLookupAndPrint(ctx, client, handler, cfg)
	case actionReport:
		if err := sendActiveReports(client, cfg); err != nil {
			return err
		}
		return waitSettle(ctx, cfg.settle)
	case actionReportClose:
		if err := sendCloseReports(client, cfg); err != nil {
			return err
		}
		return waitSettle(ctx, cfg.settle)
	case actionReportCycle:
		if err := sendActiveReports(client, cfg); err != nil {
			return err
		}
		if err := waitSettle(ctx, cfg.settle); err != nil {
			return err
		}
		if err := sendCloseReports(client, cfg); err != nil {
			return err
		}
		return waitSettle(ctx, cfg.settle)
	case actionSmoke:
		if err := sendRegister(client, cfg); err != nil {
			return err
		}
		if err := waitSettle(ctx, cfg.settle); err != nil {
			return err
		}
		if err := sendLookupAndPrint(ctx, client, handler, cfg); err != nil {
			return err
		}
		if err := sendActiveReports(client, cfg); err != nil {
			return err
		}
		if err := waitSettle(ctx, cfg.settle); err != nil {
			return err
		}
		if err := sendDeregister(client, cfg); err != nil {
			return err
		}
		return waitSettle(ctx, cfg.settle)
	default:
		return fmt.Errorf("unsupported action %q", cfg.action)
	}
}

// normalizeConfig validates flags and fills derived defaults.
func normalizeConfig(cfg clientConfig) (clientConfig, error) {
	cfg.addr = strings.TrimSpace(cfg.addr)
	cfg.action = strings.TrimSpace(cfg.action)
	cfg.serviceName = strings.TrimSpace(cfg.serviceName)
	cfg.instanceID = strings.TrimSpace(cfg.instanceID)
	cfg.privateIP = strings.TrimSpace(cfg.privateIP)
	cfg.publicIP = strings.TrimSpace(cfg.publicIP)
	cfg.innerIP = strings.TrimSpace(cfg.innerIP)
	cfg.tenantID = strings.TrimSpace(cfg.tenantID)
	cfg.deviceID = strings.TrimSpace(cfg.deviceID)
	cfg.nodeID = strings.TrimSpace(cfg.nodeID)
	cfg.nodeName = strings.TrimSpace(cfg.nodeName)
	cfg.region = strings.TrimSpace(cfg.region)
	cfg.streamID = strings.TrimSpace(cfg.streamID)
	cfg.streamName = strings.TrimSpace(cfg.streamName)
	cfg.streamPath = strings.TrimSpace(cfg.streamPath)
	cfg.sessionID = strings.TrimSpace(cfg.sessionID)
	cfg.clientID = strings.TrimSpace(cfg.clientID)
	cfg.clientIP = strings.TrimSpace(cfg.clientIP)
	cfg.clientType = strings.TrimSpace(cfg.clientType)
	cfg.userName = strings.TrimSpace(cfg.userName)
	cfg.protocol = strings.TrimSpace(cfg.protocol)
	cfg.status = strings.TrimSpace(cfg.status)
	if cfg.addr == "" {
		return cfg, errors.New("addr cannot be empty")
	}
	if cfg.serviceName == "" {
		return cfg, errors.New("service cannot be empty")
	}
	if cfg.instanceID == "" {
		cfg.instanceID = cfg.serviceName
	}
	if cfg.nodeID == "" {
		cfg.nodeID = fmt.Sprintf("node-%d", cfg.node)
	}
	if cfg.nodeName == "" {
		cfg.nodeName = cfg.nodeID
	}
	if cfg.deviceID == "" {
		cfg.deviceID = "device-" + cfg.instanceID
	}
	if cfg.streamID == "" {
		cfg.streamID = "stream-" + cfg.instanceID
	}
	if cfg.streamName == "" {
		cfg.streamName = cfg.streamID
	}
	if cfg.streamPath == "" {
		cfg.streamPath = fmt.Sprintf("rtmp://%s/live/%s", cfg.privateIP, cfg.streamID)
	}
	if cfg.sessionID == "" {
		cfg.sessionID = "session-" + cfg.instanceID
	}
	if cfg.node <= 0 {
		return cfg, errors.New("node must be positive")
	}
	if cfg.timeout <= 0 {
		return cfg, errors.New("timeout must be positive")
	}
	if cfg.settle < 0 {
		return cfg, errors.New("settle cannot be negative")
	}
	switch cfg.action {
	case actionPing:
	case actionRegister, actionDeregister, actionRegisterLookup, actionSmoke:
		if cfg.privateIP == "" {
			return cfg, errors.New("private-ip cannot be empty")
		}
		if cfg.privatePort <= 0 || cfg.privatePort > 65535 {
			return cfg, errors.New("private-port must be between 1 and 65535")
		}
	case actionLookup, actionReport, actionReportClose, actionReportCycle:
	default:
		return cfg, fmt.Errorf("unsupported action %q", cfg.action)
	}
	switch cfg.action {
	case actionReport, actionReportClose, actionReportCycle, actionSmoke:
		if cfg.tenantID == "" {
			return cfg, errors.New("tenant-id cannot be empty for report actions")
		}
		if cfg.deviceID == "" {
			return cfg, errors.New("device-id cannot be empty for report actions")
		}
		if cfg.nodeID == "" {
			return cfg, errors.New("node-id cannot be empty for report actions")
		}
		if cfg.instanceID == "" {
			return cfg, errors.New("instance-id cannot be empty for report actions")
		}
		if cfg.streamID == "" {
			return cfg, errors.New("stream-id cannot be empty for report actions")
		}
		if cfg.sessionID == "" {
			return cfg, errors.New("session-id cannot be empty for report actions")
		}
		if _, err := parseStreamProtocol(cfg.protocol); err != nil {
			return cfg, err
		}
		if _, err := parseStreamStatus(cfg.status); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

// sendPingAndPrint sends one system ping packet and waits for the echoed pong.
func sendPingAndPrint(ctx context.Context, client *network.TcpClient, handler *clientHandler) error {
	timestamp := time.Now().Unix()
	if err := client.Write(uint8(gen.CMD_SYSTEM), uint8(gen.SCMDSystem_PING), &gen.Ping{Timestamp: timestamp}); err != nil {
		return fmt.Errorf("write ping packet: %w", err)
	}
	pong, err := handler.waitPong(ctx)
	if err != nil {
		return err
	}
	if pong.GetTimestamp() != timestamp {
		return fmt.Errorf("unexpected pong timestamp %d, want %d", pong.GetTimestamp(), timestamp)
	}
	fmt.Printf("pong received timestamp=%d\n", pong.GetTimestamp())
	return nil
}

// sendRegister sends one discovery register packet to LinaPro TCP.
func sendRegister(client *network.TcpClient, cfg clientConfig) error {
	extra, err := parseExtra(cfg.extra)
	if err != nil {
		return err
	}
	instance := &gen.Instance{
		InstanceId:   cfg.instanceID,
		InstanceName: cfg.serviceName,
		PublicIp:     cfg.publicIP,
		PublicPort:   int32(cfg.publicPort),
		PrivateIp:    cfg.privateIP,
		PrivatePort:  int32(cfg.privatePort),
		InnerIp:      cfg.innerIP,
		InnerPort:    int32(cfg.innerPort),
		Weight:       float32(cfg.weight),
		Healthy:      cfg.healthy,
		Enable:       cfg.enable,
		Ephemeral:    cfg.ephemeral,
		Node:         int32(cfg.node),
		Extra:        extra,
	}
	if err := client.Write(uint8(gen.CMD_DISCOVERY), uint8(gen.SCMDDisco_REGISTER), instance); err != nil {
		return fmt.Errorf("write register packet: %w", err)
	}
	fmt.Printf("register sent service=%s node=%d private=%s:%d\n", cfg.serviceName, cfg.node, cfg.privateIP, cfg.privatePort)
	return nil
}

// sendDeregister sends one discovery deregister packet to LinaPro TCP.
func sendDeregister(client *network.TcpClient, cfg clientConfig) error {
	deregister := &gen.Deregister{
		InstanceId:   cfg.instanceID,
		InstanceName: cfg.serviceName,
		Ip:           cfg.privateIP,
		Port:         int32(cfg.privatePort),
		Node:         int32(cfg.node),
	}
	if err := client.Write(uint8(gen.CMD_DISCOVERY), uint8(gen.SCMDDisco_DEREGISTER), deregister); err != nil {
		return fmt.Errorf("write deregister packet: %w", err)
	}
	fmt.Printf("deregister sent service=%s node=%d private=%s:%d\n", cfg.serviceName, cfg.node, cfg.privateIP, cfg.privatePort)
	return nil
}

// sendLookupAndPrint sends a lookup packet and prints the returned LookupAck.
func sendLookupAndPrint(ctx context.Context, client *network.TcpClient, handler *clientHandler, cfg clientConfig) error {
	lookup := &gen.Lookup{
		ServiceName: cfg.serviceName,
		Node:        int32(cfg.node),
		Healthy:     cfg.healthy,
	}
	if err := client.Write(uint8(gen.CMD_DISCOVERY), uint8(gen.SCMDDisco_LOOKUP), lookup); err != nil {
		return fmt.Errorf("write lookup packet: %w", err)
	}
	ack, err := handler.waitLookupAck(ctx)
	if err != nil {
		return err
	}
	return printLookupAck(ack)
}

// sendActiveReports sends one complete active dashboard sample through TCP data-report packets.
func sendActiveReports(client *network.TcpClient, cfg clientConfig) error {
	now := time.Now().UnixMilli()
	startedAt := now - int64(time.Minute/time.Millisecond)
	protocol, err := parseStreamProtocol(cfg.protocol)
	if err != nil {
		return err
	}
	status, err := parseStreamStatus(cfg.status)
	if err != nil {
		return err
	}
	reportExtra := map[string]string{"device_id": cfg.deviceID}
	if err := writeReportPacket(client, gen.SCMDDataReport_MACHINE_METRIC, &gen.MachineMetric{
		MachineId:      cfg.instanceID,
		InstanceId:     cfg.instanceID,
		InstanceName:   cfg.instanceID,
		NodeId:         cfg.nodeID,
		NodeName:       cfg.nodeName,
		Region:         cfg.region,
		NodeStatus:     "healthy",
		Status:         "running",
		CpuUsage:       32.5,
		CpuCount:       4,
		MemUsed:        512 * 1024 * 1024,
		MemTotal:       1024 * 1024 * 1024,
		DiskReadBytes:  256 * 1024,
		DiskWriteBytes: 128 * 1024,
		NetworkIn:      4 * 8 * 1024,
		NetworkOut:     6 * 8 * 1024,
		StartTime:      startedAt,
		Version:        "collection-client",
		Timestamp:      now,
	}); err != nil {
		return err
	}
	if err := writeReportPacket(client, gen.SCMDDataReport_NETWORK_METRIC, &gen.NetworkMetric{
		MachineId:     cfg.nodeID,
		SourceIp:      cfg.privateIP,
		DestinationIp: cfg.clientIP,
		Rtt:           18,
		Throughput:    8 * 8 * 1024,
		Timestamp:     now,
	}); err != nil {
		return err
	}
	if err := writeReportPacket(client, gen.SCMDDataReport_STREAM_ADD, &gen.StreamMetric{
		MachineId:             cfg.instanceID,
		StreamId:              cfg.streamID,
		Status:                status,
		Protocol:              protocol,
		Bitrate:               4096,
		Width:                 1280,
		Height:                720,
		StreamPath:            cfg.streamPath,
		StreamAlias:           cfg.streamName,
		Timestamp:             now,
		TenantId:              cfg.tenantID,
		NodeName:              cfg.nodeName,
		InstanceId:            cfg.instanceID,
		InstanceName:          cfg.instanceID,
		Fps:                   25,
		PacketLoss:            0.01,
		StartTime:             startedAt,
		AvgDelay:              35,
		TotalSessionsLifetime: 1,
		CurrentActiveSessions: 1,
		WatermarkEnabled:      true,
		NodeId:                cfg.nodeID,
		Extra:                 reportExtra,
	}); err != nil {
		return err
	}
	if err := writeReportPacket(client, gen.SCMDDataReport_SESSION_ADD, &gen.SessionMetric{
		SessionId:        cfg.sessionID,
		StreamId:         cfg.streamID,
		StreamName:       cfg.streamName,
		TenantId:         cfg.tenantID,
		ClientId:         cfg.clientID,
		ClientIp:         cfg.clientIP,
		ClientType:       cfg.clientType,
		UserName:         cfg.userName,
		Protocol:         protocol,
		StartTime:        startedAt,
		CurrentFps:       25,
		CurrentBitrate:   2048,
		CurrentWidth:     1280,
		CurrentHeight:    720,
		MachineId:        cfg.instanceID,
		NodeName:         cfg.nodeName,
		InstanceId:       cfg.instanceID,
		InstanceName:     cfg.instanceID,
		TotalLinkLatency: 18,
		Timestamp:        now,
		Extra:            reportExtra,
		LinkHops: []*gen.LinkHop{{
			HopId:   "hop-" + cfg.nodeID,
			HopName: cfg.nodeName,
			HopType: "node",
			NodeId:  cfg.nodeID,
			Latency: 18,
		}},
		NodeId: cfg.nodeID,
	}); err != nil {
		return err
	}
	fmt.Printf(
		"report sent tenant=%s node=%s instance=%s stream=%s session=%s\n",
		cfg.tenantID,
		cfg.nodeID,
		cfg.instanceID,
		cfg.streamID,
		cfg.sessionID,
	)
	return nil
}

// sendCloseReports sends lifecycle delete packets for the sample stream and session.
func sendCloseReports(client *network.TcpClient, cfg clientConfig) error {
	now := time.Now().UnixMilli()
	protocol, err := parseStreamProtocol(cfg.protocol)
	if err != nil {
		return err
	}
	if err := writeReportPacket(client, gen.SCMDDataReport_SESSION_DELETE, &gen.SessionMetric{
		SessionId:    cfg.sessionID,
		StreamId:     cfg.streamID,
		StreamName:   cfg.streamName,
		TenantId:     cfg.tenantID,
		ClientId:     cfg.clientID,
		ClientIp:     cfg.clientIP,
		ClientType:   cfg.clientType,
		Protocol:     protocol,
		MachineId:    cfg.instanceID,
		NodeName:     cfg.nodeName,
		InstanceId:   cfg.instanceID,
		InstanceName: cfg.instanceID,
		Timestamp:    now,
		NodeId:       cfg.nodeID,
		Extra:        map[string]string{"device_id": cfg.deviceID},
	}); err != nil {
		return err
	}
	if err := writeReportPacket(client, gen.SCMDDataReport_STREAM_DELETE, &gen.StreamMetric{
		MachineId:    cfg.instanceID,
		StreamId:     cfg.streamID,
		Status:       gen.StreamStatus_SS_CLOSED,
		Protocol:     protocol,
		StreamPath:   cfg.streamPath,
		StreamAlias:  cfg.streamName,
		Timestamp:    now,
		TenantId:     cfg.tenantID,
		NodeName:     cfg.nodeName,
		InstanceId:   cfg.instanceID,
		InstanceName: cfg.instanceID,
		NodeId:       cfg.nodeID,
		Extra:        map[string]string{"device_id": cfg.deviceID},
	}); err != nil {
		return err
	}
	fmt.Printf("close report sent stream=%s session=%s\n", cfg.streamID, cfg.sessionID)
	return nil
}

// writeReportPacket writes one net-flux data-report packet.
func writeReportPacket(client *network.TcpClient, subcmd gen.SCMDDataReport, packet proto.Message) error {
	if err := client.Write(uint8(gen.CMD_DATA_REPORT), uint8(subcmd), packet); err != nil {
		return fmt.Errorf("write data report packet subcmd=%s: %w", subcmd.String(), err)
	}
	return nil
}

// parseStreamProtocol converts a CLI protocol token into a net-flux enum.
func parseStreamProtocol(value string) (gen.StreamProtocol, error) {
	switch normalizeEnumToken(value) {
	case "", "hls":
		return gen.StreamProtocol_SP_HLS, nil
	case "rtmp":
		return gen.StreamProtocol_SP_RTMP, nil
	case "rtsp":
		return gen.StreamProtocol_SP_RTSP, nil
	case "httpflv":
		return gen.StreamProtocol_SP_HTTP_FLV, nil
	case "wsflv":
		return gen.StreamProtocol_SP_WS_FLV, nil
	case "httpsflv":
		return gen.StreamProtocol_SP_HTTPS_FLV, nil
	case "wssflv":
		return gen.StreamProtocol_SP_WSS_FLV, nil
	case "gb28181":
		return gen.StreamProtocol_SP_GB28181, nil
	default:
		return gen.StreamProtocol_SP_UNIVERSAL, fmt.Errorf("unsupported protocol %q", value)
	}
}

// parseStreamStatus converts a CLI stream status token into a net-flux enum.
func parseStreamStatus(value string) (gen.StreamStatus, error) {
	switch normalizeEnumToken(value) {
	case "", "running":
		return gen.StreamStatus_SS_RUNNING, nil
	case "inactive":
		return gen.StreamStatus_SS_INACTIVE, nil
	case "error":
		return gen.StreamStatus_SS_ERROR, nil
	case "closed":
		return gen.StreamStatus_SS_CLOSED, nil
	case "timeout":
		return gen.StreamStatus_SS_TIMEOUT, nil
	case "cancelled", "canceled":
		return gen.StreamStatus_SS_CANCELLED, nil
	case "failed":
		return gen.StreamStatus_SS_FAILED, nil
	default:
		return gen.StreamStatus_SS_UNIVERSAL, fmt.Errorf("unsupported status %q", value)
	}
}

// normalizeEnumToken makes CLI enum values tolerant to common separators.
func normalizeEnumToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	return value
}

// waitSettle waits for the server-side fire-and-forget packet handler to settle.
func waitSettle(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wait settle: %w", ctx.Err())
	}
}

// parseExtra converts comma-separated key=value metadata into a proto map.
func parseExtra(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	pairs := strings.Split(raw, ",")
	extra := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid extra pair %q, expected key=value", pair)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("invalid extra pair %q, key cannot be empty", pair)
		}
		extra[key] = strings.TrimSpace(value)
	}
	return extra, nil
}

// newClientHandler creates the TCP callback handler used by the client.
func newClientHandler(verbose bool) *clientHandler {
	return &clientHandler{
		lookupAckCh: make(chan *gen.LookupAck, 1),
		pongCh:      make(chan *gen.Pong, 1),
		verbose:     verbose,
	}
}

// OnConnect records the remote address when verbose logging is enabled.
func (h *clientHandler) OnConnect(conn network.TCPConn) error {
	if h.verbose {
		fmt.Fprintf(os.Stderr, "connected local=%s remote=%s\n", conn.LocalAddr().String(), conn.RemoteAddr().String())
	}
	return nil
}

// OnClose records connection close events when verbose logging is enabled.
func (h *clientHandler) OnClose(conn network.TCPConn) {
	if h.verbose {
		fmt.Fprintf(os.Stderr, "closed remote=%s\n", conn.RemoteAddr().String())
	}
}

// OnCmdSystem receives Pong acknowledgements from the TCP connection.
func (h *clientHandler) OnCmdSystem(_ network.TCPConn, pkt proto.Message) error {
	pong, ok := pkt.(*gen.Pong)
	if !ok {
		return fmt.Errorf("unexpected system packet: %T", pkt)
	}
	select {
	case h.pongCh <- pong:
	default:
	}
	return nil
}

// OnCmdDiscovery receives lookup acknowledgements from the TCP connection.
func (h *clientHandler) OnCmdDiscovery(_ network.TCPConn, pkt proto.Message) error {
	ack, ok := pkt.(*gen.LookupAck)
	if !ok {
		return fmt.Errorf("unexpected discovery packet: %T", pkt)
	}
	select {
	case h.lookupAckCh <- ack:
	default:
	}
	return nil
}

// OnCmdDataReport rejects unexpected data-report callbacks on this client.
func (h *clientHandler) OnCmdDataReport(_ network.TCPConn, subcmd uint8, pkt proto.Message) error {
	return fmt.Errorf("unexpected data report packet subcmd=%d packet=%T", subcmd, pkt)
}

// OnCmdConfig rejects unexpected config callbacks on this client.
func (h *clientHandler) OnCmdConfig(_ network.TCPConn, pkt proto.Message) error {
	return fmt.Errorf("unexpected config packet: %T", pkt)
}

// OnCmdEvent rejects unexpected event callbacks on this client.
func (h *clientHandler) OnCmdEvent(_ network.TCPConn, pkt proto.Message) error {
	return fmt.Errorf("unexpected event packet: %T", pkt)
}

// OnCmdControl rejects unexpected control callbacks on this client.
func (h *clientHandler) OnCmdControl(_ network.TCPConn, pkt proto.Message) error {
	return fmt.Errorf("unexpected control packet: %T", pkt)
}

// waitLookupAck waits for one LookupAck or the command timeout.
func (h *clientHandler) waitLookupAck(ctx context.Context) (*gen.LookupAck, error) {
	select {
	case ack := <-h.lookupAckCh:
		return ack, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("wait lookup ack: %w", ctx.Err())
	}
}

// waitPong waits for one Pong or the command timeout.
func (h *clientHandler) waitPong(ctx context.Context) (*gen.Pong, error) {
	select {
	case pong := <-h.pongCh:
		return pong, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("wait pong: %w", ctx.Err())
	}
}

// printLookupAck prints the returned lookup acknowledgement in proto JSON.
func printLookupAck(ack *gen.LookupAck) error {
	payload, err := protojson.MarshalOptions{
		Multiline:     true,
		Indent:        "  ",
		UseProtoNames: true,
	}.Marshal(ack)
	if err != nil {
		return fmt.Errorf("marshal lookup ack: %w", err)
	}
	fmt.Println(string(payload))
	return nil
}

// closeClient closes the TCP client and reports close failures to stderr.
func closeClient(client *network.TcpClient) {
	if err := client.Close(); err != nil && !errors.Is(err, network.ErrClosed) && !errors.Is(err, network.ErrNoReady) {
		fmt.Fprintf(os.Stderr, "close tcp client failed: %v\n", err)
	}
}
