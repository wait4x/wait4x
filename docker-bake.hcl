// Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {}
target "docker-metadata-action-debian" {}

// Common configuration
target "_common" {
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
}

// Group to build all image variants
group "image-all" {
  targets = ["image-alpine", "image-debian"]
}

// Default image target (alpine)
target "image" {
  inherits = ["image-alpine"]
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
