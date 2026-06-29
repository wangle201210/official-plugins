# LinaPro Kubernetes 部署说明

本目录提供单文件 Kubernetes 部署清单，用于部署已包含`media`插件的`nightly-20260616`版 LinaPro 镜像。LinaPro `Deployment`默认运行`3`个副本。清单会同时部署 Nacos，因为`media`采集服务默认使用基于 Nacos 的发现能力。

## 文件说明

| 文件 | 用途 |
| ---- | ---- |
| `linapro-k8s.yaml` | 一体化部署清单，会在集群内创建 PostgreSQL 和 Nacos。 |
| `linapro-k8s-external-pgsql.yaml` | 使用已有 PostgreSQL 的部署清单，不创建 PostgreSQL 资源，会在集群内创建 Nacos。 |
| `README.md` | 英文使用说明，部署步骤与本文档保持一致。 |

## 前置条件

- Kubernetes 集群已配置默认`StorageClass`。
- 集群存在支持`ReadWriteMany`的`StorageClass`，用于`3`个副本共享 LinaPro 数据卷。
- `kubectl`已经连接到目标集群。
- 节点可以拉取`ghcr.io/wangle201210/linapro:nightly-20260616`。
- 节点可以拉取`nacos/nacos-server:v2.3.2`。
- 使用文档中的`port-forward`访问方式时，节点端口`8082`未被占用。

如果`kubectl`已配置但访问集群时报内部错误，或读取`Secret`、`PersistentVolumeClaim`、`Deployment`、`StatefulSet`时报`Forbidden`，请先切换到具备集群管理员权限的`kubeconfig`。自建控制平面节点上通常可以使用：

```bash
export KUBECONFIG=/etc/kubernetes/admin.conf
```

## 选择清单

按部署场景选择其中一份清单：

| 场景 | 清单 |
| ---- | ---- |
| 需要 Kubernetes 创建 PostgreSQL | `linapro-k8s.yaml` |
| 已经有 PostgreSQL | `linapro-k8s-external-pgsql.yaml` |

## 配置内置 PostgreSQL 版本

应用清单前，先编辑`linapro-k8s.yaml`并替换以下默认值：

| 字段 | 默认值 | 说明 |
| ---- | ------ | ---- |
| `POSTGRES_PASSWORD` | `linapro-change-me` | PostgreSQL 密码。 |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(linapro-postgres:5432)/linapro?sslmode=disable` | LinaPro 数据库连接串，需要与`POSTGRES_PASSWORD`保持一致。 |
| `jwt.secret` | `linapro-jwt-change-me` | JWT 签名密钥。 |
| `jwt.expire` | `24h` | JWT Token 有效期。 |
| `logger.level` | `info` | 运行日志级别。多副本部署保持`info`；仅在短时间排障时改为`all`。 |
| `i18n.default` | `zh-CN` | 宿主运行时必需的默认语言配置，不要删除该配置块。 |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: true` | `media`源码插件运行配置。默认采集 TCP 端口为`1911`，发现服务指向`linapro-nacos:8848`。 |
| `PersistentVolumeClaim/linapro-nacos-data.resources.requests.storage` | `5Gi` | Nacos 单机模式数据卷容量。 |
| `resources.requests.storage` | `20Gi` | PostgreSQL 和 LinaPro 数据卷容量。 |
| `plugin.autoEnable[0].withMockData` | `false` | 启动自动安装`media`插件时是否加载演示数据。生产环境保持`false`。 |
| `spec.replicas` | `3` | LinaPro 应用副本数。 |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | 未设置 | 可选。集群默认`StorageClass`不支持`ReadWriteMany`时需要设置。 |

复用已有`linapro-postgres-data` PVC 时，必须保持`POSTGRES_PASSWORD`和`database.default.link`与 PostgreSQL 数据目录内已有用户密码一致。更新 Kubernetes `Secret`不会自动修改已经初始化过的 PostgreSQL 用户密码。如果需要使用新密码干净重装，请先删除旧的 PostgreSQL PVC，或先手工修改数据库用户密码后再启动新的 LinaPro Pod。

## 配置外部 PostgreSQL 版本

使用`linapro-k8s-external-pgsql.yaml`时，先修改外部数据库占位值：

