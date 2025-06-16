package tests

import (
	"testing"
	"time"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
)

// TestComprehensiveSuite runs the complete test suite with proper cleanup and reporting
func TestComprehensiveSuite(t *testing.T) {
	suite := NewTestSuite()
	suite.Setup(t)
	
	// Validate test environment
	suite.ValidateTestEnvironment(t)
	
	// Check initial memory usage
	suite.CheckMemoryUsage(t, "start")
	
	// Run all test categories
	t.Run("BasicDataTypes", func(t *testing.T) {
		TestBasicDataTypes(t)
	})
	
	t.Run("ControlFlow", func(t *testing.T) {
		TestControlFlow(t)
	})
	
	t.Run("ArrayOperations", func(t *testing.T) {
		TestArrayOperations(t)
	})
	
	t.Run("MapOperations", func(t *testing.T) {
		TestMapOperations(t)
	})
	
	t.Run("FunctionCalls", func(t *testing.T) {
		TestFunctionCalls(t)
	})
	
	t.Run("UnaryOperations", func(t *testing.T) {
		TestUnaryOperations(t)
	})
	
	t.Run("AllExamplePrograms", func(t *testing.T) {
		TestAllExamplePrograms(t)
	})
	
	// LLVM Tests
	t.Run("LLVMBasicTypes", func(t *testing.T) {
		TestLLVMCodegenBasicTypes(t)
	})
	
	t.Run("LLVMArithmetic", func(t *testing.T) {
		TestLLVMCodegenArithmetic(t)
	})
	
	t.Run("LLVMControlFlow", func(t *testing.T) {
		TestLLVMCodegenControlFlow(t)
	})
	
	t.Run("LLVMFunctions", func(t *testing.T) {
		TestLLVMCodegenFunctions(t)
	})
	
	t.Run("LLVMCompilation", func(t *testing.T) {
		TestLLVMCodegenCompilation(t)
	})
	
	t.Run("LLVMMultiModule", func(t *testing.T) {
		TestLLVMMultiModuleCodegen(t)
	})
	
	t.Run("LLVMErrorHandling", func(t *testing.T) {
		TestLLVMCodegenErrorHandling(t)
	})
	
	// Integration Tests
	t.Run("InterpreterVsCompiler", func(t *testing.T) {
		TestInterpreterVsCompiler(t)
	})
	
	t.Run("ExampleProgramsIntegration", func(t *testing.T) {
		TestExampleProgramsIntegration(t)
	})
	
	t.Run("ValidationIntegration", func(t *testing.T) {
		TestValidationIntegration(t)
	})
	
	// Edge Cases and Error Handling
	t.Run("EdgeCaseValues", func(t *testing.T) {
		TestEdgeCaseValues(t)
	})
	
	t.Run("ErrorHandling", func(t *testing.T) {
		TestErrorHandling(t)
	})
	
	t.Run("ComplexDataStructures", func(t *testing.T) {
		TestComplexDataStructures(t)
	})
	
	t.Run("ValidationEdgeCases", func(t *testing.T) {
		TestValidationEdgeCases(t)
	})
	
	t.Run("MemoryLimits", func(t *testing.T) {
		TestMemoryLimits(t)
	})
	
	t.Run("CodegenEdgeCases", func(t *testing.T) {
		TestCodegenEdgeCases(t)
	})
	
	// Force garbage collection and check memory usage
	suite.ForceGC(t)
	suite.CheckMemoryUsage(t, "after_gc")
	
	t.Log("Comprehensive test suite completed successfully")
}

// TestParallelExecution tests running multiple interpreters in parallel
func TestParallelExecution(t *testing.T) {
	suite := NewTestSuite()
	suite.Setup(t)
	
	// Create parallel test functions
	parallelTests := []func(*testing.T){
		func(t *testing.T) { TestBasicDataTypes(t) },
		func(t *testing.T) { TestControlFlow(t) },
		func(t *testing.T) { TestArrayOperations(t) },
		func(t *testing.T) { TestMapOperations(t) },
		func(t *testing.T) { TestFunctionCalls(t) },
	}
	
	// Run tests in parallel
	runner := suite.NewParallelTestRunner(3) // Limit to 3 concurrent tests
	runner.RunTests(t, parallelTests)
}

