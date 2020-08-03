package taskworker

import (
	"context"
	"git.yayafish.com/nbagent/log"
	"testing"
	"time"
)

func init() {
	var aryConfig []WorkerConfig
	aryConfig = append(aryConfig, WorkerConfig{"test", 1})
	TaskWorkerManagerInstance().Init(aryConfig)
}

func TestGoTask(ptrTest *testing.T) {
	i := 0;
	_ = TaskWorkerManagerInstance().GoTask(context.TODO(), "test", func(ctx context.Context) int {
		for {
			i ++
			time.Sleep(time.Second)

			if i == 6 {
				break
			}
		}

		return i;
	})

	_ = TaskWorkerManagerInstance().GoTask(context.TODO(), "test", func(ctx context.Context) int {
		if i != 6 {
			ptrTest.Fatalf("worker go err!")
		}

		return 1;
	})
}

func TestPanic(ptrTest *testing.T) {
	bPass := false
	_ = TaskWorkerManagerInstance().GoTask(context.TODO(), "test", func(ctx context.Context) int {
		var a *int
		b := *a
		b++

		return b;
	})

	_ = TaskWorkerManagerInstance().DoTask(context.TODO(), "test", func(ctx context.Context) int {
		log.Info("success")
		bPass = true
		return 1;
	})


	if !bPass {
		ptrTest.Fatalf("panic not work")
	}
}
