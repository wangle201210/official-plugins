# 水印源码插件

水印源码插件提供基于远端媒体策略接口的截图水印处理能力。

插件通过配置调用`mediaopen`内部策略解析接口，不要求当前集群同时安装`media`源码插件。

## 运行配置

截图水印渲染规则保存在`media_strategy.strategy`的`snapshot_watermark`节点中：

```yaml
snapshot_watermark:
  order: LinaPro Water
  font_size: 40
  color: "#ffffff"
  opacity: 0.7
  rotate: -30
```

`media_strategy.enable`是策略是否参与水印渲染的唯一开关。未配置`opacity`时，插件默认使用`0.15`。

远端策略查询通过`water`插件运行时配置指定：

```yaml
mediaStrategy:
  baseUrl: "http://media-linapro.example.com"
  apiKey: "media"
  timeout: 10s
```

`baseUrl`指向部署了`media`插件的 LinaPro 集群。解析器调用`GET /api/v1/strategies/resolve`，并将`apiKey`作为`X-Inner-Api-Key`请求头发送。

服务端运行并发配置在宿主后端配置文件中：

```yaml
water:
  consumerCount: 1
```

`consumerCount`控制异步水印任务消费者并发数。小于`1`时回退为`1`，大于`32`时按`32`封顶。

异步任务状态快照复用宿主 `pluginhost.HostServices.Cache()` 服务，默认保留 12 小时。插件不定义专属 Redis 配置命名空间，也不维护插件自有缓存后端。
