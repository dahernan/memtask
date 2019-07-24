# Memtask

Simple async in memory task execution.

Ideal for small project or webapp.

Usage:

```
    m := memtask.NewManager(5 * time.Minute)
	id := m.Run(ctx, func(ctx context.Context, task Task) error {
		// do some work
        time.Sleep(1 * time.Second)
        // store the results of the work
		task.Data = "raw data"
		task.Store()
		return nil
	})

    // check the results
    task, ok := m.Get(id)
    if !ok {...}
    if !task.IsFinished() {
        // not finished
    }    
    // results here
    task.Data

```