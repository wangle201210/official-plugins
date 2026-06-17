# LinaPro Water Kubernetes Deployment

This directory provides a single-file Kubernetes deployment manifest for a `water`-specific LinaPro image with only the `water` source plugin auto-enabled.

The manifest deploys ten `linapro-water` pods behind one Kubernetes Service to spread CPU and memory pressure from watermark processing. It does not deploy the `media` plugin. Configure `mediaStrategy.baseUrl` to point to the LinaPro cluster where `media` is installed and where the `mediaopen` internal API is reachable.

## Files

| File | Purpose |
| ---- | ------- |
| `linapro-k8s.yaml` | All-in-one `water` deployment that creates PostgreSQL, Redis coordination, one database init Job, and ten LinaPro `water` pods inside the cluster. |
| `../docker/Dockerfile` | Dedicated `CGO` and FFmpeg/x264 image build for the `water` renderer. |
| `README.zh-CN.md` | Chinese usage guide with the same deployment steps. |

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- `kubectl` configured for the target cluster.
- The node can pull `ghcr.io/wangle201210/linapro-water:nightly-20260616`.
- The storage class used by `linapro-water-data` supports `ReadWriteMany`.
- A reachable `media` LinaPro cluster with the `media` plugin enabled.
- The `media` cluster exposes `GET /api/v1/strategies/resolve` to this `water` cluster.
- Port `30081` is available when using the bundled `NodePort` service.

## Configure

Before applying the manifest, edit `linapro-k8s.yaml` and replace these default values:

| Field | Default | Description |
| ----- | ------- | ----------- |
| `POSTGRES_PASSWORD` | `linapro-change-me` | PostgreSQL password. |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(linapro-water-postgres:5432)/linapro?sslmode=disable` | LinaPro database connection string. Keep it aligned with `POSTGRES_PASSWORD`. |
| `jwt.secret` | `linapro-jwt-change-me` | JWT signing secret. |
| `jwt.expire` | `24h` | JWT expiration duration. |
| `logger.level` | `info` | Runtime log level. Keep `info` for multi-replica deployments; use `all` only for short-lived debugging. |
| `cluster.enabled` | `true` | Enables LinaPro cluster coordination for the ten `water` pods. |
| `cluster.redis.address` | `linapro-water-redis:6379` | Redis endpoint used for cluster election, locks, and shared runtime KV cache. |
| `plugin.autoEnable` | `water` only | Only the `water` source plugin is automatically installed and enabled by this manifest. |
| `mediaStrategy.baseUrl` | `http://media-linapro.example.com` | Base URL of the remote LinaPro `media` cluster. |
| `mediaStrategy.apiKey` | `media` | Value sent as the `X-Inner-Api-Key` header to `mediaopen`. |
| `mediaStrategy.timeout` | `10s` | Timeout for one remote strategy lookup. |
| `water.consumerCount` | `2` | Number of asynchronous watermark task consumers per LinaPro pod. With ten replicas, the default total is twenty consumers. |
| `spec.replicas` | `10` | Number of `linapro-water` pods. |
| `resources.requests` | `500m` CPU, `512Mi` memory | Baseline resources for each watermark pod. Ten replicas request about `5` CPU and `5Gi` memory in total. |
| `resources.limits` | `2` CPU, `2Gi` memory | Upper resources for each watermark pod. Ten replicas can burst to `20` CPU and `20Gi` memory in total. |
| `spec.ports[0].nodePort` | `30081` | External access port for `NodePort`. |
| `linapro-water-data.resources.requests.storage` | `10Gi` | Shared `/app/data` storage for uploads, dynamic plugin artifacts, and local runtime files. |
| `resources.requests.storage` | `20Gi` | Storage size for the bundled PostgreSQL data volume. |

The `water` plugin reads its plugin-scoped runtime config from `/app/config/plugins/water/config.yaml`. The manifest mounts this file from the `linapro-water-plugin-config` Secret. The template `manifest/config/config.example.yaml` is documentation only and is not read as a runtime default.

