package ast

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestModule(t *testing.T) {
	tests := []struct {
		name    string
		module  Module
		wantErr bool
	}{
		{
			name: "basic module",
			module: Module{
				Type:      "module",
				Name:      "test",
				Functions: []Function{},
			},
		},
		{
			name: "module with exports and imports",
			module: Module{
				Type:      "module",
				Name:      "test",
				Exports:   []string{"foo", "bar"},
				Imports:   []string{"std.io", "std.math"},
				Functions: []Function{},
			},
		},
		{
			name: "module with meta",
			module: Module{
				Type:      "module",
				Name:      "test",
				Functions: []Function{},
				Meta: map[string]interface{}{
					"version":     "1.0.0",
					"description": "Test module",
				},
			},
		},
		{
			name: "module with functions",
			module: Module{
				Type: "module",
				Name: "test",
				Functions: []Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []Parameter{},
						Returns: "int",
						Body:    []Statement{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.module)
			if err != nil {
				t.Fatalf("Failed to marshal module: %v", err)
			}

			// Test JSON unmarshaling
			var got Module
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Failed to unmarshal module: %v", err)
			}

			// Compare - we can do a simple comparison since these don't have pointers
			if !reflect.DeepEqual(got, tt.module) {
				t.Errorf("Module mismatch after marshal/unmarshal\ngot:  %+v\nwant: %+v", got, tt.module)
			}
		})
	}
}

func TestFunction(t *testing.T) {
	tests := []struct {
		name     string
		function Function
	}{
		{
			name: "simple function",
			function: Function{
				Type:    "function",
				Name:    "add",
				Params:  []Parameter{},
				Returns: "int",
				Body:    []Statement{},
			},
		},
		{
			name: "function with parameters",
			function: Function{
				Type: "function",
				Name: "add",
				Params: []Parameter{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "int"},
				},
				Returns: "int",
				Body:    []Statement{},
			},
		},
		{
			name: "function with body",
			function: Function{
				Type:    "function",
				Name:    "add",
				Params:  []Parameter{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
				Returns: "int",
				Body: []Statement{
					{
						Type: StmtReturn,
						Value: &Expression{
							Type: ExprBinary,
							Op:   OpAdd,
							Left: &Expression{
								Type: ExprVariable,
								Name: "a",
							},
							Right: &Expression{
								Type: ExprVariable,
								Name: "b",
							},
						},
					},
				},
			},
		},
		{
			name: "function with meta",
			function: Function{
				Type:    "function",
				Name:    "test",
				Params:  []Parameter{},
				Returns: "void",
				Body:    []Statement{},
				Meta: map[string]interface{}{
					"description": "Test function",
					"deprecated":  false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.function)
			if err != nil {
				t.Fatalf("Failed to marshal function: %v", err)
			}

			var got Function
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Failed to unmarshal function: %v", err)
			}

			if !reflect.DeepEqual(got, tt.function) {
				t.Errorf("Function mismatch\ngot:  %+v\nwant: %+v", got, tt.function)
			}
		})
	}
}

func TestParameter(t *testing.T) {
	tests := []struct {
		name  string
		param Parameter
	}{
		{
			name:  "int parameter",
			param: Parameter{Name: "x", Type: "int"},
		},
		{
			name:  "string parameter",
			param: Parameter{Name: "msg", Type: "string"},
		},
		{
			name:  "array parameter",
			param: Parameter{Name: "nums", Type: "array"},
		},
		{
			name:  "map parameter",
			param: Parameter{Name: "config", Type: "map"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.param)
			if err != nil {
				t.Fatalf("Failed to marshal parameter: %v", err)
			}

			var got Parameter
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Failed to unmarshal parameter: %v", err)
			}

			if got != tt.param {
				t.Errorf("Parameter mismatch\ngot:  %+v\nwant: %+v", got, tt.param)
			}
		})
	}
}

func TestStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt Statement
	}{
		{
			name: "assign statement",
			stmt: Statement{
				Type:   StmtAssign,
				Target: "x",
				Value: &Expression{
					Type:  ExprLiteral,
					Value: 42,
				},
			},
		},
		{
			name: "if statement",
			stmt: Statement{
				Type: StmtIf,
				Cond: &Expression{
					Type: ExprBinary,
					Op:   OpGt,
					Left: &Expression{
						Type: ExprVariable,
						Name: "x",
					},
					Right: &Expression{
						Type:  ExprLiteral,
						Value: 0,
					},
				},
				Then: []Statement{
					{
						Type: StmtReturn,
						Value: &Expression{
							Type:  ExprLiteral,
							Value: true,
						},
					},
				},
			},
		},
		{
			name: "if-else statement",
			stmt: Statement{
				Type: StmtIf,
				Cond: &Expression{
					Type: ExprVariable,
					Name: "flag",
				},
				Then: []Statement{
					{
						Type: StmtReturn,
						Value: &Expression{
							Type:  ExprLiteral,
							Value: "yes",
						},
					},
				},
				Else: []Statement{
					{
						Type: StmtReturn,
						Value: &Expression{
							Type:  ExprLiteral,
							Value: "no",
						},
					},
				},
			},
		},
		{
			name: "while statement",
			stmt: Statement{
				Type: StmtWhile,
				Cond: &Expression{
					Type: ExprBinary,
					Op:   OpLt,
					Left: &Expression{
						Type: ExprVariable,
						Name: "i",
					},
					Right: &Expression{
						Type:  ExprLiteral,
						Value: 10,
					},
				},
				Body: []Statement{
					{
						Type:   StmtAssign,
						Target: "i",
						Value: &Expression{
							Type: ExprBinary,
							Op:   OpAdd,
							Left: &Expression{
								Type: ExprVariable,
								Name: "i",
							},
							Right: &Expression{
								Type:  ExprLiteral,
								Value: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "return statement",
			stmt: Statement{
				Type: StmtReturn,
				Value: &Expression{
					Type:  ExprLiteral,
					Value: "result",
				},
			},
		},
		{
			name: "expression statement",
			stmt: Statement{
				Type: StmtExpr,
				Value: &Expression{
					Type: ExprCall,
					Name: "print",
					Args: []Expression{
						{
							Type:  ExprLiteral,
							Value: "Hello",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.stmt)
			if err != nil {
				t.Fatalf("Failed to marshal statement: %v", err)
			}

			var got Statement
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Failed to unmarshal statement: %v", err)
			}

			// Marshal both to JSON and compare
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.stmt)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("Statement mismatch after marshal/unmarshal\ngot:  %s\nwant: %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestExpression(t *testing.T) {
	tests := []struct {
		name string
		expr Expression
	}{
		{
			name: "int literal",
			expr: Expression{
				Type:  ExprLiteral,
				Value: 42,
			},
		},
		{
			name: "float literal",
			expr: Expression{
				Type:  ExprLiteral,
				Value: 3.14,
			},
		},
		{
			name: "string literal",
			expr: Expression{
				Type:  ExprLiteral,
				Value: "hello world",
			},
		},
		{
			name: "bool literal",
			expr: Expression{
				Type:  ExprLiteral,
				Value: true,
			},
		},
		{
			name: "variable",
			expr: Expression{
				Type: ExprVariable,
				Name: "x",
			},
		},
		{
			name: "binary operation",
			expr: Expression{
				Type: ExprBinary,
				Op:   OpAdd,
				Left: &Expression{
					Type:  ExprLiteral,
					Value: 1,
				},
				Right: &Expression{
					Type:  ExprLiteral,
					Value: 2,
				},
			},
		},
		{
			name: "unary operation",
			expr: Expression{
				Type: ExprUnary,
				Op:   OpNot,
				Operand: &Expression{
					Type:  ExprLiteral,
					Value: false,
				},
			},
		},
		{
			name: "function call",
			expr: Expression{
				Type: ExprCall,
				Name: "add",
				Args: []Expression{
					{Type: ExprLiteral, Value: 1},
					{Type: ExprLiteral, Value: 2},
				},
			},
		},
		{
			name: "array literal",
			expr: Expression{
				Type: ExprArrayLit,
				Elements: []Expression{
					{Type: ExprLiteral, Value: 1},
					{Type: ExprLiteral, Value: 2},
					{Type: ExprLiteral, Value: 3},
				},
			},
		},
		{
			name: "map literal",
			expr: Expression{
				Type: ExprMapLit,
				Pairs: []MapPair{
					{
						Key:   Expression{Type: ExprLiteral, Value: "name"},
						Value: Expression{Type: ExprLiteral, Value: "Alice"},
					},
					{
						Key:   Expression{Type: ExprLiteral, Value: "age"},
						Value: Expression{Type: ExprLiteral, Value: 30},
					},
				},
			},
		},
		{
			name: "index operation",
			expr: Expression{
				Type: ExprIndex,
				Object: &Expression{
					Type: ExprVariable,
					Name: "arr",
				},
				Index: &Expression{
					Type:  ExprLiteral,
					Value: 0,
				},
			},
		},
		{
			name: "field access",
			expr: Expression{
				Type: ExprField,
				Object: &Expression{
					Type: ExprVariable,
					Name: "obj",
				},
				Name: "field",
			},
		},
		{
			name: "module call",
			expr: Expression{
				Type:   ExprModuleCall,
				Module: "std.io",
				Name:   "println",
				Args: []Expression{
					{Type: ExprLiteral, Value: "Hello"},
				},
			},
		},
		{
			name: "builtin call",
			expr: Expression{
				Type: ExprBuiltin,
				Name: "len",
				Args: []Expression{
					{Type: ExprVariable, Name: "arr"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.expr)
			if err != nil {
				t.Fatalf("Failed to marshal expression: %v", err)
			}

			var got Expression
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Failed to unmarshal expression: %v", err)
			}

			// Marshal both to JSON and compare
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.expr)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("Expression mismatch after marshal/unmarshal\ngot:  %s\nwant: %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestMapPair(t *testing.T) {
	tests := []struct {
		name string
		pair MapPair
	}{
		{
			name: "string key-value",
			pair: MapPair{
				Key:   Expression{Type: ExprLiteral, Value: "name"},
				Value: Expression{Type: ExprLiteral, Value: "Alice"},
			},
		},
		{
			name: "int key-value",
			pair: MapPair{
				Key:   Expression{Type: ExprLiteral, Value: 1},
				Value: Expression{Type: ExprLiteral, Value: "one"},
			},
		},
		{
			name: "expression key-value",
			pair: MapPair{
				Key: Expression{
					Type:  ExprBinary,
					Op:    OpAdd,
					Left:  &Expression{Type: ExprLiteral, Value: 1},
					Right: &Expression{Type: ExprLiteral, Value: 2},
				},
				Value: Expression{Type: ExprLiteral, Value: "three"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.pair)
			if err != nil {
				t.Fatalf("Failed to marshal MapPair: %v", err)
			}

			var got MapPair
			err = json.Unmarshal(data, &got)
			if err != nil {
				t.Fatalf("Failed to unmarshal MapPair: %v", err)
			}

			// Marshal both to JSON and compare
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(tt.pair)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("MapPair mismatch after marshal/unmarshal\ngot:  %s\nwant: %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Test statement type constants
	stmtTypes := []string{
		StmtAssign, StmtIf, StmtWhile, StmtFor, StmtReturn, StmtExpr,
	}
	expectedStmtTypes := []string{
		"assign", "if", "while", "for", "return", "expr",
	}
	for i, got := range stmtTypes {
		if got != expectedStmtTypes[i] {
			t.Errorf("Statement type mismatch: got %s, want %s", got, expectedStmtTypes[i])
		}
	}

	// Test expression type constants
	exprTypes := []string{
		ExprLiteral, ExprVariable, ExprBinary, ExprUnary, ExprCall,
		ExprIndex, ExprField, ExprArrayLit, ExprMapLit, ExprModuleCall, ExprBuiltin,
	}
	expectedExprTypes := []string{
		"literal", "variable", "binary", "unary", "call",
		"index", "field", "array_literal", "map_literal", "module_call", "builtin",
	}
	for i, got := range exprTypes {
		if got != expectedExprTypes[i] {
			t.Errorf("Expression type mismatch: got %s, want %s", got, expectedExprTypes[i])
		}
	}

	// Test binary operator constants
	binOps := []string{
		OpAdd, OpSub, OpMul, OpDiv, OpMod,
		OpEq, OpNe, OpLt, OpLe, OpGt, OpGe,
		OpAnd, OpOr,
	}
	expectedBinOps := []string{
		"+", "-", "*", "/", "%",
		"==", "!=", "<", "<=", ">", ">=",
		"&&", "||",
	}
	for i, got := range binOps {
		if got != expectedBinOps[i] {
			t.Errorf("Binary operator mismatch: got %s, want %s", got, expectedBinOps[i])
		}
	}

	// Test unary operator constants
	unaryOps := []string{OpNot, OpNeg}
	expectedUnaryOps := []string{"!", "-"}
	for i, got := range unaryOps {
		if got != expectedUnaryOps[i] {
			t.Errorf("Unary operator mismatch: got %s, want %s", got, expectedUnaryOps[i])
		}
	}

	// Test type constants
	types := []string{
		TypeInt, TypeFloat, TypeString, TypeBool, TypeArray, TypeMap, TypeVoid,
	}
	expectedTypes := []string{
		"int", "float", "string", "bool", "array", "map", "void",
	}
	for i, got := range types {
		if got != expectedTypes[i] {
			t.Errorf("Type constant mismatch: got %s, want %s", got, expectedTypes[i])
		}
	}
}

func TestComplexStructures(t *testing.T) {
	// Test a complex nested structure
	module := Module{
		Type: "module",
		Name: "complex",
		Functions: []Function{
			{
				Type: "function",
				Name: "factorial",
				Params: []Parameter{
					{Name: "n", Type: "int"},
				},
				Returns: "int",
				Body: []Statement{
					{
						Type: StmtIf,
						Cond: &Expression{
							Type: ExprBinary,
							Op:   OpLe,
							Left: &Expression{
								Type: ExprVariable,
								Name: "n",
							},
							Right: &Expression{
								Type:  ExprLiteral,
								Value: 1,
							},
						},
						Then: []Statement{
							{
								Type: StmtReturn,
								Value: &Expression{
									Type:  ExprLiteral,
									Value: 1,
								},
							},
						},
						Else: []Statement{
							{
								Type: StmtReturn,
								Value: &Expression{
									Type: ExprBinary,
									Op:   OpMul,
									Left: &Expression{
										Type: ExprVariable,
										Name: "n",
									},
									Right: &Expression{
										Type: ExprCall,
										Name: "factorial",
										Args: []Expression{
											{
												Type: ExprBinary,
												Op:   OpSub,
												Left: &Expression{
													Type: ExprVariable,
													Name: "n",
												},
												Right: &Expression{
													Type:  ExprLiteral,
													Value: 1,
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
	}

	// Marshal and unmarshal
	data, err := json.Marshal(module)
	if err != nil {
		t.Fatalf("Failed to marshal complex module: %v", err)
	}

	var got Module
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("Failed to unmarshal complex module: %v", err)
	}

	// Marshal both to JSON and compare
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(module)
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("Complex module mismatch after marshal/unmarshal\ngot:  %s\nwant: %s", gotJSON, wantJSON)
	}
}
