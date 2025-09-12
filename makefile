# DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
TEMP := $(shell mktemp -d)

.PHONY: tools
tools:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


.PHONY: proto
proto:
	buf generate -o "${TEMP}" src/proto
	rm -r ./src/gen
	cp -R ${TEMP}/github.com/kheina/openconman/* .

.PHONY: deps
deps:
	./scripts/deps.sh

.PHONY: install
install: export OPENCONMAN_INSTALL_BIN=1
install: build-ui


.PHONY: build-ui
build-ui: export OPENCONMAN_BUILD_UI=1
build-ui: build


.PHONY: build
build:
	./scripts/build.sh

.PHONY: test
test:
	go test -v -cover -json ./... | tparse -follow
