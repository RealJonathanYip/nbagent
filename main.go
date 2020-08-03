package main

import (
	"math/rand"
	"os"
	"path"
	"time"

	"git.yayafish.com/nbagent/config"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/nbserver"
	"git.yayafish.com/nbagent/taskworker"
)

func init() {
	_ = log.InitLog(2, log.SetTarget(config.ServerConf.LogMode), log.SetEncode("json"),
		log.LogFilePath(config.ServerConf.LogPath+path.Base(os.Args[0])+"/"), log.LogFileRotate("date"))
	_ = log.SetLogLevel(config.ServerConf.LogLevel)

	rand.Seed(time.Now().UnixNano())
}

func main() {

	taskworker.TaskWorkerManagerInstance().Init(config.ServerConf.WorkerConfigs.Workers)

	stopCh := make(chan bool, 1)
	log.Infof("config: %+v", config.ServerConf)

	ptrNBServer := nbserver.NewNBServer(config.ServerConf)
	ptrNBServer.Init()
	ptrNBServer.Run(stopCh)
}
