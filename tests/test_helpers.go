package tests

import (
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/runtime"
)

// Helper to get string representation of optimization level
func getOptLevelString(o codegen.OptimizationLevel) string {
	switch o {
	case codegen.OptNone:
		return "O0"
	case codegen.OptBasic:
		return "O1"
	case codegen.OptStandard:
		return "O2"
	case codegen.OptAggressive:
		return "O3"
	default:
		return "Unknown"
	}
}

// Helper to convert Go value to runtime.Value
func toRuntimeValue(v interface{}) runtime.Value {
	switch val := v.(type) {
	case int:
		return runtime.NewInt(int64(val))
	case int64:
		return runtime.NewInt(val)
	case float64:
		return runtime.NewFloat(val)
	case string:
		return runtime.NewString(val)
	case bool:
		return runtime.NewBool(val)
	default:
		return runtime.NewVoid()
	}
}

// Helper to convert runtime.Value to Go value
func fromRuntimeValue(v runtime.Value) interface{} {
	switch v.Type {
	case runtime.ValueTypeInt:
		val, _ := v.AsInt()
		return int(val)
	case runtime.ValueTypeFloat:
		val, _ := v.AsFloat()
		// If it's a whole number, convert to int for easier comparison
		if val == float64(int(val)) {
			return int(val)
		}
		return val
	case runtime.ValueTypeString:
		val, _ := v.AsString()
		return val
	case runtime.ValueTypeBool:
		val, _ := v.AsBool()
		return val
	case runtime.ValueTypeVoid:
		return nil
	case runtime.ValueTypeArray:
		// Return the value as-is for arrays
		return v
	case runtime.ValueTypeMap:
		// Return the value as-is for maps
		return v
	default:
		return nil
	}
}

// Helper to compare runtime values with Go values
func compareRuntimeValue(rv runtime.Value, expected interface{}) bool {
	actual := fromRuntimeValue(rv)
	return actual == expected
}