For cross-cluster access, set `mediaStrategy.baseUrl` to the `media` cluster `Ingress`, load balancer, VPN address, or `NodePort` URL. If `media` is deployed from the sibling manifest in the same Kubernetes cluster, the cluster-local URL is usually `http://linapro.linapro.svc.cluster.local:9120`. The configured `mediaStrategy.apiKey` must match the `media` cluster `innerapi.apiKey` value; when the `media` plugin does not explicitly set `innerapi.apiKey`, its default is `media`.

The bundled PostgreSQL container sets `PGDATA` to `/var/lib/postgresql/data/pgdata` so volume-root metadata from some storage classes does not interfere with first-time database initialization.

The `linapro-water` Deployment mounts `/app/data` from the `linapro-water-data` `ReadWriteMany` claim. This keeps host upload storage, dynamic plugin artifacts, and other local runtime files shared across the ten pods. If the deployment is strictly API-only and the cluster has no `ReadWriteMany` storage class, you can replace this volume with `emptyDir`, but uploaded files, dynamic plugin artifacts, and local runtime files will then be pod-local and lost on pod restart.

## Build Image

The generic `ghcr.io/wangle201210/linapro:nightly-20260616` image is not valid for real `water` rendering because the default image build disables `CGO`. The `water` renderer requires the `CGO` branch and FFmpeg/x264 runtime libraries.

Build and push the dedicated image before applying the Kubernetes manifest:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f apps/lina-plugins/water/deploy/docker/Dockerfile \
  -t ghcr.io/wangle201210/linapro-water:nightly-20260616 \
  --push .
