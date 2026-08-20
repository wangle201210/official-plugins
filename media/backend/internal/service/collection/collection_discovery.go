// This file maps media discovery configuration and packets to the upstream
// net-flux naming client lifecycle.

package collection

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/dellinger2023/net-flux/pkg/naming"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Nacos server schemes accepted in discovery host URLs.
const (
	defaultNacosServerScheme = "http"
	secureNacosServerScheme  = "https"
	discoveryRegisterRetries = 10
	discoveryRegisterBackoff = 100 * time.Millisecond
)

// newNacosDiscoSetting maps media configuration to upstream naming settings.
func newNacosDiscoSetting(cfg DiscoveryConfig) (naming.DiscoSetting, error) {
	rawHost := strings.TrimSpace(cfg.Host)
	normalizedHost := rawHost
	if !strings.Contains(rawHost, "://") {
		if strings.ContainsAny(rawHost, "/?#") {
			return naming.DiscoSetting{}, gerror.Newf("config %s must be a hostname or HTTP(S) URL", configKeyCollectionServerDiscoveryHost)
		}
	} else {
		parsed, err := url.Parse(rawHost)
		if err != nil {
			return naming.DiscoSetting{}, gerror.Wrapf(err, "parse config %s failed", configKeyCollectionServerDiscoveryHost)
		}
		scheme := strings.ToLower(parsed.Scheme)
		if scheme != defaultNacosServerScheme && scheme != secureNacosServerScheme {
			return naming.DiscoSetting{}, gerror.Newf("config %s scheme must be http or https", configKeyCollectionServerDiscoveryHost)
		}
		if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
			return naming.DiscoSetting{}, gerror.Newf("config %s URL must not contain credentials, query, fragment, or path", configKeyCollectionServerDiscoveryHost)
		}
		host := parsed.Hostname()
		if host == "" {
			return naming.DiscoSetting{}, gerror.Newf("config %s URL hostname cannot be empty", configKeyCollectionServerDiscoveryHost)
		}
		if port := parsed.Port(); port != "" && port != strconv.Itoa(cfg.Port) {
			return naming.DiscoSetting{}, gerror.Newf("config %s URL port must match %s", configKeyCollectionServerDiscoveryHost, configKeyCollectionServerDiscoveryPort)
		}
		// naming.DiscoSetting expects a bare host. Keeping the scheme in Host
		// breaks the upstream gRPC client address.
		normalizedHost = host
	}

	return naming.DiscoSetting{
		Host:         normalizedHost,
		Port:         cfg.Port,
		Namespace:    cfg.Namespace,
		LogDir:       cfg.LogDir,
		CacheDir:     cfg.CacheDir,
		PreloadCache: !cfg.NotLoadCacheAtStart,
		Timeout:      cfg.Timeout,
		GroupName:    naming.DefaultGroupName,
		Username:     cfg.Username,
		Password:     cfg.Password,
		Node:         cfg.Node,
	}, nil
}

// discoveryRuntime reuses one naming client per TCP connection and service.
type discoveryRuntime struct {
	cfg     DiscoveryConfig
	factory *discoClientFactory
	err     error
}

// newDiscoveryRuntime creates a discovery runtime for one TCP server instance.
func newDiscoveryRuntime(cfg DiscoveryConfig) *discoveryRuntime {
	settings, err := newNacosDiscoSetting(cfg)
	return &discoveryRuntime{
		cfg:     cfg,
		factory: newDiscoClientFactory(settings),
		err:     err,
	}
}

// enabled reports whether discovery commands should be handled.
func (r *discoveryRuntime) enabled() bool {
	return r != nil && r.cfg.Enabled
}

// client returns the connection/service-scoped naming client.
func (r *discoveryRuntime) client(connID uint32, serviceName string) (discoveryClient, error) {
	if r == nil || r.factory == nil {
		return nil, gerror.New("media collection discovery client factory cannot be nil")
	}
	if r.err != nil {
		return nil, r.err
	}
	return r.factory.GetDiscoClient(connID, serviceName)
}

// Register registers one media instance in Nacos and keeps its naming client.
func (r *discoveryRuntime) Register(connID uint32, instance *gen.Instance) error {
	if !r.enabled() {
		return gerror.New("media collection discovery is disabled")
	}
	normalized, err := r.normalizeInstance(instance)
	if err != nil {
		return err
	}
	client, err := r.client(connID, normalized.GetInstanceName())
	if err != nil {
		return err
	}
	return registerDiscoveryInstance(client, normalized)
}

// registerDiscoveryInstance retries only the upstream client's startup race.
func registerDiscoveryInstance(client discoveryClient, instance *gen.Instance) error {
	var lastErr error
	for attempt := 0; attempt < discoveryRegisterRetries; attempt++ {
		if err := client.RegisterInstance(instance); err == nil {
			return nil
		} else {
			lastErr = err
			if !strings.Contains(strings.ToLower(err.Error()), "client not connected") {
				return err
			}
		}
		time.Sleep(discoveryRegisterBackoff)
	}
	return gerror.Wrap(lastErr, "register media discovery instance after client startup retries failed")
}

