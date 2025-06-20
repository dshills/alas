package validator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dshills/alas/internal/ast"
)

func TestValidateModule(t *testing.T) {
	tests := []struct {
		name    string
		module  ast.Module
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid module",
			module: ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid module type",
			module: ast.Module{
				Type:      "invalid",
				Name:      "test",
				Functions: []ast.Function{{Type: "function", Name: "main", Body: []ast.Statement{}}},
			},
			wantErr: true,
			errMsg:  "module type must be 'module'",
		},
		{
			name: "empty module name",
			module: ast.Module{
				Type:      "module",
				Name:      "",
				Functions: []ast.Function{{Type: "function", Name: "main", Body: []ast.Statement{}}},
			},
			wantErr: true,
			errMsg:  "module name cannot be empty",
		},
		{
			name: "no functions",
			module: ast.Module{
				Type:      "module",
				Name:      "test",
				Functions: []ast.Function{},
			},
			wantErr: true,
			errMsg:  "module must contain at least one function",
		},
		{
			name: "duplicate function names",
			module: ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{
					{Type: "function", Name: "foo", Body: []ast.Statement{}},
					{Type: "function", Name: "foo", Body: []ast.Statement{}},
				},
			},
			wantErr: true,
			errMsg:  "duplicate function name: foo",
		},
		{
			name: "export non-existent function",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Exports: []string{"bar"},
				Functions: []ast.Function{
					{Type: "function", Name: "foo", Body: []ast.Statement{}},
				},
			},
			wantErr: true,
			errMsg:  "exported function 'bar' not found",
		},
		{
			name: "valid exports",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Exports: []string{"foo", "bar"},
				Functions: []ast.Function{
					{Type: "function", Name: "foo", Body: []ast.Statement{}},
					{Type: "function", Name: "bar", Body: []ast.Statement{}},
				},
			},
			wantErr: false,
		},
		{
			name: "empty import name",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Imports: []string{""},
				Functions: []ast.Function{
					{Type: "function", Name: "main", Body: []ast.Statement{}},
				},
			},
			wantErr: true,
			errMsg:  "import 0: name cannot be empty",
		},
		{
			name: "self import",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Imports: []string{"test"},
				Functions: []ast.Function{
					{Type: "function", Name: "main", Body: []ast.Statement{}},
				},
			},
			wantErr: true,
			errMsg:  "module cannot import itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.ValidateModule(&tt.module)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateModule() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateFunction(t *testing.T) {
	tests := []struct {
		name    string
		fn      ast.Function
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid function",
			fn: ast.Function{
				Type:    "function",
				Name:    "test",
				Params:  []ast.Parameter{{Name: "x", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type: ast.ExprVariable,
							Name: "x",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid function type",
			fn: ast.Function{
				Type: "invalid",
				Name: "test",
				Body: []ast.Statement{},
			},
			wantErr: true,
			errMsg:  "type must be 'function'",
		},
		{
			name: "empty function name",
			fn: ast.Function{
				Type: "function",
				Name: "",
				Body: []ast.Statement{},
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "duplicate parameter names",
			fn: ast.Function{
				Type: "function",
				Name: "test",
				Params: []ast.Parameter{
					{Name: "x", Type: "int"},
					{Name: "x", Type: "int"},
				},
				Body: []ast.Statement{},
			},
			wantErr: true,
			errMsg:  "duplicate parameter name: x",
		},
		{
			name: "empty parameter name",
			fn: ast.Function{
				Type:   "function",
				Name:   "test",
				Params: []ast.Parameter{{Name: "", Type: "int"}},
				Body:   []ast.Statement{},
			},
			wantErr: true,
			errMsg:  "parameter 0: name cannot be empty",
		},
		{
			name: "invalid parameter type",
			fn: ast.Function{
				Type:   "function",
				Name:   "test",
				Params: []ast.Parameter{{Name: "x", Type: ""}},
				Body:   []ast.Statement{},
			},
			wantErr: true,
			errMsg:  "invalid type ''",
		},
		{
			name: "empty return type (valid for void functions)",
			fn: ast.Function{
				Type:    "function",
				Name:    "test",
				Returns: "",
				Body:    []ast.Statement{},
			},
			wantErr: false,
		},
		{
			name: "null body",
			fn: ast.Function{
				Type: "function",
				Name: "test",
				Body: nil,
			},
			wantErr: true,
			errMsg:  "function body cannot be null",
		},
		{
			name: "custom type parameter",
			fn: ast.Function{
				Type:   "function",
				Name:   "test",
				Params: []ast.Parameter{{Name: "obj", Type: "MyCustomType"}},
				Body:   []ast.Statement{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.validateFunction(&tt.fn, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateFunction() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateStatement(t *testing.T) {
	tests := []struct {
		name    string
		stmt    ast.Statement
		scope   map[string]bool
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid assign statement",
			stmt: ast.Statement{
				Type:   ast.StmtAssign,
				Target: "x",
				Value:  &ast.Expression{Type: ast.ExprLiteral, Value: 42},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "assign without target",
			stmt: ast.Statement{
				Type:  ast.StmtAssign,
				Value: &ast.Expression{Type: ast.ExprLiteral, Value: 42},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "assign statement must have a target",
		},
		{
			name: "assign without value",
			stmt: ast.Statement{
				Type:   ast.StmtAssign,
				Target: "x",
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "assign statement must have a value",
		},
		{
			name: "valid if statement",
			stmt: ast.Statement{
				Type: ast.StmtIf,
				Cond: &ast.Expression{Type: ast.ExprLiteral, Value: true},
				Then: []ast.Statement{
					{Type: ast.StmtReturn},
				},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "if without condition",
			stmt: ast.Statement{
				Type: ast.StmtIf,
				Then: []ast.Statement{{Type: ast.StmtReturn}},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "if statement must have a condition",
		},
		{
			name: "if without then block",
			stmt: ast.Statement{
				Type: ast.StmtIf,
				Cond: &ast.Expression{Type: ast.ExprLiteral, Value: true},
				Then: []ast.Statement{},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "if statement must have a then block",
		},
		{
			name: "valid while statement",
			stmt: ast.Statement{
				Type: ast.StmtWhile,
				Cond: &ast.Expression{Type: ast.ExprLiteral, Value: true},
				Body: []ast.Statement{
					{Type: ast.StmtExpr, Value: &ast.Expression{Type: ast.ExprLiteral, Value: 1}},
				},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "while without condition",
			stmt: ast.Statement{
				Type: ast.StmtWhile,
				Body: []ast.Statement{{Type: ast.StmtReturn}},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "while statement must have a condition",
		},
		{
			name: "while without body",
			stmt: ast.Statement{
				Type: ast.StmtWhile,
				Cond: &ast.Expression{Type: ast.ExprLiteral, Value: true},
				Body: []ast.Statement{},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "while statement must have a body",
		},
		{
			name: "valid return statement",
			stmt: ast.Statement{
				Type:  ast.StmtReturn,
				Value: &ast.Expression{Type: ast.ExprLiteral, Value: 42},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "return without value",
			stmt: ast.Statement{
				Type: ast.StmtReturn,
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "valid expr statement",
			stmt: ast.Statement{
				Type:  ast.StmtExpr,
				Value: &ast.Expression{Type: ast.ExprLiteral, Value: 42},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "expr without value",
			stmt: ast.Statement{
				Type: ast.StmtExpr,
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "expression statement must have a value",
		},
		{
			name: "unknown statement type",
			stmt: ast.Statement{
				Type: "unknown",
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "unknown statement type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.validateStatement(&tt.stmt, tt.scope, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStatement() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateStatement() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		scope   map[string]bool
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid literal",
			expr:    ast.Expression{Type: ast.ExprLiteral, Value: 42},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name:    "literal without value",
			expr:    ast.Expression{Type: ast.ExprLiteral},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "literal expression must have a value",
		},
		{
			name:    "valid variable in scope",
			expr:    ast.Expression{Type: ast.ExprVariable, Name: "x"},
			scope:   map[string]bool{"x": true},
			wantErr: false,
		},
		{
			name:    "variable without name",
			expr:    ast.Expression{Type: ast.ExprVariable},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "variable expression must have a name",
		},
		{
			name:    "undefined variable",
			expr:    ast.Expression{Type: ast.ExprVariable, Name: "x"},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "undefined variable: x",
		},
		{
			name: "valid binary expression",
			expr: ast.Expression{
				Type:  ast.ExprBinary,
				Op:    ast.OpAdd,
				Left:  &ast.Expression{Type: ast.ExprLiteral, Value: 1},
				Right: &ast.Expression{Type: ast.ExprLiteral, Value: 2},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "binary without operator",
			expr: ast.Expression{
				Type:  ast.ExprBinary,
				Left:  &ast.Expression{Type: ast.ExprLiteral, Value: 1},
				Right: &ast.Expression{Type: ast.ExprLiteral, Value: 2},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "binary expression must have an operator",
		},
		{
			name: "binary with invalid operator",
			expr: ast.Expression{
				Type:  ast.ExprBinary,
				Op:    "invalid",
				Left:  &ast.Expression{Type: ast.ExprLiteral, Value: 1},
				Right: &ast.Expression{Type: ast.ExprLiteral, Value: 2},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "invalid binary operator: invalid",
		},
		{
			name: "binary without operands",
			expr: ast.Expression{
				Type: ast.ExprBinary,
				Op:   ast.OpAdd,
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "binary expression must have left and right operands",
		},
		{
			name: "valid unary expression with Operand",
			expr: ast.Expression{
				Type:    ast.ExprUnary,
				Op:      ast.OpNot,
				Operand: &ast.Expression{Type: ast.ExprLiteral, Value: true},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "valid unary expression with Right (backward compat)",
			expr: ast.Expression{
				Type:  ast.ExprUnary,
				Op:    ast.OpNeg,
				Right: &ast.Expression{Type: ast.ExprLiteral, Value: 42},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "unary without operator",
			expr: ast.Expression{
				Type:    ast.ExprUnary,
				Operand: &ast.Expression{Type: ast.ExprLiteral, Value: true},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "unary expression must have an operator",
		},
		{
			name: "unary with invalid operator",
			expr: ast.Expression{
				Type:    ast.ExprUnary,
				Op:      "invalid",
				Operand: &ast.Expression{Type: ast.ExprLiteral, Value: true},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "invalid unary operator: invalid",
		},
		{
			name: "unary without operand",
			expr: ast.Expression{
				Type: ast.ExprUnary,
				Op:   ast.OpNot,
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "unary expression must have an operand",
		},
		{
			name: "valid call expression",
			expr: ast.Expression{
				Type: ast.ExprCall,
				Name: "foo",
				Args: []ast.Expression{
					{Type: ast.ExprLiteral, Value: 42},
				},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name:    "call without name",
			expr:    ast.Expression{Type: ast.ExprCall},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "call expression must have a function name",
		},
		{
			name: "valid array literal",
			expr: ast.Expression{
				Type: ast.ExprArrayLit,
				Elements: []ast.Expression{
					{Type: ast.ExprLiteral, Value: 1},
					{Type: ast.ExprLiteral, Value: 2},
				},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "valid map literal",
			expr: ast.Expression{
				Type: ast.ExprMapLit,
				Pairs: []ast.MapPair{
					{
						Key:   ast.Expression{Type: ast.ExprLiteral, Value: "key"},
						Value: ast.Expression{Type: ast.ExprLiteral, Value: "value"},
					},
				},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "valid index expression",
			expr: ast.Expression{
				Type:   ast.ExprIndex,
				Object: &ast.Expression{Type: ast.ExprVariable, Name: "arr"},
				Index:  &ast.Expression{Type: ast.ExprLiteral, Value: 0},
			},
			scope:   map[string]bool{"arr": true},
			wantErr: false,
		},
		{
			name: "index without object",
			expr: ast.Expression{
				Type:  ast.ExprIndex,
				Index: &ast.Expression{Type: ast.ExprLiteral, Value: 0},
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "index expression must have an object",
		},
		{
			name: "index without index",
			expr: ast.Expression{
				Type:   ast.ExprIndex,
				Object: &ast.Expression{Type: ast.ExprVariable, Name: "arr"},
			},
			scope:   map[string]bool{"arr": true},
			wantErr: true,
			errMsg:  "index expression must have an index",
		},
		{
			name: "valid module call",
			expr: ast.Expression{
				Type:   ast.ExprModuleCall,
				Module: "std.io",
				Name:   "println",
				Args:   []ast.Expression{{Type: ast.ExprLiteral, Value: "hello"}},
			},
			scope:   map[string]bool{},
			wantErr: false,
		},
		{
			name: "module call without module",
			expr: ast.Expression{
				Type: ast.ExprModuleCall,
				Name: "println",
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "module call expression must have a module name",
		},
		{
			name: "module call without function name",
			expr: ast.Expression{
				Type:   ast.ExprModuleCall,
				Module: "std.io",
			},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "module call expression must have a function name",
		},
		{
			name: "valid builtin call",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "array.length",
				Args: []ast.Expression{{Type: ast.ExprVariable, Name: "arr"}},
			},
			scope:   map[string]bool{"arr": true},
			wantErr: false,
		},
		{
			name:    "builtin without name",
			expr:    ast.Expression{Type: ast.ExprBuiltin},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "builtin call expression must have a function name",
		},
		{
			name:    "unknown expression type",
			expr:    ast.Expression{Type: "unknown"},
			scope:   map[string]bool{},
			wantErr: true,
			errMsg:  "unknown expression type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.validateExpression(&tt.expr, tt.scope, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateExpression() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestArrayMapValidation(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid array literal",
			expr: ast.Expression{
				Type: ast.ExprArrayLit,
				Elements: []ast.Expression{
					{Type: ast.ExprLiteral, Value: 1},
					{Type: ast.ExprLiteral, Value: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "empty array literal",
			expr: ast.Expression{
				Type:     ast.ExprArrayLit,
				Elements: []ast.Expression{},
			},
			wantErr: false,
		},
		{
			name: "array element missing type",
			expr: ast.Expression{
				Type: ast.ExprArrayLit,
				Elements: []ast.Expression{
					{Value: 1}, // Missing Type field
				},
			},
			wantErr: true,
			errMsg:  "unknown expression type:",
		},
		{
			name: "valid map literal",
			expr: ast.Expression{
				Type: ast.ExprMapLit,
				Pairs: []ast.MapPair{
					{
						Key:   ast.Expression{Type: ast.ExprLiteral, Value: "key1"},
						Value: ast.Expression{Type: ast.ExprLiteral, Value: "value1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty map literal",
			expr: ast.Expression{
				Type:  ast.ExprMapLit,
				Pairs: []ast.MapPair{},
			},
			wantErr: false,
		},
		{
			name: "map key missing type",
			expr: ast.Expression{
				Type: ast.ExprMapLit,
				Pairs: []ast.MapPair{
					{
						Key:   ast.Expression{Value: "key1"}, // Missing Type field
						Value: ast.Expression{Type: ast.ExprLiteral, Value: "value1"},
					},
				},
			},
			wantErr: true,
			errMsg:  "unknown expression type:",
		},
		{
			name: "map value missing type",
			expr: ast.Expression{
				Type: ast.ExprMapLit,
				Pairs: []ast.MapPair{
					{
						Key:   ast.Expression{Type: ast.ExprLiteral, Value: "key1"},
						Value: ast.Expression{Value: "value1"}, // Missing Type field
					},
				},
			},
			wantErr: true,
			errMsg:  "unknown expression type:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.validateExpression(&tt.expr, map[string]bool{}, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateExpression() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestCallValidation(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid function call",
			expr: ast.Expression{
				Type: ast.ExprCall,
				Name: "foo",
				Args: []ast.Expression{{Type: ast.ExprLiteral, Value: 42}},
			},
			wantErr: false,
		},
		{
			name: "invalid function name",
			expr: ast.Expression{
				Type: ast.ExprCall,
				Name: "123invalid",
				Args: []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "invalid function name '123invalid'",
		},
		{
			name: "valid module call",
			expr: ast.Expression{
				Type:   ast.ExprModuleCall,
				Module: "std_io",
				Name:   "println",
				Args:   []ast.Expression{{Type: ast.ExprLiteral, Value: "hello"}},
			},
			wantErr: false,
		},
		{
			name: "invalid module name",
			expr: ast.Expression{
				Type:   ast.ExprModuleCall,
				Module: "123invalid",
				Name:   "println",
				Args:   []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "invalid module name '123invalid'",
		},
		{
			name: "invalid module function name",
			expr: ast.Expression{
				Type:   ast.ExprModuleCall,
				Module: "std_io",
				Name:   "123invalid",
				Args:   []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "invalid function name '123invalid'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.validateExpression(&tt.expr, map[string]bool{}, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateExpression() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid JSON module",
			json: `{
				"type": "module",
				"name": "test",
				"functions": [{
					"type": "function",
					"name": "main",
					"params": [],
					"returns": "int",
					"body": [{
						"type": "return",
						"value": {"type": "literal", "value": 42}
					}]
				}]
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			json:    `{"type": "module", invalid json`,
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name: "valid JSON but invalid module",
			json: `{
				"type": "invalid",
				"name": "test",
				"functions": []
			}`,
			wantErr: true,
			errMsg:  "module type must be 'module'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateJSON() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isValidType", func(t *testing.T) {
		validTypes := []string{
			ast.TypeInt, ast.TypeFloat, ast.TypeString, ast.TypeBool,
			ast.TypeArray, ast.TypeMap, ast.TypeVoid, "CustomType",
		}
		for _, typ := range validTypes {
			if !isValidType(typ, nil) {
				t.Errorf("isValidType(%s) = false, want true", typ)
			}
		}

		if isValidType("", nil) {
			t.Error("isValidType('') = true, want false")
		}
	})

	t.Run("isValidBinaryOp", func(t *testing.T) {
		validOps := []string{
			ast.OpAdd, ast.OpSub, ast.OpMul, ast.OpDiv, ast.OpMod,
			ast.OpEq, ast.OpNe, ast.OpLt, ast.OpLe, ast.OpGt, ast.OpGe,
			ast.OpAnd, ast.OpOr,
		}
		for _, op := range validOps {
			if !isValidBinaryOp(op) {
				t.Errorf("isValidBinaryOp(%s) = false, want true", op)
			}
		}

		invalidOps := []string{"invalid", "++", "--", "^"}
		for _, op := range invalidOps {
			if isValidBinaryOp(op) {
				t.Errorf("isValidBinaryOp(%s) = true, want false", op)
			}
		}
	})

	t.Run("isValidUnaryOp", func(t *testing.T) {
		validOps := []string{ast.OpNot, ast.OpNeg}
		for _, op := range validOps {
			if !isValidUnaryOp(op) {
				t.Errorf("isValidUnaryOp(%s) = false, want true", op)
			}
		}

		invalidOps := []string{"invalid", "++", "--", "+"}
		for _, op := range invalidOps {
			if isValidUnaryOp(op) {
				t.Errorf("isValidUnaryOp(%s) = true, want false", op)
			}
		}
	})

	t.Run("copyScope", func(t *testing.T) {
		original := map[string]bool{"x": true, "y": true}
		copied := copyScope(original)

		// Check values are copied
		for k, v := range original {
			if copied[k] != v {
				t.Errorf("copied[%s] = %v, want %v", k, copied[k], v)
			}
		}

		// Check they're independent
		copied["z"] = true
		if original["z"] {
			t.Error("modifying copied scope affected original")
		}
	})
}

func TestComplexValidation(t *testing.T) {
	// Test a complex module with nested structures
	module := ast.Module{
		Type:    "module",
		Name:    "complex",
		Imports: []string{"std.io", "std.math"},
		Exports: []string{"fibonacci"},
		Functions: []ast.Function{
			{
				Type: "function",
				Name: "fibonacci",
				Params: []ast.Parameter{
					{Name: "n", Type: "int"},
				},
				Returns: "int",
				Body: []ast.Statement{
					// if n <= 1
					{
						Type: ast.StmtIf,
						Cond: &ast.Expression{
							Type: ast.ExprBinary,
							Op:   ast.OpLe,
							Left: &ast.Expression{
								Type: ast.ExprVariable,
								Name: "n",
							},
							Right: &ast.Expression{
								Type:  ast.ExprLiteral,
								Value: 1,
							},
						},
						Then: []ast.Statement{
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type: ast.ExprVariable,
									Name: "n",
								},
							},
						},
						Else: []ast.Statement{
							// a = fibonacci(n - 1)
							{
								Type:   ast.StmtAssign,
								Target: "a",
								Value: &ast.Expression{
									Type: ast.ExprCall,
									Name: "fibonacci",
									Args: []ast.Expression{
										{
											Type: ast.ExprBinary,
											Op:   ast.OpSub,
											Left: &ast.Expression{
												Type: ast.ExprVariable,
												Name: "n",
											},
											Right: &ast.Expression{
												Type:  ast.ExprLiteral,
												Value: 1,
											},
										},
									},
								},
							},
							// b = fibonacci(n - 2)
							{
								Type:   ast.StmtAssign,
								Target: "b",
								Value: &ast.Expression{
									Type: ast.ExprCall,
									Name: "fibonacci",
									Args: []ast.Expression{
										{
											Type: ast.ExprBinary,
											Op:   ast.OpSub,
											Left: &ast.Expression{
												Type: ast.ExprVariable,
												Name: "n",
											},
											Right: &ast.Expression{
												Type:  ast.ExprLiteral,
												Value: 2,
											},
										},
									},
								},
							},
							// return a + b
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type: ast.ExprBinary,
									Op:   ast.OpAdd,
									Left: &ast.Expression{
										Type: ast.ExprVariable,
										Name: "a",
									},
									Right: &ast.Expression{
										Type: ast.ExprVariable,
										Name: "b",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	v := New()
	err := v.ValidateModule(&module)
	if err != nil {
		t.Errorf("Complex module validation failed: %v", err)
	}

	// Test JSON round-trip
	data, err := json.Marshal(module)
	if err != nil {
		t.Fatalf("Failed to marshal complex module: %v", err)
	}

	err = ValidateJSON(data)
	if err != nil {
		t.Errorf("Complex module JSON validation failed: %v", err)
	}
}

func TestScopeManagement(t *testing.T) {
	// Test that variables are properly added to scope
	module := ast.Module{
		Type: "module",
		Name: "scope_test",
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "test",
				Params:  []ast.Parameter{{Name: "param", Type: "int"}},
				Returns: "int",
				Body: []ast.Statement{
					// x = param + 1
					{
						Type:   ast.StmtAssign,
						Target: "x",
						Value: &ast.Expression{
							Type: ast.ExprBinary,
							Op:   ast.OpAdd,
							Left: &ast.Expression{
								Type: ast.ExprVariable,
								Name: "param", // Should be in scope from params
							},
							Right: &ast.Expression{
								Type:  ast.ExprLiteral,
								Value: 1,
							},
						},
					},
					// y = x * 2
					{
						Type:   ast.StmtAssign,
						Target: "y",
						Value: &ast.Expression{
							Type: ast.ExprBinary,
							Op:   ast.OpMul,
							Left: &ast.Expression{
								Type: ast.ExprVariable,
								Name: "x", // Should be in scope from previous assign
							},
							Right: &ast.Expression{
								Type:  ast.ExprLiteral,
								Value: 2,
							},
						},
					},
					// return y
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type: ast.ExprVariable,
							Name: "y", // Should be in scope
						},
					},
				},
			},
		},
	}

	v := New()
	err := v.ValidateModule(&module)
	if err != nil {
		t.Errorf("Scope management test failed: %v", err)
	}
}

func TestValidateCustomTypes(t *testing.T) {
	tests := []struct {
		name    string
		module  *ast.Module
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid struct type",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "name", Type: "string"},
								{Name: "age", Type: "int"},
							},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid enum type",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Status",
						Definition: ast.TypeDefinitionDef{
							Kind:   ast.TypeKindEnum,
							Values: []string{"active", "inactive", "pending"},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "struct with duplicate field names",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "name", Type: "string"},
								{Name: "name", Type: "int"},
							},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate field name: name",
		},
		{
			name: "enum with duplicate values",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Status",
						Definition: ast.TypeDefinitionDef{
							Kind:   ast.TypeKindEnum,
							Values: []string{"active", "inactive", "active"},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate enum value: active",
		},
		{
			name: "empty struct",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Empty",
						Definition: ast.TypeDefinitionDef{
							Kind:   ast.TypeKindStruct,
							Fields: []ast.TypeField{},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: true,
			errMsg:  "struct type 'Empty' must have at least one field",
		},
		{
			name: "empty enum",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Empty",
						Definition: ast.TypeDefinitionDef{
							Kind:   ast.TypeKindEnum,
							Values: []string{},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: true,
			errMsg:  "enum type 'Empty' must have at least one value",
		},
		{
			name: "function using custom type",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "name", Type: "string"},
								{Name: "age", Type: "int"},
							},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type: "function",
						Name: "createPerson",
						Params: []ast.Parameter{
							{Name: "name", Type: "string"},
							{Name: "age", Type: "int"},
						},
						Returns: "Person",
						Body: []ast.Statement{
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "name"},
											Value: ast.Expression{Type: ast.ExprVariable, Name: "name"},
										},
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "age"},
											Value: ast.Expression{Type: ast.ExprVariable, Name: "age"},
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
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate type names",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Types: []ast.TypeDefinition{
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "name", Type: "string"},
							},
						},
					},
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind:   ast.TypeKindEnum,
							Values: []string{"active"},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "main",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body:    []ast.Statement{},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate type name: Person",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.ValidateModule(tt.module)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				// Check if error contains the expected message
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateModule() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

// TestEnhancedValidation tests the new comprehensive validation rules.
func TestEnhancedValidation(t *testing.T) {
	tests := []struct {
		name    string
		module  ast.Module
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid module name with numbers at start",
			module: ast.Module{
				Type:      "module",
				Name:      "123invalid",
				Functions: []ast.Function{{Type: "function", Name: "main", Body: []ast.Statement{}}},
			},
			wantErr: true,
			errMsg:  "invalid module name '123invalid'",
		},
		{
			name: "invalid function name",
			module: ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{{
					Type: "function",
					Name: "123invalid",
					Body: []ast.Statement{},
				}},
			},
			wantErr: true,
			errMsg:  "invalid function name '123invalid'",
		},
		{
			name: "invalid parameter name",
			module: ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{{
					Type:   "function",
					Name:   "test",
					Params: []ast.Parameter{{Name: "123invalid", Type: "int"}},
					Body:   []ast.Statement{},
				}},
			},
			wantErr: true,
			errMsg:  "invalid name '123invalid'",
		},
		{
			name: "invalid assignment target",
			module: ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{{
					Type: "function",
					Name: "test",
					Body: []ast.Statement{{
						Type:   ast.StmtAssign,
						Target: "123invalid",
						Value:  &ast.Expression{Type: ast.ExprLiteral, Value: 42},
					}},
				}},
			},
			wantErr: true,
			errMsg:  "invalid assignment target '123invalid'",
		},
		{
			name: "invalid variable name",
			module: ast.Module{
				Type: "module",
				Name: "test",
				Functions: []ast.Function{{
					Type: "function",
					Name: "test",
					Body: []ast.Statement{{
						Type: ast.StmtExpr,
						Value: &ast.Expression{
							Type: ast.ExprVariable,
							Name: "123invalid",
						},
					}},
				}},
			},
			wantErr: true,
			errMsg:  "invalid variable name '123invalid'",
		},
		{
			name: "empty export name",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Exports: []string{"", "main"},
				Functions: []ast.Function{{
					Type: "function",
					Name: "main",
					Body: []ast.Statement{},
				}},
			},
			wantErr: true,
			errMsg:  "export 0: name cannot be empty",
		},
		{
			name: "invalid export name",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Exports: []string{"123invalid"},
				Functions: []ast.Function{{
					Type: "function",
					Name: "main",
					Body: []ast.Statement{},
				}},
			},
			wantErr: true,
			errMsg:  "export 0: invalid name '123invalid'",
		},
		{
			name: "invalid import name",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Imports: []string{"123invalid"},
				Functions: []ast.Function{{
					Type: "function",
					Name: "main",
					Body: []ast.Statement{},
				}},
			},
			wantErr: true,
			errMsg:  "import 0: invalid name '123invalid'",
		},
		{
			name: "duplicate imports",
			module: ast.Module{
				Type:    "module",
				Name:    "test",
				Imports: []string{"std_io", "std_io"},
				Functions: []ast.Function{{
					Type: "function",
					Name: "main",
					Body: []ast.Statement{},
				}},
			},
			wantErr: true,
			errMsg:  "duplicate import 'std_io'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.ValidateModule(&tt.module)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateModule() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestBuiltinValidation(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid builtin call",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "io.print",
				Args: []ast.Expression{{Type: ast.ExprLiteral, Value: "hello"}},
			},
			wantErr: false,
		},
		{
			name: "invalid builtin format - no dot",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "print",
				Args: []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "builtin name must be in format 'namespace.function'",
		},
		{
			name: "invalid builtin format - multiple dots",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "io.print.extra",
				Args: []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "builtin name must be in format 'namespace.function'",
		},
		{
			name: "empty namespace",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: ".print",
				Args: []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "builtin namespace and function cannot be empty",
		},
		{
			name: "empty function name",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "io.",
				Args: []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "builtin namespace and function cannot be empty",
		},
		{
			name: "unknown namespace",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "unknown.print",
				Args: []ast.Expression{},
			},
			wantErr: true,
			errMsg:  "unknown builtin namespace 'unknown'",
		},
		{
			name: "valid math builtin",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "math.sqrt",
				Args: []ast.Expression{{Type: ast.ExprLiteral, Value: 16.0}},
			},
			wantErr: false,
		},
		{
			name: "valid string builtin",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "string.length",
				Args: []ast.Expression{{Type: ast.ExprLiteral, Value: "hello"}},
			},
			wantErr: false,
		},
		{
			name: "valid array builtin",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "array.length",
				Args: []ast.Expression{{Type: ast.ExprVariable, Name: "arr"}},
			},
			wantErr: false,
		},
		{
			name: "valid map builtin",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "map.keys",
				Args: []ast.Expression{{Type: ast.ExprVariable, Name: "map"}},
			},
			wantErr: false,
		},
		{
			name: "valid collections builtin",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "collections.length",
				Args: []ast.Expression{{Type: ast.ExprVariable, Name: "arr"}},
			},
			wantErr: false,
		},
		{
			name: "valid type builtin",
			expr: ast.Expression{
				Type: ast.ExprBuiltin,
				Name: "type.typeOf",
				Args: []ast.Expression{{Type: ast.ExprVariable, Name: "value"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			scope := map[string]bool{"arr": true, "map": true, "value": true}
			err := v.validateExpression(&tt.expr, scope, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateExpression() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestLiteralValidation(t *testing.T) {
	tests := []struct {
		name    string
		expr    ast.Expression
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid string literal",
			expr: ast.Expression{
				Type:  ast.ExprLiteral,
				Value: "hello world",
			},
			wantErr: false,
		},
		{
			name: "valid int literal",
			expr: ast.Expression{
				Type:  ast.ExprLiteral,
				Value: 42,
			},
			wantErr: false,
		},
		{
			name: "valid float literal",
			expr: ast.Expression{
				Type:  ast.ExprLiteral,
				Value: 3.14,
			},
			wantErr: false,
		},
		{
			name: "valid bool literal",
			expr: ast.Expression{
				Type:  ast.ExprLiteral,
				Value: true,
			},
			wantErr: false,
		},
		{
			name: "null literal value",
			expr: ast.Expression{
				Type:  ast.ExprLiteral,
				Value: nil,
			},
			wantErr: true,
			errMsg:  "literal expression must have a value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.validateExpression(&tt.expr, map[string]bool{}, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateExpression() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestValidateFieldExpression(t *testing.T) {
	tests := []struct {
		name    string
		module  *ast.Module
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid field access",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "test",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   ast.StmtAssign,
								Target: "obj",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "age"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: 25},
										},
									},
								},
							},
							{
								Type:   ast.StmtAssign,
								Target: "age_value",
								Value: &ast.Expression{
									Type:   ast.ExprField,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "obj"},
									Field:  "age",
								},
							},
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type: ast.ExprVariable,
									Name: "age_value",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "field access without object",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "test",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   ast.StmtAssign,
								Target: "result",
								Value: &ast.Expression{
									Type:   ast.ExprField,
									Object: nil,
									Field:  "age",
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "field expression must have an object",
		},
		{
			name: "field access without field name",
			module: &ast.Module{
				Type: "module",
				Name: "test_module",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "test",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							{
								Type:   ast.StmtAssign,
								Target: "obj",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "age"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: 25},
										},
									},
								},
							},
							{
								Type:   ast.StmtAssign,
								Target: "result",
								Value: &ast.Expression{
									Type:   ast.ExprField,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "obj"},
									Field:  "",
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "field expression must have a field name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			err := v.ValidateModule(tt.module)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModule() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateModule() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}
