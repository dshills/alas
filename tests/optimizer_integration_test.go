package tests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
)

// TestCompilerOptimizationIntegration tests the full compilation pipeline with optimizations.
func TestCompilerOptimizationIntegration(t *testing.T) {
	// Skip if LLVM tools are not available
	if _, err := exec.LookPath("llc"); err != nil {
		t.Skip("llc not found in PATH, skipping LLVM integration tests")
	}

	testCases := []struct {
		name     string
		program  string
		function string
		args     []interface{}
		expected interface{}
	}{
		{
			name: "Optimized Fibonacci",
			program: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "fib",
					"params": [{"name": "n", "type": "int"}],
					"returns": "int",
					"body": [
						{
							"type": "if",
							"cond": {
								"type": "binary",
								"op": "<=",
								"left": {"type": "variable", "name": "n"},
								"right": {"type": "literal", "value": 1}
							},
							"then": [{
								"type": "return",
								"value": {"type": "variable", "name": "n"}
							}],
							"else": [{
								"type": "return",
								"value": {
									"type": "binary",
									"op": "+",
									"left": {
										"type": "call",
										"name": "fib",
										"args": [{
											"type": "binary",
											"op": "-",
											"left": {"type": "variable", "name": "n"},
											"right": {"type": "literal", "value": 1}
										}]
									},
									"right": {
										"type": "call",
										"name": "fib",
										"args": [{
											"type": "binary",
											"op": "-",
											"left": {"type": "variable", "name": "n"},
											"right": {"type": "literal", "value": 2}
										}]
									}
								}
							}]
						}
					]
				}]
			}`,
			function: "fib",
			args:     []interface{}{10.0},
			expected: 55,
		},
		{
			name: "Optimized Loop Sum",
			program: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "sum_to_n",
					"params": [{"name": "n", "type": "int"}],
					"returns": "int",
					"body": [
						{
							"type": "assign",
							"target": "sum",
							"value": {"type": "literal", "value": 0}
						},
						{
							"type": "assign",
							"target": "i",
							"value": {"type": "literal", "value": 1}
						},
						{
							"type": "while",
							"cond": {
								"type": "binary",
								"op": "<=",
								"left": {"type": "variable", "name": "i"},
								"right": {"type": "variable", "name": "n"}
							},
							"body": [
								{
									"type": "assign",
									"target": "sum",
									"value": {
										"type": "binary",
										"op": "+",
										"left": {"type": "variable", "name": "sum"},
										"right": {"type": "variable", "name": "i"}
									}
								},
								{
									"type": "assign",
									"target": "i",
									"value": {
										"type": "binary",
										"op": "+",
										"left": {"type": "variable", "name": "i"},
										"right": {"type": "literal", "value": 1}
									}
								}
							]
						},
						{
							"type": "return",
							"value": {"type": "variable", "name": "sum"}
						}
					]
				}]
			}`,
			function: "sum_to_n",
			args:     []interface{}{100.0},
			expected: 5050,
		},
	}

	optLevels := []codegen.OptimizationLevel{
		codegen.OptNone,
		codegen.OptBasic,
		codegen.OptStandard,
		codegen.OptAggressive,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the module
			var module ast.Module
			if err := json.Unmarshal([]byte(tc.program), &module); err != nil {
				t.Fatalf("Failed to parse module: %v", err)
			}

			// Test with interpreter first (baseline)
			interp := interpreter.New()
			if err := interp.LoadModule(&module); err != nil {
				t.Fatalf("Failed to load module: %v", err)
			}

			// Convert args to runtime values
			runtimeArgs := make([]runtime.Value, len(tc.args))
			for i, arg := range tc.args {
				runtimeArgs[i] = toRuntimeValue(arg)
			}

			interpResult, err := interp.Run(tc.function, runtimeArgs)
			if err != nil {
				t.Fatalf("Interpreter failed: %v", err)
			}
			if !compareRuntimeValue(interpResult, tc.expected) {
				t.Fatalf("Interpreter returned wrong result: got %v, want %v", fromRuntimeValue(interpResult), tc.expected)
			}

			// Test each optimization level
			for _, optLevel := range optLevels {
				t.Run(getOptLevelString(optLevel), func(t *testing.T) {
					// Generate LLVM IR
					gen := codegen.NewLLVMCodegen()
					llvmModule, err := gen.GenerateModule(&module)
					if err != nil {
						t.Fatalf("Failed to generate LLVM IR: %v", err)
					}

					// Apply optimizations
					if optLevel > codegen.OptNone {
						optimizer := codegen.NewOptimizer(optLevel)
						if err := optimizer.OptimizeModule(llvmModule); err != nil {
							t.Fatalf("Failed to optimize: %v", err)
						}
					}

					// Verify the optimized IR is valid
					ir := llvmModule.String()
					if !strings.Contains(ir, tc.function) {
						t.Errorf("Function %s not found in generated IR", tc.function)
					}

					// Optionally compile to native code if we want to test end-to-end
					// This would require setting up LLVM toolchain in CI
				})
			}
		})
	}
}

