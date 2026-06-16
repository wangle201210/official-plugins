# LinaPro Water Kubernetes 部署说明

本目录提供一份单文件 Kubernetes 部署清单，用于部署只自动启用`water`源码插件的`nightly-20260616`版 LinaPro 镜像。

该清单默认部署 10 个`linapro-water` Pod，并通过一个 Kubernetes Service 分担水印处理带来的 CPU 和内存压力。清单不部署`media`插件。需要将`mediaStrategy.baseUrl`配置为已安装`media`插件、且`mediaopen`内部接口可被当前`water`集群访问的 LinaPro 集群地址。

## 文件说明

| 文件 | 用途 |
| ---- | ---- |
| `linapro-k8s.yaml` | 一体化`water`部署清单，会在集群内创建 PostgreSQL、Redis 协调服务、数据库初始化 Job 和 10 个 LinaPro `water` Pod。 |
| `README.md` | 英文使用说明，部署步骤与本文档保持一致。 |

## 前置条件

- Kubernetes 集群已配置默认`StorageClass`。
- `kubectl`已经连接到目标集群。
- 节点可以拉取`ghcr.io/wangle201210/linapro:nightly-20260616`。
- 存在一个可访问的`media` LinaPro 集群，且已启用`media`插件。
- `media`集群向当前`water`集群开放`GET /api/v1/strategies/resolve`。
- 使用内置`NodePort`服务时，端口`30081`未被占用。

## 配置

应用清单前，先编辑`linapro-k8s.yaml`并替换以下默认值：

| 字段 | 默认值 | 说明 |
| ---- | ------ | ---- |
| `POSTGRES_PASSWORD` | `linapro-change-me` | PostgreSQL 密码。 |
| `database.default.link` | `pgsql:postgres:linapro-change-me@tcp(linapro-water-postgres:5432)/linapro?sslmode=disable` | LinaPro 数据库连接串，需要与`POSTGRES_PASSWORD`保持一致。 |
| `jwt.secret` | `linapro-jwt-change-me` | JWT 签名密钥。 |
| `jwt.expire` | `24h` | JWT 过期时间。 |
| `cluster.enabled` | `true` | 为 10 个`water`Pod 启用 LinaPro 集群协调。 |
| `cluster.redis.address` | `linapro-water-redis:6379` | Redis 协调地址，用于集群选主、锁和共享运行期 KV cache。 |
| `plugin.autoEnable` | 仅`water` | 本清单只会自动安装并启用`water`源码插件。 |
| `mediaStrategy.baseUrl` | `http://media-linapro.example.com` | 远端 LinaPro `media`集群基础地址。 |
| `mediaStrategy.apiKey` | `media` | 发送给`mediaopen`的`X-Inner-Api-Key`请求头值。 |
| `mediaStrategy.timeout` | `10s` | 单次远端策略查询超时时间。 |
| `water.consumerCount` | `2` | 每个 LinaPro Pod 内的异步水印任务消费者数量。默认 10 副本时，总消费者数量为 20。 |
| `spec.replicas` | `10` | `linapro-water`Pod 数量。 |
| `resources.requests` | `500m` CPU、`512Mi`内存 | 每个水印处理 Pod 的基础资源请求。10 副本合计约请求`5`CPU 和`5Gi`内存。 |
| `resources.limits` | `2` CPU、`2Gi`内存 | 每个水印处理 Pod 的资源上限。10 副本合计最高可使用`20`CPU 和`20Gi`内存。 |
| `spec.ports[0].nodePort` | `30081` | `NodePort`对外访问端口。 |
| `resources.requests.storage` | `20Gi` | 内置 PostgreSQL 数据卷容量。 |

`water`插件会从`/app/config/plugins/water/config.yaml`读取插件作用域运行时配置。该清单通过`linapro-water-plugin-config` Secret 挂载该文件。`manifest/config/config.example.yaml`只是配置模板，不会作为运行时默认值读取。

