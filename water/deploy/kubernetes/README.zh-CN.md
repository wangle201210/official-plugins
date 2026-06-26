# LinaPro Water Kubernetes 部署说明

本目录提供一份单文件 Kubernetes 部署清单，用于部署只自动启用`water`源码插件的 LinaPro 的`water`专用镜像。

该清单默认部署 10 个`linapro-water` Pod，并通过一个 Kubernetes Service 分担水印处理带来的 CPU 和内存压力。清单不部署`media`插件。需要将`mediaStrategy.baseUrl`配置为已安装`media`插件、且`mediaopen`内部接口可被当前`water`集群访问的 LinaPro 集群地址。

## 文件说明

| 文件 | 用途 |
| ---- | ---- |
| `linapro-k8s.yaml` | 一体化`water`部署清单，会在集群内创建 PostgreSQL、Redis 协调服务、数据库初始化 Job 和 10 个 LinaPro `water` Pod。 |
| `../docker/Dockerfile` | `water`渲染器专用的`CGO`和 FFmpeg/x264 镜像构建文件。 |
| `README.md` | 英文使用说明，部署步骤与本文档保持一致。 |

## 前置条件

- Kubernetes 集群已配置默认`StorageClass`。
- `kubectl`已经连接到目标集群。
- 节点可以拉取`ghcr.io/wangle201210/linapro-water:nightly-20260616`。
- `linapro-water-data`使用的存储类支持`ReadWriteMany`。
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
| `i18n.default` | `zh-CN` | 必需的宿主运行时默认语言配置。不要删除该配置块。 |
| `logger.level` | `info` | 运行日志级别。多副本部署保持`info`；仅在短时间排障时改为`all`。 |
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
| `linapro-water-data.resources.requests.storage` | `10Gi` | `/app/data`共享存储，用于上传文件、动态插件产物和本地运行时文件。 |
| `resources.requests.storage` | `20Gi` | 内置 PostgreSQL 数据卷容量。 |

`water`插件会从`/app/config/plugins/water/config.yaml`读取插件作用域运行时配置。该清单通过`linapro-water-plugin-config` Secret 挂载该文件。`manifest/config/config.example.yaml`只是配置模板，不会作为运行时默认值读取。

跨集群访问时，将`mediaStrategy.baseUrl`设置为`media`集群的`Ingress`、负载均衡、专线地址、VPN 地址或`NodePort`地址。如果`media`使用同级清单部署在同一个 Kubernetes 集群内，集群内地址通常是`http://linapro.linapro.svc.cluster.local:9120`。`mediaStrategy.apiKey`必须与`media`集群的`innerapi.apiKey`配置一致；`media`插件未显式配置`innerapi.apiKey`时，默认值是`media`。

内置 PostgreSQL 容器会将`PGDATA`设置为`/var/lib/postgresql/data/pgdata`，避免部分存储类在卷根目录生成的元数据影响首次数据库初始化。

`linapro-water`Deployment 会从`linapro-water-data`这个`ReadWriteMany`声明挂载`/app/data`。这样宿主上传目录、动态插件产物和其他本地运行时文件可以在 10 个 Pod 之间共享。如果该部署严格只提供 API，且集群没有`ReadWriteMany`存储类，可以将该卷替换为`emptyDir`；但上传文件、动态插件产物和本地运行时文件会变成 Pod 本地数据，并在 Pod 重启后丢失。

## 构建镜像

通用的`ghcr.io/wangle201210/linapro:nightly-20260616`镜像不能用于真实`water`渲染，因为默认镜像构建会关闭`CGO`。`water`渲染器需要`CGO`分支以及 FFmpeg/x264 运行库。

应用 Kubernetes 清单前，先构建并推送专用镜像：

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f apps/lina-plugins/water/deploy/docker/Dockerfile \
  -t ghcr.io/wangle201210/linapro-water:nightly-20260616 \
  --push .
