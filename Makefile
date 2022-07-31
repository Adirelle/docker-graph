BUILD_FLAGS = -v
GENERATE_FLAGS = -x

.PHONY: all clean build generate dev prereq modd get node_modules

all: build

clean:
	rm -f docker-graph app/public/script.*

cleaner: clean
	rm -fr app/node_modules

build: generate get
	go build $(BUILD_FLAGS) ./pkg/cli/docker-graph

generate: get node_modules
	go generate $(GENERATE_FLAGS) ./...

dev: modd get node_modules
	modd -f pkg/cli/docker-graph/modd.conf

get:
	go get -v ./...

node_modules:
	cd app && bun install

modd:
	go install github.com/cortesi/modd/cmd/modd@latest
