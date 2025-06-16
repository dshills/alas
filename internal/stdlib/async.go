package stdlib

import (
	"context"
	"fmt"
	"time"

	"github.com/dshills/alas/internal/runtime"
)

// RegisterAsyncFunctions registers all async-related standard library functions
func (r *Registry) registerAsyncFunctions() {
	// async.spawn - Spawn an async task
	r.Register("async.spawn", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.spawn: expected 1 argument, got %d", len(args))
		}
		
		// For now, we'll store the function value and execute it later
		// In a full implementation, this would need to handle function values properly
		fn := args[0]
		
		// Create a task that executes the function
		task := runtime.GetGlobalAsyncManager().SpawnTask(func(ctx context.Context) (runtime.Value, error) {
			// TODO: Execute the ALaS function here
			// For now, simulate async work
			select {
			case <-ctx.Done():
				return runtime.NewVoid(), ctx.Err()
			case <-time.After(100 * time.Millisecond):
				// Return the function argument as a placeholder
				return fn, nil
			}
		})
		
		return task.ToValue(), nil
	})
	
	// async.await - Wait for task completion
	r.Register("async.await", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.await: expected 1 argument, got %d", len(args))
		}
		
		task, err := runtime.TaskFromValue(args[0])
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("async.await: %v", err)
		}
		
		value, err := task.Wait()
		
		// Return a Result type
		result := make(map[string]runtime.Value)
		if err != nil {
			result["ok"] = runtime.NewBool(false)
			result["value"] = runtime.NewVoid()
			result["error"] = runtime.NewString(err.Error())
		} else {
			result["ok"] = runtime.NewBool(true)
			result["value"] = value
			result["error"] = runtime.NewString("")
		}
		
		return runtime.NewMap(result), nil
	})
	
	// async.awaitTimeout - Wait for task with timeout
	r.Register("async.awaitTimeout", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 2 {
			return runtime.NewVoid(), fmt.Errorf("async.awaitTimeout: expected 2 arguments, got %d", len(args))
		}
		
		task, err := runtime.TaskFromValue(args[0])
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("async.awaitTimeout: %v", err)
		}
		
		if args[1].Type != runtime.ValueTypeInt {
			return runtime.NewVoid(), fmt.Errorf("async.awaitTimeout: timeout must be int")
		}
		
		timeoutMs, _ := args[1].AsInt()
		timeout := time.Duration(timeoutMs) * time.Millisecond
		value, err, timedOut := task.WaitTimeout(timeout)
		
		// Return a Result type with timeout flag
		result := make(map[string]runtime.Value)
		result["timedOut"] = runtime.NewBool(timedOut)
		
		if timedOut {
			result["ok"] = runtime.NewBool(false)
			result["value"] = runtime.NewVoid()
			result["error"] = runtime.NewString("timeout")
		} else if err != nil {
			result["ok"] = runtime.NewBool(false)
			result["value"] = runtime.NewVoid()
			result["error"] = runtime.NewString(err.Error())
		} else {
			result["ok"] = runtime.NewBool(true)
			result["value"] = value
			result["error"] = runtime.NewString("")
		}
		
		return runtime.NewMap(result), nil
	})
	
	// async.parallel - Run tasks in parallel, wait for all
	r.Register("async.parallel", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.parallel: expected 1 argument, got %d", len(args))
		}
		
		if args[0].Type != runtime.ValueTypeArray {
			return runtime.NewVoid(), fmt.Errorf("async.parallel: expected array of tasks")
		}
		
		taskValues, _ := args[0].AsArray()
		tasks := make([]*runtime.Task, len(taskValues))
		
		// Extract all tasks
		for i, tv := range taskValues {
			task, err := runtime.TaskFromValue(tv)
			if err != nil {
				return runtime.NewVoid(), fmt.Errorf("async.parallel: task %d: %v", i, err)
			}
			tasks[i] = task
		}
		
		// Wait for all tasks
		values := make([]runtime.Value, len(tasks))
		errors := make([]runtime.Value, len(tasks))
		allOk := true
		
		for i, task := range tasks {
			value, err := task.Wait()
			if err != nil {
				allOk = false
				values[i] = runtime.NewVoid()
				errors[i] = runtime.NewString(err.Error())
			} else {
				values[i] = value
				errors[i] = runtime.NewString("")
			}
		}
		
		// Return result
		result := make(map[string]runtime.Value)
		result["ok"] = runtime.NewBool(allOk)
		result["values"] = runtime.NewArray(values)
		result["errors"] = runtime.NewArray(errors)
		
		return runtime.NewMap(result), nil
	})
	
	// async.race - Run tasks in parallel, return first completion
	r.Register("async.race", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.race: expected 1 argument, got %d", len(args))
		}
		
		if args[0].Type != runtime.ValueTypeArray {
			return runtime.NewVoid(), fmt.Errorf("async.race: expected array of tasks")
		}
		
		taskValues, _ := args[0].AsArray()
		if len(taskValues) == 0 {
			return runtime.NewVoid(), fmt.Errorf("async.race: empty task array")
		}
		
		tasks := make([]*runtime.Task, len(taskValues))
		
		// Extract all tasks
		for i, tv := range taskValues {
			task, err := runtime.TaskFromValue(tv)
			if err != nil {
				return runtime.NewVoid(), fmt.Errorf("async.race: task %d: %v", i, err)
			}
			tasks[i] = task
		}
		
		// Create channels for racing
		type raceResult struct {
			index int
			value runtime.Value
			err   error
		}
		
		resultChan := make(chan raceResult, len(tasks))
		
		// Start goroutines for each task
		for i, task := range tasks {
			go func(index int, t *runtime.Task) {
				value, err := t.Wait()
				resultChan <- raceResult{index: index, value: value, err: err}
			}(i, task)
		}
		
		// Wait for the first result
		firstResult := <-resultChan
		
		// Return result
		result := make(map[string]runtime.Value)
		result["winner"] = runtime.NewInt(int64(firstResult.index))
		
		if firstResult.err != nil {
			result["ok"] = runtime.NewBool(false)
			result["value"] = runtime.NewVoid()
			result["error"] = runtime.NewString(firstResult.err.Error())
		} else {
			result["ok"] = runtime.NewBool(true)
			result["value"] = firstResult.value
			result["error"] = runtime.NewString("")
		}
		
		return runtime.NewMap(result), nil
	})
	
	// async.sleep - Sleep for specified milliseconds
	r.Register("async.sleep", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.sleep: expected 1 argument, got %d", len(args))
		}
		
		if args[0].Type != runtime.ValueTypeInt {
			return runtime.NewVoid(), fmt.Errorf("async.sleep: duration must be int")
		}
		
		durationMs, _ := args[0].AsInt()
		duration := time.Duration(durationMs) * time.Millisecond
		
		// Create a task that sleeps
		task := runtime.GetGlobalAsyncManager().SpawnTask(func(ctx context.Context) (runtime.Value, error) {
			select {
			case <-ctx.Done():
				return runtime.NewVoid(), ctx.Err()
			case <-time.After(duration):
				return runtime.NewVoid(), nil
			}
		})
		
		return task.ToValue(), nil
	})
	
	// async.timeout - Run function with timeout
	r.Register("async.timeout", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 2 {
			return runtime.NewVoid(), fmt.Errorf("async.timeout: expected 2 arguments, got %d", len(args))
		}
		
		// First arg is function, second is timeout
		fn := args[0]
		
		if args[1].Type != runtime.ValueTypeInt {
			return runtime.NewVoid(), fmt.Errorf("async.timeout: timeout must be int")
		}
		
		timeoutMs, _ := args[1].AsInt()
		timeout := time.Duration(timeoutMs) * time.Millisecond
		
		// Create a task with timeout
		task := runtime.GetGlobalAsyncManager().SpawnTask(func(ctx context.Context) (runtime.Value, error) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			
			// TODO: Execute the ALaS function here with timeout context
			// For now, simulate work
			select {
			case <-ctx.Done():
				return runtime.NewVoid(), ctx.Err()
			case <-time.After(timeout / 2): // Simulate some work
				return fn, nil
			}
		})
		
		return task.ToValue(), nil
	})
	
	// async.cancel - Cancel a running task
	r.Register("async.cancel", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.cancel: expected 1 argument, got %d", len(args))
		}
		
		task, err := runtime.TaskFromValue(args[0])
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("async.cancel: %v", err)
		}
		
		cancelled := task.Cancel()
		return runtime.NewBool(cancelled), nil
	})
	
	// async.isRunning - Check if task is running
	r.Register("async.isRunning", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.isRunning: expected 1 argument, got %d", len(args))
		}
		
		task, err := runtime.TaskFromValue(args[0])
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("async.isRunning: %v", err)
		}
		
		return runtime.NewBool(task.IsRunning()), nil
	})
	
	// async.isCompleted - Check if task is completed
	r.Register("async.isCompleted", func(args []runtime.Value) (runtime.Value, error) {
		if len(args) != 1 {
			return runtime.NewVoid(), fmt.Errorf("async.isCompleted: expected 1 argument, got %d", len(args))
		}
		
		task, err := runtime.TaskFromValue(args[0])
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("async.isCompleted: %v", err)
		}
		
		return runtime.NewBool(task.IsCompleted()), nil
	})
}