```

该 Dockerfile 会复用`linactl build`，并使用`plugins=1`和`cgo_enabled=1`构建宿主，最终镜像会安装 FFmpeg/x264 运行库。Deployment 还会在服务启动前执行`verify-watermark-runtime`，如果镜像不是`CGO`构建或缺少 FFmpeg/x264 运行库，会直接启动失败，避免对外提供不可用的水印接口。

## 部署

执行以下命令应用清单：

```bash
kubectl apply -f linapro-k8s.yaml
```

清单会创建一个一次性的数据库初始化 Job：

| Job | 用途 |
| --- | ---- |
| `linapro-water-init-database` | 使用挂载的`/app/config.yaml`执行`./lina init --confirm=init`，创建或升级宿主表结构和必需 Seed 数据。 |

每个`linapro-water`Pod 在启动服务前会先执行四个 init container：

| Init container | 用途 |
| -------------- | ---- |
| `wait-for-postgres` | 等待 PostgreSQL 可以连接。 |
| `wait-for-redis` | 等待 Redis 可以连接，用于集群协调。 |
| `wait-for-database-init` | 等待数据库初始化 Job 创建最后一个宿主 SQL 文件中的分布式缓存修订索引。 |
| `verify-watermark-runtime` | 当镜像关闭`CGO`或缺少 FFmpeg/x264 运行库时快速失败。 |

服务启动后，`plugin.autoEnable`会自动安装并启用`water`源码插件。插件安装阶段会执行`water`插件自己的安装 SQL。启用后的`water`插件会使用挂载的插件配置调用远端`mediaopen`策略解析接口。

查看工作负载状态：

```bash
kubectl -n linapro-water get pods
kubectl -n linapro-water get job
kubectl -n linapro-water get svc
```

查看所有`water`副本的 LinaPro 日志：

```bash
kubectl -n linapro-water logs -f -l app=linapro-water -c linapro --max-log-requests=10
```

`kubectl logs deploy/linapro-water -f`只会跟随 Deployment 选择到的 Pod，可能看不到其他副本的处理日志。

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

## 问题记录与排障

| 现象 | 原因 | 处理方式 |
| ---- | ---- | -------- |
| `SetServerRoot failed: cannot find "resource/public"` | 使用的`water`镜像内没有`/app/resource/public`，但清单配置了`server.serverRoot: "resource/public"`。 | 从`linapro-water-config`的`config.yaml`中移除`server.serverRoot`，重新应用清单并重启`Deployment/linapro-water`。 |
| `runtime config i18n.default cannot be empty` | 目标镜像运行时要求宿主配置中存在`i18n.default`。 | 在`linapro-water-config`的`config.yaml`中补充`i18n.default`和`i18n.locales`配置。 |
| Pod 长时间停留在`Init:3/4` | `verify-watermark-runtime`正在对`/app/lina`执行`CGO`和 FFmpeg/x264 链接检查，新镜像或冷缓存场景下可能持续几十秒。 | 先查看`kubectl -n linapro-water logs <pod> -c verify-watermark-runtime`和事件。如果没有缺失库错误，等待检查完成；如果提示缺少 FFmpeg/x264 或非`CGO`构建，需要换用按`../docker/Dockerfile`构建的`water`镜像。 |

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

在 Kubernetes 集群内部，Service DNS 地址是：

```text
http://linapro-water.linapro-water.svc.cluster.local:9120
```

`kubectl port-forward`只适合做连通性 smoke 检查。对 Service 执行端口转发时，流量可能只进入单个后端 Pod，因此不能用于验证多 Pod 负载均衡或容量压测。需要分发到所有副本的流量，应从集群内部使用 Service DNS，或通过`NodePort`、`Ingress`、`LoadBalancer`进入。

## 异步队列与回调

每个 Pod 都有一个进程内异步任务队列，最多缓存`1024`个等待任务。只有提交流量均匀分发到多个 Pod 时，总队列容量才会随`spec.replicas`扩展。如果单个 Pod 收到的提交速度超过本 Pod 的`water.consumerCount`消费者处理速度，新提交会失败并返回`WATER_TASK_QUEUE_FULL`。扩容时需要同时评估`spec.replicas`、`water.consumerCount`、CPU 和内存，不要通过`kubectl port-forward`做容量测试。

异步任务状态存储在宿主共享 KV cache 中，因此可以从任意 Pod 查询。状态响应会故意保持`image`为空，避免大图片写入缓存超过单值大小限制；完整输出图片只通过回调返回。

提交接口支持`callbackUrl`。存在回调地址时，`water`会使用`Content-Type: application/json`发送`POST`请求，请求体结构如下：

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

回调接收端需要在超时时间内返回`2xx`状态码。如果接收端要求`image_base64`或`image_url`等字段，需要调整接收端或增加一个回调适配层；当前回调契约使用`image`字段。

## 更新镜像

编辑`linapro-k8s.yaml`中的镜像字段：

```yaml
image: ghcr.io/wangle201210/linapro-water:nightly-20260616
```

需要使用从`apps/lina-plugins/water/deploy/docker/Dockerfile`构建的`water`镜像；不要替换为通用`linapro`镜像，除非该镜像同样开启了`CGO`并包含 FFmpeg/x264 运行库。

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
