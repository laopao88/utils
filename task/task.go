package task

import (
	"encoding/json"
	"github.com/laopao88/zaia/utils"
	"sync"
)

type Task[T any] struct {
	taskLocker  sync.RWMutex
	TaskMapList map[string]*T
	projectName string
}

func NewTask[T any]() *Task[T] {
	return &Task[T]{
		taskLocker:  sync.RWMutex{},
		TaskMapList: make(map[string]*T),
		projectName: "task.json",
	}
}

func (tc *Task[T]) Set(taskId string, t *T) {
	tc.taskLocker.Lock()
	defer tc.taskLocker.Unlock()
	tc.TaskMapList[taskId] = t
	tc.Dump()
}

func (tc *Task[T]) Get(taskId string) *T {
	tc.taskLocker.Lock()
	defer tc.taskLocker.Unlock()
	v, _ := tc.TaskMapList[taskId]
	if v != nil {
		return &*v
	}
	return nil
}

func (tc *Task[T]) Remove(taskId string) {
	tc.taskLocker.Lock()
	defer tc.taskLocker.Unlock()
	v, _ := tc.TaskMapList[taskId]
	if v != nil {
		delete(tc.TaskMapList, taskId)
	}
	tc.Dump()
}

func (tc *Task[T]) Dump() {
	utils.DumpInterface(tc.projectName, tc.TaskMapList)
}

func (tc *Task[T]) Load() {
	v := make(map[string]*T)
	b := utils.ReadFileToByte(tc.projectName)
	if b != nil {
		json.Unmarshal(b, &v)
		tc.TaskMapList = v
	}
}