// TestPerformanceStress runs performance stress tests
func TestPerformanceStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	suite := NewTestSuite()
	suite.Setup(t)
	
	// Check memory before stress test
	suite.CheckMemoryUsage(t, "before_stress")
	
	t.Run("StressInterpreter", func(t *testing.T) {
		start := time.Now()
		
		// Run the same test many times to check for memory leaks
		for i := 0; i < 100; i++ {
			TestBasicDataTypes(t)
			if i%20 == 0 {
				suite.CheckMemoryUsage(t, "stress_iteration_"+string(rune('0'+i/20)))
			}
		}
		
		duration := time.Since(start)
		t.Logf("Stress test completed in %v (100 iterations)", duration)
	})
	
	// Force GC and check final memory usage
	suite.ForceGC(t)
	suite.CheckMemoryUsage(t, "after_stress")
}

// BenchmarkInterpreterPerformance benchmarks interpreter performance
func BenchmarkInterpreterPerformance(b *testing.B) {
	suite := NewTestSuite()
	
	benchHelper := suite.NewBenchmarkHelper()
	
	benchHelper.RunBenchmark(b, "BasicArithmetic", func(b *testing.B) {
		// Create a simple arithmetic module for benchmarking
		module := createBasicArithmeticModule()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			interp := interpreter.New()
			if err := interp.LoadModule(module); err != nil {
				b.Fatalf("Failed to load module: %v", err)
			}
			
			_, err := interp.Run("main", []runtime.Value{})
			if err != nil {
				b.Fatalf("Runtime error: %v", err)
			}
		}
	})
	
	benchHelper.RunBenchmark(b, "RecursiveFunction", func(b *testing.B) {
		// Create a recursive factorial module for benchmarking
		module := createFactorialModule()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			interp := interpreter.New()
			if err := interp.LoadModule(module); err != nil {
				b.Fatalf("Failed to load module: %v", err)
			}
			
			_, err := interp.Run("main", []runtime.Value{})
			if err != nil {
				b.Fatalf("Runtime error: %v", err)
			}
		}
	})
}

// BenchmarkLLVMCodegen benchmarks LLVM code generation performance
func BenchmarkLLVMCodegen(b *testing.B) {
	suite := NewTestSuite()
	
	benchHelper := suite.NewBenchmarkHelper()
	
	benchHelper.RunBenchmark(b, "SimpleModule", func(b *testing.B) {
		module := createBasicArithmeticModule()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cg := codegen.NewLLVMCodegen()
			_, err := cg.GenerateModule(module)
			if err != nil {
				b.Fatalf("Failed to generate LLVM IR: %v", err)
			}
		}
	})
	
	benchHelper.RunBenchmark(b, "ComplexModule", func(b *testing.B) {
		module := createComplexModule()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cg := codegen.NewLLVMCodegen()
			_, err := cg.GenerateModule(module)
			if err != nil {
				b.Fatalf("Failed to generate LLVM IR: %v", err)
			}
		}
	})
}

// Helper functions to create test modules

func createBasicArithmeticModule() *ast.Module {
	return &ast.Module{
		Type: "module",
		Name: "arithmetic",
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
								Left: &ast.Expression{Type: ast.ExprLiteral, Value: float64(10)},
								Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(20)},
							},
							Right: &ast.Expression{Type: ast.ExprLiteral, Value: float64(30)},
						},
					},
				},
			},
		},
	}
}

func createFactorialModule() *ast.Module {
	return &ast.Module{
		Type: "module",
		Name: "factorial",
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
	}
}

func createComplexModule() *ast.Module {
	return &ast.Module{
		Type: "module",
		Name: "complex",
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
			{
				Type:    "function",
				Name:    "helper",
				Params:  []ast.Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: "return",
						Value: &ast.Expression{
							Type: ast.ExprBinary,
							Op:   "+",
							Left: &ast.Expression{
								Type: ast.ExprCall,
								Name: "fibonacci",
								Args: []ast.Expression{
									{Type: ast.ExprVariable, Name: "a"},
								},
							},
							Right: &ast.Expression{
								Type: ast.ExprCall,
								Name: "fibonacci",
								Args: []ast.Expression{
									{Type: ast.ExprVariable, Name: "b"},
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
							Name: "helper",
							Args: []ast.Expression{
								{Type: ast.ExprLiteral, Value: float64(5)},
								{Type: ast.ExprLiteral, Value: float64(3)},
							},
						},
					},
				},
			},
		},
	}
}