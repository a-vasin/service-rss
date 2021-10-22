GO?=go
GOPATH?=$(shell go env GOPATH)
GOPACKAGES=$(shell go list ./... | grep -v generated)

GOFLAGS   = -mod=vendor
GOPROXY   = direct

export GOFLAGS   := $(GOFLAGS)
export GOPROXY   := $(GOPROXY)
export GOBIN     := $(PWD)/bin

all: build

help:
	@echo "help             - show this help"
	@echo "build            - build application from sources"
	@echo "clean            - remove build artifacts"
	@echo "fmt              - format application sources"
	@echo "check            - check code style"
	@echo "run              - start application"
	@echo "test             - run tests"
	@echo "imports          - format imports"
	@echo "deps-tidy        - go mod tidy && go mod vendor"
	@echo "toolset          - install tools to local ./bin"
	@echo "go-generate      - generate mocks"
	@echo "run-local-db     - start postgres db in docker container on port 5444"
	@echo "stop-local-db    - stop and remove docker container with db"

build: clean fmt check
	GOPATH=$(GOPATH) go build -o bin/service-entrypoint ./cmd/service

run: build
	./bin/service-entrypoint

clean:
	GOPATH=$(GOPATH) $(GO) clean
	rm -rf $(BINARY_PATH)

fmt:
	GOPATH=$(GOPATH) $(GO) fmt ${GOPACKAGES}

check:
	GOPATH=$(GOPATH) $(GO) vet $(GOPACKAGES)

test: clean
	$(GO) test $(GOPACKAGES)

imports:
	@./bin/goimports -ungroup -local service-rss -w ./internal ./cmd

deps-tidy:
	@go mod tidy
	@if [[ "`go env GOFLAGS`" =~ -mod=vendor ]]; then go mod vendor; fi

TOOLFLAGS = -mod=
toolset:
	@( \
		GOFLAGS=$(TOOLFLAGS); \
		cd tools; \
		go mod download; \
		go generate tools.go; \
	)

go-generate:
	@go generate $(GOPACKAGES)

run-local-db:
	./db/scripts/run.sh

stop-local-db:
	./db/scripts/stop.sh

build-docker:
	docker build --tag service-rss .
