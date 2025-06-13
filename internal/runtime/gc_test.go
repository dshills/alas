package runtime

import (
	"testing"
	"time"
)

func TestGCManager_AllocateArray(t *testing.T) {
	gc := NewGCManager()

	// Test array allocation
	values := []Value{NewInt(1), NewInt(2), NewInt(3)}
	obj, id := gc.AllocateArray(values)

	if obj == nil {
		t.Fatal("Expected non-nil GCObject")
	}
	if id == 0 {
		t.Fatal("Expected non-zero ObjectID")
	}
	if obj.ID != id {
		t.Errorf("Expected obj.ID (%d) to match returned ID (%d)", obj.ID, id)
	}
	if obj.Type != ObjectTypeArray {
		t.Errorf("Expected ObjectTypeArray, got %v", obj.Type)
	}
	if obj.RefCount != 1 {
		t.Errorf("Expected RefCount 1, got %d", obj.RefCount)
	}

	// Verify data
	arr, ok := obj.Data.([]Value)
	if !ok {
		t.Fatal("Expected []Value in obj.Data")
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arr))
	}
}

func TestGCManager_AllocateMap(t *testing.T) {
	gc := NewGCManager()

	// Test map allocation
	values := map[string]Value{
		"key1": NewString("value1"),
		"key2": NewInt(42),
	}
	obj, id := gc.AllocateMap(values)

	if obj == nil {
		t.Fatal("Expected non-nil GCObject")
	}
	if id == 0 {
		t.Fatal("Expected non-zero ObjectID")
	}
	if obj.Type != ObjectTypeMap {
		t.Errorf("Expected ObjectTypeMap, got %v", obj.Type)
	}
	if obj.RefCount != 1 {
		t.Errorf("Expected RefCount 1, got %d", obj.RefCount)
	}

	// Verify data
	m, ok := obj.Data.(map[string]Value)
	if !ok {
		t.Fatal("Expected map[string]Value in obj.Data")
	}
	if len(m) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(m))
	}
}

func TestGCManager_RetainRelease(t *testing.T) {
	gc := NewGCManager()

	// Allocate an array
	values := []Value{NewInt(1)}
	obj, id := gc.AllocateArray(values)

	// Initial ref count should be 1
	if obj.RefCount != 1 {
		t.Errorf("Expected initial RefCount 1, got %d", obj.RefCount)
	}

	// Retain should increase ref count
	gc.Retain(id)
	if obj.RefCount != 2 {
		t.Errorf("Expected RefCount 2 after retain, got %d", obj.RefCount)
	}

	// Release should decrease ref count
	gc.Release(id)
	if obj.RefCount != 1 {
		t.Errorf("Expected RefCount 1 after release, got %d", obj.RefCount)
	}

	// Object should still exist
	stats := gc.GetStats()
	if stats.TotalObjects != 1 {
		t.Errorf("Expected 1 object to exist, got %d", stats.TotalObjects)
	}

	// Final release should deallocate object
	gc.Release(id)

	// Object should be removed
	stats = gc.GetStats()
	if stats.TotalObjects != 0 {
		t.Errorf("Expected 0 objects after final release, got %d", stats.TotalObjects)
	}
}

func TestGCManager_NestedObjectRelease(t *testing.T) {
	gc := NewGCManager()

	// Create nested structure: array containing another GC array
	innerValues := []Value{NewInt(1), NewInt(2)}
	innerObj, innerID := gc.AllocateArray(innerValues)
	innerGCVal := &GCValue{ID: innerID, Object: innerObj}

	outerValues := []Value{
		NewInt(42),
		{Type: ValueTypeArray, Value: innerGCVal},
	}
	_, outerID := gc.AllocateArray(outerValues)

	// Should have 2 objects
	stats := gc.GetStats()
	if stats.TotalObjects != 2 {
		t.Errorf("Expected 2 objects, got %d", stats.TotalObjects)
	}

	// Release outer object - should also release inner object
	gc.Release(outerID)

	// Should have 0 objects (both should be deallocated)
	stats = gc.GetStats()
	if stats.TotalObjects != 0 {
		t.Errorf("Expected 0 objects after releasing outer, got %d", stats.TotalObjects)
	}
}

