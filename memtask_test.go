package memtask

import (
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
	"golang.org/x/net/context"
)

func TestRun(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	now := time.Now()
	m := NewManager(4 * time.Second)

	id := m.Run(ctx, func(ctx context.Context, task Task) error {
		time.Sleep(1 * time.Second)
		task.Data = "raw data"
		task.Store()
		return nil
	})

	task, ok := m.Get(id)
	is.True(ok)
	is.True(task.Started.After(now))

	// wait to finish
	wait(m, id)

	task, ok = m.Get(id)
	is.True(ok)
	is.Equal(task.Status, TaskStatusComplete)
	is.Equal(task.Err, nil)
	is.Equal(task.Data, "raw data")
	is.True(task.Finished.After(task.Started))

}

func TestError(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	now := time.Now()
	m := NewManager(4 * time.Second)

	id := m.Run(ctx, func(ctx context.Context, task Task) error {
		task.Data = "raw data"
		task.Store()
		return errors.New("some error here")
	})

	task, ok := m.Get(id)
	is.True(ok)
	is.True(task.Started.After(now))

	// wait to finish
	wait(m, id)

	task, ok = m.Get(id)
	is.True(ok)
	is.Equal(task.Status, TaskStatusFailed)
	is.Equal(task.Err.Error(), "some error here")
	is.Equal(task.Data, "raw data")
	is.True(task.Finished.After(task.Started))
}

func TestMultiRun(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	m := NewManager(1 * time.Second)

	fn := func(ctx context.Context, task Task) error {
		task.Data = task.ID
		task.Store()
		return nil
	}

	id1 := m.Run(ctx, fn)
	id2 := m.Run(ctx, fn)
	id3 := m.Run(ctx, fn)

	tasks := m.GetAll()
	is.Equal(len(tasks), 3)
	is.Equal(tasks[0].Status, TaskStatusProcessing)

	// wait to finish
	wait(m, id1)
	wait(m, id2)
	wait(m, id3)

	tasks = m.GetAll()
	is.Equal(len(tasks), 3)
	is.Equal(tasks[0].Status, TaskStatusComplete)
	is.Equal(tasks[0].Err, nil)
	is.Equal(tasks[0].Data, tasks[0].ID)

	// wait expiration
	time.Sleep(2 * time.Second)
	tasks = m.GetAll()
	is.Equal(len(tasks), 0)

}

// poll waiting for testing
func wait(m *Manager, taskID string) {
	for {
		task, ok := m.Get(taskID)
		if !ok {
			return
		}
		if task.IsFinished() {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
