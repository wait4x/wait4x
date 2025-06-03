# syntax=docker/dockerfile:1.5.1
FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.24-alpine3.21 AS base
ENV GO111MODULE=auto
ENV CGO_ENABLED=0

COPY --from=xx / /
RUN apk add --update --no-cache build-base coreutils git
WORKDIR /src

FROM base AS build
ARG TARGETPLATFORM
ARG TARGETOS

RUN --mount=type=bind,target=/src,rw \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GO_BINARY=xx-go WAIT4X_BUILD_OUTPUT=/usr/bin make build \
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
  && sha256sum "wait4x-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.tar.gz" > "wait4x-${TARGETOS}-${TARGETARCH}${TARGETVARIANT}.tar.gz.sha256sum"

FROM scratch AS artifact
COPY --from=releaser /out /

FROM alpine:3.22
RUN apk add --update --no-cache ca-certificates tzdata

COPY --from=binary /wait4x /usr/bin/wait4x

ENTRYPOINT ["wait4x"]
CMD ["help"]