| 字段 | 默认值 | 说明 |
| ---- | ------ | ---- |
| `PGSQL_HOST` | `pgsql.example.internal` | 已有 PostgreSQL 主机名或 Kubernetes Service 名称。 |
| `PGSQL_PORT` | `5432` | PostgreSQL 端口。 |
| `PGSQL_USER` | `postgres` | LinaPro 使用的 PostgreSQL 账号。 |
| `PGSQL_PASSWORD` | `linapro-change-me` | PostgreSQL 密码。 |
| `PGSQL_DATABASE` | `linapro` | LinaPro 使用的数据库名。 |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(pgsql.example.internal:5432)/linapro?sslmode=disable` | LinaPro 数据库连接串，需要与外部 PostgreSQL 配置保持一致。 |
| `jwt.secret` | `linapro-jwt-change-me` | JWT 签名密钥。 |
| `jwt.expire` | `24h` | JWT Token 有效期。 |
| `logger.level` | `info` | 运行日志级别。多副本部署保持`info`；仅在短时间排障时改为`all`。 |
| `i18n.default` | `zh-CN` | 宿主运行时必需的默认语言配置，不要删除该配置块。 |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: true` | `media`源码插件运行配置。默认采集 TCP 端口为`1911`，发现服务指向`linapro-nacos:8848`。 |
| `PersistentVolumeClaim/linapro-nacos-data.resources.requests.storage` | `5Gi` | Nacos 单机模式数据卷容量。 |
| `spec.replicas` | `3` | LinaPro 应用副本数。 |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | 未设置 | 可选。集群默认`StorageClass`不支持`ReadWriteMany`时需要设置。 |

`linapro-db-init` `Job`会执行`./lina init --confirm=init`。配置的 PostgreSQL 账号必须可以连接 PostgreSQL 维护库`postgres`，检查目标数据库是否存在，在目标数据库不存在时创建数据库，并在目标数据库中创建或更新表、索引、注释和 Seed 数据。

## 运行配置说明

清单有意把宿主`/app/config.yaml`和源码插件运行配置拆成不同`Secret`资源：

| 资源 | 挂载路径 | 用途 |
| ---- | -------- | ---- |
| `Secret/linapro-config` | `/app/config.yaml` | LinaPro 宿主运行配置。 |
| `Secret/linapro-source-plugin-configs` | `/app/config/plugins/media/config.yaml` | `media`源码插件运行配置。 |

内置`Service/linapro-nacos`会在集群内暴露 Nacos 的`8848`、`9848`、`9849`和`7848`端口。`media`插件只需要配置`collectionServer.discovery.host: "linapro-nacos"`和`collectionServer.discovery.port: 8848`；其他 Nacos 服务端口用于 Nacos 2.x 内部协议，请保持可用。

为了便于从 Kubernetes 集群外临时验证 Nacos 控制台，清单也会通过`Service/linapro-nacos-external`把管理端暴露到节点端口`30088`：

```text
http://<node-ip>:30088/nacos
```

例如：

```text
http://10.157.225.174:30088/nacos
```

生产环境应通过防火墙、VPN 或仅内网访问的入口策略限制该管理端口。

修改清单时请保留以下约束：

- 不要为`nightly-20260616`添加`server.serverRoot: "resource/public"`。该镜像内没有`/app/resource/public`，配置该路径会导致 LinaPro 启动失败。
- 保留`config.yaml`中的`i18n`配置块。该镜像运行时要求`i18n.default`非空。
- 保留插件配置挂载路径`/app/config`。该镜像的源码插件配置加载器会从`/app/config/plugins/<plugin-id>/config.yaml`读取生产运行配置。
- 保持`collectionServer.discovery.host`与`Service/linapro-nacos`一致。如果改用外部 Nacos，需要同时更新`collectionServer.discovery.host`、`collectionServer.discovery.port`、`collectionServer.discovery.namespace`、`collectionServer.discovery.username`和`collectionServer.discovery.password`。

## 多副本运行配置

两份清单都按`3`个 LinaPro 副本配置：

