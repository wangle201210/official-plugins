# LinaPro Kubernetes 部署说明

本目录提供单文件 Kubernetes 部署清单，用于部署已包含`media`插件的`nightly-20260616`版 LinaPro 镜像。LinaPro `Deployment`默认运行`3`个副本。

## 文件说明

| 文件 | 用途 |
| ---- | ---- |
| `linapro-k8s.yaml` | 一体化部署清单，会在集群内创建 PostgreSQL。 |
| `linapro-k8s-external-pgsql.yaml` | 使用已有 PostgreSQL 的部署清单，不创建 PostgreSQL 资源。 |
| `README.md` | 英文使用说明，部署步骤与本文档保持一致。 |

## 前置条件

- Kubernetes 集群已配置默认`StorageClass`。
- 集群存在支持`ReadWriteMany`的`StorageClass`，用于`3`个副本共享 LinaPro 数据卷。
- `kubectl`已经连接到目标集群。
- 节点可以拉取`ghcr.io/wangle201210/linapro:nightly-20260616`。
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
| `i18n.default` | `zh-CN` | 宿主运行时必需的默认语言配置，不要删除该配置块。 |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: false` | `media`源码插件运行配置。 |
| `linapro-source-plugin-configs/sicau-niu-config.yaml` | `token.secret: linapro-sicau-niu-token-change-me` | 镜像内置`sicau-niu`源码插件所需的最小运行配置。生产使用前必须替换 token 密钥。 |
| `resources.requests.storage` | `20Gi` | PostgreSQL 和 LinaPro 数据卷容量。 |
| `plugin.autoEnable[0].withMockData` | `false` | 启动自动安装`media`插件时是否加载演示数据。生产环境保持`false`。 |
| `spec.replicas` | `3` | LinaPro 应用副本数。 |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | 未设置 | 可选。集群默认`StorageClass`不支持`ReadWriteMany`时需要设置。 |

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
| `i18n.default` | `zh-CN` | 宿主运行时必需的默认语言配置，不要删除该配置块。 |
| `linapro-source-plugin-configs/media-config.yaml` | `collectionServer.enabled: false` | `media`源码插件运行配置。 |
| `linapro-source-plugin-configs/sicau-niu-config.yaml` | `token.secret: linapro-sicau-niu-token-change-me` | 镜像内置`sicau-niu`源码插件所需的最小运行配置。生产使用前必须替换 token 密钥。 |
| `spec.replicas` | `3` | LinaPro 应用副本数。 |
| `PersistentVolumeClaim/linapro-data.spec.storageClassName` | 未设置 | 可选。集群默认`StorageClass`不支持`ReadWriteMany`时需要设置。 |

`linapro-db-init` `Job`会执行`./lina init --confirm=init`。配置的 PostgreSQL 账号必须可以连接 PostgreSQL 维护库`postgres`，检查目标数据库是否存在，在目标数据库不存在时创建数据库，并在目标数据库中创建或更新表、索引、注释和 Seed 数据。

## 运行配置说明

清单有意把宿主`/app/config.yaml`和源码插件运行配置拆成不同`Secret`资源：

| 资源 | 挂载路径 | 用途 |
| ---- | -------- | ---- |
| `Secret/linapro-config` | `/app/config.yaml` | LinaPro 宿主运行配置。 |
| `Secret/linapro-source-plugin-configs` | `/app/config/plugins/media/config.yaml` | `media`源码插件运行配置。 |
| `Secret/linapro-source-plugin-configs` | `/app/config/plugins/sicau-niu/config.yaml` | 镜像内置`sicau-niu`源码插件的最小运行配置。 |

修改清单时请保留以下约束：

- 不要为`nightly-20260616`添加`server.serverRoot: "resource/public"`。该镜像内没有`/app/resource/public`，配置该路径会导致 LinaPro 启动失败。
- 保留`config.yaml`中的`i18n`配置块。该镜像运行时要求`i18n.default`非空。
- 保留插件配置挂载路径`/app/config`。该镜像的源码插件配置加载器会从`/app/config/plugins/<plugin-id>/config.yaml`读取生产运行配置。
- 生产使用前必须把`linapro-sicau-niu-token-change-me`替换为强随机值。`sicau-niu`源码插件已编译进 nightly 镜像，即使`plugin.autoEnable`里只配置`media`，宿主启动注册路由时也会校验`sicau-niu`的`token.secret`。
- 除非已配置真实微信小程序凭证，否则保持`sicau-niu`配置中的`wechat.mock: true`。

## 多副本运行配置

两份清单都按`3`个 LinaPro 副本配置：

| 资源 | 配置 | 用途 |
| ---- | ---- | ---- |
| `Deployment/linapro` | `replicas: 3` | 运行 3 个 LinaPro 应用 Pod。 |
| `cluster.enabled` | `true` | 启用多节点运行时协调。 |
| `Deployment/linapro-redis` | `replicas: 1` | 提供 Redis 协调能力，用于选主、分布式锁和跨实例运行时一致性。 |
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

服务启动后，`plugin.autoEnable`会自动安装并启用`media`源码插件。插件安装阶段会执行`media`插件自己的安装 SQL。除非把`withMockData`改成`true`，否则不会加载演示数据。

查看工作负载状态：

```bash
kubectl -n linapro get pods
kubectl -n linapro get job linapro-db-init
kubectl -n linapro get svc
```

查看 LinaPro 日志：

```bash
kubectl -n linapro logs deploy/linapro -f
```

如果数据库初始化失败，可以查看初始化`Job`日志：

```bash
kubectl -n linapro logs job/linapro-db-init
```

## 问题记录与排障

以下问题是在验证`nightly-20260616`镜像时实际遇到的，当前清单已经包含对应修正：

| 现象 | 原因 | 处理方式 |
| ---- | ---- | -------- |
| `SetServerRoot failed: cannot find "resource/public"` | `server.serverRoot`指向了镜像中不存在的目录。 | 保持`server.serverRoot`未配置。 |
| `runtime config i18n.default cannot be empty` | 宿主缺少`i18n`运行配置。 | 保留清单内置的`i18n.default`和`i18n.locales`配置。 |
| `Player token secret is not configured` | 镜像内置的`sicau-niu`源码插件会在宿主启动时注册路由，并要求`token.secret`非空。 | 保留`Secret/linapro-source-plugin-configs`，并把其中`sicau-niu`的 token 密钥替换为生产值。 |
| `PersistentVolumeClaim/linapro-data`一直处于`Pending` | 默认`StorageClass`不支持`ReadWriteMany`，或集群没有默认存储类。 | 设置`PersistentVolumeClaim/linapro-data.spec.storageClassName`为支持 RWX 的存储类。 |
| `kubectl`访问命名空间资源时报`Forbidden` | 当前`kubeconfig`用户权限不足。 | 使用具备集群管理员权限的`kubeconfig`，很多自建节点可使用`/etc/kubernetes/admin.conf`。 |
| 需要使用端口`8082` | Kubernetes `NodePort`通常使用`30000-32767`端口范围，因此清单默认保持`ClusterIP`。 | 使用`kubectl port-forward`、`Ingress`或`LoadBalancer`暴露`8082`。 |

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

如果需要改用`NodePort`，可以把`Service/linapro.spec.type`改成`NodePort`，并设置集群节点端口范围内的端口，例如：

```yaml
spec:
  type: NodePort
  ports:
    - name: http
      port: 9120
      targetPort: 9120
      nodePort: 30080
```

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