跨集群访问时，将`mediaStrategy.baseUrl`设置为`media`集群的`Ingress`、负载均衡、专线地址、VPN 地址或`NodePort`地址。`mediaStrategy.apiKey`必须与`media`集群的`innerapi.apiKey`配置一致。

内置 PostgreSQL 容器会将`PGDATA`设置为`/var/lib/postgresql/data/pgdata`，避免部分存储类在卷根目录生成的元数据影响首次数据库初始化。

`linapro-water`Deployment 默认将`/app/data`挂载为`emptyDir`，这样 10 个 Pod 可以同时启动，不会抢占同一个`ReadWriteOnce`数据卷。当前`water`处理路径会将水印结果作为 data URL 返回或通过回调发送，适合使用该模式。后续如果需要共享上传文件、动态插件产物或持久化本地文件，需要将`emptyDir`替换为支持多 Pod 并发访问的存储，例如`ReadWriteMany`卷或对象存储集成。

## 部署

执行以下命令应用清单：

```bash
kubectl apply -f linapro-k8s.yaml
```

清单会创建一个一次性的数据库初始化 Job：

| Job | 用途 |
| --- | ---- |
| `linapro-water-init-database` | 使用挂载的`/app/config.yaml`执行`./lina init --confirm=init`，创建或升级宿主表结构和必需 Seed 数据。 |

每个`linapro-water`Pod 在启动服务前会先执行三个 init container：

| Init container | 用途 |
| -------------- | ---- |
| `wait-for-postgres` | 等待 PostgreSQL 可以连接。 |
| `wait-for-redis` | 等待 Redis 可以连接，用于集群协调。 |
| `wait-for-database-init` | 等待数据库初始化 Job 创建服务启动所需的宿主表结构。 |

服务启动后，`plugin.autoEnable`会自动安装并启用`water`源码插件。插件安装阶段会执行`water`插件自己的安装 SQL。启用后的`water`插件会使用挂载的插件配置调用远端`mediaopen`策略解析接口。

查看工作负载状态：

```bash
kubectl -n linapro-water get pods
kubectl -n linapro-water get job
kubectl -n linapro-water get svc
```

查看 LinaPro 日志：

```bash
kubectl -n linapro-water logs deploy/linapro-water -f
```

如果数据库初始化失败，可以查看 init container 日志：

```bash
kubectl -n linapro-water logs job/linapro-water-init-database
```

修改`config.yaml`后如需手工重跑宿主数据库初始化，删除已完成的 Job，重新应用清单，然后重启 LinaPro Pod：

```bash
kubectl -n linapro-water delete job linapro-water-init-database
kubectl apply -f linapro-k8s.yaml
kubectl -n linapro-water rollout restart deploy/linapro-water
```

修改`mediaStrategy`配置后，更新`linapro-water-plugin-config` Secret 并重启部署：

```bash
kubectl -n linapro-water rollout restart deploy/linapro-water
```

## 访问

等待`linapro-water` Pod 就绪后，通过任意 Kubernetes 节点访问：

```text
http://<node-ip>:30081/admin
```

如果集群不直接暴露节点端口，可以临时转发服务到本机：

```bash
kubectl -n linapro-water port-forward svc/linapro-water 9120:9120
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

## 扩容说明

水印处理会消耗较多 CPU 和内存。该清单默认启动 10 个 Pod，每个 Pod 2 个消费者，因此默认处理宽度为 20 个消费者。需要同时调整`spec.replicas`、`resources.requests`、`resources.limits`和`water.consumerCount`。

`Ingress`或 Service 路由可以把请求分发到 10 个 Pod。宿主 KV cache 在集群模式下使用内置 Redis 协调后端，因此异步任务状态可以从任意 Pod 读取。进程内任务队列仍归属于接收提交请求的 Pod，这符合当前部署形态。

## 卸载

执行以下命令删除本清单创建的所有资源：

```bash
kubectl delete -f linapro-k8s.yaml
```
