// Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {}
target "docker-metadata-action-debian" {}
target "docker-metadata-action-alpine-nonroot" {}
target "docker-metadata-action-debian-nonroot" {}

// Common configuration
target "_common" {
  target = "runtime"
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/s390x"
  ]
}

// Alpine variant (default)
target "image-alpine" {
  inherits  = ["_common", "docker-metadata-action"]
  args = {
    BASE_VARIANT = "alpine"
  }
}

// Debian variant
target "image-debian" {
  inherits  = ["_common", "docker-metadata-action-debian"]
  args = {
    BASE_VARIANT = "debian"
  }
  // debian:13.6-slim publishes amd64, arm/v7, arm64, 386, ppc64le.
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le"
  ]
}

// Alpine non-root variant
target "image-alpine-nonroot" {
  inherits  = ["_common", "docker-metadata-action-alpine-nonroot"]
  target    = "runtime-nonroot"
  args = {
    BASE_VARIANT = "alpine"
  }
}

// Debian non-root variant
target "image-debian-nonroot" {
  inherits  = ["_common", "docker-metadata-action-debian-nonroot"]
  target    = "runtime-nonroot"
  args = {
    BASE_VARIANT = "debian"
  }
  // debian:13.6-slim publishes amd64, arm/v7, arm64, 386, ppc64le.
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le"
  ]
}

// Group to build all image variants
group "image-all" {
  targets = ["image-alpine", "image-debian", "image-alpine-nonroot", "image-debian-nonroot"]
}

// Default image target (alpine)
target "image" {
  inherits = ["image-alpine"]
}

// Default non-root image target (alpine non-root)
target "image-nonroot" {
  inherits = ["image-alpine-nonroot"]
}

target "artifact" {
  target    = "artifact"
  output    = ["./dist"]
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/mips",
    "linux/mipsle",
    "linux/mips64",
    "linux/mips64le",
    "linux/ppc64le",
    "linux/s390x",
    "windows/amd64",
    "windows/arm64",
    "darwin/amd64",
    "darwin/arm64"
  ]
}
