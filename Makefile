BUILD_FLAGS = -v
GENERATE_FLAGS = -x

.PHONY: all clean build generate dev prereq

all: build

clean:
	rm -f docker-graph app/public/script.*

cleaner: clean
	rm -fr app/node_modules

build: generate prereq
	go build $(BUILD_FLAGS) ./pkg/cli/docker-graph

generate: prereq
	go generate $(GENERATE_FLAGS) ./...

dev:
	modd -f pkg/cli/docker-graph/modd.conf

docker-dev:
	modd -f pkg/cli/docker-dev/modd.conf

prereq:
	go install github.com/cortesi/modd/cmd/modd@latest
	go get -v ./...
	cd app && bun install
