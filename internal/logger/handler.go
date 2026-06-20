package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"
	"unicode"
)

const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"
	italic  = "\033[3m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"
	gray    = "\033[90m"
	bgRed   = "\033[41m"
)

var levelStyles = map[slog.Leveler]struct {
	Label string
	Color string
}{
	slog.LevelDebug: {"DBG", magenta},
	slog.LevelInfo:  {"INF", green},
	slog.LevelWarn:  {"WRN", yellow},
	slog.LevelError: {"ERR", red},
	LevelFatal:      {"FTL", bgRed + white},
}

var levelOrder = []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError, LevelFatal}

type ColorfulHandler struct {
	opts    slog.HandlerOptions
	mu      sync.Mutex
	w       io.Writer
	attrs   []slog.Attr
	groups  []string
	once    sync.Once
}

func NewColorfulHandler(w io.Writer, opts *slog.HandlerOptions) *ColorfulHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	h := &ColorfulHandler{
		w:    w,
		opts: *opts,
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return h
}

func (h *ColorfulHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *ColorfulHandler) Handle(_ context.Context, r slog.Record) error {
	buf := make([]byte, 0, 256)

	buf = h.appendTime(buf, r.Time)

	buf = h.appendLevel(buf, r.Level)

	if h.opts.AddSource && r.PC != 0 {
		buf = h.appendSource(buf, r.PC)
	}

	buf = append(buf, ' ')
	buf = append(buf, []byte(bold+white)...)
	buf = append(buf, r.Message...)
	buf = append(buf, []byte(reset)...)

	if len(h.attrs) > 0 {
		buf = h.appendAttrs(buf, h.attrs)
	}

	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, a, h.groups, true)
		return true
	})

	buf = append(buf, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf)
	return err
}

func (h *ColorfulHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return &ColorfulHandler{
		opts:   h.opts,
		w:      h.w,
		attrs:  append(slices.Clone(h.attrs), attrs...),
		groups: slices.Clone(h.groups),
	}
}

func (h *ColorfulHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &ColorfulHandler{
		opts:   h.opts,
		w:      h.w,
		attrs:  slices.Clone(h.attrs),
		groups: append(slices.Clone(h.groups), name),
	}
}

func (h *ColorfulHandler) appendTime(buf []byte, t time.Time) []byte {
	buf = append(buf, []byte(dim)...)
	buf = t.AppendFormat(buf, "2006-01-02T15:04:05.000Z07:00")
	buf = append(buf, []byte(reset+"  ")...)
	return buf
}

func (h *ColorfulHandler) appendLevel(buf []byte, level slog.Level) []byte {
	// find best match
	style, ok := levelStyles[level]
	if !ok {
		for _, l := range levelOrder {
			if level >= l {
				if s, found := levelStyles[l]; found {
					style = s
					ok = true
				}
			}
		}
	}
	if !ok {
		style = struct {
			Label string
			Color string
		}{Label: "???", Color: white}
	}

	buf = append(buf, []byte(style.Color+bold)...)
	label := style.Label
	buf = append(buf, label...)
	if len(label) == 3 {
		buf = append(buf, ' ')
	}
	buf = append(buf, []byte(reset+"  ")...)
	return buf
}

func (h *ColorfulHandler) appendSource(buf []byte, pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return buf
	}
	file, line := fn.FileLine(pc)
	file = filepath.Base(file)
	buf = append(buf, ' ')
	buf = append(buf, []byte(gray+dim)...)
	buf = append(buf, file...)
	buf = append(buf, ':')
	buf = append(buf, strconv.Itoa(line)...)
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendAttrs(buf []byte, attrs []slog.Attr) []byte {
	for _, a := range attrs {
		buf = h.appendAttr(buf, a, h.groups, true)
	}
	return buf
}

