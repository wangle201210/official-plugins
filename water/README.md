# Water Source Plugin

The water source plugin provides image watermark processing based on the remote media strategy API.

It calls the configured `mediaopen` internal strategy resolution API and does not require the `media` source plugin to be installed in the same cluster.

## Runtime Configuration

Snapshot watermark rendering rules are stored in the `snapshot_watermark` node of `media_strategy.strategy` YAML:

```yaml
snapshot_watermark:
  order: LinaPro Water
  font_size: 40
  color: "#ffffff"
  opacity: 0.7
  rotate: -30
```

`media_strategy.enable` is the only switch for whether a strategy participates in watermark rendering. When `opacity` is omitted, the plugin uses `0.15`.

Remote strategy lookup and service concurrency are configured through the water plugin runtime config:

```yaml
mediaStrategy:
  baseUrl: "http://media-linapro.example.com"
  apiKey: "media"
  timeout: 10s
water:
  consumerCount: 1
```

`baseUrl` points to the LinaPro cluster where the `media` plugin is installed. The resolver calls `GET /api/v1/strategies/resolve` and sends `apiKey` as `X-Inner-Api-Key`.

`consumerCount` controls asynchronous watermark task consumers. Values below `1` fall back to `1`; values above `32` are capped at `32`.

Asynchronous task status snapshots reuse the host `pluginhost.HostServices.Cache()` service and stay queryable for 12 hours. The plugin does not define a plugin-specific Redis configuration namespace or maintain a plugin-owned cache backend.
