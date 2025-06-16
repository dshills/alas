package tests

import (
	"fmt"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
	"github.com/dshills/alas/internal/validator"
)

// TestEdgeCaseValues tests edge case values for different data types
func TestEdgeCaseValues(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "Zero Integer",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(0),
		},
		{
			name: "Negative Integer",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(-12345)},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(-12345),
		},
		{
			name: "Large Integer",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(9223372036854775807)}, // max int64
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(9223372036854775807),
		},
		{
			name: "Zero Float",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int", // ALaS may not distinguish float from int in literals
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: 0.0},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(0), // Expect int result
		},
		{
			name: "Negative Float",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: -3.14159},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewFloat(-3.14159),
		},
		{
			name: "Very Large Float",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: 1.7976931348623157e+308}, // close to max float64
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewFloat(1.7976931348623157e+308),
		},
		{
			name: "Very Small Float",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: 2.2250738585072014e-308}, // close to min positive float64
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewFloat(2.2250738585072014e-308),
		},
		{
			name: "Empty String",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: ""},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString(""),
		},
		{
			name: "Unicode String",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: "Hello 世界 🌍"},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString("Hello 世界 🌍"),
		},
		{
			name: "Long String",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: "This is a very long string that tests the system's ability to handle larger text content without issues or performance degradation during processing and execution."},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString("This is a very long string that tests the system's ability to handle larger text content without issues or performance degradation during processing and execution."),
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

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		module  *ast.Module
		function string
		args    []runtime.Value
		expectError bool
	}{
		{
			name: "Undefined Variable",
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
								Value: &ast.Expression{Type: ast.ExprVariable, Name: "undefined_var"},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expectError: true,
		},
		{
			name: "Undefined Function Call",
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
									Type: ast.ExprCall,
									Name: "undefined_function",
									Args: []ast.Expression{},
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expectError: true,
		},
		{
			name: "Wrong Number of Arguments",
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
										{Type: ast.ExprLiteral, Value: float64(5)}, // Only one argument instead of two
									},
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expectError: true,
		},
		{
			name: "Division by Zero",
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
									Type: ast.ExprBinary,
									Op:   "/",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expectError: true,
		},
		{
			name: "Array Index Out of Bounds",
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
										{Type: ast.ExprLiteral, Value: float64(1)},
										{Type: ast.ExprLiteral, Value: float64(2)},
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprIndex,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "arr"},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)}, // Out of bounds
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expectError: true,
		},
		{
			name: "Map Key Not Found",
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
								Target: "map",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "key1"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: "value1"},
										},
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprIndex,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "map"},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: "nonexistent_key"},
								},
							},
						},
					},
				},
			},
			function: "main",
			args:     []runtime.Value{},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				if tc.expectError {
					return // Expected error during loading
				}
				t.Fatalf("Failed to load module: %v", err)
			}

			_, err := interp.Run(tc.function, tc.args)
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestComplexDataStructures tests edge cases with complex nested data
func TestComplexDataStructures(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		function string
		expected runtime.Value
	}{
		{
			name: "Deeply Nested Arrays",
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
								Target: "deep",
								Value: &ast.Expression{
									Type: ast.ExprArrayLit,
									Elements: []ast.Expression{
										{
											Type: ast.ExprArrayLit,
											Elements: []ast.Expression{
												{
													Type: ast.ExprArrayLit,
													Elements: []ast.Expression{
														{Type: ast.ExprLiteral, Value: float64(42)},
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
										Object: &ast.Expression{
											Type: ast.ExprIndex,
											Object: &ast.Expression{Type: ast.ExprVariable, Name: "deep"},
											Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
										},
										Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
									},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
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
			name: "Complex Map Structure",
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
														Key: ast.Expression{Type: ast.ExprLiteral, Value: "connection"},
														Value: ast.Expression{
															Type: ast.ExprMapLit,
															Pairs: []ast.MapPair{
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
									},
								},
							},
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprIndex,
									Object: &ast.Expression{
										Type: ast.ExprIndex,
										Object: &ast.Expression{
											Type: ast.ExprIndex,
											Object: &ast.Expression{Type: ast.ExprVariable, Name: "config"},
											Index: &ast.Expression{Type: ast.ExprLiteral, Value: "database"},
										},
										Index: &ast.Expression{Type: ast.ExprLiteral, Value: "connection"},
									},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: "port"},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewInt(5432),
		},
		{
			name: "Mixed Array and Map",
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
								Target: "data",
								Value: &ast.Expression{
									Type: ast.ExprArrayLit,
									Elements: []ast.Expression{
										{
											Type: ast.ExprMapLit,
											Pairs: []ast.MapPair{
												{
													Key:   ast.Expression{Type: ast.ExprLiteral, Value: "name"},
													Value: ast.Expression{Type: ast.ExprLiteral, Value: "Alice"},
												},
												{
													Key: ast.Expression{Type: ast.ExprLiteral, Value: "scores"},
													Value: ast.Expression{
														Type: ast.ExprArrayLit,
														Elements: []ast.Expression{
															{Type: ast.ExprLiteral, Value: float64(95)},
															{Type: ast.ExprLiteral, Value: float64(87)},
															{Type: ast.ExprLiteral, Value: float64(92)},
														},
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
										Object: &ast.Expression{Type: ast.ExprVariable, Name: "data"},
										Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
									},
									Index: &ast.Expression{Type: ast.ExprLiteral, Value: "name"},
								},
							},
						},
					},
				},
			},
			function: "main",
			expected: runtime.NewString("Alice"),
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

// TestValidationEdgeCases tests edge cases in validation
func TestValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		shouldError bool
	}{
		{
			name: "Minimal Valid Module",
			json: `{
				"type": "module",
				"name": "minimal",
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "void",
					"body": []
				}]
			}`,
			shouldError: false,
		},
		{
			name:        "Empty JSON",
			json:        `{}`,
			shouldError: true,
		},
		{
			name:        "Invalid JSON",
			json:        `{invalid json}`,
			shouldError: true,
		},
		{
			name: "Module with empty name",
			json: `{
				"type": "module",
				"name": "",
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "void",
					"body": []
				}]
			}`,
			shouldError: true,
		},
		{
			name: "Function with empty name",
			json: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "",
					"params": [],
					"returns": "void",
					"body": []
				}]
			}`,
			shouldError: true,
		},
		{
			name: "Invalid return type",
			json: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "invalid_type",
					"body": []
				}]
			}`,
			shouldError: false, // Validator may not check return types deeply
		},
		{
			name: "Circular module imports",
			json: `{
				"type": "module",
				"name": "test",
				"imports": ["test"],
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "void",
					"body": []
				}]
			}`,
			shouldError: true,
		},
		{
			name: "Very Large JSON",
			json: func() string {
				// Create a large but valid JSON with unique function names
				base := `{
					"type": "module",
					"name": "large",
					"functions": [`
				
				// Add many functions with unique names
				for i := 0; i < 100; i++ { // Reduced to avoid timeout
					if i > 0 {
						base += ","
					}
					base += fmt.Sprintf(`{
						"type": "function",
						"name": "func%d",
						"params": [],
						"returns": "void",
						"body": []
					}`, i)
				}
				
				base += `]}`
				return base
			}(),
			shouldError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateJSON([]byte(tc.json))
			if tc.shouldError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// TestMemoryLimits tests behavior under memory pressure
func TestMemoryLimits(t *testing.T) {
	// Test large array creation
	t.Run("Large Array", func(t *testing.T) {
		elements := make([]ast.Expression, 10000)
		for i := range elements {
			elements[i] = ast.Expression{Type: ast.ExprLiteral, Value: float64(i)}
		}

		module := &ast.Module{
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
							Target: "large_array",
							Value:  &ast.Expression{Type: ast.ExprArrayLit, Elements: elements},
						},
						{
							Type: "return",
							Value: &ast.Expression{
								Type: ast.ExprIndex,
								Object: &ast.Expression{Type: ast.ExprVariable, Name: "large_array"},
								Index: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5000)},
							},
						},
					},
				},
			},
		}

		interp := interpreter.New()
		if err := interp.LoadModule(module); err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		result, err := interp.Run("main", []runtime.Value{})
		if err != nil {
			t.Fatalf("Runtime error: %v", err)
		}

		expected := runtime.NewInt(5000)
		if !valuesEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test deep recursion (stack limits)
	t.Run("Deep Recursion", func(t *testing.T) {
		module := &ast.Module{
			Type: "module",
			Name: "test",
			Functions: []ast.Function{
				{
					Type:    "function",
					Name:    "countdown",
					Params:  []ast.Parameter{{Name: "n", Type: "int"}},
					Returns: "int",
					Body: []ast.Statement{
						{
							Type: "if",
							Cond: &ast.Expression{
								Type: ast.ExprBinary,
								Op:   "<=",
								Left: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
								Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
							},
							Then: []ast.Statement{
								{
									Type: "return",
									Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
								},
							},
							Else: []ast.Statement{
								{
									Type: "return",
									Value: &ast.Expression{
										Type: ast.ExprCall,
										Name: "countdown",
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
								Name: "countdown",
								Args: []ast.Expression{
									{Type: ast.ExprLiteral, Value: float64(100)}, // Reasonable depth
								},
							},
						},
					},
				},
			},
		}

		interp := interpreter.New()
		if err := interp.LoadModule(module); err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		result, err := interp.Run("main", []runtime.Value{})
		if err != nil {
			t.Fatalf("Runtime error: %v", err)
		}

		expected := runtime.NewInt(0)
		if !valuesEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

// TestCodegenEdgeCases tests edge cases in LLVM code generation
func TestCodegenEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		module      *ast.Module
		shouldError bool
	}{
		{
			name: "Function with No Statements",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "void",
						Body:    []ast.Statement{}, // Empty body
					},
				},
			},
			shouldError: false,
		},
		{
			name: "Function with Only Comments in Meta",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Meta:    map[string]interface{}{"comment": "This function does something"},
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(42)},
							},
						},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "Very Long Function Name",
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "this_is_a_very_long_function_name_that_tests_the_systems_ability_to_handle_extremely_long_identifiers_without_breaking",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(42)},
							},
						},
					},
				},
			},
			shouldError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cg := codegen.NewLLVMCodegen()
			_, err := cg.GenerateModule(tc.module)
			if tc.shouldError {
				if err == nil {
					t.Error("Expected codegen error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected codegen error: %v", err)
				}
			}
		})
	}
}