func (h *ColorfulHandler) appendAttr(buf []byte, a slog.Attr, groups []string, topLevel bool) []byte {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return buf
	}

	if a.Value.Kind() == slog.KindGroup {
		return h.appendGroupAttr(buf, a, groups)
	}

	key := a.Key
	for _, g := range groups {
		key = g + "." + key
	}

	buf = append(buf, ' ')
	buf = append(buf, []byte(cyan)...)
	buf = append(buf, key...)
	buf = append(buf, []byte(reset+"=")...)
	buf = h.appendValue(buf, a.Value)
	return buf
}

func (h *ColorfulHandler) appendGroupAttr(buf []byte, a slog.Attr, groups []string) []byte {
	for _, ga := range a.Value.Group() {
		buf = h.appendAttr(buf, ga, append(groups, a.Key), false)
	}
	return buf
}

func (h *ColorfulHandler) appendValue(buf []byte, v slog.Value) []byte {
	switch v.Kind() {
	case slog.KindString:
		return h.appendString(buf, v.String())
	case slog.KindInt64:
		return h.appendInt64(buf, v.Int64())
	case slog.KindUint64:
		return h.appendUint64(buf, v.Uint64())
	case slog.KindFloat64:
		return h.appendFloat64(buf, v.Float64())
	case slog.KindBool:
		return h.appendBool(buf, v.Bool())
	case slog.KindTime:
		return h.appendTimeValue(buf, v.Time())
	case slog.KindDuration:
		return h.appendDuration(buf, v.Duration())
	case slog.KindGroup:
		return h.appendGroupValue(buf, v.Group())
	case slog.KindAny:
		return h.appendAny(buf, v.Any())
	default:
		return h.appendString(buf, v.String())
	}
}

func (h *ColorfulHandler) appendString(buf []byte, s string) []byte {
	buf = append(buf, []byte(yellow)...)
	needsQuote := false
	for _, r := range s {
		if !unicode.IsPrint(r) || r == '"' || r == '\\' {
			needsQuote = true
			break
		}
	}
	if needsQuote {
		buf = strconv.AppendQuote(buf, s)
	} else {
		buf = append(buf, s...)
	}
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendInt64(buf []byte, n int64) []byte {
	buf = append(buf, []byte(yellow)...)
	buf = strconv.AppendInt(buf, n, 10)
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendUint64(buf []byte, n uint64) []byte {
	buf = append(buf, []byte(yellow)...)
	buf = strconv.AppendUint(buf, n, 10)
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendFloat64(buf []byte, f float64) []byte {
	buf = append(buf, []byte(yellow)...)
	buf = strconv.AppendFloat(buf, f, 'f', -1, 64)
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendBool(buf []byte, b bool) []byte {
	buf = append(buf, []byte(yellow)...)
	if b {
		buf = append(buf, "true"...)
	} else {
		buf = append(buf, "false"...)
	}
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendTimeValue(buf []byte, t time.Time) []byte {
	buf = append(buf, []byte(yellow)...)
	buf = t.AppendFormat(buf, "2006-01-02T15:04:05.000Z07:00")
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendDuration(buf []byte, d time.Duration) []byte {
	buf = append(buf, []byte(yellow)...)
	buf = append(buf, d.String()...)
	buf = append(buf, []byte(reset)...)
	return buf
}

func (h *ColorfulHandler) appendGroupValue(buf []byte, attrs []slog.Attr) []byte {
	buf = append(buf, '{')
	for i, a := range attrs {
		if i > 0 {
			buf = append(buf, ' ')
		}
		buf = append(buf, []byte(cyan)...)
		buf = append(buf, a.Key...)
		buf = append(buf, []byte(reset+"=")...)
		buf = h.appendValue(buf, a.Value)
	}
	buf = append(buf, '}')
	return buf
}

func (h *ColorfulHandler) appendAny(buf []byte, v any) []byte {
	switch val := v.(type) {
	case error:
		buf = append(buf, []byte(red)...)
		buf = append(buf, val.Error()...)
		buf = append(buf, []byte(reset)...)
	case fmt.Stringer:
		buf = h.appendString(buf, val.String())
	default:
		buf = h.appendString(buf, fmt.Sprintf("%+v", v))
	}
	return buf
}
