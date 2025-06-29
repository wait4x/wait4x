// Special target: https://github.com/docker/metadata-action#bake-definition
target "docker-metadata-action" {}

target "image" {
  inherits  = ["docker-metadata-action"]
  platforms = [
    "linux/amd64",
    "linux/arm/v6",
    "linux/arm/v7",
    "linux/arm64",
    "linux/ppc64le",
    "linux/s390x"
  ]
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
