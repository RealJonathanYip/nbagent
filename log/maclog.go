package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)


type macZapLogInit struct {
}

func (self *macZapLogInit) loginit(config *logConf) (zap.AtomicLevel, *zap.Logger, error) {
	var (
		zapconfig zap.Config
		llevel    zap.AtomicLevel
		lzaplog   *zap.Logger
		err       error
	)

	zapconfig = zap.NewProductionConfig()
    zapconfig.Encoding = config.encodeing
	zapconfig.DisableStacktrace = true
	zapconfig.EncoderConfig.TimeKey = "timestamp"                   //"@timestamp"
	zapconfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder //epochSecondTimeEncoder //RFC3339TimeEncoder
	lzaplog, err = zapconfig.Build()
	llevel = zapconfig.Level
	return llevel, lzaplog, err
}

