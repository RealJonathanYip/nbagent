package asyncLog

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
	"github.com/fsnotify/fsnotify"
	"fmt"
)

type asyncLogType struct {
	files map[string]*LogFile // 下标是文件名

	// 避免并发new对象
	sync.RWMutex
}

type LogFile struct {
	filename   string // 原始文件名（包含完整目录）
	flag       int    // 默认为log.LstdFlags
	newlinestr string // 换行符

	// 同步设置
	sync struct {
		duration  time.Duration // 同步数据到文件的周期，默认为1秒
		beginTime time.Time     // 开始同步的时间，判断同步的耗时
		status    syncStatus    // 同步状态
	}

	// 日志的等级
	level Priority

	// 缓存
	cache struct {
		use  bool     // 是否使用缓存

		data []string // 缓存数据
		mutex sync.Mutex // 写cache时的互斥锁

		datach  chan string // 缓存队列,无锁实现
	}

	// 文件切割
	logRotate struct {
		rotate LogRotate  // 默认按小时切割
		file   *os.File   // 文件操作对象
		fw     *fsnotify.Watcher
		suffix string     // 切割后的文件名后缀
		mutex  sync.Mutex // 文件名锁
	}

	// 日志写入概率
	probability float32
}

// log同步的状态
type syncStatus int

const (
	statusInit  syncStatus = iota // 初始状态
	statusDoing                   // 同步中
	statusDone                    // 同步已经完成
)

// 日志切割的方式
type LogRotate int

const (
	RotateNone LogRotate = iota // 不切割
	RotateHour                  // 按小时切割
	RotateDate                  // 按日期切割
)

const (
	// 写日志时前缀的时间格式
	// "2006-01-02T15:04:05Z07:00"
	logTimeFormat string = time.RFC3339

	// 文件写入mode
	fileOpenMode = 0666

	// 文件Flag
	fileFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND

	// 换行符
	newlineStr  = "\n"
	newlineChar = '\n'

	// 缓存切片的初始容量
	cacheInitCap = 128
)

// 是否需要Flag信息
const (
	NoFlag  = 0
	StdFlag = log.LstdFlags
)

// 异步日志变量
var asyncLog *asyncLogType

var nowFunc = time.Now

func init() {
	asyncLog = &asyncLogType{
		files: make(map[string]*LogFile),
	}

	timer := time.NewTicker(time.Millisecond * 100)
	//timer := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				asyncLog.RLock()
				for _, file := range asyncLog.files {
					if file.sync.status != statusDoing {
						go file.flush2()
					}
				}
				asyncLog.RUnlock()
			}
		}

	}()
}

func NewLogFile(filename string) *LogFile {
	asyncLog.Lock()
	defer asyncLog.Unlock()

	if lf, ok := asyncLog.files[filename]; ok {
		return lf
	}

	lf := &LogFile{
		filename: filename,
		flag:     StdFlag,
		newlinestr: newlineStr,
	}

	asyncLog.files[filename] = lf

	// 默认按小时切割文件
	lf.logRotate.rotate = RotateNone
	lf.logRotate.file =nil
	lf.logRotate.fw = nil

	// 默认开启缓存
	lf.cache.use = true
	lf.cache.datach = make(chan string ,100000)

	// 日志写入概率，默认为1.1, 就是全部写入
	lf.probability = 1.1


	// TODO 同步的时间周期，缓存开启才有效
	lf.sync.duration = time.Second
	return lf
}

func (lf *LogFile) SetFlags(flag int) {
	lf.flag = flag
}

func (lf *LogFile) SetRotate(rotate LogRotate) {
	lf.logRotate.rotate = rotate
}

func (lf *LogFile) SetUseCache(useCache bool) {
	lf.cache.use = useCache
}

func (lf *LogFile) SetProbability(probability float32) {
	lf.probability = probability
}

func (lf *LogFile) SetNewLineStr(str string) {
	lf.newlinestr = str
}

// Write 写缓存
func (lf *LogFile) Write(msg string) error {
	if lf.flag == StdFlag {
		msg = nowFunc().Format(logTimeFormat) + " " + msg + lf.newlinestr
	} else {
		msg = msg + lf.newlinestr
	}

	if lf.cache.use {
		lf.appendCache2(msg)
		return nil
	}

	return lf.directWrite([]byte(msg))
}

// WriteJson 写入json数据
func (lf *LogFile) WriteJson(data interface{}) error {
	if lf.probability < 1.0 && rand.Float32() > lf.probability {
		// 按照概率写入
		return nil
	}

	bts, err := json.Marshal(data)
	if err != nil {
		return err
	}
	bts = append(bts, newlineChar)

	if lf.cache.use {
		lf.appendCache2(string(bts))
		return nil
	}

	return lf.directWrite(bts)
}

