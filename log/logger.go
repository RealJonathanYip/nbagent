package log

import (
	"errors"
	"fmt"
	//"git.yayafish.com/knowledge/server/common/send_logmsg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	//"git.yayafish.com/common/log_watcher"
)

type nbloger struct {
	logger *zap.Logger
	Level  zap.AtomicLevel
}

const (
	LEVEL_DEBUG = "debug"
	LEVEL_INFO  = "info"
	LEVEL_WARN  = "warn"
	LEVEL_ERROR = "error"
	LEVEL_PANIC = "panic"
	LEVEL_FATAL = "fatal"
)

var (
	defaultLog *nbloger = nil
	stdlog     *log.Logger
	once       = &sync.Once{}
	//_aryWatcher []log_watcher.LogWatcher
)

// 启动后自动初始化日志对象
func init() {
	//初始化日志为stdout
	procname := path.Base(os.Args[0])
	InitLog(2, ProcessName(procname), SetTarget("stdout"))
	Info("logger init")
}

// 根据options的设置,再次初始化日志系统。
func InitLog(skip int,options ...logOption) error {

	var (
		err    error
		logger *zap.Logger
		level  zap.AtomicLevel
	)
	config := defaultLogOptions
	for _, option := range options {
		option.apply(&config)
	}

	if level, logger, err = zapLogInit(&config); err != nil {
		fmt.Printf("ZapLogInit err:%v", err)
		return err
	}

	logger = logger.WithOptions(zap.AddCallerSkip(skip))
	if defaultLog == nil {
		defaultLog = &nbloger{logger, level}
	} else {
		defaultLog.logger = logger
		defaultLog.Level = level
	}
    // redirect go log to defaultLog.logger
	zap.RedirectStdLog(defaultLog.logger)
    stdlog=zap.NewStdLog(defaultLog.logger)

	return nil
}

func GetLogger() *nbloger {
	return defaultLog
}

func GetStdLogger() *log.Logger{
	return stdlog
}

func (l *nbloger) GetZLog(opts ...zap.Option) *zap.Logger {
	return l.logger.WithOptions(opts...)
}

//func AddWatcher(ptrWatcher log_watcher.LogWatcher) {
//	_aryWatcher = append(_aryWatcher, ptrWatcher)
//}

// Example : GetLogger().Clone(zap.AddCallerSkip(2))
func (l *nbloger) Clone(opts ...zap.Option) *nbloger {
	nl := &nbloger{
		logger: l.logger,
		Level:  l.Level,
	}

	nl.logger = l.logger.WithOptions(opts...)
	return nl
}

// implement Write for io.Writer
func (l *nbloger) Write(p []byte) (n int, err error) {
	l.writelog("info", string(p))
	return len(p), nil
}

// implement Log for github.com/go-log/log.logger
func (l *nbloger) Log(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.writelog("info", msg)
}

// implement Logf for github.com/go-log/log.logger
func (l *nbloger) Logf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.writelog("info", msg)
}

func (l *nbloger) writelog(level, msg string, fields ...zapcore.Field) {

	switch level {
	case "info":
		l.logger.Info(msg, fields...)
	case "debug":
		l.logger.Debug(msg, fields...)
	case "warn":
		l.logger.Warn(msg, fields...)
	case "error":
		l.logger.Error(msg, fields...)
	case "panic":
		l.logger.Panic(msg, fields...)
	case "dpanic":
		l.logger.DPanic(msg, fields...)
	case "fatal":
		l.logger.Fatal(msg, fields...)
	default:
		l.logger.Info(msg, fields...)
	}

	//for _, ptrWatcher := range _aryWatcher {
	//	ptrWatcher.OnMessage(level, msg)
	//}
}

// level : debug info warn error panic fatal
func Log(level string, v ...interface{}) {
	msg := fmt.Sprint(v...)
	defaultLog.writelog(level, msg)
}

// level : debug info warn error panic fatal
func LogF(szLevel string, szFormat string, v ...interface{}) {
	funcName, _, _, ok := runtime.Caller(1)
	szFuncName := runtime.FuncForPC(funcName).Name()
	szFormatFinal := ""
	if ok {
		szFormatFinal = szFuncName + " " + szFormat
	}
	msg := fmt.Sprintf(szFormatFinal, v...)

	defaultLog.writelog(szLevel, msg)
}

func LogMyF(skip int, szLevel string, szFormat string, v ...interface{}) {
	funcName, _, _, ok := runtime.Caller(skip)
	szFuncName := runtime.FuncForPC(funcName).Name()
	szFormatFinal := ""
	if ok {
		strings.Split(szFuncName,"/")

		szFormatFinal = szFuncName + " " + szFormat
	}
	msg := fmt.Sprintf(szFormatFinal, v...)

	defaultLog.writelog(szLevel, msg)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("debug", msg, fields...)
}

func Debugf(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("debug", msg)
}


// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("info", msg, fields...)
}

func Infof(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("info", msg)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("warn", msg, fields...)
}

func Warningf(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("warn", msg)
}


// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("error", msg, fields...)
}

func Errorf(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("error", msg)
}


// DPanic logs a message at DPanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("dpanic", msg, fields...)
}


func DPanicf(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("dpanic", msg)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("panic", msg, fields...)
}

func Panicf(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("panic", msg)
}


// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is disabled.
func Fatal(msg string, fields ...zapcore.Field) {
	defaultLog.writelog("fatal", msg, fields...)
}

func Fatalf(szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog("fatal", msg)
}


func CloudLog(fields ...zapcore.Field) {
	defaultLog.writelog("info", "CloudLog", fields...)
}

func Sync() error {
	return defaultLog.logger.Sync()
}

func SetLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error", "fatal":
		level = strings.ToLower(level)
	case "all":
		level = "debug"
	case "off", "none":
		level = "fatal"
	default:
		return errors.New("not support level")
	}

	defaultLog.Level.UnmarshalText([]byte(level))

	return nil
}