// Deregister removes one media instance and releases its service client.
func (r *discoveryRuntime) Deregister(connID uint32, packet *gen.Deregister) error {
	if !r.enabled() {
		return gerror.New("media collection discovery is disabled")
	}
	if packet == nil {
		return gerror.New("media collection discovery deregister packet cannot be nil")
	}
	serviceName := strings.TrimSpace(packet.GetInstanceName())
	if serviceName == "" {
		return gerror.New("media collection discovery deregister instance name cannot be empty")
	}
	ip := strings.TrimSpace(packet.GetIp())
	if ip == "" {
		return gerror.New("media collection discovery deregister ip cannot be empty")
	}
	if packet.GetPort() <= 0 {
		return gerror.New("media collection discovery deregister port must be positive")
	}
	client, err := r.client(connID, serviceName)
	if err != nil {
		return err
	}
	if err := client.DeregisterInstance(serviceName, r.groupForNode(packet.GetNode()), ip, uint64(packet.GetPort())); err != nil {
		return err
	}
	r.factory.RemoveDiscoClient(connID, serviceName)
	return nil
}

// Lookup queries all Nacos instances for one service and builds a net-flux response.
func (r *discoveryRuntime) Lookup(connID uint32, packet *gen.Lookup) (*gen.LookupAck, error) {
	if !r.enabled() {
		return nil, gerror.New("media collection discovery is disabled")
	}
	if packet == nil {
		return nil, gerror.New("media collection discovery lookup packet cannot be nil")
	}
	serviceName := strings.TrimSpace(packet.GetServiceName())
	if serviceName == "" {
		return nil, gerror.New("media collection discovery lookup service name cannot be empty")
	}
	client, err := r.client(connID, "")
	if err != nil {
		return nil, err
	}
	instances, err := client.GetServiceInstances(serviceName, r.groupForNode(packet.GetNode()), []string{})
	if err != nil {
		if isDiscoveryEmptyInstanceError(err) {
			return &gen.LookupAck{}, nil
		}
		return nil, err
	}
	if len(instances) == 0 {
		return &gen.LookupAck{}, nil
	}
	cleaned := make([]*gen.Instance, 0, len(instances))
	for _, instance := range instances {
		if instance != nil {
			cleaned = append(cleaned, cleanLookupInstance(instance))
		}
	}
	if len(cleaned) == 0 {
		return &gen.LookupAck{}, nil
	}
	return &gen.LookupAck{Services: []*gen.Service{{
		Instances: cleaned,
		Cluster:   "",
		Name:      serviceName,
		GroupName: r.groupForNode(packet.GetNode()),
		Valid:     packet.GetHealthy(),
	}}}, nil
}

// RemoveConnection releases all naming clients associated with a TCP connection.
func (r *discoveryRuntime) RemoveConnection(connID uint32) {
	if r == nil || r.factory == nil {
		return
	}
	r.factory.RemoveAll(connID)
}

// Close releases all Nacos naming clients owned by the runtime.
func (r *discoveryRuntime) Close() {
	if r == nil || r.factory == nil {
		return
	}
	r.factory.Close()
}

// normalizeInstance validates and copies one register packet.
func (r *discoveryRuntime) normalizeInstance(instance *gen.Instance) (*gen.Instance, error) {
	if instance == nil {
		return nil, gerror.New("media collection discovery instance packet cannot be nil")
	}
	if strings.TrimSpace(instance.GetInstanceName()) == "" {
		return nil, gerror.New("media collection discovery instance name cannot be empty")
	}
	if strings.TrimSpace(instance.GetPrivateIp()) == "" {
		return nil, gerror.New("media collection discovery private ip cannot be empty")
	}
	if instance.GetPrivatePort() <= 0 {
		return nil, gerror.New("media collection discovery private port must be positive")
	}

	normalized := *instance
	normalized.InstanceName = strings.TrimSpace(normalized.InstanceName)
	normalized.PrivateIp = strings.TrimSpace(normalized.PrivateIp)
	normalized.PublicIp = strings.TrimSpace(normalized.PublicIp)
	normalized.InnerIp = strings.TrimSpace(normalized.InnerIp)
	if normalized.Node <= 0 {
		normalized.Node = int32(r.cfg.Node)
	}
	if len(instance.Extra) > 0 {
		normalized.Extra = make(map[string]string, len(instance.Extra))
		for key, value := range instance.Extra {
			normalized.Extra[key] = value
		}
	}
	return &normalized, nil
}

// groupForNode returns the net-flux node group used by discovery packets.
func (r *discoveryRuntime) groupForNode(node int32) string {
	if node <= 0 {
		node = int32(r.cfg.Node)
	}
	return nodeGroup(node)
}

// nodeGroup returns the net-flux node group used by the sample server.
func nodeGroup(node int32) string {
	return strconv.Itoa(int(node))
}

// cleanNacosServiceName removes the Nacos group prefix returned as group@@service.
func cleanNacosServiceName(serviceName string) string {
	if _, name, ok := strings.Cut(serviceName, "@@"); ok {
		return name
	}
	return serviceName
}

// cleanLookupInstance removes Nacos transport-only naming details from responses.
func cleanLookupInstance(instance *gen.Instance) *gen.Instance {
	cleanName := cleanNacosServiceName(instance.GetInstanceName())
	if cleanName == instance.GetInstanceName() {
		return instance
	}
	copied := *instance
	copied.InstanceName = cleanName
	return &copied
}

// isDiscoveryEmptyInstanceError recognizes Nacos' empty lookup response.
func isDiscoveryEmptyInstanceError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "instance list is empty")
}