func TestGCManager_Stats(t *testing.T) {
	gc := NewGCManager()

	// Allocate some objects
	arrayValues := []Value{NewInt(1)}
	mapValues := map[string]Value{"key": NewString("value")}

	_, arrayID := gc.AllocateArray(arrayValues)
	_, mapID := gc.AllocateMap(mapValues)
	_, arrayID2 := gc.AllocateArray(arrayValues)

	stats := gc.GetStats()

	if stats.TotalObjects != 3 {
		t.Errorf("Expected 3 total objects, got %d", stats.TotalObjects)
	}
	if stats.ArrayObjects != 2 {
		t.Errorf("Expected 2 array objects, got %d", stats.ArrayObjects)
	}
	if stats.MapObjects != 1 {
		t.Errorf("Expected 1 map object, got %d", stats.MapObjects)
	}
	if !stats.GCEnabled {
		t.Error("Expected GC to be enabled")
	}

	// Release some objects
	gc.Release(arrayID)
	gc.Release(mapID)

	stats = gc.GetStats()
	if stats.TotalObjects != 1 {
		t.Errorf("Expected 1 total object after releases, got %d", stats.TotalObjects)
	}
	if stats.ArrayObjects != 1 {
		t.Errorf("Expected 1 array object after releases, got %d", stats.ArrayObjects)
	}
	if stats.MapObjects != 0 {
		t.Errorf("Expected 0 map objects after releases, got %d", stats.MapObjects)
	}

	// Clean up
	gc.Release(arrayID2)
}

func TestGCManager_GCThreshold(t *testing.T) {
	gc := NewGCManager()
	gc.SetGCThreshold(2) // Trigger GC after 2 allocations

	// Allocate objects and immediately release them
	values := []Value{NewInt(1)}

	_, id1 := gc.AllocateArray(values)
	gc.Release(id1) // RefCount = 0, eligible for collection

	_, id2 := gc.AllocateArray(values)
	gc.Release(id2) // RefCount = 0, eligible for collection

	// This allocation should trigger GC
	_, id3 := gc.AllocateArray(values)

	// Give GC goroutine time to run
	time.Sleep(10 * time.Millisecond)

	stats := gc.GetStats()
	// Should only have the last object (id3)
	if stats.TotalObjects != 1 {
		t.Errorf("Expected 1 object after GC, got %d", stats.TotalObjects)
	}

	// Clean up
	gc.Release(id3)
}

func TestGCManager_DisabledGC(t *testing.T) {
	gc := NewGCManager()
	gc.enabled = false // Disable GC

	// Try to allocate - should return nil
	values := []Value{NewInt(1)}
	obj, id := gc.AllocateArray(values)

	if obj != nil {
		t.Error("Expected nil object when GC disabled")
	}
	if id != 0 {
		t.Error("Expected zero ID when GC disabled")
	}

	stats := gc.GetStats()
	if stats.TotalObjects != 0 {
		t.Errorf("Expected 0 objects when GC disabled, got %d", stats.TotalObjects)
	}
}

func TestGlobalGCFunctions(t *testing.T) {
	// Test global convenience functions
	SetGlobalGCEnabled(true)

	// Test array allocation
	values := []Value{NewInt(1), NewInt(2)}
	obj, id := AllocateArray(values)

	if obj == nil || id == 0 {
		t.Fatal("Global AllocateArray failed")
	}

	// Test retain/release
	Retain(id)
	Release(id)
	Release(id) // Should deallocate

	// Test map allocation
	mapValues := map[string]Value{"test": NewString("value")}
	mapObj, mapID := AllocateMap(mapValues)

	if mapObj == nil || mapID == 0 {
		t.Fatal("Global AllocateMap failed")
	}

	Release(mapID)

	// Test stats
	stats := GetGCStats()
	if stats.TotalObjects != 0 {
		t.Errorf("Expected 0 objects after cleanup, got %d", stats.TotalObjects)
	}
}

func TestValue_GCMethods(t *testing.T) {
	// Test Value GC helper methods
	values := []Value{NewInt(1), NewInt(2)}

	// Test regular array (non-GC)
	regularArray := NewArray(values)
	if regularArray.IsGCValue() {
		t.Error("Regular array should not be a GC value")
	}

	// Test GC array
	gcArray := NewGCArray(values)
	if !gcArray.IsGCValue() {
		t.Error("GC array should be a GC value")
	}

	// Test AsArray for GC value
	arr, err := gcArray.AsArray()
	if err != nil {
		t.Errorf("AsArray failed for GC array: %v", err)
	}
	if len(arr) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(arr))
	}

	// Test map
	mapValues := map[string]Value{"key": NewString("value")}
	gcMap := NewGCMap(mapValues)
	if !gcMap.IsGCValue() {
		t.Error("GC map should be a GC value")
	}

	m, err := gcMap.AsMap()
	if err != nil {
		t.Errorf("AsMap failed for GC map: %v", err)
	}
	if len(m) != 1 {
		t.Errorf("Expected 1 element, got %d", len(m))
	}

	// Test string representation
	gcArrayStr := gcArray.String()
	if gcArrayStr == "" {
		t.Error("GC array string should not be empty")
	}

	// Test IsTruthy
	if !gcArray.IsTruthy() {
		t.Error("Non-empty GC array should be truthy")
	}

	emptyGCArray := NewGCArray([]Value{})
	if emptyGCArray.IsTruthy() {
		t.Error("Empty GC array should not be truthy")
	}

	// Clean up
	gcArray.Release()
	gcMap.Release()
	emptyGCArray.Release()
}
