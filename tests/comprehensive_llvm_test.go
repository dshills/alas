package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/validator"
)

// TestLLVMCodegenBasicTypes tests LLVM code generation for basic data types
func TestLLVMCodegenBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		expected string // Expected substring in LLVM IR
	}{
		{
			name: "Integer Return",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(42)},
							},
						},
					},
				},
			},
			expected: "ret i64 42",
		},
		{
			name: "Float Return",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: 3.14},
							},
						},
					},
				},
			},
			expected: "ret double",
		},
		{
			name: "Boolean Return",
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
								Value: &ast.Expression{Type: ast.ExprLiteral, Value: true},
							},
						},
					},
				},
			},
			expected: "ret i1 true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cg := codegen.NewLLVMCodegen()
			llvmModule, err := cg.GenerateModule(tc.module)
			if err != nil {
				t.Fatalf("Failed to generate LLVM IR: %v", err)
			}

			llvmIR := llvmModule.String()
			if !strings.Contains(llvmIR, tc.expected) {
				t.Errorf("Expected LLVM IR to contain '%s', but got:\n%s", tc.expected, llvmIR)
			}
		})
	}
}

// TestLLVMCodegenArithmetic tests arithmetic operations in LLVM IR
func TestLLVMCodegenArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		expected []string // Expected substrings in LLVM IR
	}{
		{
			name: "Integer Addition",
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
									Op:   "+",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(20)},
								},
							},
						},
					},
				},
			},
			expected: []string{"add i64"},
		},
		{
			name: "Integer Multiplication",
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
									Op:   "*",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(6)},
								},
							},
						},
					},
				},
			},
			expected: []string{"mul i64"},
		},
		{
			name: "Integer Division", // ALaS treats numeric literals as integers
			module: &ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int", // Changed to int
						Body: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "/",
									Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(22)},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(7)},
								},
							},
						},
					},
				},
			},
			expected: []string{"sdiv i64"}, // Changed expectation
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cg := codegen.NewLLVMCodegen()
			llvmModule, err := cg.GenerateModule(tc.module)
			if err != nil {
				t.Fatalf("Failed to generate LLVM IR: %v", err)
			}

			llvmIR := llvmModule.String()
			for _, expected := range tc.expected {
				if !strings.Contains(llvmIR, expected) {
					t.Errorf("Expected LLVM IR to contain '%s', but got:\n%s", expected, llvmIR)
				}
			}
		})
	}
}

// TestLLVMCodegenControlFlow tests control flow in LLVM IR
func TestLLVMCodegenControlFlow(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		expected []string
	}{
		{
			name: "If Statement",
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
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
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
			expected: []string{"icmp sgt", "br i1", "then:", "else:"},
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
								Target: "i",
								Value:  &ast.Expression{Type: ast.ExprLiteral, Value: float64(0)},
							},
							{
								Type: "while",
								Cond: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "<",
									Left: &ast.Expression{Type: ast.ExprVariable, Name: "i"},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
								},
								Body: []ast.Statement{
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
								Value: &ast.Expression{Type: ast.ExprVariable, Name: "i"},
							},
						},
					},
				},
			},
			expected: []string{"while.cond:", "while.body:", "while.end:", "br label %while.cond"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cg := codegen.NewLLVMCodegen()
			llvmModule, err := cg.GenerateModule(tc.module)
			if err != nil {
				t.Fatalf("Failed to generate LLVM IR: %v", err)
			}

			llvmIR := llvmModule.String()
			for _, expected := range tc.expected {
				if !strings.Contains(llvmIR, expected) {
					t.Errorf("Expected LLVM IR to contain '%s', but got:\n%s", expected, llvmIR)
				}
			}
		})
	}
}

// TestLLVMCodegenFunctions tests function calls and definitions
func TestLLVMCodegenFunctions(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		expected []string
	}{
		{
			name: "Function Call",
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
										{Type: ast.ExprLiteral, Value: float64(5)},
										{Type: ast.ExprLiteral, Value: float64(3)},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{
				"define i64 @add(i64 %a, i64 %b)",
				"define i64 @main()",
				"call i64 @add",
			},
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
				},
			},
			expected: []string{
				"define i64 @factorial(i64 %n)",
				"call i64 @factorial",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cg := codegen.NewLLVMCodegen()
			llvmModule, err := cg.GenerateModule(tc.module)
			if err != nil {
				t.Fatalf("Failed to generate LLVM IR: %v", err)
			}

			llvmIR := llvmModule.String()
			for _, expected := range tc.expected {
				if !strings.Contains(llvmIR, expected) {
					t.Errorf("Expected LLVM IR to contain '%s', but got:\n%s", expected, llvmIR)
				}
			}
		})
	}
}

