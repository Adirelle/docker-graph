package logging

import (
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/inconshreveable/log15"
	"github.com/mattn/go-isatty"
)

type (
	Config struct {
		Modules     ModuleLevels
		Color       ColorFlag
		StderrLevel Level
		Filename    string
	}

	Level log.Lvl

	ModuleLevels map[string]log.Lvl

	ColorFlag int
)

const (
	ModuleKey  = "module"
	MainModule = ""

	ColorAuto   ColorFlag = 0
	ColorAlways ColorFlag = 1
	ColorNever  ColorFlag = 2
)

var (
	_ flag.Value = (*ColorFlag)(nil)
	_ flag.Value = (*Level)(nil)
	_ flag.Value = (ModuleLevels)(nil)

	Log = log.New()
)

func (c *Config) SetupFlags() {
	flag.Var(c.Modules, "log", "Set logging levels")
	flag.Var(&c.StderrLevel, "logStderr", "Set the minimum level to log to stderr")
	flag.Var(&c.Color, "logColor", "Control the format of stderr logs")
	flag.StringVar(&c.Filename, "logFile", "", "Write logs to file")
}

func (c *Config) Apply(mainLogger log.Logger) error {
	if handler, err := c.createHandler(); err == nil {
		mainLogger.SetHandler(c.wrapHandler(handler))
		return nil
	} else {
		return err
	}
}

func (c *Config) createHandler() (log.Handler, error) {
	stderrHandler := c.createStderrHandler()
	if fileHandler, err := c.createFileHandler(); err != nil {
		return nil, err
	} else if fileHandler != nil {
		return log.MultiHandler(stderrHandler, fileHandler), nil
	}
	return stderrHandler, nil
}

func (c *Config) createStderrHandler() log.Handler {
	stdErrFormat := log.LogfmtFormat()
	if c.Color == ColorAlways || (c.Color == ColorAuto && isatty.IsTerminal(os.Stderr.Fd())) {
		stdErrFormat = log.TerminalFormat()
	}
	return log.LvlFilterHandler(log.Lvl(c.StderrLevel), log.StreamHandler(os.Stderr, stdErrFormat))
}

func (c *Config) createFileHandler() (log.Handler, error) {
	if c.Filename == "" {
		return nil, nil
	}

	fileFormat := log.LogfmtFormat()
	if strings.HasSuffix(c.Filename, ".json") {
		fileFormat = log.JsonFormat()
	}
	return log.FileHandler(c.Filename, fileFormat)
}

func (c *Config) accept(r *log.Record) bool {
	minLevel := c.Modules[MainModule]
	l := len(r.Ctx)
	for i := 0; i < l; i = i + 2 {
		if r.Ctx[i] == ModuleKey {
			if modName, ok := r.Ctx[i+1].(string); ok {
				if modLevel, found := c.Modules[modName]; found {
					minLevel = modLevel
				}
			}
		}
	}
	return r.Lvl <= minLevel
}

func (l ModuleLevels) String() string {
	buf := strings.Builder{}
	first := true
	for key, level := range l {
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

func (l ModuleLevels) Set(config string) error {
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
			l[key] = lvl
		} else {
			return err
		}
	}
	return nil
}

func (l *Level) String() string {
	return log.Lvl(*l).String()
}

func (l *Level) Set(value string) error {
	lvl, err := log.LvlFromString(value)
	if err == nil {
		*l = Level(lvl)
	}
	return err
}

func (c ColorFlag) String() string {
	switch c {
	case ColorNever:
		return "never"
	case ColorAuto:
		return "auto"
	case ColorAlways:
		return "always"
	}
	panic(fmt.Sprintf("invalid logColor value: %d", c))
}

func (c *ColorFlag) Set(value string) error {
	switch value {
	case "never":
		*c = ColorNever
	case "auto":
		*c = ColorAuto
	case "always":
		*c = ColorAlways
	default:
		return fmt.Errorf("invalid value: %s", value)
	}
	return nil
}
