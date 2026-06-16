# Water Source Plugin

The water source plugin provides image watermark processing based on the remote media strategy API.

It calls the configured `mediaopen` internal strategy resolution API and does not require the `media` source plugin to be installed in the same cluster.

## Runtime Configuration

Snapshot watermark rendering rules are stored in the `snapshot_watermark` node of `media_strategy.strategy` YAML:

```yaml
snapshot_watermark:
  text: LinaPro Water
  fontSize: 40
  color: "#ffffff"
  align: bottomRight
  opacity: 0.7
```

`media_strategy.enable` is the only switch for whether a strategy participates in watermark rendering. When `opacity` is omitted, the plugin uses `0.15`.

Remote strategy lookup is configured through the water plugin runtime config:

```yaml
mediaStrategy:
  baseUrl: "http://media-linapro.example.com"
  apiKey: "media"
  timeout: 10s
```

`baseUrl` points to the LinaPro cluster where the `media` plugin is installed. The resolver calls `GET /api/v1/strategies/resolve` and sends `apiKey` as `X-Inner-Api-Key`.

The service runtime concurrency is configured in the host backend config:

```yaml
water:
  consumerCount: 1
```

`consumerCount` controls asynchronous watermark task consumers. Values below `1` fall back to `1`; values above `32` are capped at `32`.

Asynchronous task status snapshots reuse the host `pluginhost.HostServices.Cache()` service and stay queryable for 12 hours. The plugin does not define a plugin-specific Redis configuration namespace or maintain a plugin-owned cache backend.
