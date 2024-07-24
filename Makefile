# Values to install the provider locally for testing purposes
HOSTNAME=registry.terraform.io
NAMESPACE=qernal
NAME=qernal
BINARY=terraform-provider-${NAME}
VERSION=1.0.0
OS_ARCH=$(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)



local-build:
	go build -o ${BINARY}

local-install: local-build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
