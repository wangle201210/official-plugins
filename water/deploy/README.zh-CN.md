# Water 镜像构建

本目录存放`water`源码插件专用 LinaPro 镜像的部署资源。

`water`必须使用专用镜像，因为真实水印渲染需要`CGO`以及 FFmpeg/x264 运行库。除非通用 LinaPro 镜像也按相同运行时要求构建，否则不要用通用镜像替代。

## 文件说明

| 路径 | 用途 |
| ---- | ---- |
| `docker/Dockerfile` | 启用`CGO`并安装 FFmpeg/x264 的`water`专用镜像构建文件。 |
| `kubernetes/` | `water`镜像的 Kubernetes 清单与部署说明。 |

## 前置条件

- 在仓库根目录执行命令。
- 已安装 Docker，且可以访问目标镜像仓库。
- 多平台发布需要可用的`docker buildx`。
- 推送前已登录目标镜像仓库，例如`ghcr.io`。
- 构建环境可以下载 Dockerfile 使用的 Go、Node、npm、pnpm、Alpine 和包依赖。

## 构建并推送

使用插件 Dockerfile 构建并推送`water`专用镜像：

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f apps/lina-plugins/water/deploy/docker/Dockerfile \
  -t ghcr.io/wangle201210/linapro-water:nightly-20260617-f749194 \
  --push .
```

该 Dockerfile 会按以下配置构建 LinaPro 宿主：

| 配置 | 值 | 用途 |
| ---- | -- | ---- |
| `plugins` | `1` | 将源码插件资源打包进宿主二进制。 |
| `cgo_enabled` | `1` | 启用链接原生库的`water`渲染器分支。 |
| 运行时包 | `ffmpeg-libs`、`x264-libs` | 提供水印渲染所需的 FFmpeg/x264 运行库。 |

如果只需要本地单平台 smoke 构建：

```bash
docker buildx build \
  --platform linux/amd64 \
  -f apps/lina-plugins/water/deploy/docker/Dockerfile \
  -t ghcr.io/wangle201210/linapro-water:local-water \
  --load .
```

多平台远端镜像使用`--push`，单平台本地测试才使用`--load`。

## 验证

推送多平台镜像后，检查远端 manifest：

```bash
docker buildx imagetools inspect ghcr.io/wangle201210/linapro-water:nightly-20260617-f749194
```

Kubernetes 清单包含`verify-watermark-runtime`初始化容器。它会检查`/app/lina`是否链接 FFmpeg/x264，并在镜像关闭`CGO`或缺少运行库时快速失败。

如果 Kubernetes 部署复用同一个 tag，由于清单使用`imagePullPolicy: IfNotPresent`，需要在节点上拉取新镜像并重启部署：

```bash
docker pull ghcr.io/wangle201210/linapro-water:nightly-20260617-f749194
kubectl -n linapro-water rollout restart deploy/linapro-water
kubectl -n linapro-water rollout status deploy/linapro-water
```

然后确认运行中的 Pod 使用预期 digest：

```bash
kubectl -n linapro-water get pods -l app=linapro-water \
  -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].imageID}{"\n"}{end}'
```

## 运行时说明

- 通用`ghcr.io/wangle201210/linapro`镜像通常关闭`CGO`，不能用于生产水印渲染。
- Pod 可能在`Init:3/4`停留一段时间，因为`verify-watermark-runtime`正在检查二进制链接；冷镜像场景下可能持续几十秒。
- 如果初始化容器提示缺少 FFmpeg/x264 运行库或不是`CGO`构建，需要使用`docker/Dockerfile`重新构建。
- 除非镜像确实包含`/app/resource/public`，不要配置`server.serverRoot: "resource/public"`。
- 对要求`i18n.default`的镜像，Kubernetes 清单中需要保留`i18n`运行配置块。
