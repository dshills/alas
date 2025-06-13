package stdlib

import (
	"fmt"

	"github.com/dshills/alas/internal/runtime"
)

// registerResultFunctions registers all std.result builtin functions.
func (r *Registry) registerResultFunctions() {
	r.Register("result.ok", resultOk)
	r.Register("result.error", resultError)
	r.Register("result.isOk", resultIsOk)
	r.Register("result.isError", resultIsError)
	r.Register("result.getValue", resultGetValue)
	r.Register("result.getError", resultGetError)
}

// resultOk implements result.ok builtin function.
// Creates a Result type with ok=true and the provided value.
func resultOk(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("result.ok expects 1 argument, got %d", len(args))
	}

	result := make(map[string]runtime.Value)
	result["ok"] = runtime.NewBool(true)
	result["value"] = args[0]
	result["error"] = runtime.NewString("")

	return runtime.NewGCMap(result), nil
}

// resultError implements result.error builtin function.
// Creates a Result type with ok=false and the provided error message.
func resultError(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("result.error expects 1 argument, got %d", len(args))
	}

	errorMsg, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("result.error: error message must be string")
	}

	result := make(map[string]runtime.Value)
	result["ok"] = runtime.NewBool(false)
	result["value"] = runtime.NewVoid()
	result["error"] = runtime.NewString(errorMsg)

	return runtime.NewGCMap(result), nil
}

// resultIsOk implements result.isOk builtin function.
// Returns true if the Result represents a success.
func resultIsOk(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("result.isOk expects 1 argument, got %d", len(args))
	}

	if args[0].Type != runtime.ValueTypeMap {
		return runtime.NewVoid(), fmt.Errorf("result.isOk: argument must be a map (Result type)")
	}

	resultMap, err := args[0].AsMap()
	if err != nil {
		return runtime.NewVoid(), err
	}

	okValue, exists := resultMap["ok"]
	if !exists {
		return runtime.NewVoid(), fmt.Errorf("result.isOk: invalid Result type (missing 'ok' field)")
	}

	isOk, err := okValue.AsBool()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("result.isOk: invalid Result type ('ok' field must be boolean)")
	}

	return runtime.NewBool(isOk), nil
}

// resultIsError implements result.isError builtin function.
// Returns true if the Result represents an error.
func resultIsError(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("result.isError expects 1 argument, got %d", len(args))
	}

	isOkResult, err := resultIsOk(args)
	if err != nil {
		return runtime.NewVoid(), err
	}

	isOk, _ := isOkResult.AsBool()
	return runtime.NewBool(!isOk), nil
}

// resultGetValue implements result.getValue builtin function.
// Returns the value from a successful Result.
func resultGetValue(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("result.getValue expects 1 argument, got %d", len(args))
	}

	if args[0].Type != runtime.ValueTypeMap {
		return runtime.NewVoid(), fmt.Errorf("result.getValue: argument must be a map (Result type)")
	}

	resultMap, err := args[0].AsMap()
	if err != nil {
		return runtime.NewVoid(), err
	}

	okValue, exists := resultMap["ok"]
	if !exists {
		return runtime.NewVoid(), fmt.Errorf("result.getValue: invalid Result type (missing 'ok' field)")
	}

	isOk, err := okValue.AsBool()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("result.getValue: invalid Result type ('ok' field must be boolean)")
	}

	if !isOk {
		return runtime.NewVoid(), fmt.Errorf("result.getValue: cannot get value from error Result")
	}

	value, exists := resultMap["value"]
	if !exists {
		return runtime.NewVoid(), fmt.Errorf("result.getValue: invalid Result type (missing 'value' field)")
	}

	return value, nil
}

// resultGetError implements result.getError builtin function.
// Returns the error message from a failed Result.
func resultGetError(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("result.getError expects 1 argument, got %d", len(args))
	}

	if args[0].Type != runtime.ValueTypeMap {
		return runtime.NewVoid(), fmt.Errorf("result.getError: argument must be a map (Result type)")
	}

	resultMap, err := args[0].AsMap()
	if err != nil {
		return runtime.NewVoid(), err
	}

	okValue, exists := resultMap["ok"]
	if !exists {
		return runtime.NewVoid(), fmt.Errorf("result.getError: invalid Result type (missing 'ok' field)")
	}

	isOk, err := okValue.AsBool()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("result.getError: invalid Result type ('ok' field must be boolean)")
	}

	if isOk {
		return runtime.NewVoid(), fmt.Errorf("result.getError: cannot get error from successful Result")
	}

	errorValue, exists := resultMap["error"]
	if !exists {
		return runtime.NewVoid(), fmt.Errorf("result.getError: invalid Result type (missing 'error' field)")
	}

	return errorValue, nil
}
