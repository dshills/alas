package runtime

import (
	"sync"
	"sync/atomic"
)

// GCManager manages garbage collection for ALaS runtime objects.
type GCManager struct {
	objects     map[ObjectID]*GCObject
	mu          sync.RWMutex
	nextID      int64
	gcThreshold int64
	enabled     bool
	gcRunning   int32 // Atomic flag to prevent concurrent GC runs
}

// ObjectID uniquely identifies a garbage-collected object.
type ObjectID int64

// GCObject represents a garbage-collected runtime object.
type GCObject struct {
	Data     interface{}
	Finalize func()
	RefCount int64
	Size     int64
	ID       ObjectID
	Type     ObjectType
}

// ObjectType defines the type of garbage-collected object.
type ObjectType int

const (
	ObjectTypeArray ObjectType = iota
	ObjectTypeMap
)

// Global GC manager instance.
var globalGC = &GCManager{
	objects:     make(map[ObjectID]*GCObject),
	gcThreshold: 1000, // Run GC after 1000 allocations
	enabled:     true,
}

// NewGCManager creates a new garbage collection manager.
func NewGCManager() *GCManager {
	return &GCManager{
		objects:     make(map[ObjectID]*GCObject),
		gcThreshold: 1000,
		enabled:     true,
	}
}

// SetGlobalGCEnabled enables or disables garbage collection globally.
func SetGlobalGCEnabled(enabled bool) {
	globalGC.mu.Lock()
	defer globalGC.mu.Unlock()
	globalGC.enabled = enabled
}

// AllocateArray creates a new garbage-collected array.
func (gc *GCManager) AllocateArray(values []Value) (*GCObject, ObjectID) {
	if !gc.enabled {
		return nil, 0
	}

	id := ObjectID(atomic.AddInt64(&gc.nextID, 1))
	size := int64(len(values)) * 64 // Approximate size in bytes

	obj := &GCObject{
		ID:       id,
		RefCount: 1,
		Type:     ObjectTypeArray,
		Data:     values,
		Size:     size,
	}

	gc.mu.Lock()
	gc.objects[id] = obj
	gc.mu.Unlock()

	gc.checkGCThreshold()
	return obj, id
}

// AllocateMap creates a new garbage-collected map.
func (gc *GCManager) AllocateMap(values map[string]Value) (*GCObject, ObjectID) {
	if !gc.enabled {
		return nil, 0
	}

	id := ObjectID(atomic.AddInt64(&gc.nextID, 1))
	size := int64(len(values)) * 96 // Approximate size in bytes (key + value)

	obj := &GCObject{
		ID:       id,
		RefCount: 1,
		Type:     ObjectTypeMap,
		Data:     values,
		Size:     size,
	}

	gc.mu.Lock()
	gc.objects[id] = obj
	gc.mu.Unlock()

	gc.checkGCThreshold()
	return obj, id
}

// Retain increases the reference count for an object.
func (gc *GCManager) Retain(id ObjectID) {
	if !gc.enabled || id == 0 {
		return
	}

	gc.mu.RLock()
	obj, exists := gc.objects[id]
	gc.mu.RUnlock()

	if exists {
		atomic.AddInt64(&obj.RefCount, 1)
	}
}

// Release decreases the reference count and deallocates if zero.
func (gc *GCManager) Release(id ObjectID) {
	if !gc.enabled || id == 0 {
		return
	}

	gc.mu.RLock()
	obj, exists := gc.objects[id]
	gc.mu.RUnlock()

	if !exists {
		return
	}

	newRefCount := atomic.AddInt64(&obj.RefCount, -1)
	if newRefCount <= 0 {
		gc.deallocate(id, obj)
	}
}

// deallocate removes an object from memory.
func (gc *GCManager) deallocate(id ObjectID, obj *GCObject) {
	// Run finalizer if present
	if obj.Finalize != nil {
		obj.Finalize()
	}

	// Release references to nested objects
	gc.releaseNestedRefs(obj)

	// Remove from objects map
	gc.mu.Lock()
	delete(gc.objects, id)
	gc.mu.Unlock()
}

// releaseNestedRefs releases references to objects contained within this object.
func (gc *GCManager) releaseNestedRefs(obj *GCObject) {
	switch obj.Type {
	case ObjectTypeArray:
		if arr, ok := obj.Data.([]Value); ok {
			for _, val := range arr {
				if val.Type == ValueTypeArray || val.Type == ValueTypeMap {
					if gcVal, ok := val.Value.(*GCValue); ok {
						gc.Release(gcVal.ID)
					}
				}
			}
		}
	case ObjectTypeMap:
		if m, ok := obj.Data.(map[string]Value); ok {
			for _, val := range m {
				if val.Type == ValueTypeArray || val.Type == ValueTypeMap {
					if gcVal, ok := val.Value.(*GCValue); ok {
						gc.Release(gcVal.ID)
					}
				}
			}
		}
	}
}

// checkGCThreshold runs garbage collection if threshold is exceeded.
func (gc *GCManager) checkGCThreshold() {
	gc.mu.RLock()
	objectCount := int64(len(gc.objects))
	gc.mu.RUnlock()

	if objectCount > gc.gcThreshold {
		// Ensure only one GC run is in progress
		if atomic.CompareAndSwapInt32(&gc.gcRunning, 0, 1) {
			go func() {
				gc.RunGC()
				atomic.StoreInt32(&gc.gcRunning, 0) // Reset flag after GC completes
			}()
		}
	}
}

// RunGC performs a garbage collection sweep.
func (gc *GCManager) RunGC() {
	if !gc.enabled {
		return
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Collect objects with zero reference count
	var toDelete []ObjectID
	for id, obj := range gc.objects {
		if atomic.LoadInt64(&obj.RefCount) <= 0 {
			toDelete = append(toDelete, id)
		}
	}

	// Remove collected objects
	for _, id := range toDelete {
		if obj, exists := gc.objects[id]; exists {
			if obj.Finalize != nil {
				obj.Finalize()
			}
			delete(gc.objects, id)
		}
	}
}

// GetStats returns garbage collection statistics.
func (gc *GCManager) GetStats() GCStats {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	var totalSize int64
	var arrayCount, mapCount int64

	for _, obj := range gc.objects {
		totalSize += obj.Size
		switch obj.Type {
		case ObjectTypeArray:
			arrayCount++
		case ObjectTypeMap:
			mapCount++
		}
	}

	return GCStats{
		TotalObjects: int64(len(gc.objects)),
		TotalSize:    totalSize,
		ArrayObjects: arrayCount,
		MapObjects:   mapCount,
		GCThreshold:  gc.gcThreshold,
		GCEnabled:    gc.enabled,
	}
}

// GCStats provides statistics about garbage collection.
type GCStats struct {
	TotalObjects int64
	TotalSize    int64
	ArrayObjects int64
	MapObjects   int64
	GCThreshold  int64
	GCEnabled    bool
}

// SetGCThreshold sets the garbage collection threshold.
func (gc *GCManager) SetGCThreshold(threshold int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.gcThreshold = threshold
}

// Global GC functions for convenience.
func AllocateArray(values []Value) (*GCObject, ObjectID) {
	return globalGC.AllocateArray(values)
}

func AllocateMap(values map[string]Value) (*GCObject, ObjectID) {
	return globalGC.AllocateMap(values)
}

func Retain(id ObjectID) {
	globalGC.Retain(id)
}

func Release(id ObjectID) {
	globalGC.Release(id)
}

func RunGC() {
	globalGC.RunGC()
}

func GetGCStats() GCStats {
	return globalGC.GetStats()
}
