GO?=go
GOPATH?=$(shell go env GOPATH)
GOPACKAGES=$(shell go list ./... | grep -v generated)

GOFLAGS   = -mod=vendor
GOPRIVATE = go.avito.ru
GOPROXY   = direct

export GOFLAGS   := $(GOFLAGS)
export GOPRIVATE := $(GOPRIVATE)
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
	@echo "imports-install  - install goimport to local ./bin"
	@echo "imports          - format imports"
	@echo "deps-tidy        - go mod tidy && go mod vendor"

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

coverage: clean
	$(GO) test -v -cover $(GOPACKAGES)

imports-install: toolset

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

