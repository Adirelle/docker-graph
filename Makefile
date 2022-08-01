export GOPATH = $(abspath .direnv/go)

GO_SOURCES = $(shell find src/go/lib -name *.go)
GO_TARGET = docker-graph
GO_CLI_SRC = src/go/cli/docker-graph

TS_SOURCES = $(shell find src/ts -name *.ts)
ASSETS = $(shell find public)

# PHONY targets

.PHONY: all build clean cleaner

all: build

build: $(GO_TARGET)

clean:
	rm -fr $(GO_TARGET)

cleaner: clean
	if [ -d .direnv ]; then chmod -R u+w .direnv; fi
	rm -fr node_modules .direnv

# File targets

$(GO_TARGET): $(GO_SOURCES) $(GO_CLI_SRC) $(TS_SOURCES) $(ASSETS)
	go generate -x ./...
	go build ./$(GO_CLI_SRC)

node_modules: package.json bun.lockb
	bun install
