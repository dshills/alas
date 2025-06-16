package stdlib

import (
	"strings"
	"testing"
	"time"

	"github.com/dshills/alas/internal/runtime"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestAsyncSpawn(t *testing.T) {
	registry := NewRegistry()
	
	// Test spawn
	args := []runtime.Value{
		runtime.NewString("test function"), // Placeholder for function
	}
	
	result, err := registry.Call("async.spawn", args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != runtime.ValueTypeMap {
		t.Errorf("expected map type, got %v", result.Type)
	}
	
	// Check task structure
	taskMap, _ := result.AsMap()
	typeVal, _ := taskMap["type"].AsString()
	if typeVal != "task" {
		t.Errorf("expected type 'task', got %v", typeVal)
	}
	idVal, _ := taskMap["id"].AsString()
	if idVal == "" {
		t.Error("expected non-empty task id")
	}
	status, _ := taskMap["status"].AsString()
	if status != "pending" && status != "running" {
		t.Errorf("expected status 'pending' or 'running', got %v", status)
	}
}

func TestAsyncAwait(t *testing.T) {
	registry := NewRegistry()
	
	// First spawn a task
	spawnArgs := []runtime.Value{
		runtime.NewString("test function"),
	}
	task, err := registry.Call("async.spawn", spawnArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Wait for the task
	awaitArgs := []runtime.Value{task}
	result, err := registry.Call("async.await", awaitArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != runtime.ValueTypeMap {
		t.Errorf("expected map type, got %v", result.Type)
	}
	
	// Check result structure
	resultMap, _ := result.AsMap()
	okVal, _ := resultMap["ok"].AsBool()
	if okVal != true {
		t.Errorf("expected ok=true, got %v", okVal)
	}
	valueStr, _ := resultMap["value"].AsString()
	if valueStr != "test function" {
		t.Errorf("expected value='test function', got %v", valueStr)
	}
	errStr, _ := resultMap["error"].AsString()
	if errStr != "" {
		t.Errorf("expected empty error, got %v", errStr)
	}
}

func TestAsyncAwaitTimeout(t *testing.T) {
	registry := NewRegistry()
	
	// Create a sleep task that takes 200ms
	sleepArgs := []runtime.Value{
		runtime.NewInt(200),
	}
	task, err := registry.Call("async.sleep", sleepArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Try to wait with 50ms timeout (should timeout)
	awaitArgs := []runtime.Value{
		task,
		runtime.NewInt(50),
	}
	result, err := registry.Call("async.awaitTimeout", awaitArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	resultMap, _ := result.AsMap()
	timedOut, _ := resultMap["timedOut"].AsBool()
	if timedOut != true {
		t.Errorf("expected timedOut=true, got %v", timedOut)
	}
	okVal, _ := resultMap["ok"].AsBool()
	if okVal != false {
		t.Errorf("expected ok=false, got %v", okVal)
	}
	errStr, _ := resultMap["error"].AsString()
	if errStr != "timeout" {
		t.Errorf("expected error='timeout', got %v", errStr)
	}
}

func TestAsyncSleep(t *testing.T) {
	registry := NewRegistry()
	
	start := time.Now()
	
	// Create sleep task
	sleepArgs := []runtime.Value{
		runtime.NewInt(100), // 100ms
	}
	task, err := registry.Call("async.sleep", sleepArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Wait for it
	awaitArgs := []runtime.Value{task}
	_, err = registry.Call("async.await", awaitArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Errorf("expected elapsed time >= 100ms, got %v", elapsed)
	}
}

func TestAsyncCancel(t *testing.T) {
	registry := NewRegistry()
	
	// Create a long-running sleep task
	sleepArgs := []runtime.Value{
		runtime.NewInt(1000), // 1 second
	}
	task, err := registry.Call("async.sleep", sleepArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Cancel it
	cancelArgs := []runtime.Value{task}
	cancelled, err := registry.Call("async.cancel", cancelArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancelledBool, _ := cancelled.AsBool()
	if cancelledBool != true {
		t.Errorf("expected cancelled=true, got %v", cancelledBool)
	}
	
	// Wait should return quickly with cancellation error
	awaitArgs := []runtime.Value{task}
	result, err := registry.Call("async.await", awaitArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	resultMap, _ := result.AsMap()
	okVal, _ := resultMap["ok"].AsBool()
	if okVal != false {
		t.Errorf("expected ok=false, got %v", okVal)
	}
	errStr, _ := resultMap["error"].AsString()
	if !contains(errStr, "context canceled") {
		t.Errorf("expected error to contain 'context canceled', got %v", errStr)
	}
}

func TestAsyncIsRunning(t *testing.T) {
	registry := NewRegistry()
	
	// Create a sleep task
	sleepArgs := []runtime.Value{
		runtime.NewInt(100),
	}
	task, err := registry.Call("async.sleep", sleepArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Check if running (might be pending or running)
	checkArgs := []runtime.Value{task}
	isRunning, err := registry.Call("async.isRunning", checkArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Task should be either pending or running
	
	// Wait for completion
	awaitArgs := []runtime.Value{task}
	_, err = registry.Call("async.await", awaitArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Now should not be running
	isRunning, err = registry.Call("async.isRunning", checkArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	isRunningBool, _ := isRunning.AsBool()
	if isRunningBool != false {
		t.Errorf("expected isRunning=false, got %v", isRunningBool)
	}
}

func TestAsyncIsCompleted(t *testing.T) {
	registry := NewRegistry()
	
	// Create a sleep task
	sleepArgs := []runtime.Value{
		runtime.NewInt(50),
	}
	task, err := registry.Call("async.sleep", sleepArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Initially not completed
	checkArgs := []runtime.Value{task}
	isCompleted, err := registry.Call("async.isCompleted", checkArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	isCompletedBool, _ := isCompleted.AsBool()
	if isCompletedBool != false {
		t.Errorf("expected isCompleted=false, got %v", isCompletedBool)
	}
	
	// Wait for completion
	awaitArgs := []runtime.Value{task}
	_, err = registry.Call("async.await", awaitArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Now should be completed
	isCompleted, err = registry.Call("async.isCompleted", checkArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	isCompletedBool, _ = isCompleted.AsBool()
	if isCompletedBool != true {
		t.Errorf("expected isCompleted=true, got %v", isCompletedBool)
	}
}

func TestAsyncParallel(t *testing.T) {
	registry := NewRegistry()
	
	// Create multiple tasks
	tasks := []runtime.Value{}
	for i := 0; i < 3; i++ {
		sleepArgs := []runtime.Value{
			runtime.NewInt(int64(50 * (i + 1))), // 50ms, 100ms, 150ms
		}
		task, err := registry.Call("async.sleep", sleepArgs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tasks = append(tasks, task)
	}
	
	// Run parallel
	start := time.Now()
	parallelArgs := []runtime.Value{
		runtime.NewArray(tasks),
	}
	result, err := registry.Call("async.parallel", parallelArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	
	// Should take about 150ms (the longest task), not 300ms (sum of all)
	if elapsed >= 200*time.Millisecond {
		t.Errorf("expected elapsed time < 200ms, got %v", elapsed)
	}
	
	resultMap, _ := result.AsMap()
	okVal, _ := resultMap["ok"].AsBool()
	if okVal != true {
		t.Errorf("expected ok=true, got %v", okVal)
	}
	valuesArray, _ := resultMap["values"].AsArray()
	if len(valuesArray) != 3 {
		t.Errorf("expected 3 values, got %v", len(valuesArray))
	}
}

func TestAsyncRace(t *testing.T) {
	registry := NewRegistry()
	
	// Create multiple tasks with different durations
	tasks := []runtime.Value{}
	durations := []int{100, 50, 150} // 50ms should win
	
	for _, duration := range durations {
		sleepArgs := []runtime.Value{
			runtime.NewInt(int64(duration)),
		}
		task, err := registry.Call("async.sleep", sleepArgs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tasks = append(tasks, task)
	}
	
	// Race them
	start := time.Now()
	raceArgs := []runtime.Value{
		runtime.NewArray(tasks),
	}
	result, err := registry.Call("async.race", raceArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	
	// Should complete in about 50ms (the fastest task)
	if elapsed >= 80*time.Millisecond {
		t.Errorf("expected elapsed time < 80ms, got %v", elapsed)
	}
	
	resultMap, _ := result.AsMap()
	okVal, _ := resultMap["ok"].AsBool()
	if okVal != true {
		t.Errorf("expected ok=true, got %v", okVal)
	}
	winnerInt, _ := resultMap["winner"].AsInt()
	if winnerInt != 1 {
		t.Errorf("expected winner=1, got %v", winnerInt)
	}
}