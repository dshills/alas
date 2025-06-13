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
		{
			name:     "Simple Array Access",
			file:     "../examples/programs/simple_array.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(20),
		},
		{
			name:     "Module Demo",
			file:     "../examples/programs/module_demo.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(30), // (10 + 5) * 2 = 30
		},
		{
			name:     "Complex Modules",
			file:     "../examples/programs/complex_modules.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewString("Number: 42"),
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
	case runtime.ValueTypeArray:
		aa, _ := a.AsArray()
		ba, _ := b.AsArray()
		if len(aa) != len(ba) {
			return false
		}
		for i := range aa {
			if !valuesEqual(aa[i], ba[i]) {
				return false
			}
		}
		return true
	case runtime.ValueTypeMap:
		am, _ := a.AsMap()
		bm, _ := b.AsMap()
		if len(am) != len(bm) {
			return false
		}
		for k, v := range am {
			if bv, ok := bm[k]; !ok || !valuesEqual(v, bv) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// TestArrayOperations tests array literal creation and indexing.
func TestArrayOperations(t *testing.T) {
	interp := interpreter.New()

	// Create a simple array program
	module := &ast.Module{
		Type: "module",
		Name: "test_arrays",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "test_array",
				Params:  []ast.Parameter{},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type:   "assign",
						Target: "arr",
						Value: &ast.Expression{
							Type: ast.ExprArrayLit,
							Elements: []ast.Expression{
								{Type: ast.ExprLiteral, Value: float64(10)},
								{Type: ast.ExprLiteral, Value: float64(20)},
								{Type: ast.ExprLiteral, Value: float64(30)},
							},
						},
					},
					{
						Type: "return",
						Value: &ast.Expression{
							Type: ast.ExprIndex,
							Object: &ast.Expression{
								Type: ast.ExprVariable,
								Name: "arr",
							},
							Index: &ast.Expression{
								Type:  ast.ExprLiteral,
								Value: float64(1),
							},
						},
					},
				},
			},
		},
	}

	if err := interp.LoadModule(module); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	result, err := interp.Run("test_array", []runtime.Value{})
	if err != nil {
		t.Fatalf("Runtime error: %v", err)
	}

	expected := runtime.NewInt(20)
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestMapOperations tests map literal creation and key access.
func TestMapOperations(t *testing.T) {
	interp := interpreter.New()

	// Create a simple map program
	module := &ast.Module{
		Type: "module",
		Name: "test_maps",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "test_map",
				Params:  []ast.Parameter{},
				Returns: "string",
				Body: []ast.Statement{
					{
						Type:   "assign",
						Target: "person",
						Value: &ast.Expression{
							Type: ast.ExprMapLit,
							Pairs: []ast.MapPair{
								{
									Key:   ast.Expression{Type: ast.ExprLiteral, Value: "name"},
									Value: ast.Expression{Type: ast.ExprLiteral, Value: "Alice"},
								},
								{
									Key:   ast.Expression{Type: ast.ExprLiteral, Value: "age"},
									Value: ast.Expression{Type: ast.ExprLiteral, Value: float64(30)},
								},
							},
						},
					},
					{
						Type: "return",
						Value: &ast.Expression{
							Type: ast.ExprIndex,
							Object: &ast.Expression{
								Type: ast.ExprVariable,
								Name: "person",
							},
							Index: &ast.Expression{
								Type:  ast.ExprLiteral,
								Value: "name",
							},
						},
					},
				},
			},
		},
	}

	if err := interp.LoadModule(module); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	result, err := interp.Run("test_map", []runtime.Value{})
	if err != nil {
		t.Fatalf("Runtime error: %v", err)
	}

	expected := runtime.NewString("Alice")
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestModuleSystem tests module import/export functionality.
func TestModuleSystem(t *testing.T) {
	interp := interpreter.New()

	// Create math_utils module
	mathModule := &ast.Module{
		Type:    "module",
		Name:    "math_utils",
		Exports: []string{"add", "multiply"},
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "add",
				Params:  []ast.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: "return",
						Value: &ast.Expression{
							Type:  ast.ExprBinary,
							Op:    "+",
							Left:  &ast.Expression{Type: ast.ExprVariable, Name: "a"},
							Right: &ast.Expression{Type: ast.ExprVariable, Name: "b"},
						},
					},
				},
			},
			{
				Type:    "function",
				Name:    "multiply",
				Params:  []ast.Parameter{{Name: "x", Type: "int"}, {Name: "y", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: "return",
						Value: &ast.Expression{
							Type:  ast.ExprBinary,
							Op:    "*",
							Left:  &ast.Expression{Type: ast.ExprVariable, Name: "x"},
							Right: &ast.Expression{Type: ast.ExprVariable, Name: "y"},
						},
					},
				},
			},
		},
	}

	// Create main module that imports math_utils
	mainModule := &ast.Module{
		Type:    "module",
		Name:    "main",
		Imports: []string{"math_utils"},
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "test_module",
				Params:  []ast.Parameter{},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type:   "assign",
						Target: "sum",
						Value: &ast.Expression{
							Type:   ast.ExprModuleCall,
							Module: "math_utils",
							Name:   "add",
							Args: []ast.Expression{
								{Type: ast.ExprLiteral, Value: float64(10)},
								{Type: ast.ExprLiteral, Value: float64(5)},
							},
						},
					},
					{
						Type: "return",
						Value: &ast.Expression{
							Type:   ast.ExprModuleCall,
							Module: "math_utils",
							Name:   "multiply",
							Args: []ast.Expression{
								{Type: ast.ExprVariable, Name: "sum"},
								{Type: ast.ExprLiteral, Value: float64(2)},
							},
						},
					},
				},
			},
		},
	}

	// Load math_utils first
	if err := interp.LoadModule(mathModule); err != nil {
		t.Fatalf("Failed to load math_utils module: %v", err)
	}

	// Load main module
	if err := interp.LoadModule(mainModule); err != nil {
		t.Fatalf("Failed to load main module: %v", err)
	}

	// Test module function call
	result, err := interp.Run("test_module", []runtime.Value{})
	if err != nil {
		t.Fatalf("Runtime error: %v", err)
	}

	expected := runtime.NewInt(30) // (10 + 5) * 2 = 30
	if !valuesEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestModuleExportValidation tests that non-exported functions cannot be called.