```

The Dockerfile reuses `linactl build` with `plugins=1` and `cgo_enabled=1`, then installs FFmpeg/x264 runtime libraries in the final image. The Deployment also runs `verify-watermark-runtime` before the server starts, so replacing the image with a non-`CGO` build fails fast instead of serving broken watermark APIs.

## Deploy

Apply the manifest:

```bash
kubectl apply -f linapro-k8s.yaml
```

The manifest creates a one-shot database initialization Job:

| Job | Purpose |
| --- | ------- |
| `linapro-water-init-database` | Runs `./lina init --confirm=init` once with the mounted `/app/config.yaml`. This creates or upgrades the host schema and required seed data. |

Each `linapro-water` pod runs four init containers before starting the server:

| Init container | Purpose |
| -------------- | ------- |
| `wait-for-postgres` | Waits until PostgreSQL accepts connections. |
| `wait-for-redis` | Waits until Redis accepts connections for cluster coordination. |
| `wait-for-database-init` | Waits until the init Job has created the final distributed-cache revision index from the last host SQL file. |
| `verify-watermark-runtime` | Fails fast when the selected image was built without `CGO` or is missing FFmpeg/x264 runtime libraries. |

After the server starts, `plugin.autoEnable` automatically installs and enables the `water` source plugin. The plugin install phase executes the `water` install SQL. The enabled `water` plugin then uses the mounted plugin config to call the remote `mediaopen` strategy resolver.

Check workload status:

```bash
kubectl -n linapro-water get pods
kubectl -n linapro-water get job
kubectl -n linapro-water get svc
```

Follow LinaPro logs from all `water` replicas:

```bash
kubectl -n linapro-water logs -f -l app=linapro-water -c linapro --max-log-requests=10
```

`kubectl logs deploy/linapro-water -f` follows a deployment-selected pod and can miss activity from the other replicas.

If database initialization fails, inspect the init container logs:

```bash
kubectl -n linapro-water logs job/linapro-water-init-database
```

To manually rerun the host database initialization after changing `config.yaml`, delete the completed Job, apply the manifest again, and restart the LinaPro pods:

```bash
kubectl -n linapro-water delete job linapro-water-init-database
kubectl apply -f linapro-k8s.yaml
kubectl -n linapro-water rollout restart deploy/linapro-water
```

After changing `mediaStrategy` values, update the `linapro-water-plugin-config` Secret and restart the deployment:

```bash
kubectl -n linapro-water rollout restart deploy/linapro-water
```

## Troubleshooting

| Symptom | Cause | Fix |
| ------- | ----- | --- |
| `SetServerRoot failed: cannot find "resource/public"` | The selected `water` image does not contain `/app/resource/public`, but the manifest configures `server.serverRoot: "resource/public"`. | Remove `server.serverRoot` from the `linapro-water-config` `config.yaml`, apply the manifest again, and restart `Deployment/linapro-water`. |
| `runtime config i18n.default cannot be empty` | The selected image requires `i18n.default` in the host runtime config. | Add `i18n.default` and `i18n.locales` to the `linapro-water-config` `config.yaml`. |
| Pods stay at `Init:3/4` for a long time | `verify-watermark-runtime` is checking `/app/lina` for `CGO` and FFmpeg/x264 linkage. This can take tens of seconds with a new image or cold cache. | Inspect `kubectl -n linapro-water logs <pod> -c verify-watermark-runtime` and pod events. If there is no missing-library error, wait for the check to finish. If it reports missing FFmpeg/x264 libraries or a non-`CGO` build, switch to a `water` image built from `../docker/Dockerfile`. |

## Access

When the `linapro-water` pod is ready, access LinaPro through any Kubernetes node:

```text
http://<node-ip>:30081/admin
```

If the cluster does not expose node ports directly, forward the service locally:

```bash
kubectl -n linapro-water port-forward svc/linapro-water 9120:9120
```

Then open:

```text
http://127.0.0.1:9120/admin
```

Inside the Kubernetes cluster, the Service DNS name is:

```text
http://linapro-water.linapro-water.svc.cluster.local:9120
```

Use `kubectl port-forward` only for smoke checks. Port forwarding a Service can tunnel to a single backend pod, so it is not a valid way to verify multi-pod load balancing or run capacity tests. Use the Service DNS name from inside the cluster, or use `NodePort`, `Ingress`, or `LoadBalancer` for traffic that should be distributed across all replicas.

## Async Queue and Callback

Each pod owns one in-process asynchronous task queue with a capacity of `1024` waiting tasks. Total queue capacity scales with `spec.replicas` only when submit traffic is balanced across pods. If one pod receives submissions faster than its `water.consumerCount` workers can drain them, new submit requests fail with `WATER_TASK_QUEUE_FULL`. Scale `spec.replicas`, `water.consumerCount`, CPU, and memory together, and avoid testing capacity through `kubectl port-forward`.

Asynchronous task status is stored in the shared host KV cache, so status can be queried from any pod. The status response intentionally leaves `image` empty to keep cache values bounded; the full output image is delivered only through the callback.

The submit API accepts `callbackUrl`. When a callback URL is present, `water` sends a `POST` request with `Content-Type: application/json` and this payload shape:

```json
{
  "error_code": "",
  "deviceCode": "device-a",
  "channelCode": "channel-a",
  "deviceIdx": "1",
  "image": "data:image/png;base64,...",
  "imageName": "snap.png",
  "imagePath": "/tmp/snap.png",
  "accessNode": "",
  "acceptNode": "",
  "uploadUrl": ""
}
```

The receiver must return a `2xx` response within the callback timeout. If the receiver expects fields such as `image_base64` or `image_url`, adapt the receiver or add a small callback adapter; the current callback contract uses `image`.

## Update Image

Edit the image field in `linapro-k8s.yaml`:

```yaml
image: ghcr.io/wangle201210/linapro-water:nightly-20260616
```

Use a `water` image built from `apps/lina-plugins/water/deploy/docker/Dockerfile`; do not replace it with the generic `linapro` image unless that image was also built with `CGO` and FFmpeg/x264 runtime libraries.

Then apply the manifest again:

```bash
kubectl apply -f linapro-k8s.yaml
```

## Scale Notes

Watermark processing is CPU and memory intensive. This manifest starts with ten pods and two consumers per pod, so the default processing width is twenty consumers. Tune `spec.replicas`, `resources.requests`, `resources.limits`, and `water.consumerCount` together.

Ingress or Service routing can distribute requests across the ten pods. The host KV cache uses the bundled Redis coordination backend in cluster mode, so asynchronous task status can be read from any pod. The in-process task queue remains local to the pod that accepted the submit request, which is expected for this deployment shape.

The manifest defaults `logger.level` to `info`. With `cluster.enabled: true`, setting the level to `all` prints debug election messages from follower pods, including `[cluster] not leader, waiting for lease expiry`, on every renew interval.

## Uninstall

Delete all resources created by this manifest:

```bash
kubectl delete -f linapro-k8s.yaml
```
