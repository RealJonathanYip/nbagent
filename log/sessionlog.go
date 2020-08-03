package log

import (
	"context"
	"go.uber.org/zap/zapcore"
)

type logKey struct{}

type logVal struct {
	fields []zapcore.Field
}

// start a session log
func LogStart(ctx context.Context, fields ...zapcore.Field) context.Context {
	slog, ok := ctx.Value(logKey{}).(*logVal)
	if ok {
		if len(fields) > 0 {
			slog.fields = append(slog.fields, fields...)
		}
	} else {
		slog = &logVal{make([]zapcore.Field, 0)}
		if len(fields) > 0 {
			slog.fields = append(slog.fields, fields...)
		}
		ctx = context.WithValue(ctx, logKey{}, slog)
	}
	return ctx
}

// append a session log
func LogAppend(ctx context.Context, fields ...zapcore.Field) {
	slog, ok := ctx.Value(logKey{}).(*logVal)
	if ok {
		if len(fields) > 0 {
			slog.fields = append(slog.fields, fields...)
		}
	} else {
		// should start first
	}

}

// flush a session log
func LogFlush(ctx context.Context, mkey string, fields ...zapcore.Field) {
	slog, ok := ctx.Value(logKey{}).(*logVal)
	if ok {
		if len(fields) > 0 {
			slog.fields = append(slog.fields, fields...)
		}
	} else {
		slog = &logVal{make([]zapcore.Field, 0)}
		if len(fields) > 0 {
			slog.fields = append(slog.fields, fields...)
		}
	}
	defaultLog.writelog("info", mkey, slog.fields...)
}
