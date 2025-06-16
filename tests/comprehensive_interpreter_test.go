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

// TestBasicDataTypes tests all basic data type operations
func TestBasicDataTypes(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "Integer Arithmetic",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   "assign",
								Target: "result",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "+",
									Left: &ast.Expression{
										Type: ast.ExprBinary,
										Op:   "*",
										Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
										Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
									},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(3)},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{Type: ast.ExprVariable, Name: "result"},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(53), // (10 * 5) + 3 = 53
		},
		{
			name: "Float Operations",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "float",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "/",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: 22.0},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: 7.0},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(3), // Integer division in ALaS
		},
		{
			name: "String Concatenation",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "string",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "+",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: "Hello, "},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: "World!"},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString("Hello, World!"),
		},
		{
			name: "Boolean Logic",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "bool",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "&&",
									Left: &ast.Expression{
										Type: ast.ExprBinary,
										Op:   ">",
										Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
										Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
									},
									Right: &ast.Expression{
										Type: ast.ExprBinary,
										Op:   "==",
										Left: &ast.Expression{Type: ast.ExprLiteral, Value: "test"},
										Right: &ast.Expression{Type: ast.ExprLiteral, Value: "test"},
									},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewBool(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, []runtime.Value{})
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestControlFlow tests all control flow constructs
func TestControlFlow(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "If Statement - True Branch",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "if",
								Cond: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   ">",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
								},
								Then: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(42)},
									},
								},
								Else: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
									},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(42),
		},
		{
			name: "If Statement - False Branch",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "if",
								Cond: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "<",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
								},
								Then: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(42)},
									},
								},
								Else: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(99)},
									},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(99),
		},
		{
			name: "While Loop",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   "assign",
								Target: "sum",
								Value:  &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
							},
							{
								Type:   "assign",
								Target: "i",
								Value:  &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
							},
							{
								Type: "while",
								Cond: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "<=",
									Left: &ast.Expression{Type: ast.ExprVariable, Name: "i"},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
								},
								Body: []ast.Statement{
									{
										Type:   "assign",
										Target: "sum",
										Value: &ast.Expression{
											Type: ast.ExprBinary,
											Op:   "+",
											Left: &ast.Expression{Type: ast.ExprVariable, Name: "sum"},
											Right: &ast.Expression{Type: ast.ExprVariable, Name: "i"},
										},
									},
									{
										Type:   "assign",
										Target: "i",
										Value: &ast.Expression{
											Type: ast.ExprBinary,
											Op:   "+",
											Left: &ast.Expression{Type: ast.ExprVariable, Name: "i"},
											Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
										},
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{Type: ast.ExprVariable, Name: "sum"},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(15), // 1+2+3+4+5 = 15
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, []runtime.Value{})
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestArrayOperations tests comprehensive array functionality
func TestArrayOperations(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "Array Creation and Access",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   "assign",
								Target: "arr",
								Value: &ast.Expression{
									Type: ast.ExprArrayLit,
									Elements: []ast.Expression{
										{Type: ast.ExprLiteral, Value: float64(100)},
										{Type: ast.ExprLiteral, Value: float64(200)},
										{Type: ast.ExprLiteral, Value: float64(300)},
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprIndex,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "arr"},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(2)},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(300),
		},
		{
			name: "Nested Arrays",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   "assign",
								Target: "matrix",
								Value: &ast.Expression{
									Type: ast.ExprArrayLit,
									Elements: []ast.Expression{
										{
											Type: ast.ExprArrayLit,
											Elements: []ast.Expression{
												{Type: ast.ExprLiteral, Value: float64(1)},
												{Type: ast.ExprLiteral, Value: float64(2)},
											},
										},
										{
											Type: ast.ExprArrayLit,
											Elements: []ast.Expression{
												{Type: ast.ExprLiteral, Value: float64(3)},
												{Type: ast.ExprLiteral, Value: float64(4)},
											},
										},
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprIndex,
									Object: &ast.Expression{
										Type: ast.ExprIndex,
										Object: &ast.Expression{Type: ast.ExprVariable, Name: "matrix"},
										Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
									},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, []runtime.Value{})
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestMapOperations tests comprehensive map functionality
func TestMapOperations(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "Map Creation and Access",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
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
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "firstName"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: "John"},
										},
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "lastName"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: "Doe"},
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
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "person"},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: "firstName"},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString("John"),
		},
		{
			name: "Nested Maps",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "string",
						Body: []ast.Statement{
							{
								Type:   "assign",
								Target: "config",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key: ast.Expression{Type: ast.ExprLiteral, Value: "database"},
											Value: ast.Expression{
												Type: ast.ExprMapLit,
												Pairs: []ast.MapPair{
													{
														Key:   ast.Expression{Type: ast.ExprLiteral, Value: "host"},
														Value: ast.Expression{Type: ast.ExprLiteral, Value: "localhost"},
													},
													{
														Key:   ast.Expression{Type: ast.ExprLiteral, Value: "port"},
														Value: ast.Expression{Type: ast.ExprLiteral, Value: float64(5432)},
													},
												},
											},
										},
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprIndex,
									Object: &ast.Expression{
										Type: ast.ExprIndex,
										Object: &ast.Expression{Type: ast.ExprVariable, Name: "config"},
										Index: &ast.Expression{Type: ast.ExprLiteral, Value: "database"},
									},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: "host"},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString("localhost"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, []runtime.Value{})
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestFunctionCalls tests function call functionality
func TestFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		args     []runtime.Value
		expected runtime.Value
	}{
		{
			name: "Simple Function Call",
			module: &ast.Module{
				Type: "module",
				Name: "test",
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
									Type: ast.ExprBinary,
									Op:   "+",
									Left: &ast.Expression{Type: ast.ExprVariable, Name: "a"},
									Right: &ast.Expression{Type: ast.ExprVariable, Name: "b"},
								},
							},
						},
					},
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprCall,
									Name: "add",
									Args: []ast.Expression{
										{Type: ast.ExprLiteral, Value: float64(15)},
										{Type: ast.ExprLiteral, Value: float64(25)},
									},
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(40),
		},
		{
			name: "Recursive Function",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "factorial",
						Params:  []ast.Parameter{{Name: "n", Type: "int"}},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "if",
								Cond: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "<=",
									Left: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
								},
								Then: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
									},
								},
								Else: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{
											Type: ast.ExprBinary,
											Op:   "*",
											Left: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
											Right: &ast.Expression{
												Type: ast.ExprCall,
												Name: "factorial",
												Args: []ast.Expression{
													{
														Type: ast.ExprBinary,
														Op:   "-",
														Left: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
														Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprCall,
									Name: "factorial",
									Args: []ast.Expression{
										{Type: ast.ExprLiteral, Value: float64(6)},
									},
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(720), // 6! = 720
		},
		{
			name: "Function with Parameters",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "greet",
						Params:  []ast.Parameter{{Name: "name", Type: "string"}},
						Returns: "string",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "+",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: "Hello, "},
									Right: &ast.Expression{Type: ast.ExprVariable, Name: "name"},
								},
							},
						},
					},
				},
			},
			function: "greet",
			args:     []runtime.Value{runtime.NewString("Alice")},
			expected: runtime.NewString("Hello, Alice"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestUnaryOperations tests unary operators
func TestUnaryOperations(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "Negation",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprUnary,
									Op:   "-",
									Operand: &ast.Expression{Type: ast.ExprLiteral, Value: float64(42)},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(-42),
		},
		{
			name: "Logical NOT",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "bool",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprUnary,
									Op:   "!",
									Operand: &ast.Expression{Type: ast.ExprLiteral, Value: false},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewBool(true),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			result, err := interp.Run(tc.function, []runtime.Value{})
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			if !valuesEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestAllExamplePrograms tests all example programs systematically
func TestAllExamplePrograms(t *testing.T) {
	exampleTests := []struct {
		file     string
		function string
		args     []runtime.Value
		expected runtime.Value
	}{
		{
			file:     "examples/programs/hello.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewString("Hello, ALaS!"),
		},
		{
			file:     "examples/programs/fibonacci.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(55), // fibonacci(10)
		},
		{
			file:     "examples/programs/factorial.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(120), // factorial(5)
		},
		{
			file:     "examples/programs/loops.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(55), // sum 1 to 10
		},
		{
			file:     "examples/programs/simple_array.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(20),
		},
	}

	for _, tc := range exampleTests {
		t.Run(tc.file, func(t *testing.T) {
			// Try to read the file with current path, fallback to ../
			var data []byte
			var err error
			
			data, err = os.ReadFile(tc.file)
			if err != nil {
				// Try with ../ prefix in case we're still in tests directory
				altFile := "../" + tc.file
				data, err = os.ReadFile(altFile)
				if err != nil {
					t.Skipf("Skipping test, file not found: %s or %s", tc.file, altFile)
					return
				}
			}

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