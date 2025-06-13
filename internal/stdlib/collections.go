package stdlib

import (
	"fmt"

	"github.com/dshills/alas/internal/runtime"
)

// registerCollectionsFunctions registers all std.collections builtin functions.
func (r *Registry) registerCollectionsFunctions() {
	r.Register("collections.length", collectionsLength)
	r.Register("collections.append", collectionsAppend)
	r.Register("collections.contains", collectionsContains)
	r.Register("collections.indexOf", collectionsIndexOf)
	r.Register("collections.slice", collectionsSlice)
}

// validateSliceArgs validates slice arguments and returns start/end indices.
func validateSliceArgs(args []runtime.Value, maxLength int64) (int64, int64, error) {
	start, err := args[1].AsInt()
	if err != nil {
		return 0, 0, fmt.Errorf("collections.slice: start index must be integer")
	}

	end := maxLength
	if len(args) == 3 {
		end, err = args[2].AsInt()
		if err != nil {
			return 0, 0, fmt.Errorf("collections.slice: end index must be integer")
		}
	}

	// Bounds checking
	if start < 0 || start > maxLength || end < start || end > maxLength {
		return 0, 0, fmt.Errorf("collections.slice: index out of bounds")
	}

	return start, end, nil
}

// collectionsLength implements collections.length builtin function.
func collectionsLength(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("collections.length expects 1 argument, got %d", len(args))
	}

	val := args[0]
	switch val.Type {
	case runtime.ValueTypeArray:
		arr, err := val.AsArray()
		if err != nil {
			return runtime.NewVoid(), err
		}
		return runtime.NewInt(int64(len(arr))), nil
	case runtime.ValueTypeMap:
		m, err := val.AsMap()
		if err != nil {
			return runtime.NewVoid(), err
		}
		return runtime.NewInt(int64(len(m))), nil
	case runtime.ValueTypeString:
		str, err := val.AsString()
		if err != nil {
			return runtime.NewVoid(), err
		}
		return runtime.NewInt(int64(len(str))), nil
	case runtime.ValueTypeInt, runtime.ValueTypeFloat, runtime.ValueTypeBool, runtime.ValueTypeVoid:
		return runtime.NewVoid(), fmt.Errorf("collections.length: argument must be array, map, or string")
	default:
		return runtime.NewVoid(), fmt.Errorf("collections.length: argument must be array, map, or string")
	}
}

// collectionsAppend implements collections.append builtin function.
func collectionsAppend(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("collections.append expects 2 arguments, got %d", len(args))
	}

	if args[0].Type != runtime.ValueTypeArray {
		return runtime.NewVoid(), fmt.Errorf("collections.append: first argument must be an array")
	}

	arr, err := args[0].AsArray()
	if err != nil {
		return runtime.NewVoid(), err
	}

	// Create a new array with the appended element
	newElements := make([]runtime.Value, len(arr)+1)
	copy(newElements, arr)
	newElements[len(arr)] = args[1]

	return runtime.NewGCArray(newElements), nil
}

// collectionsContains implements collections.contains builtin function.
func collectionsContains(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("collections.contains expects 2 arguments, got %d", len(args))
	}

	container := args[0]
	switch container.Type {
	case runtime.ValueTypeArray:
		arr, err := container.AsArray()
		if err != nil {
			return runtime.NewVoid(), err
		}
		searchValue := args[1]
		for _, elem := range arr {
			if Equal(elem, searchValue) {
				return runtime.NewBool(true), nil
			}
		}
		return runtime.NewBool(false), nil
	case runtime.ValueTypeMap:
		m, err := container.AsMap()
		if err != nil {
			return runtime.NewVoid(), err
		}
		keyStr, err := args[1].AsString()
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("collections.contains: map key must be string")
		}
		_, exists := m[keyStr]
		return runtime.NewBool(exists), nil
	case runtime.ValueTypeString:
		str, err := container.AsString()
		if err != nil {
			return runtime.NewVoid(), err
		}
		substr, err := args[1].AsString()
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("collections.contains: search value must be string")
		}
		contains := StringContains(str, substr)
		return runtime.NewBool(contains), nil
	case runtime.ValueTypeInt, runtime.ValueTypeFloat, runtime.ValueTypeBool, runtime.ValueTypeVoid:
		return runtime.NewVoid(), fmt.Errorf("collections.contains: first argument must be array, map, or string")
	default:
		return runtime.NewVoid(), fmt.Errorf("collections.contains: first argument must be array, map, or string")
	}
}

// collectionsIndexOf implements collections.indexOf builtin function.
func collectionsIndexOf(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("collections.indexOf expects 2 arguments, got %d", len(args))
	}

	container := args[0]
	switch container.Type {
	case runtime.ValueTypeArray:
		arr, err := container.AsArray()
		if err != nil {
			return runtime.NewVoid(), err
		}
		searchValue := args[1]
		for i, elem := range arr {
			if Equal(elem, searchValue) {
				return runtime.NewInt(int64(i)), nil
			}
		}
		return runtime.NewInt(-1), nil // Not found
	case runtime.ValueTypeString:
		str, err := container.AsString()
		if err != nil {
			return runtime.NewVoid(), err
		}
		substr, err := args[1].AsString()
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("collections.indexOf: search value must be string")
		}
		index := StringIndexOf(str, substr)
		return runtime.NewInt(int64(index)), nil
	case runtime.ValueTypeInt, runtime.ValueTypeFloat, runtime.ValueTypeBool, runtime.ValueTypeMap, runtime.ValueTypeVoid:
		return runtime.NewVoid(), fmt.Errorf("collections.indexOf: first argument must be array or string")
	default:
		return runtime.NewVoid(), fmt.Errorf("collections.indexOf: first argument must be array or string")
	}
}

// collectionsSlice implements collections.slice builtin function.
func collectionsSlice(args []runtime.Value) (runtime.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return runtime.NewVoid(), fmt.Errorf("collections.slice expects 2 or 3 arguments, got %d", len(args))
	}

	container := args[0]
	switch container.Type {
	case runtime.ValueTypeArray:
		arr, err := container.AsArray()
		if err != nil {
			return runtime.NewVoid(), err
		}
		start, end, err := validateSliceArgs(args, int64(len(arr)))
		if err != nil {
			return runtime.NewVoid(), err
		}
		sliced := arr[start:end]
		return runtime.NewGCArray(sliced), nil

	case runtime.ValueTypeString:
		str, err := container.AsString()
		if err != nil {
			return runtime.NewVoid(), err
		}
		start, end, err := validateSliceArgs(args, int64(len(str)))
		if err != nil {
			return runtime.NewVoid(), err
		}
		sliced := str[start:end]
		return runtime.NewString(sliced), nil

	case runtime.ValueTypeInt, runtime.ValueTypeFloat, runtime.ValueTypeBool, runtime.ValueTypeMap, runtime.ValueTypeVoid:
		return runtime.NewVoid(), fmt.Errorf("collections.slice: first argument must be array or string")
	default:
		return runtime.NewVoid(), fmt.Errorf("collections.slice: first argument must be array or string")
	}
}
