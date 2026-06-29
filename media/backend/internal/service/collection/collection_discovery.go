// This file bridges net-flux discovery packets to the optional Nacos client.

package collection

import (
	"strconv"
	"strings"
	"sync"

	"github.com/dellinger2023/net-flux/gen"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// discoveryClient is the Nacos-backed discovery adapter used by collection server.
type discoveryClient interface {
	RegisterInstance(instance *gen.Instance) error
	DeregisterInstance(serviceName, groupName, ip string, port uint64) error
	GetServiceInstanceByGroup(serviceName, groupName string) (*gen.Instance, error)
	Close()
}

// discoveryClientFactory creates a discovery client. Tests replace this with a fake factory.
type discoveryClientFactory func(cfg DiscoveryConfig) (discoveryClient, error)

// newDiscoveryClient creates the production Nacos discovery client.
var newDiscoveryClient discoveryClientFactory = func(cfg DiscoveryConfig) (discoveryClient, error) {
	nacosClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": []constant.ServerConfig{
			{
				IpAddr: cfg.Host,
				Port:   uint64(cfg.Port),
			},
		},
		"clientConfig": newNacosClientConfig(cfg),
	})
	if err != nil {
		return nil, err
	}
	return &nacosDiscoveryClient{client: nacosClient}, nil
}

// newNacosClientConfig builds the Nacos SDK config used by short-lived discovery clients.
func newNacosClientConfig(cfg DiscoveryConfig) constant.ClientConfig {
	return constant.ClientConfig{
		NamespaceId:          cfg.Namespace,
		TimeoutMs:            uint64(cfg.Timeout),
		NotLoadCacheAtStart:  cfg.NotLoadCacheAtStart,
		UpdateCacheWhenEmpty: true,
		LogDir:               cfg.LogDir,
		CacheDir:             cfg.CacheDir,
		LogLevel:             "info",
		Username:             cfg.Username,
		Password:             cfg.Password,
	}
}

// nacosDiscoveryClient adapts Nacos SDK naming operations to net-flux packets.
type nacosDiscoveryClient struct {
	client naming_client.INamingClient
}

// RegisterInstance registers one net-flux instance in Nacos.
func (c *nacosDiscoveryClient) RegisterInstance(instance *gen.Instance) error {
	_, err := c.client.RegisterInstance(newRegisterInstanceParam(instance))
	return err
}

// newRegisterInstanceParam builds a pod-independent Nacos registration request.
func newRegisterInstanceParam(instance *gen.Instance) vo.RegisterInstanceParam {
	return vo.RegisterInstanceParam{
		ServiceName: instance.GetInstanceName(),
		GroupName:   nodeGroup(instance.GetNode()),
		Ip:          instance.GetPrivateIp(),
		Port:        uint64(instance.GetPrivatePort()),
		Enable:      true,
		Healthy:     true,
		Weight:      1.0,
		Ephemeral:   false,
		Metadata:    instanceMetadata(instance),
	}
}

// DeregisterInstance deregisters one instance from Nacos.
func (c *nacosDiscoveryClient) DeregisterInstance(serviceName, groupName, ip string, port uint64) error {
	_, err := c.client.DeregisterInstance(newDeregisterInstanceParam(serviceName, groupName, ip, port))
	return err
}

// newDeregisterInstanceParam builds the Nacos deregistration request matching
// the persistent instances registered by this collection server.
func newDeregisterInstanceParam(serviceName, groupName, ip string, port uint64) vo.DeregisterInstanceParam {
	return vo.DeregisterInstanceParam{
		ServiceName: serviceName,
		GroupName:   groupName,
		Ip:          ip,
		Port:        port,
		Ephemeral:   false,
	}
}

// GetServiceInstanceByGroup queries one healthy Nacos instance and converts it to net-flux format.
func (c *nacosDiscoveryClient) GetServiceInstanceByGroup(serviceName, groupName string) (*gen.Instance, error) {
	instance, err := c.client.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
		GroupName:   groupName,
	})
	if err != nil {
		return nil, err
	}
	return nacosInstanceToProto(instance)
}

// Close releases Nacos SDK naming resources.
func (c *nacosDiscoveryClient) Close() {
	if c == nil || c.client == nil {
		return
	}
	c.client.CloseClient()
}

// discoveryRuntime creates short-lived Nacos discovery clients for TCP packets.
type discoveryRuntime struct {
	cfg     DiscoveryConfig
	factory discoveryClientFactory
	mu      sync.Mutex
	client  discoveryClient
}

// newDiscoveryRuntime creates a discovery runtime for one TCP server instance.
func newDiscoveryRuntime(cfg DiscoveryConfig) *discoveryRuntime {
	return &discoveryRuntime{
		cfg:     cfg,
		factory: newDiscoveryClient,
	}
}

// enabled reports whether discovery commands should be handled.
func (r *discoveryRuntime) enabled() bool {
	return r != nil && r.cfg.Enabled
}

// Register registers one media instance in Nacos.
func (r *discoveryRuntime) Register(instance *gen.Instance) error {
	if !r.enabled() {
		return gerror.New("media collection discovery is disabled")
	}
	normalized, err := r.normalizeInstance(instance)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	client, err := r.clientLocked()
	if err != nil {
		return err
	}
	defer r.closeClientLocked()
	return client.RegisterInstance(normalized)
}

