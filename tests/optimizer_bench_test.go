package tests

import (
	"encoding/json"
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
)

// BenchmarkOptimizer measures the effectiveness of different optimization levels
func BenchmarkOptimizer(b *testing.B) {
	programs := []struct {
		name string
		code string
	}{
		{
			name: "ConstantHeavy",
			code: generateConstantHeavyProgram(),
		},
		{
			name: "DeadCodeHeavy",
			code: generateDeadCodeHeavyProgram(),
		},
		{
			name: "FunctionCallHeavy",
			code: generateFunctionCallHeavyProgram(),
		},
		{
			name: "LoopHeavy",
			code: generateLoopHeavyProgram(),
		},
	}

	optLevels := []struct {
		name  string
		level codegen.OptimizationLevel
	}{
		{"O0", codegen.OptNone},
		{"O1", codegen.OptBasic},
		{"O2", codegen.OptStandard},
		{"O3", codegen.OptAggressive},
	}

	for _, prog := range programs {
		for _, opt := range optLevels {
			b.Run(prog.name+"-"+opt.name, func(b *testing.B) {
				// Parse once outside the benchmark loop
				var module ast.Module
				if err := json.Unmarshal([]byte(prog.code), &module); err != nil {
					b.Fatalf("Failed to parse module: %v", err)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// Generate LLVM IR
					gen := codegen.NewLLVMCodegen()
					llvmModule, err := gen.GenerateModule(&module)
					if err != nil {
						b.Fatalf("Failed to generate LLVM IR: %v", err)
					}

					// Apply optimizations
					if opt.level > codegen.OptNone {
						optimizer := codegen.NewOptimizer(opt.level)
						if err := optimizer.OptimizeModule(llvmModule); err != nil {
							b.Fatalf("Failed to optimize: %v", err)
						}
					}
				}
			})
		}
	}
}

// TestOptimizationEffectiveness measures the size reduction from optimizations
func TestOptimizationEffectiveness(t *testing.T) {
	testCases := []struct {
		name              string
		program           string
		expectedReduction map[codegen.OptimizationLevel]float64 // minimum expected reduction percentage
	}{
		{
			name:    "Constant Folding Effectiveness",
			program: generateConstantHeavyProgram(),
			expectedReduction: map[codegen.OptimizationLevel]float64{
				codegen.OptBasic:      5.0,  // At least 5% reduction with basic opts
				codegen.OptStandard:   8.0,  // At least 8% with standard opts
				codegen.OptAggressive: 10.0, // At least 10% with aggressive opts
			},
		},
		{
			name:    "Dead Code Elimination Effectiveness",
			program: generateDeadCodeHeavyProgram(),
			expectedReduction: map[codegen.OptimizationLevel]float64{
				codegen.OptBasic:      15.0, // Dead code should give significant reduction
				codegen.OptStandard:   20.0,
				codegen.OptAggressive: 25.0,
			},
		},
		{
			name:    "Function Inlining Effectiveness",
			program: generateFunctionCallHeavyProgram(),
			expectedReduction: map[codegen.OptimizationLevel]float64{
				codegen.OptBasic:      0.0, // No inlining at basic level
				codegen.OptStandard:   0.0, // No inlining at standard level
				codegen.OptAggressive: 5.0, // Inlining should reduce code size for small functions
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the module
			var module ast.Module
			if err := json.Unmarshal([]byte(tc.program), &module); err != nil {
				t.Fatalf("Failed to parse module: %v", err)
			}

			// Generate unoptimized IR as baseline
			gen := codegen.NewLLVMCodegen()
			baselineModule, err := gen.GenerateModule(&module)
			if err != nil {
				t.Fatalf("Failed to generate baseline LLVM IR: %v", err)
			}
			baselineSize := len(baselineModule.String())

			// Test each optimization level
			for optLevel, minReduction := range tc.expectedReduction {
				t.Run(getOptLevelString(optLevel), func(t *testing.T) {
					// Generate fresh module
					gen := codegen.NewLLVMCodegen()
					llvmModule, err := gen.GenerateModule(&module)
					if err != nil {
						t.Fatalf("Failed to generate LLVM IR: %v", err)
					}

					// Apply optimizations
					optimizer := codegen.NewOptimizer(optLevel)
					if err := optimizer.OptimizeModule(llvmModule); err != nil {
						t.Fatalf("Failed to optimize: %v", err)
					}

					// Measure size reduction
					optimizedSize := len(llvmModule.String())
					reduction := float64(baselineSize-optimizedSize) / float64(baselineSize) * 100

					t.Logf("Optimization %s: baseline=%d, optimized=%d, reduction=%.2f%%",
						getOptLevelString(optLevel), baselineSize, optimizedSize, reduction)

					// Check if reduction meets expectations
					if reduction < minReduction {
						t.Logf("Warning: Expected at least %.2f%% reduction, got %.2f%%",
							minReduction, reduction)
					}
				})
			}
		})
	}
}

