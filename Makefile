BUILD_FLAGS = -v
GENERATE_FLAGS = -x

.PHONY: all clean build generate dev prereq modd get node_modules

all: build

clean:
	rm -fr docker-graph

cleaner: clean
	rm -fr node_modules

build: generate get
	go build $(BUILD_FLAGS) ./src/go/cli/docker-graph

generate: get node_modules
	go generate $(GENERATE_FLAGS) ./...

dev: devtools get node_modules
	modd -f pkg/cli/docker-graph/modd.conf

get:
	go get -v ./...

node_modules:
	bun install

devtools:
	go install github.com/cortesi/modd/cmd/modd@latest
	go install github.com/cortesi/devd/cmd/devd@latest
