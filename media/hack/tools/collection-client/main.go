// Command collection-client calls the media plugin net-flux TCP collection server.
//
// It is a small cross-platform development tool for verifying the LinaPro media
// TCP discovery path without calling Nacos directly. The command sends net-flux
// discovery packets to collectionServer.addr and prints lookup acknowledgements.
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
	actionRegister       = "register"
	actionLookup         = "lookup"
	actionDeregister     = "deregister"
	actionRegisterLookup = "register-lookup"
)

// clientConfig holds command-line options for one TCP discovery call.
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
	timeout     time.Duration // timeout bounds connect and lookup waiting time.
	settle      time.Duration // settle waits after fire-and-forget writes.
	verbose     bool          // verbose prints connection callbacks.
}

// clientHandler receives asynchronous net-flux packets from the TCP client.
type clientHandler struct {
	lookupAckCh chan *gen.LookupAck
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
	flag.StringVar(&cfg.action, "action", actionRegisterLookup, "action: register, lookup, deregister, register-lookup")
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
	flag.BoolVar(&cfg.ephemeral, "ephemeral", true, "instance ephemeral flag")
	flag.StringVar(&cfg.extra, "extra", "", "extra metadata, comma-separated key=value pairs")
	flag.DurationVar(&cfg.timeout, "timeout", 5*time.Second, "connection and lookup timeout")
	flag.DurationVar(&cfg.settle, "settle", 500*time.Millisecond, "wait after register or deregister before lookup or exit")
	flag.BoolVar(&cfg.verbose, "v", false, "print connection callbacks")
	flag.Parse()
	return cfg
}

// run opens one net-flux TCP client and executes the selected discovery action.
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
	if cfg.addr == "" {
		return cfg, errors.New("addr cannot be empty")
	}
	if cfg.serviceName == "" {
		return cfg, errors.New("service cannot be empty")
	}
	if cfg.instanceID == "" {
		cfg.instanceID = cfg.serviceName
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
	case actionRegister, actionDeregister, actionRegisterLookup:
		if cfg.privateIP == "" {
			return cfg, errors.New("private-ip cannot be empty")
		}
		if cfg.privatePort <= 0 || cfg.privatePort > 65535 {
			return cfg, errors.New("private-port must be between 1 and 65535")
		}
	}
	return cfg, nil
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

// OnCmdSystem accepts system callbacks from the TCP connection.
func (h *clientHandler) OnCmdSystem(_ network.TCPConn, _ proto.Message) error {
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
