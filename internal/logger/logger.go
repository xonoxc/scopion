package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const LevelFatal = slog.Level(12)

var Default *slog.Logger

func init() {
	Default = New(Config{
		Level:  "info",
		Format: "auto",
		Output: os.Stderr,
	})
}

type Config struct {
	Level     string
	Format    string
	AddSource bool
	Output    io.Writer
}

func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	format := resolveFormat(cfg.Format, cfg.Output)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(cfg.Output, opts)
	default:
		handler = NewColorfulHandler(cfg.Output, opts)
	}

	return slog.New(handler)
}

func parseLevel(s string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug", "dbg":
		return slog.LevelDebug
	case "info", "inf":
		return slog.LevelInfo
	case "warn", "warning", "wrn":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	case "fatal", "ftl":
		return LevelFatal
	default:
		return slog.LevelInfo
	}
}

func resolveFormat(format string, w io.Writer) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return "json"
	case "text", "color", "colored", "pretty":
		return "text"
	case "auto":
		return detectFormat(w)
	default:
		return "text"
	}
}

func detectFormat(w io.Writer) string {
	if f, ok := w.(*os.File); ok {
		info, err := f.Stat()
		if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			return "text"
		}
	}
	return "json"
}

func L() *slog.Logger {
	if Default == nil {
		Default = New(Config{Level: "info", Format: "auto", Output: os.Stderr})
	}
	return Default
}

func SetDefault(l *slog.Logger) {
	Default = l
	slog.SetDefault(l)
}
