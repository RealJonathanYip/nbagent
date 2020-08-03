// +build !windows,!nacl,!plan9

package log

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/syslog"
)

//
// Core for syslog
//

type core struct {
	zapcore.LevelEnabler

	encoder zapcore.Encoder
	writer  *syslog.Writer

	fields []zapcore.Field
}

func newCore(enab zapcore.LevelEnabler, encoder zapcore.Encoder, writer *syslog.Writer) *core {
	return &core{
		LevelEnabler: enab,
		encoder:      encoder,
		writer:       writer,
	}
}

func (core *core) With(fields []zapcore.Field) zapcore.Core {
	// Clone core.
	clone := *core

	// Clone encoder.
	clone.encoder = core.encoder.Clone()

	// append fields.
	for i := range fields {
		fields[i].AddTo(clone.encoder)
	}
	// Done.
	return &clone
}

func (core *core) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}
	return checked
}

func (core *core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Generate the message.
	buffer, err := core.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return errors.Wrap(err, "failed to encode log entry")
	}

	message := buffer.String()

	// Write the message.
	switch entry.Level {
	case zapcore.DebugLevel:
		return core.writer.Debug(message)

	case zapcore.InfoLevel:
		return core.writer.Info(message)

	case zapcore.WarnLevel:
		return core.writer.Warning(message)

	case zapcore.ErrorLevel:
		return core.writer.Err(message)

	case zapcore.DPanicLevel:
		return core.writer.Crit(message)

	case zapcore.PanicLevel:
		return core.writer.Crit(message)

	case zapcore.FatalLevel:
		return core.writer.Crit(message)

	default:
		return errors.Errorf("unknown log level: %v", entry.Level)
	}
}

func (core *core) Sync() error {
	return nil
}

type syslogZapLogInit struct {
}

func (self *syslogZapLogInit) loginit(config *logConf) (zap.AtomicLevel, *zap.Logger, error) {
	var (
		llevel  zap.AtomicLevel
		lzaplog *zap.Logger
	)
	//writer , err := syslog.Dial("udp","123.125.189.167:514",syslog.LOG_ERR|syslog.LOG_LOCAL0, config.processName)
	writer, err := syslog.New(syslog.LOG_ERR|syslog.LOG_LOCAL0, config.processName)
	if err != nil {
		//fmt.Println("syslog dial err",err)
		return llevel, lzaplog, err
	}
	// Initialize Zap.
	encconf := zap.NewProductionEncoderConfig()
	encconf.TimeKey = "timestamp"             //"@timestamp"
	encconf.EncodeTime = epochFullTimeEncoder //epochSecondTimeEncoder //RFC3339TimeEncoder
	var encoder zapcore.Encoder
	if config.encodeing == "console" {
		encoder = zapcore.NewConsoleEncoder(encconf)
	} else {
		encoder = zapcore.NewJSONEncoder(encconf)
	}
	if config.testenv {
		llevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		llevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	core := newCore(llevel, encoder, writer)

	lzaplog = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.DPanicLevel))

	return llevel, lzaplog, nil
}
