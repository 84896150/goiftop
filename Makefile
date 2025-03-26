.PHONY: build clean

default: build

BINARY=goiftop
GIT_VERSION := $(shell git rev-parse HEAD)
GO_VERSION := $(shell go version)
BUILD_TIME := $(shell date +%FT%T%z)

LDFLAGS=-ldflags '-s -X "github.com/fs714/goiftop/utils/version.GitVersion=${GIT_VERSION}" -X "github.com/fs714/goiftop/utils/version.GoVersion=${GO_VERSION}" -X "github.com/fs714/goiftop/utils/version.BuildTime=${BUILD_TIME}"'

# 默认目标平台：linux/arm64
build:
	env GOOS=linux GOARCH=arm64 go build -o bin/${BINARY}-linux-arm64 ${LDFLAGS}

# 针对 32 位 ARM 架构（例如 ARMv7）
build-armv7:
	env GOOS=linux GOARCH=arm GOARM=7 go build -o bin/${BINARY}-linux-armv7 ${LDFLAGS}

# 针对 64 位 ARM 架构
build-arm64:
	env GOOS=linux GOARCH=arm64 go build -o bin/${BINARY}-linux-arm64 ${LDFLAGS}

# 清理生成的二进制文件
clean:
	rm -rf bin/
