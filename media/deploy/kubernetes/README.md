# LinaPro Kubernetes Deployment

This directory provides single-file Kubernetes deployment manifests for the `nightly-20260616` LinaPro image with the `media` plugin included. The LinaPro `Deployment` runs `3` replicas by default. The manifests also deploy Nacos because the `media` collection service uses Nacos-backed discovery by default.

## Files

| File | Purpose |
| ---- | ------- |
| `linapro-k8s.yaml` | All-in-one deployment that creates PostgreSQL and Nacos inside the cluster. |
| `linapro-k8s-external-pgsql.yaml` | Deployment that uses an existing PostgreSQL service and creates Nacos inside the cluster. |
| `README.zh-CN.md` | Chinese usage guide with the same deployment steps. |

## Prerequisites

- A Kubernetes cluster with a default `StorageClass`.
- A `StorageClass` that supports `ReadWriteMany` for the shared LinaPro data volume used by `3` replicas.
- `kubectl` configured for the target cluster.
- The node can pull `ghcr.io/wangle201210/linapro:nightly-20260616`.
- The node can pull `nacos/nacos-server:v2.3.2`.
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
| `logger.level` | `info` | Runtime log level. Keep `info` for multi-replica deployments; use `all` only for short-lived debugging. |
| `i18n.default` | `zh-CN` | Required host runtime default locale. Do not remove this block. |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: true` | Runtime config for the `media` source plugin. The default collection TCP port is `1911`, and discovery points to `linapro-nacos:8848`. |
| `PersistentVolumeClaim/linapro-nacos-data.resources.requests.storage` | `5Gi` | Nacos standalone data volume capacity. |
| `resources.requests.storage` | `20Gi` | Storage size for PostgreSQL and LinaPro data. |
| `plugin.autoEnable[0].withMockData` | `false` | Whether to load `media` mock demo data during startup auto-install. Keep `false` for production. |
| `spec.replicas` | `3` | LinaPro application replica count. |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | unset | Optional. Set this when the cluster default `StorageClass` does not support `ReadWriteMany`. |

When reusing an existing `linapro-postgres-data` PVC, keep `POSTGRES_PASSWORD` and `database.default.link` aligned with the password already stored in the PostgreSQL data directory. Updating the Kubernetes `Secret` does not rotate the password of an already initialized PostgreSQL user. For a clean redeploy with a new password, delete the old PostgreSQL PVC or change the database user password before starting new LinaPro pods.

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
| `logger.level` | `info` | Runtime log level. Keep `info` for multi-replica deployments; use `all` only for short-lived debugging. |
| `i18n.default` | `zh-CN` | Required host runtime default locale. Do not remove this block. |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: true` | Runtime config for the `media` source plugin. The default collection TCP port is `1911`, and discovery points to `linapro-nacos:8848`. |
| `PersistentVolumeClaim/linapro-nacos-data.resources.requests.storage` | `5Gi` | Nacos standalone data volume capacity. |
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

The bundled `Service/linapro-nacos` exposes Nacos ports `8848`, `9848`, `9849`, and `7848` inside the cluster. The `media` plugin only needs `collectionServer.discovery.host: "linapro-nacos"` and `collectionServer.discovery.port: 8848`; keep the other Nacos service ports available for Nacos 2.x internal protocols.

For temporary validation of the Nacos console from outside the Kubernetes cluster, the manifest also exposes `Service/linapro-nacos-external` on node port `30088`:

```text
http://<node-ip>:30088/nacos
```

For example:

```text
http://10.157.225.174:30088/nacos
```

Production environments should restrict this management port with firewall rules, VPN access, or an internal-only ingress policy.

Keep these constraints when editing the manifests:

- Do not add `server.serverRoot: "resource/public"` for `nightly-20260616`. The image does not contain `/app/resource/public`, and LinaPro fails during startup if that path is configured.
- Keep the `i18n` block in `config.yaml`. This image requires `i18n.default` at runtime.
- Keep the plugin config mount at `/app/config`. The source plugin config loader resolves production plugin config files from `/app/config/plugins/<plugin-id>/config.yaml` in this image.
- Keep `collectionServer.discovery.host` aligned with `Service/linapro-nacos`. If you replace the bundled Nacos with an external Nacos service, update `collectionServer.discovery.host`, `collectionServer.discovery.port`, `collectionServer.discovery.namespace`, `collectionServer.discovery.username`, and `collectionServer.discovery.password`.

## Multi-Replica Runtime

Both manifests are configured for `3` LinaPro replicas:

