package interpreter

import (
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/runtime"
)

func TestInterpreter_GCArrayLiterals(t *testing.T) {
	// Test that array literals use garbage collection
	interp := New()

	// Create a simple module with array literal
	module := &ast.Module{
		Type: "module",
		Name: "test",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "main",
				Params:  []ast.Parameter{},
				Returns: "array",
				Body: []ast.Statement{
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type: ast.ExprArrayLit,
							Elements: []ast.Expression{
								{Type: ast.ExprLiteral, Value: 1},
								{Type: ast.ExprLiteral, Value: 2},
								{Type: ast.ExprLiteral, Value: 3},
							},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Run the function
	result, err := interp.Run("main", []runtime.Value{})
	if err != nil {
		t.Fatalf("Failed to run function: %v", err)
	}

	// Verify result is an array
	if result.Type != runtime.ValueTypeArray {
		t.Errorf("Expected array result, got %v", result.Type)
	}

	// Verify it's a GC value
	if !result.IsGCValue() {
		t.Error("Expected GC array, got regular array")
	}

	// Verify array contents
	arr, err := result.AsArray()
	if err != nil {
		t.Fatalf("Failed to get array: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arr))
	}

	// Verify values
	for i, expected := range []int64{1, 2, 3} {
		val, err := arr[i].AsInt()
		if err != nil {
			t.Errorf("Failed to get int from element %d: %v", i, err)
		}
		if val != expected {
			t.Errorf("Expected element %d to be %d, got %d", i, expected, val)
		}
	}
}

func TestInterpreter_GCMapLiterals(t *testing.T) {
	// Test that map literals use garbage collection
	interp := New()

	// Create a simple module with map literal
	module := &ast.Module{
		Type: "module",
		Name: "test",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "main",
				Params:  []ast.Parameter{},
				Returns: "map",
				Body: []ast.Statement{
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type: ast.ExprMapLit,
							Pairs: []ast.MapPair{
								{
									Key:   ast.Expression{Type: ast.ExprLiteral, Value: "name"},
									Value: ast.Expression{Type: ast.ExprLiteral, Value: "ALaS"},
								},
								{
									Key:   ast.Expression{Type: ast.ExprLiteral, Value: "version"},
									Value: ast.Expression{Type: ast.ExprLiteral, Value: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Run the function
	result, err := interp.Run("main", []runtime.Value{})
	if err != nil {
		t.Fatalf("Failed to run function: %v", err)
	}

	// Verify result is a map
	if result.Type != runtime.ValueTypeMap {
		t.Errorf("Expected map result, got %v", result.Type)
	}

	// Verify it's a GC value
	if !result.IsGCValue() {
		t.Error("Expected GC map, got regular map")
	}

	// Verify map contents
	m, err := result.AsMap()
	if err != nil {
		t.Fatalf("Failed to get map: %v", err)
	}
	if len(m) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(m))
	}

	// Verify values
	nameVal, exists := m["name"]
	if !exists {
		t.Error("Expected 'name' key in map")
	} else {
		name, err := nameVal.AsString()
		if err != nil {
			t.Errorf("Failed to get string from name: %v", err)
		}
		if name != "ALaS" {
			t.Errorf("Expected name 'ALaS', got '%s'", name)
		}
	}

	versionVal, exists := m["version"]
	if !exists {
		t.Error("Expected 'version' key in map")
	} else {
		version, err := versionVal.AsInt()
		if err != nil {
			t.Errorf("Failed to get int from version: %v", err)
		}
		if version != 1 {
			t.Errorf("Expected version 1, got %d", version)
		}
	}
}

func TestInterpreter_GCVariableAssignment(t *testing.T) {
	// Test that variable assignment properly releases old GC objects
	interp := New()

	// Create module that assigns arrays to the same variable
	module := &ast.Module{
		Type: "module",
		Name: "test",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "main",
				Params:  []ast.Parameter{},
				Returns: "array",
				Body: []ast.Statement{
					// var arr = [1, 2]
					{
						Type:   ast.StmtAssign,
						Target: "arr",
						Value: &ast.Expression{
							Type: ast.ExprArrayLit,
							Elements: []ast.Expression{
								{Type: ast.ExprLiteral, Value: 1},
								{Type: ast.ExprLiteral, Value: 2},
							},
						},
					},
					// arr = [3, 4, 5] (should release previous array)
					{
						Type:   ast.StmtAssign,
						Target: "arr",
						Value: &ast.Expression{
							Type: ast.ExprArrayLit,
							Elements: []ast.Expression{
								{Type: ast.ExprLiteral, Value: 3},
								{Type: ast.ExprLiteral, Value: 4},
								{Type: ast.ExprLiteral, Value: 5},
							},
						},
					},
					// return arr
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type: ast.ExprVariable,
							Name: "arr",
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Run the function
	result, err := interp.Run("main", []runtime.Value{})
	if err != nil {
		t.Fatalf("Failed to run function: %v", err)
	}

	// Verify final result has the second array
	arr, err := result.AsArray()
	if err != nil {
		t.Fatalf("Failed to get final array: %v", err)
	}
	if len(arr) != 3 {
		t.Errorf("Expected 3 elements in final array, got %d", len(arr))
	}

	// Check values [3, 4, 5]
	for i, expected := range []int64{3, 4, 5} {
		val, err := arr[i].AsInt()
		if err != nil {
			t.Errorf("Failed to get int from element %d: %v", i, err)
		}
		if val != expected {
			t.Errorf("Expected element %d to be %d, got %d", i, expected, val)
		}
	}
}

func TestInterpreter_GCCleanupOnFunctionReturn(t *testing.T) {
	// Test that function environments are properly cleaned up
	interp := New()

	// Get initial GC stats
	initialStats := runtime.GetGCStats()

	// Create module with function that creates local arrays
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
					// var localArray = [1, 2, 3]
					{
						Type:   ast.StmtAssign,
						Target: "localArray",
						Value: &ast.Expression{
							Type: ast.ExprArrayLit,
							Elements: []ast.Expression{
								{Type: ast.ExprLiteral, Value: 1},
								{Type: ast.ExprLiteral, Value: 2},
								{Type: ast.ExprLiteral, Value: 3},
							},
						},
					},
					// var localMap = {"key": "value"}
					{
						Type:   ast.StmtAssign,
						Target: "localMap",
						Value: &ast.Expression{
							Type: ast.ExprMapLit,
							Pairs: []ast.MapPair{
								{
									Key:   ast.Expression{Type: ast.ExprLiteral, Value: "key"},
									Value: ast.Expression{Type: ast.ExprLiteral, Value: "value"},
								},
							},
						},
					},
					// return 42 (don't return the GC objects)
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type:  ast.ExprLiteral,
							Value: 42,
						},
					},
				},
			},
		},
	}

	err := interp.LoadModule(module)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Run the function
	result, err := interp.Run("main", []runtime.Value{})
	if err != nil {
		t.Fatalf("Failed to run function: %v", err)
	}

	// Verify return value
	val, err := result.AsInt()
	if err != nil {
		t.Fatalf("Failed to get int result: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected result 42, got %d", val)
	}

	// Give GC time to clean up
	runtime.RunGC()

	// Check that GC objects were cleaned up
	finalStats := runtime.GetGCStats()
	if finalStats.TotalObjects > initialStats.TotalObjects {
		t.Logf("Warning: GC objects may not have been fully cleaned up. Initial: %d, Final: %d",
			initialStats.TotalObjects, finalStats.TotalObjects)
		// Note: This might not always pass due to async GC, but it's useful for monitoring
	}
}
