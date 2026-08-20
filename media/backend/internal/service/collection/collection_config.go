// This file loads and validates media collection server plugin configuration.

package collection

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/plugincap"
)

// Media collection server configuration keys and defaults.
const (
	// configKeyCollectionServer identifies the collection server configuration section.
	configKeyCollectionServer = "collectionServer"
	// configKeyCollectionServerEnabled identifies the collection server enable switch.
	configKeyCollectionServerEnabled = "collectionServer.enabled"
	// configKeyCollectionServerAddr identifies the TCP listen address.
	configKeyCollectionServerAddr = "collectionServer.addr"
	// configKeyCollectionServerDiscoveryEnabled identifies the discovery enable switch.
	configKeyCollectionServerDiscoveryEnabled = "collectionServer.discovery.enabled"
	// configKeyCollectionServerDiscoveryHost identifies the Nacos host.
	configKeyCollectionServerDiscoveryHost = "collectionServer.discovery.host"
	// configKeyCollectionServerDiscoveryPort identifies the Nacos port.
	configKeyCollectionServerDiscoveryPort = "collectionServer.discovery.port"
	// configKeyCollectionServerDiscoveryNamespace identifies the Nacos namespace.
	configKeyCollectionServerDiscoveryNamespace = "collectionServer.discovery.namespace"
	// configKeyCollectionServerDiscoveryLogDir identifies the Nacos SDK log directory.
	configKeyCollectionServerDiscoveryLogDir = "collectionServer.discovery.logDir"
	// configKeyCollectionServerDiscoveryCacheDir identifies the Nacos SDK cache directory.
	configKeyCollectionServerDiscoveryCacheDir = "collectionServer.discovery.cacheDir"
	// configKeyCollectionServerDiscoveryNotLoadCacheAtStart identifies whether Nacos skips local cache loading.
	configKeyCollectionServerDiscoveryNotLoadCacheAtStart = "collectionServer.discovery.notLoadCacheAtStart"
	// configKeyCollectionServerDiscoveryTimeout identifies the Nacos request timeout in milliseconds.
	configKeyCollectionServerDiscoveryTimeout = "collectionServer.discovery.timeout"
	// configKeyCollectionServerDiscoveryUsername identifies the Nacos username.
	configKeyCollectionServerDiscoveryUsername = "collectionServer.discovery.username"
	// configKeyCollectionServerDiscoveryPassword identifies the Nacos password.
	configKeyCollectionServerDiscoveryPassword = "collectionServer.discovery.password"
	// configKeyCollectionServerDiscoveryNode identifies the default node ID.
	configKeyCollectionServerDiscoveryNode = "collectionServer.discovery.node"
	// defaultAddr is the net-flux example-compatible listen address.
	defaultAddr = ":1911"
	// defaultDiscoveryHost is the net-flux example-compatible Nacos host.
	defaultDiscoveryHost = "127.0.0.1"
	// defaultDiscoveryPort is the net-flux example-compatible Nacos port.
	defaultDiscoveryPort = 8848
	// defaultDiscoveryNamespace is the net-flux example-compatible Nacos namespace.
	defaultDiscoveryNamespace = "public"
	// defaultDiscoveryLogDir is the net-flux example-compatible Nacos SDK log directory.
	defaultDiscoveryLogDir = "./logs"
	// defaultDiscoveryCacheDir is the net-flux example-compatible Nacos SDK cache directory.
	defaultDiscoveryCacheDir = "./cache"
	// defaultDiscoveryTimeout is the net-flux example-compatible Nacos timeout in milliseconds.
	defaultDiscoveryTimeout = 5000
	// defaultDiscoveryUsername is the net-flux example-compatible Nacos username.
	defaultDiscoveryUsername = "nacos"
	// defaultDiscoveryPassword is the net-flux example-compatible Nacos password.
	defaultDiscoveryPassword = "nacos"
	// defaultDiscoveryNode is the net-flux example-compatible default node.
	defaultDiscoveryNode = 1
)

// Config holds media collection server configuration.
type Config struct {
	// Enabled controls whether the TCP collection server starts with the host.
	Enabled bool `json:"enabled"`
	// Addr is the TCP listen address used by net-flux clients.
	Addr string `json:"addr"`
	// Discovery controls optional Nacos discovery command handling.
	Discovery DiscoveryConfig `json:"discovery"`
}