func TestModuleExportValidation(t *testing.T) {
	interp := interpreter.New()

	// Create module with private function
	mathModule := &ast.Module{
		Type:    "module",
		Name:    "math_utils",
		Exports: []string{"add"}, // Only export add, not subtract
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "add",
				Params:  []ast.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: "return",
						Value: &ast.Expression{
							Type:  ast.ExprBinary,
							Op:    "+",
							Left:  &ast.Expression{Type: ast.ExprVariable, Name: "a"},
							Right: &ast.Expression{Type: ast.ExprVariable, Name: "b"},
						},
					},
				},
			},
			{
				Type:    "function",
				Name:    "subtract", // Not exported
				Params:  []ast.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: "return",
						Value: &ast.Expression{
							Type:  ast.ExprBinary,
							Op:    "-",
							Left:  &ast.Expression{Type: ast.ExprVariable, Name: "a"},
							Right: &ast.Expression{Type: ast.ExprVariable, Name: "b"},
						},
					},
				},
			},
		},
	}

	if err := interp.LoadModule(mathModule); err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Try to call exported function - should work
	_, err := interp.RunModuleFunction("math_utils", "add", []runtime.Value{
		runtime.NewInt(10), runtime.NewInt(5),
	})
	if err != nil {
		t.Errorf("Expected exported function to work, got error: %v", err)
	}

	// Try to call non-exported function - should fail
	_, err = interp.RunModuleFunction("math_utils", "subtract", []runtime.Value{
		runtime.NewInt(10), runtime.NewInt(5),
	})
	if err == nil {
		t.Error("Expected non-exported function to fail, but it succeeded")
	}
}
