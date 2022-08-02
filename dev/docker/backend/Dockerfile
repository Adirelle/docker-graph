FROM golang:1.18-bullseye

RUN apt-get update -yq \
  && apt-get install -yq entr

RUN mkdir /src

WORKDIR /src
CMD find src/go | entr -n -r -d go run -tags dev ./src/go/cli/docker-graph -bind 0.0.0.0:8080 -logLevel debug -logColor always
