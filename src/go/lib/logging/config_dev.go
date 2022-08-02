//go:build dev

package logging

import (
	log "github.com/inconshreveable/log15"
)

func (c *Config) wrapHandler(handler log.Handler) log.Handler {
	return log.CallerFileHandler(handler)
}
