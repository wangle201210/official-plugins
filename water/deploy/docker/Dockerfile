ARG GO_IMAGE=golang:1.25-alpine3.22
ARG RUNTIME_IMAGE=alpine:3.22

FROM ${GO_IMAGE} AS builder

WORKDIR /src

RUN apk add --no-cache \
    build-base \
    ffmpeg-dev \
    git \
    nodejs \
    npm \
    pkgconf \
    x264-dev

RUN npm install -g corepack@latest \
    && corepack enable \
    && corepack prepare pnpm@10.30.3 --activate

COPY . .

RUN cd apps/lina-vben \
    && pnpm install --frozen-lockfile

RUN go run ./hack/tools/linactl build \
    plugins=1 \
    cgo_enabled=1 \
    platforms=linux/$(go env GOARCH)

FROM ${RUNTIME_IMAGE}

ENV WORKDIR=/app

RUN apk add --no-cache \
    ca-certificates \
    ffmpeg-libs \
    tzdata \
    x264-libs

WORKDIR ${WORKDIR}

COPY --from=builder /src/temp/output/lina ${WORKDIR}/lina
COPY .github/image/config.runtime.yaml ${WORKDIR}/config.yaml

EXPOSE 9120

CMD ["./lina"]
