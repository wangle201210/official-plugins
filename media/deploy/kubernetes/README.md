# LinaPro Kubernetes Deployment

This directory provides single-file Kubernetes deployment manifests for the `nightly-20260616` LinaPro image with the `media` plugin included. The LinaPro `Deployment` runs `3` replicas by default.

## Files

| File | Purpose |
| ---- | ------- |
| `linapro-k8s.yaml` | All-in-one deployment that creates PostgreSQL inside the cluster. |
| `linapro-k8s-external-pgsql.yaml` | Deployment that uses an existing PostgreSQL service and does not create PostgreSQL resources. |
| `README.zh-CN.md` | Chinese usage guide with the same deployment steps. |

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- A `StorageClass` that supports `ReadWriteMany` for the shared LinaPro data volume used by `3` replicas.
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
| `jwt.secret` | `linapro-jwt-change-me` | JWT signing secret. |
| `jwt.expire` | `24h` | JWT token validity duration. |
| `spec.ports[0].nodePort` | `30080` | External access port for `NodePort`. |
| `resources.requests.storage` | `20Gi` | Storage size for PostgreSQL and LinaPro data. |
| `plugin.autoEnable[0].withMockData` | `false` | Whether to load `media` mock demo data during startup auto-install. Keep `false` for production. |
| `spec.replicas` | `3` | LinaPro application replica count. |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | unset | Optional. Set this when the cluster default `StorageClass` does not support `ReadWriteMany`. |

## Configure External PostgreSQL

When using `linapro-k8s-external-pgsql.yaml`, first update the external database placeholders:

| Field | Default | Description |
| ----- | ------- | ----------- |
| `PGSQL_HOST` | `pgsql.example.internal` | Existing PostgreSQL host or Kubernetes service name. |
| `PGSQL_PORT` | `5432` | PostgreSQL port. |
| `PGSQL_USER` | `postgres` | PostgreSQL user used by LinaPro. |
| `PGSQL_PASSWORD` | `linapro-change-me` | PostgreSQL password. |
| `PGSQL_DATABASE` | `linapro` | Database name used by LinaPro. |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(pgsql.example.internal:5432)/linapro?sslmode=disable` | LinaPro database connection string. Keep it aligned with the external PostgreSQL values. |
| `jwt.secret` | `linapro-jwt-change-me` | JWT signing secret. |
| `jwt.expire` | `24h` | JWT token validity duration. |
| `spec.replicas` | `3` | LinaPro application replica count. |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | unset | Optional. Set this when the cluster default `StorageClass` does not support `ReadWriteMany`. |

The `linapro-db-init` `Job` runs `./lina init --confirm=init`. The configured PostgreSQL account must be able to connect to the PostgreSQL maintenance database `postgres`, check whether the target database exists, create the target database when it is missing, and create or update tables, indexes, comments, and seed data in the target database.

## Multi-Replica Runtime

Both manifests are configured for `3` LinaPro replicas:

| Resource | Setting | Purpose |
| -------- | ------- | ------- |
| `Deployment/linapro` | `replicas: 3` | Runs three LinaPro application pods. |
| `cluster.enabled` | `true` | Enables multi-node runtime coordination. |
| `Deployment/linapro-redis` | `replicas: 1` | Provides Redis coordination for election, distributed locks, and cross-instance runtime consistency. |
| `PersistentVolumeClaim/linapro-data` | `ReadWriteMany` | Shares upload and plugin runtime data across the three LinaPro pods. |

If the target cluster does not provide a default `ReadWriteMany` storage class, set `PersistentVolumeClaim/linapro-data.spec.storageClassName` to a shared-mount storage class before applying the manifest:

```yaml
spec:
  storageClassName: nfs-rwx
  accessModes:
    - ReadWriteMany
```

Use the actual RWX-capable storage class name from your cluster, such as NFS, CephFS, or EFS. Without an RWX-capable storage class, `PersistentVolumeClaim/linapro-data` can remain `Pending`.

## Deploy

Apply the manifest:

```bash
kubectl apply -f linapro-k8s.yaml
```

For an existing PostgreSQL deployment, use:

```bash
kubectl apply -f linapro-k8s-external-pgsql.yaml
```

The manifest creates one database initialization `Job`:

| Resource | Purpose |
| -------- | ------- |
| `Job/linapro-db-init` | Runs `./lina init --confirm=init` once with the mounted `/app/config.yaml`. This creates or upgrades the host schema and required seed data. |

The `linapro` pods then wait for PostgreSQL, Redis, and the initialized host schema before starting the server. The startup wait checks the required plugin and cache tables, including `sys_plugin`, `sys_kv_cache`, and `sys_cache_revision`, to avoid starting before the database initialization `Job` has finished the host runtime schema.

After the server starts, `plugin.autoEnable` automatically installs and enables the `media` source plugin. The plugin install phase executes the `media` install SQL. Mock data is not loaded unless `withMockData` is changed to `true`.

Check workload status:

```bash
kubectl -n linapro get pods
kubectl -n linapro get job linapro-db-init
kubectl -n linapro get svc
```

Follow LinaPro logs:

```bash
kubectl -n linapro logs deploy/linapro -f
```

If database initialization fails, inspect the initialization `Job` logs:

```bash
kubectl -n linapro logs job/linapro-db-init
```

To manually rerun the host database initialization after changing `config.yaml`, delete the completed `Job` and apply the manifest again:

```bash
kubectl -n linapro delete job linapro-db-init
kubectl apply -f <selected-manifest>.yaml
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
