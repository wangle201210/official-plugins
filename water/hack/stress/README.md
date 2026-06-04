# Water Stress Tool

This tool sends concurrent HTTP requests to the LinaPro water plugin APIs and prints success rate, RPS, and response-time metrics.

## Build

```bash
cd apps/lina-plugins/water
go build -o /tmp/water-stress ./hack/stress
```

## Preview API

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

## Submit API

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

Use `-data` instead of `-image` to send a fully custom JSON body.
