package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/adirelle/docker-graph/src/go/lib/docker/connections"
	"github.com/adirelle/docker-graph/src/go/lib/docker/containers"
	"github.com/adirelle/docker-graph/src/go/lib/docker/events"
	"github.com/adirelle/docker-graph/src/go/lib/docker/listeners"
	log "github.com/inconshreveable/log15"
)

type (
	logLevelMap map[string]log.Lvl

	logLevelFilter struct {
		levels logLevelMap
		next   log.Handler
	}
)

var (
	Log = log.New()

	logLevels = logLevelMap{"": log.LvlWarn}

	_ log.Handler = (*logLevelFilter)(nil)
)

func init() {
	dockerLogger := Log.New("module", "docker")
	connections.Log = dockerLogger.New("module", "connections")
	containers.Log = dockerLogger.New("module", "containers")
	events.Log = dockerLogger.New("module", "events")
	listeners.Log = dockerLogger.New("module", "listeners")

	flag.Var(logLevels, "log", "Setup logging")
}

func SetupLogging() {
	fmt.Printf("levels: %#v", logLevels)

	backend := Log.GetHandler()

	Log.SetHandler(&logLevelFilter{logLevels, backend})
}

func (m logLevelMap) String() string {
	buf := strings.Builder{}
	first := true
	for key, level := range m {
		if first {
			first = false
		} else {
			buf.WriteString(",")
		}
		if key != "" {
			buf.WriteString(key)
			buf.WriteString(":")
		}
		buf.WriteString(level.String())

	}
	return buf.String()
}

func (m logLevelMap) Set(config string) error {
	var key, levelStr string
	for _, part := range strings.Split(config, ",") {
		subParts := strings.SplitN(part, ":", 2)
		switch len(subParts) {
		case 2:
			key = subParts[0]
			levelStr = subParts[1]
		case 1:
			key = ""
			levelStr = subParts[0]
		default:
			return fmt.Errorf("invalid log level: %q", part)
		}
		if lvl, err := log.LvlFromString(levelStr); err == nil {
			m[key] = lvl
		} else {
			return err
		}
	}
	return nil
}

func (f *logLevelFilter) Log(r *log.Record) error {
	minLevel := f.levels[""]
	l := len(r.Ctx)
	for i := 0; i < l; i = i + 2 {
		if r.Ctx[i] == "module" {
			if modName, ok := r.Ctx[i+1].(string); ok {
				if modLevel, found := f.levels[modName]; found {
					minLevel = modLevel
				}
			}
		}
	}
	if r.Lvl > minLevel {
		return nil
	}
	return f.next.Log(r)
}
