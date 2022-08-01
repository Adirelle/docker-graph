package docker_graph

import "embed"

//go:generate esbuild src/ts/index.ts --minify --bundle --platform=browser --outfile=public/js/index.js
//go:embed public
var Assets embed.FS
