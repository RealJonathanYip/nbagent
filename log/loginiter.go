package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type zapLogIniter interface {
	loginit(conf *logConf) (zap.AtomicLevel, *zap.Logger, error)
}

func zapLogInit(config *logConf) (zap.AtomicLevel, *zap.Logger, error) {
	var (
		zapinit zapLogIniter
		level   zap.AtomicLevel
		llog    *zap.Logger
		err     error
	)

	if config.targetName=="syslog" {
		zapinit = &syslogZapLogInit{}
	} else if config.targetName=="asyncfile" {
		zapinit = &asyncFileZapLogInit{}
	} else  {
		zapinit = &stdZapLogInit{}
	}

	if level, llog, err = zapinit.loginit(config); err != nil {
		fmt.Printf("loginit err:%v", err)
		return level, llog, err
	}

	if config.withPid {
		llog = llog.With(zap.Int("pid", os.Getpid()))
	}
	if config.ElkTemplateName != "" {
		llog = llog.With(zap.String("procname", config.ElkTemplateName))
	}

	if config.HostName != "" {
		llog = llog.With(zap.String("hostname", config.HostName))
	}


	return level, llog, nil
}

func epochMillisTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	nanos := t.UnixNano()
	millis := nanos / int64(time.Millisecond)
	enc.AppendInt64(millis)
}

func epochSecondTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.Unix())
}

func epochFullTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}