// TestLLVMCodegenCompilation tests that generated LLVM IR compiles successfully
func TestLLVMCodegenCompilation(t *testing.T) {
	// Skip if llc is not available
	if _, err := exec.LookPath("llc"); err != nil {
		t.Skip("llc not found, skipping compilation test")
	}

	exampleFiles := []string{
		"examples/programs/hello.alas.json",
		"examples/programs/factorial.alas.json",
		"examples/programs/fibonacci.alas.json",
		"examples/programs/loops.alas.json",
		"examples/programs/simple_array.alas.json",
	}

	for _, file := range exampleFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Try to read the file with current path, fallback to ../
			var data []byte
			var err error
			
			data, err = os.ReadFile(file)
			if err != nil {
				// Try with ../ prefix in case we're still in tests directory
				altFile := "../" + file
				data, err = os.ReadFile(altFile)
				if err != nil {
					t.Skipf("Skipping test, file not found: %s or %s", file, altFile)
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

			// Generate LLVM IR
			cg := codegen.NewLLVMCodegen()
			llvmModule, err := cg.GenerateModule(&module)
			if err != nil {
				t.Fatalf("Failed to generate LLVM IR: %v", err)
			}

			llvmIR := llvmModule.String()

			// Write LLVM IR to temporary file
			tmpDir := t.TempDir()
			llvmFile := filepath.Join(tmpDir, "test.ll")
			objFile := filepath.Join(tmpDir, "test.o")

			if err := os.WriteFile(llvmFile, []byte(llvmIR), 0644); err != nil {
				t.Fatalf("Failed to write LLVM IR file: %v", err)
			}

			// Compile with llc
			cmd := exec.Command("llc", llvmFile, "-o", objFile)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("LLC compilation failed: %v\nOutput: %s\nLLVM IR:\n%s", err, output, llvmIR)
			}

			// Check that object file was created
			if _, err := os.Stat(objFile); os.IsNotExist(err) {
				t.Fatalf("Object file was not created")
			}
		})
	}
}

// TestLLVMMultiModuleCodegen tests multi-module compilation
func TestLLVMMultiModuleCodegen(t *testing.T) {
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
				Name:    "main",
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

	// Test multi-module compilation
	multiCodegen := codegen.NewMultiModuleCodegen()

	if err := multiCodegen.AddModule(mathModule); err != nil {
		t.Fatalf("Failed to add math module: %v", err)
	}

	if err := multiCodegen.AddModule(mainModule); err != nil {
		t.Fatalf("Failed to add main module: %v", err)
	}

	compiledModules, err := multiCodegen.CompileModules()
	if err != nil {
		t.Fatalf("Failed to compile modules: %v", err)
	}

	if len(compiledModules) != 2 {
		t.Errorf("Expected 2 compiled modules, got %d", len(compiledModules))
	}

	// Check that both modules were compiled
	if _, exists := compiledModules["math_utils"]; !exists {
		t.Error("math_utils module not found in compiled modules")
	}

	if _, exists := compiledModules["main"]; !exists {
		t.Error("main module not found in compiled modules")
	}

	// Check that the main module contains calls to math_utils functions
	if mainModule, exists := compiledModules["main"]; exists {
		mainIR := mainModule.String()
		expectedCalls := []string{
			"call i64 @math_utils__add",
			"call i64 @math_utils__multiply",
		}

		for _, expected := range expectedCalls {
			if !strings.Contains(mainIR, expected) {
				t.Errorf("Expected main module to contain '%s', but got:\n%s", expected, mainIR)
			}
		}
	}
}

// TestLLVMCodegenErrorHandling tests error conditions in LLVM code generation
func TestLLVMCodegenErrorHandling(t *testing.T) {
	tests := []struct {
		name   string
		module *ast.Module
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
								Value: &ast.Expression{Type: ast.ExprVariable, Name: "undefined"},
							},
						},
					},
				},
			},
		},
		{
			name: "Undefined Function",
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cg := codegen.NewLLVMCodegen()
			_, err := cg.GenerateModule(tc.module)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// BenchmarkLLVMCodegenPerformance benchmarks LLVM code generation performance
func BenchmarkLLVMCodegenPerformance(b *testing.B) {
	// Create a complex module for benchmarking
	module := &ast.Module{
		Type: "module",
		Name: "benchmark",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "fibonacci",
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
								Value: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
							},
						},
						Else: []ast.Statement{
							{
								Type: "return",
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   "+",
									Left: &ast.Expression{
										Type: ast.ExprCall,
										Name: "fibonacci",
										Args: []ast.Expression{
											{
												Type: ast.ExprBinary,
												Op:   "-",
												Left: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
												Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(1)},
											},
										},
									},
									Right: &ast.Expression{
										Type: ast.ExprCall,
										Name: "fibonacci",
										Args: []ast.Expression{
											{
												Type: ast.ExprBinary,
												Op:   "-",
												Left: &ast.Expression{Type: ast.ExprVariable, Name: "n"},
												Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(2)},
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg := codegen.NewLLVMCodegen()
		_, err := cg.GenerateModule(module)
		if err != nil {
			b.Fatalf("Failed to generate LLVM IR: %v", err)
		}
	}
}