// Deregister removes one media instance from Nacos.
func (r *discoveryRuntime) Deregister(packet *gen.Deregister) error {
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

	r.mu.Lock()
	defer r.mu.Unlock()

	client, err := r.clientLocked()
	if err != nil {
		return err
	}
	defer r.closeClientLocked()
	return client.DeregisterInstance(serviceName, r.groupForNode(packet.GetNode()), ip, uint64(packet.GetPort()))
}

// Lookup queries one media service instance from Nacos and builds the net-flux response.
func (r *discoveryRuntime) Lookup(packet *gen.Lookup) (*gen.LookupAck, error) {
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

	r.mu.Lock()
	defer r.mu.Unlock()

	client, err := r.clientLocked()
	if err != nil {
		return nil, err
	}
	defer r.closeClientLocked()
	instance, err := client.GetServiceInstanceByGroup(serviceName, r.groupForNode(packet.GetNode()))
	if err != nil {
		if isDiscoveryEmptyInstanceError(err) {
			return &gen.LookupAck{}, nil
		}
		return nil, err
	}
	if instance == nil {
		return &gen.LookupAck{}, nil
	}
	instance = cleanLookupInstance(instance)

	groupName := r.groupForNode(instance.GetNode())
	return &gen.LookupAck{
		Services: []*gen.Service{
			{
				Instances: []*gen.Instance{instance},
				Cluster:   "",
				Name:      instance.GetInstanceName(),
				GroupName: groupName,
				Valid:     packet.GetHealthy(),
			},
		},
	}, nil
}

// Close closes the current Nacos discovery client.
func (r *discoveryRuntime) Close() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closeClientLocked()
}

// clientLocked returns the current client or creates one. Caller must hold r.mu.
func (r *discoveryRuntime) clientLocked() (discoveryClient, error) {
	if r.client != nil {
		return r.client, nil
	}
	if r.factory == nil {
		return nil, gerror.New("media collection discovery client factory cannot be nil")
	}
	client, err := r.factory(r.cfg)
	if err != nil {
		return nil, gerror.Wrap(err, "create media collection discovery client failed")
	}
	r.client = client
	return r.client, nil
}

// closeClientLocked closes and clears the current client. Caller must hold r.mu.
func (r *discoveryRuntime) closeClientLocked() {
	if r.client == nil {
		return
	}
	r.client.Close()
	r.client = nil
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

// groupForNode returns the net-flux node group used by the sample server.
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

// instanceMetadata builds the Nacos metadata used by net-flux discovery.
func instanceMetadata(instance *gen.Instance) map[string]string {
	metadata := make(map[string]string, len(instance.GetExtra())+5)
	for key, value := range instance.GetExtra() {
		metadata[key] = value
	}
	metadata["inner_ip"] = instance.GetInnerIp()
	metadata["inner_port"] = strconv.Itoa(int(instance.GetInnerPort()))
	metadata["public_ip"] = instance.GetPublicIp()
	metadata["public_port"] = strconv.Itoa(int(instance.GetPublicPort()))
	metadata["node"] = nodeGroup(instance.GetNode())
	return metadata
}

// nacosInstanceToProto converts a Nacos instance to the net-flux Instance message.
func nacosInstanceToProto(instance *model.Instance) (*gen.Instance, error) {
	if instance == nil {
		return nil, nil
	}
	metadata := instance.Metadata
	innerPort, err := metadataInt32(metadata, "inner_port")
	if err != nil {
		return nil, err
	}
	publicPort, err := metadataInt32(metadata, "public_port")
	if err != nil {
		return nil, err
	}
	node, err := metadataInt32(metadata, "node")
	if err != nil {
		return nil, err
	}
	return &gen.Instance{
		InstanceId:   instance.InstanceId,
		InstanceName: cleanNacosServiceName(instance.ServiceName),
		PrivateIp:    instance.Ip,
		PrivatePort:  int32(instance.Port),
		InnerIp:      metadata["inner_ip"],
		InnerPort:    innerPort,
		PublicIp:     metadata["public_ip"],
		PublicPort:   publicPort,
		Weight:       float32(instance.Weight),
		Healthy:      instance.Healthy,
		Enable:       instance.Enable,
		Ephemeral:    instance.Ephemeral,
		Node:         node,
		Extra:        metadata,
	}, nil
}

// metadataInt32 reads one int32 value from Nacos metadata.
func metadataInt32(metadata map[string]string, key string) (int32, error) {
	value, err := strconv.Atoi(metadata[key])
	if err != nil {
		return 0, gerror.Wrapf(err, "read media collection discovery metadata %s failed", key)
	}
	return int32(value), nil
}

// cleanNacosServiceName removes the Nacos group prefix returned as group@@service.
func cleanNacosServiceName(serviceName string) string {
	if _, name, ok := strings.Cut(serviceName, "@@"); ok {
		return name
	}
	return serviceName
}

// cleanLookupInstance removes Nacos transport-only naming details from LookupAck payloads.
func cleanLookupInstance(instance *gen.Instance) *gen.Instance {
	cleanName := cleanNacosServiceName(instance.GetInstanceName())
	if cleanName == instance.GetInstanceName() {
		return instance
	}
	copied := *instance
	copied.InstanceName = cleanName
	return &copied
}

// isDiscoveryEmptyInstanceError recognizes Nacos' "no healthy instance" response as an empty lookup result.
func isDiscoveryEmptyInstanceError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "instance list is empty")
}
