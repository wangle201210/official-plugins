# LinaPro Kubernetes Deployment

This directory provides a single-file Kubernetes deployment manifest for the `nightly-20260616` LinaPro image with the `media` plugin included.

## Files

| File | Purpose |
| ---- | ------- |
| `linapro-k8s.yaml` | All-in-one deployment that creates PostgreSQL inside the cluster. |
| `linapro-k8s-external-pgsql.yaml` | Deployment that uses an existing PostgreSQL service and does not create PostgreSQL resources. |
| `README.zh-CN.md` | Chinese usage guide with the same deployment steps. |

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- `kubectl` configured for the target cluster.
- The node can pull `ghcr.io/wangle201210/linapro:nightly-20260616`.
- Port `30080` is available when using the bundled `NodePort` service.

## Choose a Manifest

Use one of these manifests:

| Scenario | Manifest |
| -------- | -------- |
| Need Kubernetes to create PostgreSQL | `linapro-k8s.yaml` |
| Already have PostgreSQL | `linapro-k8s-external-pgsql.yaml` |

## Configure All-in-One PostgreSQL

Before applying the manifest, edit `linapro-k8s.yaml` and replace these default values:

| Field | Default | Description |
| ----- | ------- | ----------- |
| `POSTGRES_PASSWORD` | `linapro-change-me` | PostgreSQL password. |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(linapro-postgres:5432)/linapro?sslmode=disable` | LinaPro database connection string. Keep it aligned with `POSTGRES_PASSWORD`. |
| `auth.jwt.secret` | `linapro-jwt-change-me` | JWT signing secret. |
| `spec.ports[0].nodePort` | `30080` | External access port for `NodePort`. |
| `resources.requests.storage` | `20Gi` | Storage size for PostgreSQL and LinaPro data. |
| `plugin.autoEnable[0].withMockData` | `false` | Whether to load `media` mock demo data during startup auto-install. Keep `false` for production. |

## Configure External PostgreSQL

When using `linapro-k8s-external-pgsql.yaml`, first update the external database placeholders:

| Field | Default | Description |
| ----- | ------- | ----------- |
| `PGSQL_HOST` | `pgsql.example.internal` | Existing PostgreSQL host or Kubernetes service name. |
| `PGSQL_PORT` | `5432` | PostgreSQL port. |
| `PGSQL_USER` | `postgres` | PostgreSQL user used by LinaPro. |
| `PGSQL_PASSWORD` | `linapro-change-me` | PostgreSQL password. |
| `PGSQL_DATABASE` | `linapro` | Database name used by LinaPro. |
| `PGSQL_SSLMODE` | `disable` | SSL mode expected by the PostgreSQL endpoint. |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(pgsql.example.internal:5432)/linapro?sslmode=disable` | LinaPro database connection string. Keep it aligned with the external PostgreSQL values. |
| `auth.jwt.secret` | `linapro-jwt-change-me` | JWT signing secret. |

The `init-database` init container runs `./lina init --confirm=init`, so the configured PostgreSQL account must be able to create or update tables, indexes, comments, and seed data in the target database.

## Deploy

Apply the manifest:

```bash
kubectl apply -f linapro-k8s.yaml
```

For an existing PostgreSQL deployment, use:

```bash
kubectl apply -f linapro-k8s-external-pgsql.yaml
```

The `linapro` pod runs two init containers before starting the server:

| Init container | Purpose |
| -------------- | ------- |
| `wait-for-postgres` or `wait-for-pgsql` | Waits until PostgreSQL accepts connections. |
| `init-database` | Runs `./lina init --confirm=init` with the mounted `/app/config.yaml`. This creates or upgrades the host schema and required seed data. |

After the server starts, `plugin.autoEnable` automatically installs and enables the `media` source plugin. The plugin install phase executes the `media` install SQL. Mock data is not loaded unless `withMockData` is changed to `true`.

Check workload status:

```bash
kubectl -n linapro get pods
kubectl -n linapro get svc
```

Follow LinaPro logs:

```bash
kubectl -n linapro logs deploy/linapro -f
```

If database initialization fails, inspect the init container logs:

```bash
kubectl -n linapro logs deploy/linapro -c init-database
```

To manually rerun the host database initialization after changing `config.yaml`, restart the LinaPro pod so the init container runs again:

```bash
kubectl -n linapro rollout restart deploy/linapro
```

## Access

When the `linapro` pod is ready, access LinaPro through any Kubernetes node:

```text
http://<node-ip>:30080/admin
```

If the cluster does not expose node ports directly, forward the service locally:

```bash
kubectl -n linapro port-forward svc/linapro 9120:9120
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

## Uninstall

Delete all resources created by this manifest:

```bash
kubectl delete -f linapro-k8s.yaml
```

For the external PostgreSQL deployment, use:

```bash
kubectl delete -f linapro-k8s-external-pgsql.yaml
```

Deleting the external PostgreSQL manifest removes only LinaPro Kubernetes resources. It does not delete the external PostgreSQL instance or database.
