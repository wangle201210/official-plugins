# LinaPro Water Kubernetes Deployment

This directory provides a single-file Kubernetes deployment manifest for the `nightly-20260616` LinaPro image with only the `water` source plugin auto-enabled.

The manifest deploys ten `linapro-water` pods behind one Kubernetes Service to spread CPU and memory pressure from watermark processing. It does not deploy the `media` plugin. Configure `mediaStrategy.baseUrl` to point to the LinaPro cluster where `media` is installed and where the `mediaopen` internal API is reachable.

## Files

| File | Purpose |
| ---- | ------- |
| `linapro-k8s.yaml` | All-in-one `water` deployment that creates PostgreSQL, Redis coordination, one database init Job, and ten LinaPro `water` pods inside the cluster. |
| `README.zh-CN.md` | Chinese usage guide with the same deployment steps. |

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- `kubectl` configured for the target cluster.
- The node can pull `ghcr.io/wangle201210/linapro:nightly-20260616`.
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
| `resources.requests.storage` | `20Gi` | Storage size for the bundled PostgreSQL data volume. |

The `water` plugin reads its plugin-scoped runtime config from `/app/config/plugins/water/config.yaml`. The manifest mounts this file from the `linapro-water-plugin-config` Secret. The template `manifest/config/config.example.yaml` is documentation only and is not read as a runtime default.

For cross-cluster access, set `mediaStrategy.baseUrl` to the `media` cluster `Ingress`, load balancer, VPN address, or `NodePort` URL. The configured `mediaStrategy.apiKey` must match the `media` cluster `innerapi.apiKey` value.

The bundled PostgreSQL container sets `PGDATA` to `/var/lib/postgresql/data/pgdata` so volume-root metadata from some storage classes does not interfere with first-time database initialization.

The `linapro-water` Deployment mounts `/app/data` as `emptyDir` so ten pods can start without competing for one `ReadWriteOnce` volume. This is suitable for the current `water` processing path, where watermark results are returned as data URLs or sent by callback. If you later use shared uploaded files, dynamic plugin artifacts, or persistent local files, replace `emptyDir` with a storage backend that supports concurrent multi-pod access, such as a `ReadWriteMany` volume or object storage integration.

## Deploy

Apply the manifest:

```bash
kubectl apply -f linapro-k8s.yaml
```

The manifest creates a one-shot database initialization Job:

| Job | Purpose |
| --- | ------- |
| `linapro-water-init-database` | Runs `./lina init --confirm=init` once with the mounted `/app/config.yaml`. This creates or upgrades the host schema and required seed data. |

Each `linapro-water` pod runs three init containers before starting the server:

| Init container | Purpose |
| -------------- | ------- |
| `wait-for-postgres` | Waits until PostgreSQL accepts connections. |
| `wait-for-redis` | Waits until Redis accepts connections for cluster coordination. |
| `wait-for-database-init` | Waits until the init Job has created the host schema needed by server startup. |

After the server starts, `plugin.autoEnable` automatically installs and enables the `water` source plugin. The plugin install phase executes the `water` install SQL. The enabled `water` plugin then uses the mounted plugin config to call the remote `mediaopen` strategy resolver.

Check workload status:

```bash
kubectl -n linapro-water get pods
kubectl -n linapro-water get job
kubectl -n linapro-water get svc
```

Follow LinaPro logs:

```bash
kubectl -n linapro-water logs deploy/linapro-water -f
```

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

## Update Image

Edit the image field in `linapro-k8s.yaml`:

```yaml
image: ghcr.io/wangle201210/linapro:nightly-20260616
```

Then apply the manifest again:

```bash
kubectl apply -f linapro-k8s.yaml
```

## Scale Notes

Watermark processing is CPU and memory intensive. This manifest starts with ten pods and two consumers per pod, so the default processing width is twenty consumers. Tune `spec.replicas`, `resources.requests`, `resources.limits`, and `water.consumerCount` together.

Ingress or Service routing can distribute requests across the ten pods. The host KV cache uses the bundled Redis coordination backend in cluster mode, so asynchronous task status can be read from any pod. The in-process task queue remains local to the pod that accepted the submit request, which is expected for this deployment shape.

## Uninstall

Delete all resources created by this manifest:

```bash
kubectl delete -f linapro-k8s.yaml
```
