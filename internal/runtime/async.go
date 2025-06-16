package runtime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskStatus represents the status of an async task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCanceled TaskStatus = "canceled"
)

// Task represents an async task in ALaS.
type Task struct {
	ID       string
	Status   TaskStatus
	Result   Value
	Error    error
	cancel   context.CancelFunc
	done     chan struct{}
	mu       sync.RWMutex
}

// AsyncManager manages async tasks.
type AsyncManager struct {
	tasks    map[string]*Task
	mu       sync.RWMutex
	idCounter atomic.Uint64
}

var (
	globalAsyncManager = NewAsyncManager()
)

// NewAsyncManager creates a new async manager.
func NewAsyncManager() *AsyncManager {
	return &AsyncManager{
		tasks: make(map[string]*Task),
	}
}

// GetGlobalAsyncManager returns the global async manager instance.
func GetGlobalAsyncManager() *AsyncManager {
	return globalAsyncManager
}

// generateTaskID generates a unique task ID
func (am *AsyncManager) generateTaskID() string {
	id := am.idCounter.Add(1)
	return fmt.Sprintf("task_%d_%d", time.Now().UnixNano(), id)
}

// SpawnTask spawns a new async task
func (am *AsyncManager) SpawnTask(fn func(context.Context) (Value, error)) *Task {
	taskID := am.generateTaskID()
	ctx, cancel := context.WithCancel(context.Background())
	
	task := &Task{
		ID:     taskID,
		Status: TaskStatusPending,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	
	am.mu.Lock()
	am.tasks[taskID] = task
	am.mu.Unlock()
	
	// Start the task in a goroutine
	go func() {
		task.mu.Lock()
		task.Status = TaskStatusRunning
		task.mu.Unlock()
		
		// Execute the function
		result, err := fn(ctx)
		
		task.mu.Lock()
		if err != nil {
			task.Status = TaskStatusFailed
			task.Error = err
		} else if ctx.Err() != nil {
			task.Status = TaskStatusCanceled
			task.Error = ctx.Err()
		} else {
			task.Status = TaskStatusCompleted
			task.Result = result
		}
		task.mu.Unlock()
		
		close(task.done)
	}()
	
	return task
}

// GetTask retrieves a task by ID
func (am *AsyncManager) GetTask(taskID string) (*Task, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()
	task, ok := am.tasks[taskID]
	return task, ok
}

// Wait waits for a task to complete
func (t *Task) Wait() (Value, error) {
	<-t.done
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Result, t.Error
}

// WaitTimeout waits for a task to complete with timeout
func (t *Task) WaitTimeout(timeout time.Duration) (Value, bool, error) {
	select {
	case <-t.done:
		t.mu.RLock()
		defer t.mu.RUnlock()
		return t.Result, false, t.Error
	case <-time.After(timeout):
		return NewVoid(), true, nil
	}
}

// Cancel cancels a running task
func (t *Task) Cancel() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.Status == TaskStatusRunning || t.Status == TaskStatusPending {
		t.cancel()
		return true
	}
	return false
}

// IsRunning checks if a task is running
func (t *Task) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == TaskStatusRunning
}

// IsCompleted checks if a task is completed (success, failed, or cancelled)
func (t *Task) IsCompleted() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed || t.Status == TaskStatusCanceled
}

// ToValue converts a Task to a runtime Value (map)
func (t *Task) ToValue() Value {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	taskMap := make(map[string]Value)
	taskMap["type"] = NewString("task")
	taskMap["id"] = NewString(t.ID)
	taskMap["status"] = NewString(string(t.Status))
	
	return NewMap(taskMap)
}

// TaskFromValue extracts task information from a Value
func TaskFromValue(v Value) (*Task, error) {
	if v.Type != ValueTypeMap {
		return nil, fmt.Errorf("expected map for task, got %v", v.Type)
	}
	
	taskMap, err := v.AsMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get task map: %v", err)
	}
	
	// Extract task ID
	idValue, ok := taskMap["id"]
	if !ok || idValue.Type != ValueTypeString {
		return nil, fmt.Errorf("task missing valid id field")
	}
	
	taskID, err := idValue.AsString()
	if err != nil {
		return nil, fmt.Errorf("failed to get task id: %v", err)
	}
	
	// Get the task from the manager
	task, ok := globalAsyncManager.GetTask(taskID)
	if !ok {
		return nil, fmt.Errorf("task with id %s not found", taskID)
	}
	
	return task, nil
}