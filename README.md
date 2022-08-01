# docker-graph

## Building from sources

# Prereqs

- [Go 1.18+](https://go.dev/dl/)
- [Bun](https://bun.sh/)
- Make

# Building

This generates the all-in-one binary `docker-graph`:

```shell
make build
```

# Developping

This launches a developpement server with live-reload listening on http://localost:8080:

```shell
make serve
```

# License

Unless stated otherwise, all files in this repository are licensed under the [MIT license](./LICENSE.md).
