package stdlib

import (
	"strings"

	"github.com/dshills/alas/internal/runtime"
)

// Equal compares two runtime values for equality.
func Equal(a, b runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case runtime.ValueTypeInt:
		aVal, _ := a.AsInt()
		bVal, _ := b.AsInt()
		return aVal == bVal
	case runtime.ValueTypeFloat:
		aVal, _ := a.AsFloat()
		bVal, _ := b.AsFloat()
		return aVal == bVal
	case runtime.ValueTypeString:
		aVal, _ := a.AsString()
		bVal, _ := b.AsString()
		return aVal == bVal
	case runtime.ValueTypeBool:
		aVal, _ := a.AsBool()
		bVal, _ := b.AsBool()
		return aVal == bVal
	case runtime.ValueTypeVoid:
		return true
	case runtime.ValueTypeArray, runtime.ValueTypeMap:
		// For simplicity, only compare by reference for complex types
		// A full deep comparison would be more complex
		return false
	default:
		return false
	}
}

// StringContains checks if a string contains a substring.
func StringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// StringIndexOf returns the index of the first occurrence of substr in s, or -1 if not found.
func StringIndexOf(s, substr string) int {
	return strings.Index(s, substr)
}
