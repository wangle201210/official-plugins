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
- 使用内置`NodePort`服务时，端口`30080`未被占用。

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
| `auth.jwt.secret` | `linapro-jwt-change-me` | JWT 签名密钥。 |
| `spec.ports[0].nodePort` | `30080` | `NodePort`对外访问端口。 |
| `resources.requests.storage` | `20Gi` | PostgreSQL 和 LinaPro 数据卷容量。 |
| `plugin.autoEnable[0].withMockData` | `false` | 启动自动安装`media`插件时是否加载演示数据。生产环境保持`false`。 |
| `spec.replicas` | `3` | LinaPro 应用副本数。 |

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
| `auth.jwt.secret` | `linapro-jwt-change-me` | JWT 签名密钥。 |
| `spec.replicas` | `3` | LinaPro 应用副本数。 |

`linapro-db-init` `Job`会执行`./lina init --confirm=init`。配置的 PostgreSQL 账号必须可以连接 PostgreSQL 维护库`postgres`，检查目标数据库是否存在，在目标数据库不存在时创建数据库，并在目标数据库中创建或更新表、索引、注释和 Seed 数据。

## 多副本运行配置

两份清单都按`3`个 LinaPro 副本配置：

| 资源 | 配置 | 用途 |
| ---- | ---- | ---- |
| `Deployment/linapro` | `replicas: 3` | 运行 3 个 LinaPro 应用 Pod。 |
| `cluster.enabled` | `true` | 启用多节点运行时协调。 |
| `Deployment/linapro-redis` | `replicas: 1` | 提供 Redis 协调能力，用于选主、分布式锁和跨实例运行时一致性。 |
| `PersistentVolumeClaim/linapro-data` | `ReadWriteMany` | 让 3 个 LinaPro Pod 共享上传文件和插件运行时数据。 |

如果目标集群没有支持`ReadWriteMany`的存储类，需要先把 PVC 的存储类替换为支持共享挂载的存储类，再应用清单。

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

随后`linapro` Pod 会等待 PostgreSQL、Redis 和已初始化的宿主表结构就绪，再启动服务。

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

修改`config.yaml`后如需手工重跑宿主数据库初始化，删除已完成的`Job`后重新应用清单：

```bash
kubectl -n linapro delete job linapro-db-init
kubectl apply -f <selected-manifest>.yaml
```

## 访问

等待`linapro` Pod 就绪后，通过任意 Kubernetes 节点访问：

```text
http://<node-ip>:30080/admin
```

如果集群不直接暴露节点端口，可以临时转发服务到本机：

```bash
kubectl -n linapro port-forward svc/linapro 9120:9120
```

然后访问：

```text
http://127.0.0.1:9120/admin
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