| 资源 | 配置 | 用途 |
| ---- | ---- | ---- |
| `Deployment/linapro` | `replicas: 3` | 运行 3 个 LinaPro 应用 Pod。 |
| `cluster.enabled` | `true` | 启用多节点运行时协调。 |
| `Deployment/linapro-redis` | `replicas: 1` | 提供 Redis 协调能力，用于选主、分布式锁和跨实例运行时一致性。 |
| `Deployment/linapro-nacos` | `replicas: 1` | 为`media`采集服务提供 Nacos 单机模式发现能力。 |
| `PersistentVolumeClaim/linapro-data` | `ReadWriteMany` | 让 3 个 LinaPro Pod 共享上传文件和插件运行时数据。 |

如果目标集群没有默认支持`ReadWriteMany`的存储类，需要先把`PersistentVolumeClaim/linapro-data.spec.storageClassName`设置为支持共享挂载的存储类，再应用清单：

```yaml
spec:
  storageClassName: nfs-rwx
  accessModes:
    - ReadWriteMany
```

请使用目标集群里真实存在且支持`ReadWriteMany`的存储类，例如 NFS、CephFS 或 EFS。否则`PersistentVolumeClaim/linapro-data`可能会一直处于`Pending`状态。

## 部署

执行以下命令应用清单：

```bash
kubectl apply -f linapro-k8s.yaml
```

如果使用已有 PostgreSQL，执行：

```bash
kubectl apply -f linapro-k8s-external-pgsql.yaml
```

清单会创建一个数据库初始化`Job`：

| 资源 | 用途 |
| ---- | ---- |
| `Job/linapro-db-init` | 使用挂载的`/app/config.yaml`执行一次`./lina init --confirm=init`，创建或升级宿主表结构和必需 Seed 数据。 |

随后`linapro` Pod 会等待 PostgreSQL、Redis 和已初始化的宿主表结构就绪，再启动服务。启动等待会检查插件和缓存必需表，包括`sys_plugin`、`sys_kv_cache`和`sys_cache_revision`，避免数据库初始化`Job`尚未完成宿主运行时表结构时应用提前启动。

服务启动后，`plugin.autoEnable`会自动安装并启用`media`源码插件。插件安装阶段会执行`media`插件自己的安装 SQL。除非把`withMockData`改成`true`，否则不会加载演示数据。`media`插件会在`1911`端口启动采集 TCP 服务，并使用`linapro-nacos:8848`作为发现服务。

查看工作负载状态：

```bash
kubectl -n linapro get pods
kubectl -n linapro get job linapro-db-init
kubectl -n linapro get svc
kubectl -n linapro logs deploy/linapro-nacos
```

查看所有 LinaPro 副本日志：

```bash
kubectl -n linapro logs -f -l app=linapro -c linapro --max-log-requests=3
```

`kubectl logs deploy/linapro -f`只会跟随 Deployment 选择到的 Pod，可能看不到其他副本日志。

如果数据库初始化失败，可以查看初始化`Job`日志：

```bash
kubectl -n linapro logs job/linapro-db-init
```

## 问题记录与排障

以下问题是在验证早期`nightly-20260616`部署时实际遇到的，当前清单和新镜像修复已经覆盖：

| 现象 | 原因 | 处理方式 |
| ---- | ---- | -------- |
| `SetServerRoot failed: cannot find "resource/public"` | `server.serverRoot`指向了镜像中不存在的目录。 | 保持`server.serverRoot`未配置。 |
| `runtime config i18n.default cannot be empty` | 宿主缺少`i18n`运行配置。 | 保留清单内置的`i18n.default`和`i18n.locales`配置。 |
| `PersistentVolumeClaim/linapro-data`一直处于`Pending` | 默认`StorageClass`不支持`ReadWriteMany`，或集群没有默认存储类。 | 设置`PersistentVolumeClaim/linapro-data.spec.storageClassName`为支持 RWX 的存储类。 |
| `kubectl`访问命名空间资源时报`Forbidden` | 当前`kubeconfig`用户权限不足。 | 使用具备集群管理员权限的`kubeconfig`，很多自建节点可使用`/etc/kubernetes/admin.conf`。 |
| 需要使用端口`8082` | Kubernetes `NodePort`通常使用`30000-32767`端口范围，因此清单默认保持`ClusterIP`。 | 使用`kubectl port-forward`、`Ingress`或`LoadBalancer`暴露`8082`。 |
| 重新应用清单后，`wait-for-database-init`持续输出`password authentication failed for user "postgres"` | 已有 PostgreSQL PVC 是用另一个密码初始化的。重新应用`Secret`只会改变后续环境变量，不会修改已有 PostgreSQL 数据目录内的用户密码。 | 保持`POSTGRES_PASSWORD`和`database.default.link`与已有数据库密码一致，手工修改数据库用户密码，或删除`PersistentVolumeClaim/linapro-postgres-data`后干净重装。 |
| 首次多副本启动时部分副本短暂输出`Startup auto-enable plugin media timed out`并重启一次 | 多个`linapro`副本并发启动，`media`插件自动安装和启用由其中一个副本先完成，其他副本可能在首轮等待窗口内超时。 | 先等待`Deployment/linapro`达到`3/3`，再检查`sys_plugin`中`media`的`installed`、`status`、`desired_state`和`current_state`。如果状态为已安装且`desired_state/current_state`均为`enabled`，且后续日志无持续错误，则无需处理。 |
| `media collection discovery`请求失败 | 采集服务无法访问 Nacos，或 Nacos 命名空间、账号密码与部署服务不一致。 | 检查`Deployment/linapro-nacos`、`Service/linapro-nacos`以及`linapro-source-plugin-configs`中的`collectionServer.discovery.*`配置。 |

