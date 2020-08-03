package log

import (
	"os"
	"path"
)

const logpath = "/log"

type logConf struct {
	testenv         bool
	processName     string
	withPid         bool
	logPath         string
	listenAddr      string
	encodeing       string // json console [todo add more]
	targetName      string // stdout syslog asyncfile ...
	logFilePath     string // default "../log"
	logFileRotate   string // default no option [date,hour]
	HostName        string
	ElkTemplateName string //区分不同业务
}

var defaultLogOptions logConf = logConf{
	testenv:         false,
	processName:     path.Base(os.Args[0]),
	withPid:         true,
	logPath:      logpath,
	listenAddr:      "127.0.0.1:0",
	encodeing:  "json",
	targetName :  "syslog",
	logFilePath : "../log",
	ElkTemplateName: path.Base(os.Args[0]),
}

type logOption interface {
	apply(*logConf)
}

type logOptionFunc func(*logConf)

func (self logOptionFunc) apply(option *logConf) {
	self(option)
}

// ListenAddr设置logserver的http端口,用来管理日志级别。
// 默认监听127.0.0.1下的随机端口
func ListenAddr(addr string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.listenAddr = addr
	})
}

// LogApiPath设置logserver的api名字。
// 默认为 /log。
func LogApiPath(apipath string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.logPath = apipath
	})
}
// stdout syslog asyncfile
func SetTarget(name string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.targetName = name
	})
}

// need logfile when asyncfile
func LogFilePath(path string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.logFilePath = path
	})
}

func LogFileRotate(r string) logOption {
	return logOptionFunc(func(option *logConf){
		if r=="hour" || r=="date" {
			option.logFileRotate = r
		}
	})
}

// json console
func SetEncode(enc string) logOption {
	return logOptionFunc(func(option *logConf) {
			option.encodeing = enc
	})
}

// WithPid设置日志输出中是否加入pid的项。
// 默认为true。
func WithPid(yes bool) logOption {
	return logOptionFunc(func(option *logConf) {
		option.withPid = yes
	})
}

// ProcessName设置输出的进程名字。
// 默认去当前执行文件的名字。
func ProcessName(pname string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.processName = pname
	})
}

// TestEnv设置是否测试环境。
// 默认为true,测试环境。
func TestEnv(yes bool) logOption {
	return logOptionFunc(func(option *logConf) {
		option.testenv = yes
	})
}

// HostName设置日志机器的ip地址,方便定位。
// 默认不输出。
func HostName(hostname string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.HostName = hostname
	})
}

// ElkTmeplateName用来设置输出到elk中的唯一名字。
// 默认不输出。
func ElkTmeplateName(name string) logOption {
	return logOptionFunc(func(option *logConf) {
		option.ElkTemplateName = name
	})
}