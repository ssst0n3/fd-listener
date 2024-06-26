variable "APT_MIRROR" {
  default = "cdn-fastly.deb.debian.org"
#  default = "repo.huaweicloud.com"
}

variable "GOPROXY" {
  default = "https://goproxy.io,https://goproxy.cn,direct"
  #  default = "repo.huaweicloud.com"
}

variable "SLIM_LDFLAGS" {
  default = "-s -w"
}

group "default" {
  targets = ["binary"]
}

target "_common" {
  args = {
    APT_MIRROR = APT_MIRROR
    GOPROXY = GOPROXY
    SLIM_LDFLAGS = SLIM_LDFLAGS
  }
}

target "binary" {
  dockerfile = "Dockerfile_dev"
  inherits = ["_common"]
  target = "binary"
  output = ["bin/release"]
}