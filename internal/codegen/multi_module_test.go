package codegen

import (
	"errors"
	"testing"

	"github.com/llir/llvm/ir/types"

	"github.com/dshills/alas/internal/ast"
)

func TestMultiModuleCodegen_AddModule(t *testing.T) {
	codegen := NewMultiModuleCodegen()

	module := &ast.Module{
		Name: "test_module",
		Functions: []ast.Function{
			{
				Name:    "test_func",
				Params:  []ast.Parameter{{Name: "x", Type: "int"}},
				Returns: "int",
				Body:    []ast.Statement{},
			},
		},
	}

	err := codegen.AddModule(module)
	if err != nil {
		t.Fatalf("AddModule failed: %v", err)
	}

	if len(codegen.modules) != 1 {
		t.Errorf("Expected 1 module, got %d", len(codegen.modules))
	}

	if codegen.modules["test_module"] != module {
		t.Errorf("Module not stored correctly")
	}
}

func TestMultiModuleCodegen_ResolveDependencies(t *testing.T) {
	codegen := NewMultiModuleCodegen()

	// Create modules with dependencies: A -> B -> C
	moduleC := &ast.Module{
		Name:      "moduleC",
		Imports:   []string{},
		Functions: []ast.Function{{Name: "funcC", Returns: "int"}},
	}

	moduleB := &ast.Module{
		Name:      "moduleB",
		Imports:   []string{"moduleC"},
		Functions: []ast.Function{{Name: "funcB", Returns: "int"}},
	}

	moduleA := &ast.Module{
		Name:      "moduleA",
		Imports:   []string{"moduleB"},
		Functions: []ast.Function{{Name: "funcA", Returns: "int"}},
	}

	// Add modules
	err := codegen.AddModule(moduleA)
	if err != nil {
		t.Fatalf("Failed to add moduleA: %v", err)
	}

	err = codegen.AddModule(moduleB)
	if err != nil {
		t.Fatalf("Failed to add moduleB: %v", err)
	}

	err = codegen.AddModule(moduleC)
	if err != nil {
		t.Fatalf("Failed to add moduleC: %v", err)
	}

	// Resolve dependencies
	order, err := codegen.ResolveDependencies()
	if err != nil {
		t.Fatalf("ResolveDependencies failed: %v", err)
	}

	// Should compile in order: C, B, A (dependencies first)
	expectedOrder := []string{"moduleC", "moduleB", "moduleA"}
	if len(order) != len(expectedOrder) {
		t.Fatalf("Expected %d modules, got %d", len(expectedOrder), len(order))
	}

	for i, expected := range expectedOrder {
		if order[i] != expected {
			t.Errorf("Expected order[%d] = %s, got %s", i, expected, order[i])
		}
	}
}

func TestMultiModuleCodegen_CircularDependency(t *testing.T) {
	codegen := NewMultiModuleCodegen()

	// Create circular dependency: A -> B -> A
	moduleA := &ast.Module{
		Name:      "moduleA",
		Imports:   []string{"moduleB"},
		Functions: []ast.Function{{Name: "funcA", Returns: "int"}},
	}

	moduleB := &ast.Module{
		Name:      "moduleB",
		Imports:   []string{"moduleA"},
		Functions: []ast.Function{{Name: "funcB", Returns: "int"}},
	}

	err := codegen.AddModule(moduleA)
	if err != nil {
		t.Fatalf("Failed to add moduleA: %v", err)
	}

	err = codegen.AddModule(moduleB)
	if err != nil {
		t.Fatalf("Failed to add moduleB: %v", err)
	}

	// Should detect circular dependency
	_, err = codegen.ResolveDependencies()
	if err == nil {
		t.Fatal("Expected circular dependency error, got nil")
	}
}

func TestMultiModuleCodegen_GetQualifiedFunctionName(t *testing.T) {
	codegen := NewMultiModuleCodegen()

	qualifiedName := codegen.GetQualifiedFunctionName("math_utils", "add")
	expected := "math_utils__add"

	if qualifiedName != expected {
		t.Errorf("Expected %s, got %s", expected, qualifiedName)
	}
}

func TestMultiModuleCodegen_RegisterModuleLoader(t *testing.T) {
	codegen := NewMultiModuleCodegen()

	testModule := &ast.Module{
		Name: "test_loader_module",
	}

	loader := func(name string) (*ast.Module, error) {
		if name == "test_loader_module" {
			return testModule, nil
		}
		return nil, errors.New("module not found")
	}

	codegen.RegisterModuleLoader("test_loader_module", loader)

	if len(codegen.moduleLoaders) != 1 {
		t.Errorf("Expected 1 module loader, got %d", len(codegen.moduleLoaders))
	}

	// Test loading
	loadedModule, err := codegen.LoadModule("test_loader_module")
	if err != nil {
		t.Fatalf("LoadModule failed: %v", err)
	}

	if loadedModule != testModule {
		t.Errorf("Loaded module is not the expected module")
	}
}

func TestLLVMCodegen_DeclareExternalFunction(t *testing.T) {
	codegen := NewLLVMCodegen()

	// Test declaring an external function
	paramTypes, err := convertALaSTypesToLLVM([]string{"int", "int"})
	if err != nil {
		t.Fatalf("Failed to convert parameter types: %v", err)
	}

	returnType, err := codegen.convertType("int")
	if err != nil {
		t.Fatalf("Failed to convert return type: %v", err)
	}

	externalFunc, err := codegen.DeclareExternalFunction("math_utils", "add", paramTypes, returnType)
	if err != nil {
		t.Fatalf("DeclareExternalFunction failed: %v", err)
	}

	if externalFunc == nil {
		t.Fatal("External function is nil")
	}

	expectedName := "math_utils__add"
	if externalFunc.Name() != expectedName {
		t.Errorf("Expected function name %s, got %s", expectedName, externalFunc.Name())
	}

	// Test that redeclaring returns the same function
	externalFunc2, err := codegen.DeclareExternalFunction("math_utils", "add", paramTypes, returnType)
	if err != nil {
		t.Fatalf("Redeclaring external function failed: %v", err)
	}

	if externalFunc != externalFunc2 {
		t.Errorf("Redeclaring should return the same function instance")
	}
}

// Helper function to convert ALaS types to LLVM types for testing.
func convertALaSTypesToLLVM(alasTypes []string) ([]types.Type, error) {
	codegen := NewLLVMCodegen()
	result := make([]types.Type, len(alasTypes))
	for i, alasType := range alasTypes {
		llvmType, err := codegen.convertType(alasType)
		if err != nil {
			return nil, err
		}
		result[i] = llvmType
	}
	return result, nil
}

func TestMultiModuleCodegen_CompileModules_Simple(t *testing.T) {
	codegen := NewMultiModuleCodegen()

	// Create a simple module without dependencies
	module := &ast.Module{
		Name: "simple_module",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "simple_func",
				Params:  []ast.Parameter{{Name: "x", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: "return",
						Value: &ast.Expression{
							Type: "variable",
							Name: "x",
						},
					},
				},
			},
		},
	}

	err := codegen.AddModule(module)
	if err != nil {
		t.Fatalf("AddModule failed: %v", err)
	}

	// Compile modules
	compiledModules, err := codegen.CompileModules()
	if err != nil {
		t.Fatalf("CompileModules failed: %v", err)
	}

	if len(compiledModules) != 1 {
		t.Errorf("Expected 1 compiled module, got %d", len(compiledModules))
	}

	if _, exists := compiledModules["simple_module"]; !exists {
		t.Errorf("simple_module not found in compiled modules")
	}
}
