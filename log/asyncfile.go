package log

import (
	"fmt"
	"git.yayafish.com/nbagent/log/asyncfile"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

//
// Core for syslog
//

type filecore struct {
	zapcore.LevelEnabler

	encoder zapcore.Encoder
	logfile  *asyncLog.LogFile

	fields []zapcore.Field
}

func newFileCore(enab zapcore.LevelEnabler, encoder zapcore.Encoder, logfile *asyncLog.LogFile) *filecore {
	return &filecore{
		LevelEnabler: enab,
		encoder:      encoder,
		logfile:      logfile,
	}
}

func (core *filecore) With(fields []zapcore.Field) zapcore.Core {
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

func (core *filecore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}
	return checked
}

func (core *filecore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Generate the message.
	buffer, err := core.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return errors.Wrap(err, "failed to encode log entry")
	}
	_ = core.logfile.Write(buffer.String())

	return nil
}

func (core *filecore) Sync() error {
	return nil
}


type asyncFileZapLogInit struct {
}

func (self *asyncFileZapLogInit) loginit(config *logConf) (zap.AtomicLevel, *zap.Logger, error) {
	var (
		llevel  zap.AtomicLevel
		lzaplog *zap.Logger
	)

	logfilename :=config.processName+".log"
	if config.logFilePath !="" {
		if _, err := os.Stat(config.logFilePath); os.IsNotExist(err) {
			// create path
			err = os.MkdirAll(config.logFilePath,os.ModePerm)
			if err != nil {
				fmt.Println("create log path err",config.logFilePath)
				return llevel,lzaplog,err
			}

		}
		logfilename = config.logFilePath+"/"+logfilename
	}

	lf := asyncLog.NewLogFile(logfilename)

	switch config.logFileRotate {
	case "date":
		lf.SetRotate(asyncLog.RotateDate)
	case "hour":
		lf.SetRotate(asyncLog.RotateHour)
	}
	lf.SetFlags(asyncLog.NoFlag) // 不输出时间
	lf.SetNewLineStr("") // 去掉换行符
	//lf.SetUseCache(false)

	// Initialize Zap.
	encconf := zap.NewProductionEncoderConfig()
	encconf.TimeKey = "timestamp"             //"@timestamp"
	encconf.EncodeTime = epochFullTimeEncoder //epochSecondTimeEncoder //RFC3339TimeEncoder
	var encoder zapcore.Encoder
	if config.encodeing=="console" {
		encoder = zapcore.NewConsoleEncoder(encconf)
	} else {
		encoder = zapcore.NewJSONEncoder(encconf)
	}
	if config.testenv {
		llevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		llevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	core := newFileCore(llevel, encoder, lf)

	lzaplog = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.DPanicLevel))

	return llevel, lzaplog, nil
}