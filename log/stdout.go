package log

import (
	"go.uber.org/zap"
)


type stdZapLogInit struct {
}

func (self *stdZapLogInit) loginit(config *logConf) (zap.AtomicLevel, *zap.Logger, error) {
	//fmt.Println("log init for stdout and encoding",config.encodeing)
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
	zapconfig.EncoderConfig.EncodeTime = epochFullTimeEncoder //epochSecondTimeEncoder //RFC3339TimeEncoder
	lzaplog, err = zapconfig.Build()
	llevel = zapconfig.Level
	return llevel, lzaplog, err
}

