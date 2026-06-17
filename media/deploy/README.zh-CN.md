# Media 镜像构建

本目录存放用于部署已嵌入并自动启用`media`源码插件的 LinaPro 镜像资源。

`media`不需要专用插件运行时镜像。请使用仓库标准 LinaPro 镜像构建链路，并通过`plugins=1`打包源码插件资源。

## 文件说明

| 路径 | 用途 |
| ---- | ---- |
| `kubernetes/` | `media`版 LinaPro 镜像的 Kubernetes 清单与部署说明。 |

## 前置条件

- 在仓库根目录执行命令。
- 已安装 Docker，且可以访问目标镜像仓库。
- 多平台发布需要可用的`docker buildx`。
- 工作区包含预期的`apps/lina-plugins/media`插件版本和资源。
- 推送前已登录目标镜像仓库，例如`ghcr.io`。

## 构建并推送

发布镜像使用标准镜像构建入口：

```bash
make image \
  plugins=1 \
  platforms=linux/amd64,linux/arm64 \
  registry=ghcr.io/wangle201210 \
  image=linapro \
  tag=nightly-20260616 \
  push=1
```

该命令会构建包含源码插件资源的宿主二进制，并发布：

```text
ghcr.io/wangle201210/linapro:nightly-20260616
```

如果只需要本地单平台 smoke 构建且不推送：

```bash
make image \
  plugins=1 \
  platforms=linux/amd64 \
  registry=ghcr.io/wangle201210 \
  image=linapro \
  tag=local-media \
  push=0
```

多平台构建必须使用`push=1`，因为`docker buildx`会发布远端 manifest，而不是把单个镜像加载到本地 Docker。

## 验证

推送多平台镜像后，检查远端 manifest：

```bash
docker buildx imagetools inspect ghcr.io/wangle201210/linapro:nightly-20260616
```

manifest 应包含请求的平台，例如`linux/amd64`和`linux/arm64`。

如果 Kubernetes 部署复用同一个 tag，由于清单使用`imagePullPolicy: IfNotPresent`，需要在节点上拉取新镜像并重启部署：

```bash
docker pull ghcr.io/wangle201210/linapro:nightly-20260616
kubectl -n linapro rollout restart deploy/linapro
kubectl -n linapro rollout status deploy/linapro
```

然后确认运行中的 Pod 使用预期 digest：

```bash
kubectl -n linapro get pods -l app=linapro \
  -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].imageID}{"\n"}{end}'
```

## 运行时说明

- 保留`plugins=1`；否则镜像可能缺少 Kubernetes 清单预期的源码插件资源。
- 标准 LinaPro 镜像适用于`media`，不需要`CGO`、FFmpeg 或 x264 运行库。
- 除非镜像确实包含`/app/resource/public`，不要配置`server.serverRoot: "resource/public"`。
- 对要求`i18n.default`的镜像，Kubernetes 清单中需要保留`i18n`运行配置块。
