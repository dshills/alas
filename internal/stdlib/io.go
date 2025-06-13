package stdlib

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dshills/alas/internal/runtime"
)

// registerIOFunctions registers all std.io builtin functions.
func (r *Registry) registerIOFunctions() {
	r.Register("io.readFile", ioReadFile)
	r.Register("io.writeFile", ioWriteFile)
	r.Register("io.print", ioPrint)
	r.Register("io.readLine", ioReadLine)
}

// ioReadFile implements io.readFile builtin function.
// Returns a map with structure: {ok: bool, data: string, error: string}.
func ioReadFile(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("io.readFile expects 1 argument, got %d", len(args))
	}

	pathStr, err := args[0].AsString()
	if err != nil {
		return createIOResult(false, "", "path must be a string"), nil
	}

	data, err := os.ReadFile(pathStr)
	if err != nil {
		return createIOResult(false, "", err.Error()), nil
	}

	return createIOResult(true, string(data), ""), nil
}

// ioWriteFile implements io.writeFile builtin function.
// Returns a map with structure: {ok: bool, error: string}.
func ioWriteFile(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 2 {
		return runtime.NewVoid(), fmt.Errorf("io.writeFile expects 2 arguments, got %d", len(args))
	}

	pathStr, err := args[0].AsString()
	if err != nil {
		return createIOWriteResult(false, "path must be a string"), nil
	}

	dataStr, err := args[1].AsString()
	if err != nil {
		return createIOWriteResult(false, "data must be a string"), nil
	}

	err = os.WriteFile(pathStr, []byte(dataStr), 0600)
	if err != nil {
		return createIOWriteResult(false, err.Error()), nil
	}

	return createIOWriteResult(true, ""), nil
}

// ioPrint implements io.print builtin function.
// Prints value to stdout, returns void.
func ioPrint(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("io.print expects 1 argument, got %d", len(args))
	}

	// Convert the value to a string representation
	val := args[0]
	switch val.Type {
	case runtime.ValueTypeInt:
		intVal, _ := val.AsInt()
		fmt.Print(intVal)
	case runtime.ValueTypeFloat:
		floatVal, _ := val.AsFloat()
		fmt.Print(floatVal)
	case runtime.ValueTypeString:
		strVal, _ := val.AsString()
		fmt.Print(strVal)
	case runtime.ValueTypeBool:
		boolVal, _ := val.AsBool()
		if boolVal {
			fmt.Print("true")
		} else {
			fmt.Print("false")
		}
	case runtime.ValueTypeArray:
		arr, _ := val.AsArray()
		fmt.Print("[")
		for i, elem := range arr {
			if i > 0 {
				fmt.Print(", ")
			}
			// Recursively print element
			ioPrint([]runtime.Value{elem})
		}
		fmt.Print("]")
	case runtime.ValueTypeMap:
		m, _ := val.AsMap()
		fmt.Print("{")
		first := true
		for key, value := range m {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("%s: ", key)
			ioPrint([]runtime.Value{value})
			first = false
		}
		fmt.Print("}")
	case runtime.ValueTypeVoid:
		fmt.Print("<void>")
	default:
		fmt.Print("<void>")
	}

	return runtime.NewVoid(), nil
}

// ioReadLine implements io.readLine builtin function.
// Reads a line from stdin, returns string.
func ioReadLine(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 0 {
		return runtime.NewVoid(), fmt.Errorf("io.readLine expects 0 arguments, got %d", len(args))
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return runtime.NewString(scanner.Text()), nil
	}

	if err := scanner.Err(); err != nil {
		return runtime.NewString(""), fmt.Errorf("error reading from stdin: %v", err)
	}

	// EOF reached
	return runtime.NewString(""), nil
}

// Helper function to create standard I/O result maps for readFile.
func createIOResult(ok bool, data, errorMsg string) runtime.Value {
	result := make(map[string]runtime.Value)
	result["ok"] = runtime.NewBool(ok)
	result["data"] = runtime.NewString(data)
	result["error"] = runtime.NewString(errorMsg)
	return runtime.NewGCMap(result)
}

// Helper function to create standard I/O result maps for writeFile.
func createIOWriteResult(ok bool, errorMsg string) runtime.Value {
	result := make(map[string]runtime.Value)
	result["ok"] = runtime.NewBool(ok)
	result["error"] = runtime.NewString(errorMsg)
	return runtime.NewGCMap(result)
}
