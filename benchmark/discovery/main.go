package main

import (
	"fmt"
	"os"

	"git.yayafish.com/nbagent/benchmark/discovery/cmd"
	"git.yayafish.com/nbagent/log"
	"git.yayafish.com/nbagent/taskworker"
)

func init() {
	_ = log.SetLogLevel("warn")

	//<worker name="tcp_msg" size="100" queue="100"/>
	//<worker name="network_task_flag" size="100" queue="100"/>
	var aryWorker []taskworker.WorkerConfig
	aryWorker = append(aryWorker, taskworker.WorkerConfig{Name: "tcp_msg", Size: 1000})
	aryWorker = append(aryWorker, taskworker.WorkerConfig{Name: "network_task_flag", Size: 1000})
	taskworker.TaskWorkerManagerInstance().Init(aryWorker)
}

func Run(stopCh chan bool) {
	for {
		select {
		case <-stopCh:
			{
				return
			}
		}
	}
}

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	stopCh := make(chan bool, 1)
	Run(stopCh)
}
