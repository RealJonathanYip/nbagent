package taskworker

import (
	"context"
	"fmt"
	"git.yayafish.com/nbagent/log"
	"golang.org/x/sync/semaphore"
	"runtime"
	"strings"
)

type WorkerConfig struct {
	Name  string `xml:"name,attr"`
	Size  int    `xml:"size,attr"`
}

type WorkerConfigs struct {
	Workers []WorkerConfig `xml:"workers>worker"`
}

type TaskFunc func(context.Context) int

type Task struct {
	f    TaskFunc
	ctx  context.Context
	done chan int
}

type TaskWorker struct {
	config    WorkerConfig
	ptrWeight *semaphore.Weighted
}

func (t TaskWorker) id() string {
	return fmt.Sprintf("worker_%v", t.config.Name)
}

func NewTaskWorker(config WorkerConfig) *TaskWorker {
	t := &TaskWorker{
		config:    config,
		ptrWeight: semaphore.NewWeighted(int64(config.Size)),
	}

	return t
}

func (ptrTaskWorker *TaskWorker) GoTask(ctx context.Context, f TaskFunc) error {
	if anyErr := ptrTaskWorker.ptrWeight.Acquire(ctx, 1); anyErr != nil {
		log.Errorf("go task err! szTaskName:%s anyErr:%v", ptrTaskWorker.config.Name, anyErr)
		return anyErr
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				stackInfo := fmt.Sprintf("%s", buf[:n])
				stackInfo = strings.Replace(stackInfo, "\n", "\\n", -1)
				log.Errorf("%v panic %v %v", ptrTaskWorker.id(), err, stackInfo)
			}

			ptrTaskWorker.ptrWeight.Release(1)
		}()

		f(ctx);
	}()

	return nil
}

func (ptrTaskWorker *TaskWorker) DoTask(ctx context.Context, f TaskFunc) error {
	if anyErr := ptrTaskWorker.ptrWeight.Acquire(ctx, 1); anyErr != nil {
		log.Errorf("go task err! szTaskName:%s anyErr:%v", ptrTaskWorker.config.Name, anyErr)
		return anyErr
	}

	done := make(chan int, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				stackInfo := fmt.Sprintf("%s", buf[:n])
				stackInfo = strings.Replace(stackInfo, "\n", "\\n", -1)
				log.Errorf("%v panic %v %v", ptrTaskWorker.id(), err, stackInfo)
			}
			ptrTaskWorker.ptrWeight.Release(1)

			done <- 1
		}()

		f(ctx);
	}()

	<-done

	return nil
}
