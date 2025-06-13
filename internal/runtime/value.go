package runtime

import (
	"fmt"
)

// GCValue wraps a garbage-collected object with its ID.
type GCValue struct {
	Object *GCObject
	ID     ObjectID
}

// ValueType represents the type of a runtime value.
type ValueType int

const (
	ValueTypeInt ValueType = iota
	ValueTypeFloat
	ValueTypeString
	ValueTypeBool
	ValueTypeArray
	ValueTypeMap
	ValueTypeVoid
)

// Value represents a runtime value in ALaS.
type Value struct {
	Value interface{}
	Type  ValueType
}

// NewInt creates a new integer value.
func NewInt(v int64) Value {
	return Value{Type: ValueTypeInt, Value: v}
}

// NewFloat creates a new float value.
func NewFloat(v float64) Value {
	return Value{Type: ValueTypeFloat, Value: v}
}

// NewString creates a new string value.
func NewString(v string) Value {
	return Value{Type: ValueTypeString, Value: v}
}

// NewBool creates a new boolean value.
func NewBool(v bool) Value {
	return Value{Type: ValueTypeBool, Value: v}
}

// NewArray creates a new array value.
func NewArray(v []Value) Value {
	return Value{Type: ValueTypeArray, Value: v}
}

// NewGCArray creates a new garbage-collected array value.
func NewGCArray(v []Value) Value {
	obj, id := AllocateArray(v)
	gcVal := &GCValue{ID: id, Object: obj}
	return Value{Type: ValueTypeArray, Value: gcVal}
}

// NewMap creates a new map value.
func NewMap(v map[string]Value) Value {
	return Value{Type: ValueTypeMap, Value: v}
}

// NewGCMap creates a new garbage-collected map value.
func NewGCMap(v map[string]Value) Value {
	obj, id := AllocateMap(v)
	gcVal := &GCValue{ID: id, Object: obj}
	return Value{Type: ValueTypeMap, Value: gcVal}
}

// NewVoid creates a void value.
func NewVoid() Value {
	return Value{Type: ValueTypeVoid, Value: nil}
}

// AsInt returns the value as an integer.
func (v Value) AsInt() (int64, error) {
	switch v.Type {
	case ValueTypeInt:
		return v.Value.(int64), nil
	case ValueTypeFloat:
		return int64(v.Value.(float64)), nil
	case ValueTypeString, ValueTypeBool, ValueTypeArray, ValueTypeMap, ValueTypeVoid:
		return 0, fmt.Errorf("cannot convert %v to int", v.Type)
	default:
		return 0, fmt.Errorf("cannot convert %v to int", v.Type)
	}
}

// AsFloat returns the value as a float.
func (v Value) AsFloat() (float64, error) {
	switch v.Type {
	case ValueTypeFloat:
		return v.Value.(float64), nil
	case ValueTypeInt:
		return float64(v.Value.(int64)), nil
	case ValueTypeString, ValueTypeBool, ValueTypeArray, ValueTypeMap, ValueTypeVoid:
		return 0, fmt.Errorf("cannot convert %v to float", v.Type)
	default:
		return 0, fmt.Errorf("cannot convert %v to float", v.Type)
	}
}

// AsString returns the value as a string.
func (v Value) AsString() (string, error) {
	if v.Type != ValueTypeString {
		return "", fmt.Errorf("value is not a string")
	}
	return v.Value.(string), nil
}

// AsBool returns the value as a boolean.
func (v Value) AsBool() (bool, error) {
	if v.Type != ValueTypeBool {
		return false, fmt.Errorf("value is not a boolean")
	}
	return v.Value.(bool), nil
}

// AsArray returns the value as an array.
func (v Value) AsArray() ([]Value, error) {
	if v.Type != ValueTypeArray {
		return nil, fmt.Errorf("value is not an array")
	}

	// Handle garbage-collected arrays
	if gcVal, ok := v.Value.(*GCValue); ok {
		if arr, ok := gcVal.Object.Data.([]Value); ok {
			return arr, nil
		}
		return nil, fmt.Errorf("invalid garbage-collected array data")
	}

	// Handle regular arrays
	return v.Value.([]Value), nil
}

// AsMap returns the value as a map.
func (v Value) AsMap() (map[string]Value, error) {
	if v.Type != ValueTypeMap {
		return nil, fmt.Errorf("value is not a map")
	}

	// Handle garbage-collected maps
	if gcVal, ok := v.Value.(*GCValue); ok {
		if m, ok := gcVal.Object.Data.(map[string]Value); ok {
			return m, nil
		}
		return nil, fmt.Errorf("invalid garbage-collected map data")
	}

	// Handle regular maps
	return v.Value.(map[string]Value), nil
}

// IsTruthy returns whether the value is truthy.
func (v Value) IsTruthy() bool {
	switch v.Type {
	case ValueTypeBool:
		return v.Value.(bool)
	case ValueTypeInt:
		return v.Value.(int64) != 0
	case ValueTypeFloat:
		return v.Value.(float64) != 0
	case ValueTypeString:
		return v.Value.(string) != ""
	case ValueTypeArray:
		if gcVal, ok := v.Value.(*GCValue); ok {
			if arr, ok := gcVal.Object.Data.([]Value); ok {
				return len(arr) > 0
			}
			return false
		}
		return len(v.Value.([]Value)) > 0
	case ValueTypeMap:
		if gcVal, ok := v.Value.(*GCValue); ok {
			if m, ok := gcVal.Object.Data.(map[string]Value); ok {
				return len(m) > 0
			}
			return false
		}
		return len(v.Value.(map[string]Value)) > 0
	case ValueTypeVoid:
		return false
	default:
		return false
	}
}

// String returns a string representation of the value.
func (v Value) String() string {
	switch v.Type {
	case ValueTypeInt:
		return fmt.Sprintf("%d", v.Value.(int64))
	case ValueTypeFloat:
		return fmt.Sprintf("%f", v.Value.(float64))
	case ValueTypeString:
		return v.Value.(string)
	case ValueTypeBool:
		return fmt.Sprintf("%t", v.Value.(bool))
	case ValueTypeArray:
		if gcVal, ok := v.Value.(*GCValue); ok {
			return fmt.Sprintf("GCArray[%d](%v)", gcVal.ID, gcVal.Object.Data)
		}
		return fmt.Sprintf("%v", v.Value)
	case ValueTypeMap:
		if gcVal, ok := v.Value.(*GCValue); ok {
			return fmt.Sprintf("GCMap[%d](%v)", gcVal.ID, gcVal.Object.Data)
		}
		return fmt.Sprintf("%v", v.Value)
	case ValueTypeVoid:
		return "void"
	default:
		return "unknown"
	}
}

// Release releases any garbage-collected objects contained in this value.
func (v Value) Release() {
	if v.Type == ValueTypeArray || v.Type == ValueTypeMap {
		if gcVal, ok := v.Value.(*GCValue); ok {
			Release(gcVal.ID)
		}
	}
}

// Retain retains any garbage-collected objects contained in this value.
func (v Value) Retain() {
	if v.Type == ValueTypeArray || v.Type == ValueTypeMap {
		if gcVal, ok := v.Value.(*GCValue); ok {
			Retain(gcVal.ID)
		}
	}
}

// IsGCValue returns true if this value contains a garbage-collected object.
func (v Value) IsGCValue() bool {
	if v.Type == ValueTypeArray || v.Type == ValueTypeMap {
		_, ok := v.Value.(*GCValue)
		return ok
	}
	return false
}
