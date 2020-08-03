package taskworker

import (
	"context"
	"fmt"
)

type TaskWorkerManager struct {
	taskworkers map[string]*TaskWorker
}

func (m *TaskWorkerManager) Init(configs []WorkerConfig) {
	m.taskworkers = make(map[string]*TaskWorker)

	for _, c := range configs {
		m.taskworkers[c.Name] = NewTaskWorker(c)
	}
}

func (m *TaskWorkerManager) GoTask(ctx context.Context, worker string, f TaskFunc) error {
	taskworker, ok := m.taskworkers[worker]
	if !ok {
		panic(fmt.Sprintf("taskworker: %v not found", worker))
	}

	return taskworker.GoTask(ctx, f)
}

func (m *TaskWorkerManager) DoTask(ctx context.Context, worker string, f TaskFunc) error {
	taskworker, ok := m.taskworkers[worker]
	if !ok {
		panic(fmt.Sprintf("taskworker: %v not found", worker))
	}

	return taskworker.DoTask(ctx, f)
}

var gTaskWorkerManager = &TaskWorkerManager{}

func TaskWorkerManagerInstance() *TaskWorkerManager {
	return gTaskWorkerManager
}
