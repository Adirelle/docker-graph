# PHONY targets

.PHONY: all build server clean cleaner

all: build

build:
	bun install
	go get ./...
	go generate -x ./...
	go build ./src/go/cli/docker-graph

serve:
	docker composer up --build --detach

clean:
	rm -fr docker-graph public/js/index.js node_modules.bun

cleaner: clean
	rm -fr node_modules