修改`config.yaml`后如需手工重跑宿主数据库初始化，删除已完成的`Job`后重新应用清单：

```bash
kubectl -n linapro delete job linapro-db-init
kubectl apply -f <selected-manifest>.yaml
```

## 访问

等待`linapro` Pod 就绪后，把 Service 转发到节点`8082`端口：

```bash
kubectl -n linapro port-forward --address 0.0.0.0 svc/linapro 8082:9120
```

然后访问：

```text
http://<node-ip>:8082/admin
```

如果只需要从运行`kubectl`的机器本地检查，可以使用：

```bash
kubectl -n linapro port-forward svc/linapro 9120:9120
```

然后访问：

```text
http://127.0.0.1:9120/admin
```

生产流量建议使用`Ingress`或`LoadBalancer`，不要依赖手工运行的`port-forward`进程。

在 Kubernetes 集群内部，HTTP Service DNS 地址是：

```text
http://linapro.linapro.svc.cluster.local:9120
```

采集 TCP Service 地址是：

```text
linapro.linapro.svc.cluster.local:1911
```

Kubernetes 集群外部的服务请使用专门的采集`NodePort`服务：

```text
<node-ip>:30091
```

例如节点 IP 是`10.157.225.174`时，外部采集服务连接：

```text
10.157.225.174:30091
```

`kubectl port-forward`只适合做连通性 smoke 检查。对 Service 执行端口转发时，流量可能只进入单个后端 Pod，因此不能用于验证多 Pod 负载均衡。需要分发到所有副本的流量，应从集群内部使用 Service DNS，或通过`Ingress`、`LoadBalancer`、`NodePort`进入。

清单会保持 HTTP 服务为`ClusterIP`，并通过`Service/linapro-collection-external`暴露采集 TCP 端口：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: linapro-collection-external
  namespace: linapro
spec:
  type: NodePort
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800
  selector:
    app: linapro
  ports:
    - name: collection
      port: 1911
      targetPort: 1911
      nodePort: 30091
```

`sessionAffinity: ClientIP`会让同一个采集端的连续 TCP 连接保持在同一个后端 Pod 上。发现注册使用由注册 Pod 的客户端会话维持的 Nacos 临时实例，因此同一个采集端的`register`、`deregister`、`lookup`检查不应被分发到不同副本。

## 更新镜像

编辑`linapro-k8s.yaml`中的镜像字段：

```yaml
image: ghcr.io/wangle201210/linapro:nightly-20260616
```

然后重新应用清单：

```bash
kubectl apply -f linapro-k8s.yaml
```

## 卸载

执行以下命令删除本清单创建的所有资源：

```bash
kubectl delete -f linapro-k8s.yaml
```

外部 PostgreSQL 版本使用：

```bash
kubectl delete -f linapro-k8s-external-pgsql.yaml
```

删除外部 PostgreSQL 版本清单只会删除 LinaPro 相关 Kubernetes 资源，不会删除外部 PostgreSQL 实例或数据库。