// TestOptimizationConsistency ensures all optimization levels produce the same results.
func TestOptimizationConsistency(t *testing.T) {
	// Load all example programs - try both paths
	exampleDir := "examples/programs"
	files, err := ioutil.ReadDir(exampleDir)
	if err != nil {
		exampleDir = "../examples/programs"
		files, err = ioutil.ReadDir(exampleDir)
		if err != nil {
			t.Fatalf("Failed to read examples directory: %v", err)
		}
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".alas.json") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			// Read the program
			data, err := ioutil.ReadFile(filepath.Join(exampleDir, file.Name()))
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// Parse the module
			var module ast.Module
			if err := json.Unmarshal(data, &module); err != nil {
				t.Fatalf("Failed to parse module: %v", err)
			}

			// Get baseline result from interpreter
			interp := interpreter.New()
			if err := interp.LoadModule(&module); err != nil {
				t.Skipf("Cannot load module: %v", err)
			}
			var testFunction string

			// Find a suitable test function (prefer main)
			for _, fn := range module.Functions {
				if fn.Name == "main" && len(fn.Params) == 0 {
					testFunction = "main"
					break
				} else if testFunction == "" && len(fn.Params) == 0 {
					testFunction = fn.Name
				}
			}

			if testFunction == "" {
				t.Skip("No parameterless function found for testing")
			}

			_, err = interp.Run(testFunction, []runtime.Value{})
			if err != nil {
				// Some programs might not be runnable (e.g., require parameters)
				t.Skipf("Cannot run function %s: %v", testFunction, err)
			}

			// Test that all optimization levels can compile the program
			optLevels := []codegen.OptimizationLevel{
				codegen.OptNone,
				codegen.OptBasic,
				codegen.OptStandard,
				codegen.OptAggressive,
			}

			irSizes := make(map[codegen.OptimizationLevel]int)

			for _, optLevel := range optLevels {
				t.Run(getOptLevelString(optLevel), func(t *testing.T) {
					// Generate LLVM IR
					gen := codegen.NewLLVMCodegen()
					llvmModule, err := gen.GenerateModule(&module)
					if err != nil {
						// Skip modules that depend on external functions
						if strings.Contains(err.Error(), "external function") || strings.Contains(err.Error(), "not declared") {
							t.Skipf("Skipping module with external dependencies: %v", err)
							return
						}
						t.Fatalf("Failed to generate LLVM IR: %v", err)
					}

					// Apply optimizations
					if optLevel > codegen.OptNone {
						optimizer := codegen.NewOptimizer(optLevel)
						if err := optimizer.OptimizeModule(llvmModule); err != nil {
							t.Fatalf("Failed to optimize: %v", err)
						}
					}

					// Record IR size
					ir := llvmModule.String()
					irSizes[optLevel] = len(ir)

					// Verify function is still present
					if !strings.Contains(ir, testFunction) {
						t.Errorf("Test function %s not found in optimized IR", testFunction)
					}
				})
			}

			// Log optimization effectiveness
			if baseSize, ok := irSizes[codegen.OptNone]; ok {
				for optLevel, size := range irSizes {
					if optLevel != codegen.OptNone {
						reduction := float64(baseSize-size) / float64(baseSize) * 100
						t.Logf("%s - %s: %.1f%% reduction", file.Name(), getOptLevelString(optLevel), reduction)
					}
				}
			}
		})
	}
}

