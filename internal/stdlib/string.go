package stdlib

import (
	"fmt"
	"strings"

	"github.com/dshills/alas/internal/runtime"
)

// registerStringFunctions registers all std.string builtin functions.
func (r *Registry) registerStringFunctions() {
	r.Register("string.length", stringLength)
	r.Register("string.split", stringSplit)
	r.Register("string.join", stringJoin)
	r.Register("string.toUpper", stringToUpper)
	r.Register("string.toLower", stringToLower)
	r.Register("string.trim", stringTrim)
	r.Register("string.replace", stringReplace)
}

// stringLength implements string.length builtin function.
func stringLength(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("string.length expects 1 argument, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.length: %v", err)
	}

	return runtime.NewInt(int64(len(str))), nil
}

// stringSplit implements string.split builtin function.
func stringSplit(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("string.split expects 2 arguments, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.split: %v", err)
	}

	separator, err := args[1].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.split: separator must be string")
	}

	parts := strings.Split(str, separator)
	elements := make([]runtime.Value, len(parts))
	for i, part := range parts {
		elements[i] = runtime.NewString(part)
	}

	return runtime.NewGCArray(elements), nil
}

// stringJoin implements string.join builtin function.
func stringJoin(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("string.join expects 2 arguments, got %d", len(args))
	}

	if args[0].Type != runtime.ValueTypeArray {
		return runtime.NewVoid(), fmt.Errorf("string.join: first argument must be array")
	}

	arr, err := args[0].AsArray()
	if err != nil {
		return runtime.NewVoid(), err
	}

	separator, err := args[1].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.join: separator must be string")
	}

	// Convert array elements to strings
	parts := make([]string, len(arr))
	for i, elem := range arr {
		str, err := elem.AsString()
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("string.join: array element %d is not a string", i)
		}
		parts[i] = str
	}

	result := strings.Join(parts, separator)
	return runtime.NewString(result), nil
}

// stringToUpper implements string.toUpper builtin function.
func stringToUpper(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("string.toUpper expects 1 argument, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.toUpper: %v", err)
	}

	return runtime.NewString(strings.ToUpper(str)), nil
}

// stringToLower implements string.toLower builtin function.
func stringToLower(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("string.toLower expects 1 argument, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.toLower: %v", err)
	}

	return runtime.NewString(strings.ToLower(str)), nil
}

// stringTrim implements string.trim builtin function.
func stringTrim(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("string.trim expects 1 argument, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.trim: %v", err)
	}

	return runtime.NewString(strings.TrimSpace(str)), nil
}

// stringReplace implements string.replace builtin function.
func stringReplace(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 3 {
		return runtime.NewVoid(), fmt.Errorf("string.replace expects 3 arguments, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.replace: %v", err)
	}

	old, err := args[1].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.replace: old value must be string")
	}

	new, err := args[2].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("string.replace: new value must be string")
	}

	result := strings.ReplaceAll(str, old, new)
	return runtime.NewString(result), nil
}
