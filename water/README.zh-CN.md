# 水印源码插件

水印源码插件提供基于媒体策略表的截图水印处理能力。

插件读取 `media_strategy`、`media_strategy_tenant`、`media_strategy_device`、`media_strategy_device_tenant`，不创建自有存储表，也不做宿主租户隔离存储。

## 运行配置

水印渲染规则保存在 `media_strategy.strategy` 的 YAML 内容中。服务端运行并发配置在宿主后端配置文件中：

```yaml
water:
  consumerCount: 1
```

`consumerCount` 控制异步水印任务消费者并发数。小于 `1` 时回退为 `1`，大于 `32` 时按 `32` 封顶。