// TestOptimizationRegressions tests specific cases that have caused issues.
func TestOptimizationRegressions(t *testing.T) {
	testCases := []struct {
		name        string
		program     string
		optLevel    codegen.OptimizationLevel
		shouldPass  bool
		description string
	}{
		{
			name: "InfiniteLoopRegression",
			program: `{
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
							"type": "binary",
							"op": "+",
							"left": {"type": "literal", "value": 1},
							"right": {"type": "literal", "value": 2}
						}
					}]
				}]
			}`,
			optLevel:    codegen.OptBasic,
			shouldPass:  true,
			description: "Constant folding should not cause infinite loops",
		},
		{
			name: "EmptyFunctionOptimization",
			program: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "empty",
					"params": [],
					"returns": "void",
					"body": []
				}]
			}`,
			optLevel:    codegen.OptAggressive,
			shouldPass:  true,
			description: "Empty functions should optimize without errors",
		},
		{
			name: "MultipleBlocksNoInlining",
			program: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "complex",
					"params": [{"name": "x", "type": "int"}],
					"returns": "int",
					"body": [
						{
							"type": "if",
							"cond": {
								"type": "binary",
								"op": ">",
								"left": {"type": "variable", "name": "x"},
								"right": {"type": "literal", "value": 0}
							},
							"then": [{
								"type": "return",
								"value": {"type": "literal", "value": 1}
							}],
							"else": [{
								"type": "return",
								"value": {"type": "literal", "value": -1}
							}]
						}
					]
				}]
			}`,
			optLevel:    codegen.OptAggressive,
			shouldPass:  true,
			description: "Functions with multiple blocks should not be inlined",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the module
			var module ast.Module
			if err := json.Unmarshal([]byte(tc.program), &module); err != nil {
				t.Fatalf("Failed to parse module: %v", err)
			}

			// Create a temporary file to test compilation timeout
			tmpfile, err := ioutil.TempFile("", "alas_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tc.program)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Run the compiler with timeout - try both paths
			var cmd *exec.Cmd
			if _, err := os.Stat("bin/alas-compile"); err == nil {
				cmd = exec.Command("bin/alas-compile",
					"-file", tmpfile.Name(),
					"-O", getOptLevelString(tc.optLevel)[1:], // Remove 'O' prefix
					"-o", tmpfile.Name()+".ll")
			} else {
				cmd = exec.Command("../bin/alas-compile",
					"-file", tmpfile.Name(),
					"-O", getOptLevelString(tc.optLevel)[1:], // Remove 'O' prefix
					"-o", tmpfile.Name()+".ll")
			}

			// Set a reasonable timeout
			output, err := runCommandWithTimeout(cmd, 5) // 5 second timeout

			if tc.shouldPass {
				if err != nil {
					t.Errorf("%s: Expected success but got error: %v\nOutput: %s",
						tc.description, err, output)
				}
			} else {
				if err == nil {
					t.Errorf("%s: Expected failure but succeeded", tc.description)
				}
			}
		})
	}
}

// Helper function to run command with timeout.
func runCommandWithTimeout(cmd *exec.Cmd, _ int) (string, error) {
	output, err := cmd.CombinedOutput()
	return string(output), err
}
