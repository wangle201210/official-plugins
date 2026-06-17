# Media Image Build

This directory contains deployment assets for the LinaPro image that embeds and auto-enables the `media` source plugin.

`media` does not need a dedicated plugin runtime image. Build it with the repository standard LinaPro image pipeline and enable source plugin packaging with `plugins=1`.

## Files

| Path | Purpose |
| ---- | ------- |
| `kubernetes/` | Kubernetes manifests and deployment guide for the `media`-enabled LinaPro image. |

## Prerequisites

- Run commands from the repository root.
- Docker is installed and can access the target registry.
- `docker buildx` is available for multi-platform publishing.
- The working tree contains the intended `apps/lina-plugins/media` plugin version and resources.
- You are logged in to the registry before pushing, for example `ghcr.io`.

## Build And Push

Use the standard image builder for release images:

```bash
make image \
  plugins=1 \
  platforms=linux/amd64,linux/arm64 \
  registry=ghcr.io/wangle201210 \
  image=linapro \
  tag=nightly-20260616 \
  push=1
```

The command builds host binaries with source plugins packaged, then publishes:

```text
ghcr.io/wangle201210/linapro:nightly-20260616
```

For a local single-platform smoke build without pushing:

```bash
make image \
  plugins=1 \
  platforms=linux/amd64 \
  registry=ghcr.io/wangle201210 \
  image=linapro \
  tag=local-media \
  push=0
```

Multi-platform builds require `push=1` because `docker buildx` publishes a registry manifest instead of loading a single local image.

## Verify

After pushing a multi-platform image, inspect the registry manifest:

```bash
docker buildx imagetools inspect ghcr.io/wangle201210/linapro:nightly-20260616
```

The manifest should include the requested platforms, such as `linux/amd64` and `linux/arm64`.

To verify a Kubernetes rollout that reuses the same tag, pull the tag on the node and restart the deployment because the manifest uses `imagePullPolicy: IfNotPresent`:

```bash
docker pull ghcr.io/wangle201210/linapro:nightly-20260616
kubectl -n linapro rollout restart deploy/linapro
kubectl -n linapro rollout status deploy/linapro
```

Then confirm the running pods use the expected digest:

```bash
kubectl -n linapro get pods -l app=linapro \
  -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.status.containerStatuses[0].imageID}{"\n"}{end}'
```

## Runtime Notes

- Keep `plugins=1`; otherwise the image may not contain the source plugin resources expected by the Kubernetes manifest.
- The standard LinaPro image is valid for `media`; it does not require `CGO`, FFmpeg, or x264 runtime libraries.
- Do not configure `server.serverRoot: "resource/public"` unless the image actually contains `/app/resource/public`.
- Keep the `i18n` runtime config block in Kubernetes manifests for images that require `i18n.default`.
