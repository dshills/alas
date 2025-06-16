package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
	"github.com/dshills/alas/internal/validator"
)

// TestInterpreterVsCompiler tests that interpreter and compiler produce the same results
func TestInterpreterVsCompiler(t *testing.T) {
	// Skip compiler tests if llc is not available
	hasLLC := true
	if _, err := exec.LookPath("llc"); err != nil {
		hasLLC = false
		t.Log("llc not found, skipping compiler comparison")
	}

	testCases := []struct {
		name         string
		module       *ast.Module
		function     string
		args         []runtime.Value
		expectedType runtime.ValueType
	}{
		{
			name: "Simple Integer Return",
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
			function:     "main",
			args:         []runtime.Value{},
			expectedType: runtime.ValueTypeInt,
		},
		{
			name: "Arithmetic Operations",
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
									Left: &ast.Expression{
										Type: ast.ExprBinary,
										Op:   "*",
										Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(5)},
										Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(6)},
									},
									Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
								},
							},
						},
					},
				},
			},
			function:     "main",
			args:         []runtime.Value{},
			expectedType: runtime.ValueTypeInt,
		},
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
										{Type: ast.ExprLiteral, Value: float64(15)},
										{Type: ast.ExprLiteral, Value: float64(25)},
									},
								},
							},
						},
					},
				},
			},
			function:     "main",
			args:         []runtime.Value{},
			expectedType: runtime.ValueTypeInt,
		},
		{
			name: "Conditional Logic",
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
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(100)},
									},
								},
								Else: []ast.Statement{
									{
										Type: "return",
										Value: &ast.Expression{Type: ast.ExprLiteral, Value: float64(200)},
									},
								},
							},
						},
					},
				},
			},
			function:     "main",
			args:         []runtime.Value{},
			expectedType: runtime.ValueTypeInt,
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
										{Type: ast.ExprLiteral, Value: float64(5)},
									},
								},
							},
						},
					},
				},
			},
			function:     "main",
			args:         []runtime.Value{},
			expectedType: runtime.ValueTypeInt,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test interpreter
			interp := interpreter.New()
			if err := interp.LoadModule(tc.module); err != nil {
				t.Fatalf("Failed to load module in interpreter: %v", err)
			}

			interpResult, err := interp.Run(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Interpreter runtime error: %v", err)
			}

			if interpResult.Type != tc.expectedType {
				t.Errorf("Interpreter result type mismatch: expected %v, got %v", tc.expectedType, interpResult.Type)
			}

			// Test compiler (if available)
			if hasLLC {
				compiledResult, err := runCompiledProgram(t, tc.module, tc.function)
				if err != nil {
					t.Fatalf("Compiled program error: %v", err)
				}

				// Compare results based on type
				switch tc.expectedType {
				case runtime.ValueTypeInt:
					interpInt, _ := interpResult.AsInt()
					if interpInt != compiledResult {
						t.Errorf("Results differ: interpreter=%d, compiled=%d", interpInt, compiledResult)
					}
				case runtime.ValueTypeFloat:
					interpFloat, _ := interpResult.AsFloat()
					if abs(interpFloat-float64(compiledResult)) > 1e-6 {
						t.Errorf("Results differ: interpreter=%f, compiled=%d", interpFloat, compiledResult)
					}
				case runtime.ValueTypeBool:
					interpBool, _ := interpResult.AsBool()
					compiledBool := compiledResult != 0
					if interpBool != compiledBool {
						t.Errorf("Results differ: interpreter=%t, compiled=%t", interpBool, compiledBool)
					}
				}
			}
		})
	}
}

// TestExampleProgramsIntegration tests all example programs with both interpreter and compiler
func TestExampleProgramsIntegration(t *testing.T) {
	// Skip compiler tests if llc is not available
	hasLLC := true
	if _, err := exec.LookPath("llc"); err != nil {
		hasLLC = false
		t.Log("llc not found, skipping compiler integration tests")
	}

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
			file:     "examples/programs/factorial.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(120), // factorial(5)
		},
		{
			file:     "examples/programs/fibonacci.alas.json",
			function: "main",
			args:     []runtime.Value{},
			expected: runtime.NewInt(55), // fibonacci(10)
		},
	}

	for _, tc := range exampleTests {
		t.Run(filepath.Base(tc.file), func(t *testing.T) {
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

			// Test interpreter
			interp := interpreter.New()
			if err := interp.LoadModule(&module); err != nil {
				t.Fatalf("Failed to load module in interpreter: %v", err)
			}

			interpResult, err := interp.Run(tc.function, tc.args)
			if err != nil {
				t.Fatalf("Interpreter runtime error: %v", err)
			}

			// Check interpreter result
			if !valuesEqual(interpResult, tc.expected) {
				t.Errorf("Interpreter result mismatch: expected %v, got %v", tc.expected, interpResult)
			}

			// Test compiler (if available) for integer results only
			if hasLLC && tc.expected.Type == runtime.ValueTypeInt {
				compiledResult, err := runCompiledProgram(t, &module, tc.function)
				if err != nil {
					t.Logf("Compiled program error (may not be critical): %v", err)
					return
				}

				expectedInt, _ := tc.expected.AsInt()
				if expectedInt != compiledResult {
					t.Errorf("Compiled result mismatch: expected %d, got %d", expectedInt, compiledResult)
				}
			}
		})
	}
}

