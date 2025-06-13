package tests

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
	"github.com/dshills/alas/internal/validator"
)

func TestStandardLibraryIntegration(t *testing.T) {
	tests := []testCase{
		{
			name:     "Builtin Functions Basic",
			file:     "../examples/programs/builtin_test.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewVoid(), // print functions return void
		},
		{
			name:     "Comprehensive Standard Library",
			file:     "../examples/programs/stdlib_comprehensive_test.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewVoid(), // print functions return void
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Read file
			data, err := os.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("Failed to read file %s: %v", tc.file, err)
			}

			// Validate
			if err := validator.ValidateJSON(data); err != nil {
				t.Fatalf("Validation failed: %v", err)
			}

			// Parse
			var module ast.Module
			if err := json.Unmarshal(data, &module); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Interpret
			interp := interpreter.New()
			if err := interp.LoadModule(&module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			// Check result
			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestStandardLibraryMath(t *testing.T) {
	interp := interpreter.New()

	// Test individual math functions
	mathTests := []struct {
		name     string
		function string
		args     []runtime.Value
		expected runtime.Value
	}{
		{
			name:     "math.abs",
			function: "math.abs",
			args:     []runtime.Value{runtime.NewFloat(-5.5)},
			expected: runtime.NewFloat(5.5),
		},
		{
			name:     "math.sqrt",
			function: "math.sqrt",
			args:     []runtime.Value{runtime.NewFloat(25.0)},
			expected: runtime.NewFloat(5.0),
		},
		{
			name:     "math.max",
			function: "math.max",
			args:     []runtime.Value{runtime.NewFloat(3.14), runtime.NewFloat(2.71)},
			expected: runtime.NewFloat(3.14),
		},
		{
			name:     "math.min",
			function: "math.min",
			args:     []runtime.Value{runtime.NewFloat(3.14), runtime.NewFloat(2.71)},
			expected: runtime.NewFloat(2.71),
		},
	}

	for _, tc := range mathTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := interp.CallBuiltinFunction(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", tc.function, err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestStandardLibraryCollections(t *testing.T) {
	interp := interpreter.New()

	// Test collections functions
	arr := runtime.NewGCArray([]runtime.Value{
		runtime.NewString("hello"),
		runtime.NewString("world"),
		runtime.NewString("test"),
	})

	collectionsTests := []struct {
		name     string
		function string
		args     []runtime.Value
		expected runtime.Value
	}{
		{
			name:     "collections.length array",
			function: "collections.length",
			args:     []runtime.Value{arr},
			expected: runtime.NewInt(3),
		},
		{
			name:     "collections.length string",
			function: "collections.length",
			args:     []runtime.Value{runtime.NewString("hello")},
			expected: runtime.NewInt(5),
		},
		{
			name:     "collections.contains array",
			function: "collections.contains",
			args:     []runtime.Value{arr, runtime.NewString("world")},
			expected: runtime.NewBool(true),
		},
		{
			name:     "collections.contains string",
			function: "collections.contains",
			args:     []runtime.Value{runtime.NewString("hello world"), runtime.NewString("world")},
			expected: runtime.NewBool(true),
		},
	}

	for _, tc := range collectionsTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := interp.CallBuiltinFunction(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", tc.function, err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestStandardLibraryString(t *testing.T) {
	interp := interpreter.New()

	stringTests := []struct {
		name     string
		function string
		args     []runtime.Value
		expected runtime.Value
	}{
		{
			name:     "string.toUpper",
			function: "string.toUpper",
			args:     []runtime.Value{runtime.NewString("hello")},
			expected: runtime.NewString("HELLO"),
		},
		{
			name:     "string.toLower",
			function: "string.toLower",
			args:     []runtime.Value{runtime.NewString("WORLD")},
			expected: runtime.NewString("world"),
		},
		{
			name:     "string.length",
			function: "string.length",
			args:     []runtime.Value{runtime.NewString("test")},
			expected: runtime.NewInt(4),
		},
	}

	for _, tc := range stringTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := interp.CallBuiltinFunction(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", tc.function, err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestStandardLibraryType(t *testing.T) {
	interp := interpreter.New()

	typeTests := []struct {
		name     string
		function string
		args     []runtime.Value
		expected runtime.Value
	}{
		{
			name:     "type.typeOf int",
			function: "type.typeOf",
			args:     []runtime.Value{runtime.NewInt(42)},
			expected: runtime.NewString("int"),
		},
		{
			name:     "type.typeOf string",
			function: "type.typeOf",
			args:     []runtime.Value{runtime.NewString("hello")},
			expected: runtime.NewString("string"),
		},
		{
			name:     "type.isInt true",
			function: "type.isInt",
			args:     []runtime.Value{runtime.NewInt(42)},
			expected: runtime.NewBool(true),
		},
		{
			name:     "type.isInt false",
			function: "type.isInt",
			args:     []runtime.Value{runtime.NewString("hello")},
			expected: runtime.NewBool(false),
		},
	}

	for _, tc := range typeTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := interp.CallBuiltinFunction(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", tc.function, err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