// DiscoveryConfig holds optional Nacos discovery configuration.
type DiscoveryConfig struct {
	// Enabled controls whether discovery commands create and use a Nacos client.
	Enabled bool `json:"enabled"`
	// Host is the Nacos server host.
	Host string `json:"host"`
	// Port is the Nacos server port.
	Port int `json:"port"`
	// Namespace is the Nacos namespace.
	Namespace string `json:"namespace"`
	// LogDir is the Nacos SDK log directory.
	LogDir string `json:"logDir"`
	// CacheDir is the Nacos SDK local cache directory.
	CacheDir string `json:"cacheDir"`
	// NotLoadCacheAtStart controls whether Nacos SDK skips loading local disk cache at client creation.
	NotLoadCacheAtStart bool `json:"notLoadCacheAtStart"`
	// Timeout is the Nacos SDK request timeout in milliseconds.
	Timeout int `json:"timeout"`
	// Username is the optional Nacos username.
	Username string `json:"username"`
	// Password is the optional Nacos password.
	Password string `json:"password"`
	// Node is the fallback node ID for discovery packets that omit node.
	Node int `json:"node"`
}

// LoadConfig reads and validates collection server configuration.
func LoadConfig(ctx context.Context, reader plugincap.ConfigService) (*Config, error) {
	if reader == nil {
		return nil, gerror.New("media collection config reader cannot be nil")
	}

	cfg := defaultConfig()
	raw := &Config{}
	if err := reader.Scan(ctx, configKeyCollectionServer, raw); err != nil {
		return nil, gerror.Wrap(err, "scan media collection config failed")
	}
	cfg.Enabled = raw.Enabled
	if strings.TrimSpace(raw.Addr) != "" {
		cfg.Addr = strings.TrimSpace(raw.Addr)
	}

	enabled, err := reader.Bool(ctx, configKeyCollectionServerEnabled, cfg.Enabled)
	if err != nil {
		return nil, gerror.Wrap(err, "read media collection enabled config failed")
	}
	cfg.Enabled = enabled

	addrValue, err := reader.Get(ctx, configKeyCollectionServerAddr, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "read media collection addr config failed")
	}
	if addrValue != nil && !addrValue.IsNil() && strings.TrimSpace(addrValue.String()) == "" {
		return nil, gerror.Newf("config %s cannot be empty", configKeyCollectionServerAddr)
	}
	addr, err := reader.String(ctx, configKeyCollectionServerAddr, cfg.Addr)
	if err != nil {
		return nil, gerror.Wrap(err, "read media collection addr config failed")
	}
	cfg.Addr = strings.TrimSpace(addr)

	if err := loadDiscoveryConfig(ctx, reader, cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// defaultConfig returns the collection server defaults.
func defaultConfig() *Config {
	return &Config{
		Enabled:   false,
		Addr:      defaultAddr,
		Discovery: defaultDiscoveryConfig(),
	}
}

// defaultDiscoveryConfig returns the optional Nacos discovery defaults.
func defaultDiscoveryConfig() DiscoveryConfig {
	return DiscoveryConfig{
		Enabled:             false,
		Host:                defaultDiscoveryHost,
		Port:                defaultDiscoveryPort,
		Namespace:           defaultDiscoveryNamespace,
		LogDir:              defaultDiscoveryLogDir,
		CacheDir:            defaultDiscoveryCacheDir,
		NotLoadCacheAtStart: true,
		Timeout:             defaultDiscoveryTimeout,
		Username:            defaultDiscoveryUsername,
		Password:            defaultDiscoveryPassword,
		Node:                defaultDiscoveryNode,
	}
}

// loadDiscoveryConfig reads optional Nacos discovery settings.
func loadDiscoveryConfig(ctx context.Context, reader plugincap.ConfigService, cfg *Config) error {
	enabled, err := reader.Bool(ctx, configKeyCollectionServerDiscoveryEnabled, cfg.Discovery.Enabled)
	if err != nil {
		return gerror.Wrap(err, "read media collection discovery enabled config failed")
	}
	cfg.Discovery.Enabled = enabled

	if cfg.Discovery.Host, err = readStringConfig(ctx, reader, configKeyCollectionServerDiscoveryHost, cfg.Discovery.Host, true); err != nil {
		return err
	}
	if cfg.Discovery.Namespace, err = readStringConfig(ctx, reader, configKeyCollectionServerDiscoveryNamespace, cfg.Discovery.Namespace, true); err != nil {
		return err
	}
	if cfg.Discovery.LogDir, err = readStringConfig(ctx, reader, configKeyCollectionServerDiscoveryLogDir, cfg.Discovery.LogDir, true); err != nil {
		return err
	}
	if cfg.Discovery.CacheDir, err = readStringConfig(ctx, reader, configKeyCollectionServerDiscoveryCacheDir, cfg.Discovery.CacheDir, true); err != nil {
		return err
	}
	if cfg.Discovery.Username, err = readStringConfig(ctx, reader, configKeyCollectionServerDiscoveryUsername, cfg.Discovery.Username, true); err != nil {
		return err
	}
	if cfg.Discovery.Password, err = readStringConfig(ctx, reader, configKeyCollectionServerDiscoveryPassword, cfg.Discovery.Password, true); err != nil {
		return err
	}

	if cfg.Discovery.Port, err = reader.Int(ctx, configKeyCollectionServerDiscoveryPort, cfg.Discovery.Port); err != nil {
		return gerror.Wrap(err, "read media collection discovery port config failed")
	}
	if cfg.Discovery.Timeout, err = reader.Int(ctx, configKeyCollectionServerDiscoveryTimeout, cfg.Discovery.Timeout); err != nil {
		return gerror.Wrap(err, "read media collection discovery timeout config failed")
	}
	if cfg.Discovery.Node, err = reader.Int(ctx, configKeyCollectionServerDiscoveryNode, cfg.Discovery.Node); err != nil {
		return gerror.Wrap(err, "read media collection discovery node config failed")
	}
	if cfg.Discovery.NotLoadCacheAtStart, err = reader.Bool(ctx, configKeyCollectionServerDiscoveryNotLoadCacheAtStart, cfg.Discovery.NotLoadCacheAtStart); err != nil {
		return gerror.Wrap(err, "read media collection discovery not load cache at start config failed")
	}
	return nil
}

// readStringConfig reads one string key while preserving explicit blank values when allowed.
func readStringConfig(
	ctx context.Context,
	reader plugincap.ConfigService,
	key string,
	defaultValue string,
	allowBlank bool,
) (string, error) {
	value, err := reader.Get(ctx, key, nil)
	if err != nil {
		return "", gerror.Wrapf(err, "read media collection config %s failed", key)
	}
	if value != nil && !value.IsNil() {
		text := strings.TrimSpace(value.String())
		if text == "" && !allowBlank {
			return "", gerror.Newf("config %s cannot be empty", key)
		}
		return text, nil
	}

	text, err := reader.String(ctx, key, defaultValue)
	if err != nil {
		return "", gerror.Wrapf(err, "read media collection config %s failed", key)
	}
	return strings.TrimSpace(text), nil
}

// validate verifies collection server configuration constraints.
func (c *Config) validate() error {
	if c == nil {
		return gerror.New("media collection config cannot be nil")
	}
	if strings.TrimSpace(c.Addr) == "" {
		return gerror.Newf("config %s cannot be empty", configKeyCollectionServerAddr)
	}
	if err := c.Discovery.validate(); err != nil {
		return err
	}
	return nil
}

// validate verifies optional Nacos discovery configuration constraints.
func (c DiscoveryConfig) validate() error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.Host) == "" {
		return gerror.Newf("config %s cannot be empty when discovery is enabled", configKeyCollectionServerDiscoveryHost)
	}
	if c.Port <= 0 || c.Port > 65535 {
		return gerror.Newf("config %s must be between 1 and 65535 when discovery is enabled", configKeyCollectionServerDiscoveryPort)
	}
	if _, err := newNacosDiscoSetting(c); err != nil {
		return err
	}
	if strings.TrimSpace(c.Namespace) == "" {
		return gerror.Newf("config %s cannot be empty when discovery is enabled", configKeyCollectionServerDiscoveryNamespace)
	}
	if strings.TrimSpace(c.LogDir) == "" {
		return gerror.Newf("config %s cannot be empty when discovery is enabled", configKeyCollectionServerDiscoveryLogDir)
	}
	if strings.TrimSpace(c.CacheDir) == "" {
		return gerror.Newf("config %s cannot be empty when discovery is enabled", configKeyCollectionServerDiscoveryCacheDir)
	}
	if c.Timeout <= 0 {
		return gerror.Newf("config %s must be positive when discovery is enabled", configKeyCollectionServerDiscoveryTimeout)
	}
	if c.Node <= 0 {
		return gerror.Newf("config %s must be positive when discovery is enabled", configKeyCollectionServerDiscoveryNode)
	}
	return nil
}
