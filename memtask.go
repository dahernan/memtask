// Simple memory task manager, to run task async on memory and track the status of the task
// the tasks are deleted from memory when they finished after the expiration time
// so you can store the results of the task to be collected
package memtask

import (
	"context"
	"crypto/rand"
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	TaskStatusProcessing = "processing"
	TaskStatusComplete   = "complete"
	TaskStatusFailed     = "failed"
)

type Manager struct {
	tasks      sync.Map
	expireTime time.Duration
}

type Task struct {
	ID           string      `json:"id,omitempty"`
	Status       string      `json:"status,omitempty"`
	Err          error       `json:"-"`
	ErrorMessage string      `json:"error"`
	Started      time.Time   `json:"started,omitempty"`
	Finished     time.Time   `json:"finished,omitempty"`
	Data         interface{} `json:"-"`

	manager *Manager
}

func NewManager(expireTime time.Duration) *Manager {
	return &Manager{
		expireTime: expireTime,
	}
}

func (m *Manager) Run(ctx context.Context, fn func(ctx context.Context, task Task) error) string {
	id := timestampShortUUID()
	task := Task{
		ID:      id,
		Status:  TaskStatusProcessing,
		Started: time.Now(),
		manager: m,
	}
	m.Store(task)
	go func() {
		err := fn(ctx, task)
		// refresh the task (it may be changes)
		task, ok := m.Get(task.ID)
		if !ok {
			return
		}
		task.Status = TaskStatusComplete
		task.Finished = time.Now()
		if err != nil {
			task.Err = err
			task.ErrorMessage = err.Error()
			task.Status = TaskStatusFailed
		}
		m.Store(task)
	}()
	return id
}

func (m *Manager) Get(ID string) (Task, bool) {
	v, ok := m.tasks.Load(ID)
	if !ok {
		return Task{}, ok
	}
	return v.(Task), ok
}

func (m *Manager) Delete(ID string) {
	_, ok := m.tasks.Load(ID)
	if !ok {
		return
	}
	m.tasks.Delete(ID)
}

func (m *Manager) GetAll() []Task {
	var taskKeys []string
	m.tasks.Range(func(key, value interface{}) bool {
		taskKeys = append(taskKeys, key.(string))
		return true
	})
	tasks := []Task{}
	sort.Slice(taskKeys, func(i, j int) bool {
		return taskKeys[i] > taskKeys[j]
	})
	for i := range taskKeys {
		taskObj, ok := m.tasks.Load(taskKeys[i])
		if !ok {
			continue
		}
		task := taskObj.(Task)
		// if expired -> delete
		if task.Status != TaskStatusProcessing && task.Finished.Add(m.expireTime).Before(time.Now()) {
			m.tasks.Delete(task.ID)
		} else {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (m *Manager) Store(task Task) {
	m.tasks.Store(task.ID, task)
}

func (tk Task) Store() {
	tk.manager.Store(tk)
}

func (tk Task) IsFinished() bool {
	if tk.Status == TaskStatusComplete {
		return true
	}
	if tk.Status == TaskStatusFailed {
		return true
	}
	return false
}

func timestampShortUUID() string {
	now := uint32(time.Now().UTC().Unix())

	b := make([]byte, 4)
	count, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	if count != len(b) {
		panic("not enough random bytes")
	}
	return fmt.Sprintf("%08x%x", now, b)
}
