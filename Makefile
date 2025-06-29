PROJECT := visigoth
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git describe --tags 2>/dev/null || echo "dev")
SOURCE_FILES ?=./...
TEST_OPTIONS := -v -failfast -race
TEST_PATTERN ?=.
BENCH_OPTIONS ?= -v -bench=. -benchmem
CLEAN_OPTIONS ?=-modcache -testcache
TEST_TIMEOUT ?=2m
XC_OS := linux
XC_ARCH := amd64
LD_FLAGS := -X main.version=$(VERSION) -s -w

.PHONY: all
all: help

.PHONY: help
help:
	@echo "$(PROJECT) v$(VERSION)"
	@echo ""
	@echo "make build       - build $(PROJECT) binaries"
	@echo "make build-dev   - build $(PROJECT) for current platform"
	@echo "make build-docker - build $(PROJECT) for docker"
	@echo "make fmt         - use gofmt, goimports"
	@echo "make test        - run go test including race detection"
	@echo "make bench       - run go test including benchmarking"
	@echo "make clean       - clean build artifacts and test cache"
	@echo "make dist        - build and create distribution packages"
	@echo "make docker      - create docker image"
	@echo "make run-server  - run development server"
	@echo "make run-client  - run development client"
	@echo "make setup       - install development dependencies"

.PHONY: build
build:
	$(info Make: Build $(PROJECT) v$(VERSION))
	@scripts/build.sh

.PHONY: build-dev
build-dev:
	$(info Make: Build $(PROJECT) for development)
	@scripts/build.sh dev

.PHONY: build-docker
build-docker:
	$(info Make: Build $(PROJECT) for docker)
	@scripts/build.sh docker

.PHONY: fmt
fmt:
	$(info Make: Format)
	gofmt -w .
	goimports -w .
	golines -w .

.PHONY: test
test:
	$(info Make: Test)
	CGO_ENABLED=1 go test ${TEST_OPTIONS} ${SOURCE_FILES} -run ${TEST_PATTERN} -timeout=${TEST_TIMEOUT}

.PHONY: bench
bench:
	$(info Make: Benchmark)
	CGO_ENABLED=1 go test ${BENCH_OPTIONS} ${SOURCE_FILES} -run ${TEST_PATTERN} -timeout=${TEST_TIMEOUT}

.PHONY: clean
clean:
	$(info Make: Clean)
	@scripts/clean.sh

.PHONY: dist
dist:
	$(info Make: Distribution)
	@scripts/dist.sh

.PHONY: docker
docker: build-docker
	$(info Make: Docker)
	@scripts/docker.sh

.PHONY: run-server
run-server:
	$(info Make: Run server)
	go run cmd/server/server.go

.PHONY: run-client
run-client:
	$(info Make: Run client)
	go run cmd/cmd_client.go

.PHONY: setup
setup:
	$(info Make: Setup development environment)
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/segmentio/golines@latest
	@scripts/add-pre-commit.sh
