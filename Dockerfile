# syntax=docker/dockerfile:1.26.0@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

# Global so later FROM lines can interpolate it. An ARG after COPY is
# scoped to that stage, which made BASE_VARIANT empty and produced "runtime-".
ARG BASE_VARIANT=alpine

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.10.0@sha256:40fe21f44bb8cddf1eb1226db190d0a984118f62ee02a91b77d99448d28264a6 AS xx

FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine3.24@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS base
ENV GO111MODULE=auto
ENV CGO_ENABLED=0

COPY --from=xx / /
RUN apk --update upgrade --no-cache \
    && apk add --no-cache build-base coreutils git
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

FROM alpine:3.24.4@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS runtime-alpine
RUN apk --update upgrade --no-cache \
    && apk add --no-cache ca-certificates tzdata

FROM debian:13.7-slim@sha256:a72b4433de1cf7d6bae0c4cee74d0547f58601a643b3375f070c51600d63b6b2 AS runtime-debian
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y --no-install-recommends \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

FROM runtime-${BASE_VARIANT} AS runtime

COPY --from=binary /wait4x /usr/bin/wait4x

ENTRYPOINT ["wait4x"]
CMD ["help"]
