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
- Port `8082` is available on the node if using the documented `port-forward` access method.

If `kubectl` is already configured but reports cluster internal errors or `Forbidden`
responses for `Secret`, `PersistentVolumeClaim`, `Deployment`, or `StatefulSet`,
switch to a kubeconfig with cluster-admin rights before applying the manifest. On
self-managed control-plane nodes this is often:

```bash
export KUBECONFIG=/etc/kubernetes/admin.conf
```

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
| `i18n.default` | `zh-CN` | Required host runtime default locale. Do not remove this block. |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: false` | Runtime config for the `media` source plugin. |
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
| `i18n.default` | `zh-CN` | Required host runtime default locale. Do not remove this block. |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: false` | Runtime config for the `media` source plugin. |
| `spec.replicas` | `3` | LinaPro application replica count. |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | unset | Optional. Set this when the cluster default `StorageClass` does not support `ReadWriteMany`. |

The `linapro-db-init` `Job` runs `./lina init --confirm=init`. The configured PostgreSQL account must be able to connect to the PostgreSQL maintenance database `postgres`, check whether the target database exists, create the target database when it is missing, and create or update tables, indexes, comments, and seed data in the target database.

## Runtime Configuration Notes

The manifests intentionally keep the host `/app/config.yaml` and the source
plugin runtime configs as separate `Secret` resources:

| Resource | Mounted path | Purpose |
| -------- | ------------ | ------- |
| `Secret/linapro-config` | `/app/config.yaml` | Host LinaPro runtime config. |
| `Secret/linapro-source-plugin-configs` | `/app/config/plugins/media/config.yaml` | `media` source plugin runtime config. |

Keep these constraints when editing the manifests:

- Do not add `server.serverRoot: "resource/public"` for `nightly-20260616`. The image does not contain `/app/resource/public`, and LinaPro fails during startup if that path is configured.
- Keep the `i18n` block in `config.yaml`. This image requires `i18n.default` at runtime.
- Keep the plugin config mount at `/app/config`. The source plugin config loader resolves production plugin config files from `/app/config/plugins/<plugin-id>/config.yaml` in this image.
- No `sicau-niu` runtime config is required when the image includes the missing `token.secret` warning-only fix. The bundled `sicau-niu` source plugin will log a warning when `token.secret` is absent, but it will not block `media` startup.

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

## Troubleshooting

These issues were found during validation of earlier `nightly-20260616`
deployments. The current manifests and the new image fix cover them:

| Symptom | Cause | Fix |
| ------- | ----- | --- |
| `SetServerRoot failed: cannot find "resource/public"` | `server.serverRoot` pointed to a directory not present in the image. | Leave `server.serverRoot` unset. |
| `runtime config i18n.default cannot be empty` | Host `i18n` runtime config was missing. | Keep the included `i18n.default` and `i18n.locales` block. |
| `Player token secret is not configured` | Older images treated missing `sicau-niu` `token.secret` as a startup error. | Use an image that includes the warning-only fix. If an older image must be used, temporarily mount `/app/config/plugins/sicau-niu/config.yaml` with a production `token.secret`. |
| `PersistentVolumeClaim/linapro-data` stays `Pending` | The default `StorageClass` does not support `ReadWriteMany` or no default storage class exists. | Set `PersistentVolumeClaim/linapro-data.spec.storageClassName` to an RWX-capable storage class. |
| `kubectl` reports `Forbidden` for namespace resources | The active kubeconfig user does not have enough rights. | Use a kubeconfig with cluster-admin permissions, for example `/etc/kubernetes/admin.conf` on many self-managed nodes. |
| Port `8082` is required | Kubernetes `NodePort` normally uses the `30000-32767` range, so the manifest keeps the service as `ClusterIP`. | Use `kubectl port-forward`, `Ingress`, or `LoadBalancer` to expose `8082`. |

To manually rerun the host database initialization after changing `config.yaml`, delete the completed `Job` and apply the manifest again:

```bash
kubectl -n linapro delete job linapro-db-init
kubectl apply -f <selected-manifest>.yaml
```

## Access

When the `linapro` pod is ready, forward the service to node port `8082`:

```bash
kubectl -n linapro port-forward --address 0.0.0.0 svc/linapro 8082:9120
```

Then open:

```text
http://<node-ip>:8082/admin
```

For a local-only check from the machine running `kubectl`, use:

```bash
kubectl -n linapro port-forward svc/linapro 9120:9120
```

Then open:

```text
http://127.0.0.1:9120/admin
```

For production traffic, prefer `Ingress` or `LoadBalancer` instead of a manual
`port-forward` process.

To use `NodePort` instead, change `Service/linapro.spec.type` to `NodePort` and
set a port in the cluster's node-port range, for example:

```yaml
spec:
  type: NodePort
  ports:
    - name: http
      port: 9120
      targetPort: 9120
      nodePort: 30080
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
