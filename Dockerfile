# syntax=docker/dockerfile:1.26.0@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

# Global so later FROM lines can interpolate it. An ARG after COPY is
# scoped to that stage, which made BASE_VARIANT empty and produced "runtime-".
ARG BASE_VARIANT=alpine

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.9.0@sha256:c64defb9ed5a91eacb37f96ccc3d4cd72521c4bd18d5442905b95e2226b0e707 AS xx

FROM --platform=$BUILDPLATFORM golang:1.26.6-alpine3.24@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS base
ENV GO111MODULE=auto
ENV CGO_ENABLED=0

COPY --from=xx / /
RUN apk add --update --no-cache build-base coreutils git
WORKDIR /src

FROM base AS build
ARG TARGETPLATFORM
ARG TARGETOS
ARG COMMIT_HASH
ARG COMMIT_REF_SLUG

RUN --mount=type=bind,target=/src,rw \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GO_BINARY=xx-go WAIT4X_BUILD_OUTPUT=/usr/bin WAIT4X_COMMIT_HASH=${COMMIT_HASH} WAIT4X_COMMIT_REF_SLUG=${COMMIT_REF_SLUG} make build \
    && xx-verify --static /usr/bin/wait4x*

FROM scratch AS binary
COPY --from=build /usr/bin/wait4x* /

FROM base AS releaser
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /work
RUN --mount=from=binary,target=/build \
  --mount=type=bind,target=/src \
  mkdir -p /out \
  && cp /build/wait4x* /src/README.md /src/LICENSE . \
  && tar -czvf "/out/wait4x-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.tar.gz" * \
  # Change dir to "/out" to prevent adding "/out" in the sha256sum command output.
  && cd /out \
  # Note: This will be removed in v4.0.0. Please use the SHA256SUMS file going forward.
  && sha256sum "wait4x-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.tar.gz" > "wait4x-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.tar.gz.sha256sum" \
  && sha256sum "wait4x-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.tar.gz" >> "SHA256SUMS"

FROM scratch AS artifact
COPY --from=releaser /out /

FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS runtime-alpine
RUN apk --update upgrade --no-cache \
    && apk add --no-cache ca-certificates tzdata

FROM debian:bookworm-slim@sha256:abd67ffcfa541b485a3dff59865ab629aa048a6c613e639d36e7456b0b229241 AS runtime-debian
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y --no-install-recommends \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

FROM runtime-${BASE_VARIANT} AS runtime

COPY --from=binary /wait4x /usr/bin/wait4x

ENTRYPOINT ["wait4x"]
CMD ["help"]
