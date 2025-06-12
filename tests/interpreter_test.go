package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
	"github.com/dshills/alas/internal/validator"
)

type testCase struct {
	name     string
	file     string
	function string
	args     []runtime.Value
	expected runtime.Value
}

func TestInterpreter(t *testing.T) {
	tests := []testCase{
		{
			name:     "Hello World",
			file:     "../examples/programs/hello.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewString("Hello, ALaS!"),
		},
		{
			name:     "Fibonacci(10)",
			file:     "../examples/programs/fibonacci.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(55),
		},
		{
			name:     "Factorial(5)",
			file:     "../examples/programs/factorial.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(120),
		},
		{
			name:     "Sum to 10",
			file:     "../examples/programs/loops.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(55),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Read the file
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

func TestValidation(t *testing.T) {
	// Test valid programs
	validFiles, err := filepath.Glob("../examples/programs/*.alas.json")
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	for _, file := range validFiles {
		t.Run("Valid: "+filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			if err := validator.ValidateJSON(data); err != nil {
				t.Errorf("Validation failed for valid file: %v", err)
			}
		})
	}

	// Test invalid programs
	invalidPrograms := []struct {
		name string
		json string
	}{
		{
			name: "Missing module type",
			json: `{"name": "test", "functions": []}`,
		},
		{
			name: "Empty functions array",
			json: `{"type": "module", "name": "test", "functions": []}`,
		},
		{
			name: "Function without body",
			json: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "void",
					"body": null
				}]
			}`,
		},
		{
			name: "Undefined variable",
			json: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "int",
					"body": [{
						"type": "return",
						"value": {
							"type": "variable",
							"name": "undefined"
						}
					}]
				}]
			}`,
		},
	}

	for _, tc := range invalidPrograms {
		t.Run("Invalid: "+tc.name, func(t *testing.T) {
			if err := validator.ValidateJSON([]byte(tc.json)); err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
		})
	}
}

func valuesEqual(a, b runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case runtime.ValueTypeInt:
		ai, _ := a.AsInt()
		bi, _ := b.AsInt()
		return ai == bi
	case runtime.ValueTypeFloat:
		af, _ := a.AsFloat()
		bf, _ := b.AsFloat()
		return af == bf
	case runtime.ValueTypeString:
		as, _ := a.AsString()
		bs, _ := b.AsString()
		return as == bs
	case runtime.ValueTypeBool:
		ab, _ := a.AsBool()
		bb, _ := b.AsBool()
		return ab == bb
	case runtime.ValueTypeVoid:
		return true
	case runtime.ValueTypeArray, runtime.ValueTypeMap:
		// TODO: Implement array and map comparison
		return false
	default:
		return false
	}
}
