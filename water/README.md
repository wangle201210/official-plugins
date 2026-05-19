# Water Source Plugin

The water source plugin provides image watermark processing based on the shared media strategy tables.

It reads `media_strategy`, `media_strategy_tenant`, `media_strategy_device`, and `media_strategy_device_tenant`; it does not create its own storage tables or tenant-isolated storage.

## Runtime Configuration

Watermark rendering rules are stored in `media_strategy.strategy` YAML. The service runtime concurrency is configured in the host backend config:

```yaml
water:
  consumerCount: 1
```

`consumerCount` controls asynchronous watermark task consumers. Values below `1` fall back to `1`; values above `32` are capped at `32`.

Asynchronous task status snapshots reuse the host `pluginhost.HostServices.Cache()` service and stay queryable for 12 hours. The plugin does not define a plugin-specific Redis configuration namespace or maintain a plugin-owned cache backend.