//*********************** 以下是私有函数 ************************************

func (lf *LogFile) appendCache(msg string) {
	lf.cache.mutex.Lock()
	lf.cache.data = append(lf.cache.data, msg)
	lf.cache.mutex.Unlock()
}

func (lf *LogFile) appendCache2(msg string) {

	if len(lf.cache.datach) < cap(lf.cache.datach) {
		lf.cache.datach <- msg
	} else {
		// todo handle channel full
	}
}

// 同步缓存到文件中
func (lf *LogFile) flush() error {
	lf.sync.status = statusDoing
	defer func() {
		lf.sync.status = statusDone
	}()

	// 写入log文件
	file, err := lf.openFileNoCache()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 获取缓存数据
	lf.cache.mutex.Lock()
	cache := lf.cache.data
	lf.cache.data = make([]string, 0, cacheInitCap)
	lf.cache.mutex.Unlock()

	if len(cache) == 0 {
		return nil
	}

	_, err = file.WriteString(strings.Join(cache, ""))
	if err != nil {
		// 重试
		_, err = file.WriteString(strings.Join(cache, ""))
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func (lf *LogFile) flush2() error {
	lf.sync.status = statusDoing
	defer func() {
		lf.sync.status = statusDone
	}()

	// 写入log文件
	file, err := lf.openFileNoCache()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	count:=0
    for {
    	count++
    	if count>10000 {
    		return nil
		}
    	select{
    	case msg := <-lf.cache.datach:
			_,err = file.WriteString(msg)
			if err !=nil {
				return err
			}
		default:
			return nil
		}
	}
}

// 获取文件名的后缀
func (lf *LogFile) getFilenameSuffix() string {
	if lf.logRotate.rotate ==RotateNone {
		return ""
	}
	if lf.logRotate.rotate == RotateDate {
		return nowFunc().Format("20060102")
	}
	return nowFunc().Format("2006010215")
}

// 直接写入日志文件
func (lf *LogFile) directWrite(msg []byte) error {
	file, err := lf.openFile()
	if err != nil {
		panic(err)
	}

	lf.logRotate.mutex.Lock()
	_, err = file.Write(msg)
	lf.logRotate.mutex.Unlock()

	return err
}

// 打开日志文件
func (lf *LogFile) openFile() (*os.File, error) {
	suffix := lf.getFilenameSuffix()
	logFilename :=lf.filename
	if suffix !="" {
		logFilename = lf.filename + "." + suffix
	}

	lf.logRotate.mutex.Lock()
	defer lf.logRotate.mutex.Unlock()

	if suffix == lf.logRotate.suffix && lf.logRotate.file !=nil {
		// check file status events
		if lf.logRotate.fw !=nil {
			select {
			case event :=<-lf.logRotate.fw.Events:
				if event.Op != fsnotify.Create && event.Op !=fsnotify.Remove && event.Op != fsnotify.Rename {
					// file still exists
					return lf.logRotate.file, nil
				}
			default:
				return lf.logRotate.file, nil
			}
		} else {
			return lf.logRotate.file, nil
		}
	}

	file, err := os.OpenFile(logFilename, fileFlag, fileOpenMode)
	if err != nil {
		// 重试
		file, err = os.OpenFile(logFilename, fileFlag, fileOpenMode)
		if err != nil {
			return file, err
		}
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("new file status watcher err")
		return nil, err
	}
	fw.Add(file.Name())
    fmt.Println("open log file:",file.Name())

	// 关闭旧的文件
	if lf.logRotate.file != nil {
		lf.logRotate.file.Close()
		lf.logRotate.fw.Close()
	}

	lf.logRotate.file = file
	lf.logRotate.suffix = suffix
	lf.logRotate.fw = fw
	return file, nil
}

// 打开日志文件(不缓存句柄)
func (lf *LogFile) openFileNoCache() (*os.File, error) {
	logFilename :=lf.filename
	suff := lf.getFilenameSuffix()
	if suff !="" {
		logFilename = lf.filename + "." + suff
	}

	lf.logRotate.mutex.Lock()
	defer lf.logRotate.mutex.Unlock()

	file, err := os.OpenFile(logFilename, fileFlag, fileOpenMode)
	if err != nil {
		// 重试
		file, err = os.OpenFile(logFilename, fileFlag, fileOpenMode)
		if err != nil {
			return file, err
		}
	}

	return file, nil
}
