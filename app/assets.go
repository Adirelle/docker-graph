package app

import "embed"

//go:generate bun run build
//go:embed public
var Assets embed.FS