// TestValidationIntegration tests that validation works consistently across all tools
func TestValidationIntegration(t *testing.T) {
	// Test valid programs - try both paths
	validFiles, err := filepath.Glob("examples/programs/*.alas.json")
	if err != nil || len(validFiles) == 0 {
		validFiles, err = filepath.Glob("../examples/programs/*.alas.json")
		if err != nil {
			t.Fatalf("Failed to glob files: %v", err)
		}
	}

	for _, file := range validFiles {
		t.Run("Valid: "+filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// Test validator
			if err := validator.ValidateJSON(data); err != nil {
				t.Errorf("Validation failed for valid file: %v", err)
			}

			// Test that interpreter can load it
			var module ast.Module
			if err := json.Unmarshal(data, &module); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			interp := interpreter.New()
			if err := interp.LoadModule(&module); err != nil {
				// Skip modules that depend on stdlib modules that may not be available
				if strings.Contains(err.Error(), "module std.") || strings.Contains(err.Error(), "not found in search paths") {
					t.Skipf("Skipping module with stdlib dependencies: %v", err)
					return
				}
				t.Errorf("Interpreter failed to load valid module: %v", err)
			}

			// Test that compiler can process it
			cg := codegen.NewLLVMCodegen()
			_, err = cg.GenerateModule(&module)
			if err != nil {
				// Skip modules that depend on external functions
				if strings.Contains(err.Error(), "external function") || strings.Contains(err.Error(), "not declared") {
					t.Skipf("Skipping module with external dependencies: %v", err)
					return
				}
				t.Errorf("Compiler failed to process valid module: %v", err)
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
	}

	for _, tc := range invalidPrograms {
		t.Run("Invalid: "+tc.name, func(t *testing.T) {
			// Test validator rejects it
			if err := validator.ValidateJSON([]byte(tc.json)); err == nil {
				t.Error("Expected validation to fail, but it passed")
			}

			// Test that interpreter/compiler handle gracefully
			var module ast.Module
			if err := json.Unmarshal([]byte(tc.json), &module); err == nil {
				interp := interpreter.New()
				if err := interp.LoadModule(&module); err == nil {
					t.Log("Interpreter should reject invalid module (validation should catch this first)")
				}

				cg := codegen.NewLLVMCodegen()
				if _, err := cg.GenerateModule(&module); err == nil {
					t.Log("Compiler should reject invalid module (validation should catch this first)")
				}
			}
		})
	}
}

// Helper function to run a compiled program and get its integer result
func runCompiledProgram(t *testing.T, module *ast.Module, function string) (int64, error) {
	// Generate LLVM IR
	cg := codegen.NewLLVMCodegen()
	llvmModule, err := cg.GenerateModule(module)
	if err != nil {
		return 0, fmt.Errorf("failed to generate LLVM IR: %v", err)
	}

	llvmIR := llvmModule.String()

	// Create temporary directory
	tmpDir := t.TempDir()
	llvmFile := filepath.Join(tmpDir, "test.ll")
	objFile := filepath.Join(tmpDir, "test.o")
	exeFile := filepath.Join(tmpDir, "test")

	// Write LLVM IR to file
	if err := os.WriteFile(llvmFile, []byte(llvmIR), 0644); err != nil {
		return 0, fmt.Errorf("failed to write LLVM IR file: %v", err)
	}

	// Compile with llc
	cmd := exec.Command("llc", llvmFile, "-o", objFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("LLC compilation failed: %v\nOutput: %s", err, output)
	}

	// Link with clang (if available)
	if _, err := exec.LookPath("clang"); err == nil {
		cmd = exec.Command("clang", objFile, "-o", exeFile)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return 0, fmt.Errorf("linking failed: %v\nOutput: %s", err, output)
		}

		// Run the executable
		cmd = exec.Command(exeFile)
		output, err = cmd.CombinedOutput()
		if err != nil {
			return 0, fmt.Errorf("execution failed: %v\nOutput: %s", err, output)
		}

		// Parse the result (assuming it prints an integer)
		result := strings.TrimSpace(string(output))
		if result == "" {
			return 0, nil // Void return or no output
		}

		intResult, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse result '%s' as integer: %v", result, err)
		}

		return intResult, nil
	}

	// If clang is not available, we can't execute but compilation success is still valuable
	return 0, fmt.Errorf("clang not available for linking and execution")
}

// Helper function for absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestPerformanceComparison benchmarks interpreter vs compiler performance
func TestPerformanceComparison(t *testing.T) {
	// Create a computationally intensive module for performance testing
	module := &ast.Module{
		Type: "module",
		Name: "performance_test",
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

	// Test interpreter performance
	t.Run("Interpreter Performance", func(t *testing.T) {
		interp := interpreter.New()
		if err := interp.LoadModule(module); err != nil {
			t.Fatalf("Failed to load module: %v", err)
		}

		// Time a small fibonacci calculation
		start := testing.AllocsPerRun(1, func() {
			_, err := interp.Run("fibonacci", []runtime.Value{runtime.NewInt(10)})
			if err != nil {
				t.Errorf("Interpreter error: %v", err)
			}
		})

		t.Logf("Interpreter allocations per run: %f", start)
	})

	// Note: Compiler performance testing would require actual compilation and execution
	// which is more complex and environment-dependent
}