// +build windows

package log


import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)


type syslogZapLogInit struct {
}

func (self *syslogZapLogInit) loginit(config *logConf) (zap.AtomicLevel, *zap.Logger, error) {

	var (
		zapconfig zap.Config
		llevel    zap.AtomicLevel
		lzaplog   *zap.Logger
		err       error
	)
	if config.testenv {
		zapconfig = zap.NewDevelopmentConfig()
	} else {
		zapconfig = zap.NewProductionConfig()
	}
	zapconfig.OutputPaths =[]string{"stdout"}
	zapconfig.ErrorOutputPaths =[]string{"stderr"}
	zapconfig.Encoding = config.encodeing
	zapconfig.DisableStacktrace = true
	zapconfig.EncoderConfig.TimeKey = "timestamp"                   //"@timestamp"
	zapconfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder //epochSecondTimeEncoder //RFC3339TimeEncoder
	lzaplog, err = zapconfig.Build()
	llevel = zapconfig.Level
	return llevel, lzaplog, err
}

