# Water Image Build

This directory contains deployment assets for the LinaPro image dedicated to the `water` source plugin.

`water` must use a dedicated image because real watermark rendering requires `CGO` and FFmpeg/x264 runtime libraries. Do not use the generic LinaPro image unless it was also built with the same runtime requirements.

## Files

| Path | Purpose |
| ---- | ------- |
| `docker/Dockerfile` | Dedicated `water` image build with `CGO`, FFmpeg, and x264. |
| `kubernetes/` | Kubernetes manifests and deployment guide for the `water` image. |

## Prerequisites

- Run commands from the repository root.
- Docker is installed and can access the target registry.
- `docker buildx` is available for multi-platform publishing.
- You are logged in to the registry before pushing, for example `ghcr.io`.
- The build environment can download Go, Node, npm, pnpm, Alpine, and package dependencies used by the Dockerfile.

## Build And Push

Build and push the dedicated `water` image with the plugin Dockerfile:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f apps/lina-plugins/water/deploy/docker/Dockerfile \
  -t ghcr.io/wangle201210/linapro-water:nightly-20260617-f749194 \
  --push .
```

The Dockerfile builds the LinaPro host with:

| Setting | Value | Purpose |
| ------- | ----- | ------- |
| `plugins` | `1` | Includes source plugin resources in the host binary. |
| `cgo_enabled` | `1` | Enables the `water` renderer branch that links native libraries. |
| Runtime packages | `ffmpeg-libs`, `x264-libs` | Provides FFmpeg/x264 libraries required by watermark rendering. |

For a local single-platform smoke build:

```bash
docker buildx build \
  --platform linux/amd64 \
  -f apps/lina-plugins/water/deploy/docker/Dockerfile \
  -t ghcr.io/wangle201210/linapro-water:local-water \
  --load .
```

Use `--push` for multi-platform registry images and `--load` only for single-platform local testing.

## Verify

After pushing a multi-platform image, inspect the registry manifest:

```bash
docker buildx imagetools inspect ghcr.io/wangle201210/linapro-water:nightly-20260617-f749194
```

The Kubernetes manifest includes `verify-watermark-runtime` as an init container. It checks that `/app/lina` is linked with FFmpeg/x264 and fails fast if the image was built without `CGO` or required runtime libraries.

To update Kubernetes when reusing the same tag, pull the tag on the node and restart the deployment because the manifest uses `imagePullPolicy: IfNotPresent`:

```bash
docker pull ghcr.io/wangle201210/linapro-water:nightly-20260617-f749194
kubectl -n linapro-water rollout restart deploy/linapro-water
kubectl -n linapro-water rollout status deploy/linapro-water
```

Then confirm the running pods use the expected digest:

```bash
kubectl -n linapro-water get pods -l app=linapro-water \
  -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].imageID}{"\n"}{end}'
```

## Runtime Notes

- The generic `ghcr.io/wangle201210/linapro` image normally has `CGO` disabled and is not valid for production `water` rendering.
- Pods can stay at `Init:3/4` while `verify-watermark-runtime` checks binary linkage. This can take tens of seconds on a cold image.
- If the init container reports missing FFmpeg/x264 libraries or a non-`CGO` build, rebuild with `docker/Dockerfile`.
- Do not configure `server.serverRoot: "resource/public"` unless the image actually contains `/app/resource/public`.
- Keep the `i18n` runtime config block in Kubernetes manifests for images that require `i18n.default`.