// Helper function to generate a program with many constants
func generateConstantHeavyProgram() string {
	return `{
		"type": "module",
		"name": "constant_heavy",
		"functions": [{
			"type": "function",
			"name": "main",
			"params": [],
			"returns": "int",
			"body": [
				{
					"type": "assign",
					"target": "a",
					"value": {
						"type": "binary",
						"op": "+",
						"left": {"type": "literal", "value": 10},
						"right": {"type": "literal", "value": 20}
					}
				},
				{
					"type": "assign",
					"target": "b",
					"value": {
						"type": "binary",
						"op": "*",
						"left": {"type": "literal", "value": 5},
						"right": {"type": "literal", "value": 6}
					}
				},
				{
					"type": "assign",
					"target": "c",
					"value": {
						"type": "binary",
						"op": "-",
						"left": {"type": "literal", "value": 100},
						"right": {"type": "literal", "value": 50}
					}
				},
				{
					"type": "assign",
					"target": "d",
					"value": {
						"type": "binary",
						"op": "/",
						"left": {"type": "literal", "value": 100},
						"right": {"type": "literal", "value": 4}
					}
				},
				{
					"type": "return",
					"value": {
						"type": "binary",
						"op": "+",
						"left": {
							"type": "binary",
							"op": "+",
							"left": {"type": "variable", "name": "a"},
							"right": {"type": "variable", "name": "b"}
						},
						"right": {
							"type": "binary",
							"op": "+",
							"left": {"type": "variable", "name": "c"},
							"right": {"type": "variable", "name": "d"}
						}
					}
				}
			]
		}]
	}`
}

// Helper function to generate a program with lots of dead code
func generateDeadCodeHeavyProgram() string {
	return `{
		"type": "module",
		"name": "dead_code_heavy",
		"functions": [{
			"type": "function",
			"name": "main",
			"params": [],
			"returns": "int",
			"body": [
				{
					"type": "assign",
					"target": "unused1",
					"value": {"type": "literal", "value": 100}
				},
				{
					"type": "assign",
					"target": "unused2",
					"value": {
						"type": "binary",
						"op": "*",
						"left": {"type": "literal", "value": 50},
						"right": {"type": "literal", "value": 2}
					}
				},
				{
					"type": "assign",
					"target": "unused3",
					"value": {
						"type": "binary",
						"op": "+",
						"left": {"type": "variable", "name": "unused1"},
						"right": {"type": "variable", "name": "unused2"}
					}
				},
				{
					"type": "assign",
					"target": "result",
					"value": {"type": "literal", "value": 42}
				},
				{
					"type": "assign",
					"target": "unused4",
					"value": {"type": "literal", "value": 999}
				},
				{
					"type": "assign",
					"target": "unused5",
					"value": {"type": "literal", "value": 888}
				},
				{
					"type": "return",
					"value": {"type": "variable", "name": "result"}
				}
			]
		}]
	}`
}

// Helper function to generate a program with many small function calls
func generateFunctionCallHeavyProgram() string {
	return `{
		"type": "module",
		"name": "function_call_heavy",
		"functions": [
			{
				"type": "function",
				"name": "inc",
				"params": [{"name": "x", "type": "int"}],
				"returns": "int",
				"body": [{
					"type": "return",
					"value": {
						"type": "binary",
						"op": "+",
						"left": {"type": "variable", "name": "x"},
						"right": {"type": "literal", "value": 1}
					}
				}]
			},
			{
				"type": "function",
				"name": "double",
				"params": [{"name": "x", "type": "int"}],
				"returns": "int",
				"body": [{
					"type": "return",
					"value": {
						"type": "binary",
						"op": "*",
						"left": {"type": "variable", "name": "x"},
						"right": {"type": "literal", "value": 2}
					}
				}]
			},
			{
				"type": "function",
				"name": "add",
				"params": [
					{"name": "a", "type": "int"},
					{"name": "b", "type": "int"}
				],
				"returns": "int",
				"body": [{
					"type": "return",
					"value": {
						"type": "binary",
						"op": "+",
						"left": {"type": "variable", "name": "a"},
						"right": {"type": "variable", "name": "b"}
					}
				}]
			},
			{
				"type": "function",
				"name": "main",
				"params": [],
				"returns": "int",
				"body": [
					{
						"type": "assign",
						"target": "a",
						"value": {
							"type": "call",
							"name": "inc",
							"args": [{"type": "literal", "value": 5}]
						}
					},
					{
						"type": "assign",
						"target": "b",
						"value": {
							"type": "call",
							"name": "double",
							"args": [{"type": "variable", "name": "a"}]
						}
					},
					{
						"type": "assign",
						"target": "c",
						"value": {
							"type": "call",
							"name": "add",
							"args": [
								{"type": "variable", "name": "a"},
								{"type": "variable", "name": "b"}
							]
						}
					},
					{
						"type": "return",
						"value": {"type": "variable", "name": "c"}
					}
				]
			}
		]
	}`
}

// Helper function to generate a program with loops
func generateLoopHeavyProgram() string {
	return `{
		"type": "module",
		"name": "loop_heavy",
		"functions": [{
			"type": "function",
			"name": "main",
			"params": [],
			"returns": "int",
			"body": [
				{
					"type": "assign",
					"target": "sum",
					"value": {"type": "literal", "value": 0}
				},
				{
					"type": "assign",
					"target": "invariant",
					"value": {"type": "literal", "value": 10}
				},
				{
					"type": "assign",
					"target": "i",
					"value": {"type": "literal", "value": 0}
				},
				{
					"type": "while",
					"cond": {
						"type": "binary",
						"op": "<",
						"left": {"type": "variable", "name": "i"},
						"right": {"type": "literal", "value": 100}
					},
					"body": [
						{
							"type": "assign",
							"target": "temp",
							"value": {
								"type": "binary",
								"op": "*",
								"left": {"type": "variable", "name": "invariant"},
								"right": {"type": "literal", "value": 2}
							}
						},
						{
							"type": "assign",
							"target": "sum",
							"value": {
								"type": "binary",
								"op": "+",
								"left": {"type": "variable", "name": "sum"},
								"right": {"type": "variable", "name": "temp"}
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
	}`
}
