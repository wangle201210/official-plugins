# 水印并发压测工具

该工具用于向 LinaPro water 插件接口发送并发 HTTP 请求，并输出成功率、RPS 和响应时间统计。

## 构建

```bash
cd apps/lina-plugins/water
go build -o /tmp/water-stress ./hack/stress
```

## 同步预览接口

```bash
go run ./hack/stress \
  -mode preview \
  -base-url http://127.0.0.1:8080/api/v1 \
  -token '<jwt>' \
  -image ./backend/internal/library/watermark/input.jpg \
  -tenant tenant-a \
  -device-id 34020000001320000001 \
  -concurrency 20 \
  -requests 200
```

## 异步提交接口

```bash
go run ./hack/stress \
  -mode submit \
  -base-url http://127.0.0.1:8080/api/v1 \
  -token '<jwt>' \
  -image ./backend/internal/library/watermark/input.jpg \
  -tenant tenant-a \
  -device-type gb \
  -device-id 34020000001320000001 \
  -concurrency 20 \
  -requests 200
```

如果需要完全自定义请求体，可以用 `-data` 替代 `-image`。
