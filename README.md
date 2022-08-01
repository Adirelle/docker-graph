# docker-graph

## Building from sources

# Building

Prereqs:

- [Go 1.18+](https://go.dev/dl/)
- [Bun](https://bun.sh/)
- Make

This generates the all-in-one binary `docker-graph`:

```shell
make build
```

# Developping

Prereqs:

- docker-compose

This launches a developpement server with live-reload listening on http://localost:8080:

```shell
docker-compose up --build -d
```

# License

Unless stated otherwise, all files in this repository are licensed under the [MIT license](./LICENSE.md).