| Resource | Setting | Purpose |
| -------- | ------- | ------- |
| `Deployment/linapro` | `replicas: 3` | Runs three LinaPro application pods. |
| `cluster.enabled` | `true` | Enables multi-node runtime coordination. |
| `Deployment/linapro-redis` | `replicas: 1` | Provides Redis coordination for election, distributed locks, and cross-instance runtime consistency. |
| `Deployment/linapro-nacos` | `replicas: 1` | Provides Nacos standalone discovery for the `media` collection server. |
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

After the server starts, `plugin.autoEnable` automatically installs and enables the `media` source plugin. The plugin install phase executes the `media` install SQL. Mock data is not loaded unless `withMockData` is changed to `true`. The `media` plugin starts its collection TCP server on port `1911` and uses `linapro-nacos:8848` for discovery.

Check workload status:

```bash
kubectl -n linapro get pods
kubectl -n linapro get job linapro-db-init
kubectl -n linapro get svc
kubectl -n linapro logs deploy/linapro-nacos
```

Follow LinaPro logs:

```bash
kubectl -n linapro logs -f -l app=linapro -c linapro --max-log-requests=3
```

`kubectl logs deploy/linapro -f` follows a deployment-selected pod and can miss activity from the other replicas.

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
| `PersistentVolumeClaim/linapro-data` stays `Pending` | The default `StorageClass` does not support `ReadWriteMany` or no default storage class exists. | Set `PersistentVolumeClaim/linapro-data.spec.storageClassName` to an RWX-capable storage class. |
| `kubectl` reports `Forbidden` for namespace resources | The active kubeconfig user does not have enough rights. | Use a kubeconfig with cluster-admin permissions, for example `/etc/kubernetes/admin.conf` on many self-managed nodes. |
| Port `8082` is required | Kubernetes `NodePort` normally uses the `30000-32767` range, so the manifest keeps the service as `ClusterIP`. | Use `kubectl port-forward`, `Ingress`, or `LoadBalancer` to expose `8082`. |
| `wait-for-database-init` keeps reporting `password authentication failed for user "postgres"` after reapplying the manifest | The existing PostgreSQL PVC was initialized with a different password. Applying a new `Secret` only changes future environment values and does not update the password inside the existing PostgreSQL data directory. | Keep `POSTGRES_PASSWORD` and `database.default.link` consistent with the existing database password, change the database user password manually, or delete `PersistentVolumeClaim/linapro-postgres-data` for a clean redeploy. |
| Repeated `[cluster] not leader, waiting for lease expiry` logs | `cluster.enabled: true` is running with debug-level logging enabled. Follower pods log election wait messages on every renew interval. | Keep the manifest default `logger.level: "info"` for normal multi-replica deployments. |
| Some replicas briefly log `Startup auto-enable plugin media timed out` and restart once during the first multi-replica startup | Multiple `linapro` replicas start concurrently. One replica may finish the `media` plugin auto-install and enable flow first, while the others time out during their first wait window. | Wait for `Deployment/linapro` to reach `3/3`, then check `sys_plugin` for the `media` plugin `installed`, `status`, `desired_state`, and `current_state` values. If it is installed and both `desired_state` and `current_state` are `enabled`, and later logs show no continuous errors, no action is required. |
| `media collection discovery` requests fail | The collection server cannot reach Nacos, or the Nacos namespace or credentials do not match the deployed service. | Check `Deployment/linapro-nacos`, `Service/linapro-nacos`, and `collectionServer.discovery.*` in `linapro-source-plugin-configs`. |

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

Inside the Kubernetes cluster, the HTTP Service DNS name is:

```text
http://linapro.linapro.svc.cluster.local:9120
```

The collection TCP Service endpoint is:

```text
linapro.linapro.svc.cluster.local:1911
```

For services outside the Kubernetes cluster, use the dedicated collection
`NodePort` service:

```text
<node-ip>:30091
```

For example, if the node IP is `10.157.225.174`, external media collectors should
connect to:

```text
10.157.225.174:30091
```

Use `kubectl port-forward` only for smoke checks. Port forwarding a Service can tunnel to a single backend pod, so it is not a valid way to verify multi-pod load balancing. Use the Service DNS name from inside the cluster, or use `Ingress`, `LoadBalancer`, or `NodePort` for traffic that should be distributed across all replicas.

The manifest keeps the HTTP service as `ClusterIP` and exposes the collection TCP
port through `Service/linapro-collection-external`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: linapro-collection-external
  namespace: linapro
spec:
  type: NodePort
  selector:
    app: linapro
  ports:
    - name: collection
      port: 1911
      targetPort: 1911
      nodePort: 30091
```

Discovery registration uses persistent Nacos instance records, so
`register`/`deregister`/`lookup` requests from one collector do not require
Kubernetes `sessionAffinity` and may be routed to different `linapro` replicas.

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
