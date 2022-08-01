package assets

import "embed"

//go:generate bun run build
//go:embed *
var Assets embed.FS
