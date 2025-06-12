package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dshills/alas/internal/ast"
)

// Validator validates ALaS AST structures.
type Validator struct {
	errors []string
}

// New creates a new validator.
func New() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// ValidateModule validates a complete module.
func (v *Validator) ValidateModule(m *ast.Module) error {
	v.errors = make([]string, 0)

	// Validate module type
	if m.Type != "module" {
		v.addError("module type must be 'module', got '%s'", m.Type)
	}

	// Validate module name
	if m.Name == "" {
		v.addError("module name cannot be empty")
	}

	// Validate functions
	if len(m.Functions) == 0 {
		v.addError("module must contain at least one function")
	}

	functionNames := make(map[string]bool)
	for i, fn := range m.Functions {
		if err := v.validateFunction(&fn); err != nil {
			v.addError("function %d: %v", i, err)
		}
		if functionNames[fn.Name] {
			v.addError("duplicate function name: %s", fn.Name)
		}
		functionNames[fn.Name] = true
	}

	// Validate exports reference actual functions
	for _, export := range m.Exports {
		if !functionNames[export] {
			v.addError("exported function '%s' not found in module", export)
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("validation errors:\n%s", strings.Join(v.errors, "\n"))
	}

	return nil
}

// validateFunction validates a function definition.
func (v *Validator) validateFunction(fn *ast.Function) error {
	if fn.Type != "function" {
		return fmt.Errorf("type must be 'function', got '%s'", fn.Type)
	}

	if fn.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Validate parameters
	paramNames := make(map[string]bool)
	for i, param := range fn.Params {
		if param.Name == "" {
			return fmt.Errorf("parameter %d: name cannot be empty", i)
		}
		if paramNames[param.Name] {
			return fmt.Errorf("duplicate parameter name: %s", param.Name)
		}
		paramNames[param.Name] = true

		if !isValidType(param.Type) {
			return fmt.Errorf("parameter %s: invalid type '%s'", param.Name, param.Type)
		}
	}

	// Validate return type
	if fn.Returns != "" && !isValidType(fn.Returns) {
		return fmt.Errorf("invalid return type '%s'", fn.Returns)
	}

	// Validate body exists
	if fn.Body == nil {
		return fmt.Errorf("function body cannot be null")
	}

	// Validate body statements
	for i, stmt := range fn.Body {
		if err := v.validateStatement(&stmt, paramNames); err != nil {
			return fmt.Errorf("statement %d: %v", i, err)
		}
	}

	return nil
}

// validateStatement validates a statement.
func (v *Validator) validateStatement(stmt *ast.Statement, scope map[string]bool) error {
	switch stmt.Type {
	case ast.StmtAssign:
		if stmt.Target == "" {
			return fmt.Errorf("assign statement must have a target")
		}
		if stmt.Value == nil {
			return fmt.Errorf("assign statement must have a value")
		}
		if err := v.validateExpression(stmt.Value, scope); err != nil {
			return fmt.Errorf("assign value: %v", err)
		}
		// Add target to scope
		scope[stmt.Target] = true

	case ast.StmtIf:
		if stmt.Cond == nil {
			return fmt.Errorf("if statement must have a condition")
		}
		if err := v.validateExpression(stmt.Cond, scope); err != nil {
			return fmt.Errorf("if condition: %v", err)
		}
		if len(stmt.Then) == 0 {
			return fmt.Errorf("if statement must have a then block")
		}
		// Validate then block
		thenScope := copyScope(scope)
		for i, s := range stmt.Then {
			if err := v.validateStatement(&s, thenScope); err != nil {
				return fmt.Errorf("then block statement %d: %v", i, err)
			}
		}
		// Validate else block if present
		if len(stmt.Else) > 0 {
			elseScope := copyScope(scope)
			for i, s := range stmt.Else {
				if err := v.validateStatement(&s, elseScope); err != nil {
					return fmt.Errorf("else block statement %d: %v", i, err)
				}
			}
		}

	case ast.StmtWhile:
		if stmt.Cond == nil {
			return fmt.Errorf("while statement must have a condition")
		}
		if err := v.validateExpression(stmt.Cond, scope); err != nil {
			return fmt.Errorf("while condition: %v", err)
		}
		if len(stmt.Body) == 0 {
			return fmt.Errorf("while statement must have a body")
		}
		// Validate body
		bodyScope := copyScope(scope)
		for i, s := range stmt.Body {
			if err := v.validateStatement(&s, bodyScope); err != nil {
				return fmt.Errorf("while body statement %d: %v", i, err)
			}
		}

	case ast.StmtReturn:
		if stmt.Value != nil {
			if err := v.validateExpression(stmt.Value, scope); err != nil {
				return fmt.Errorf("return value: %v", err)
			}
		}

	case ast.StmtExpr:
		if stmt.Value == nil {
			return fmt.Errorf("expression statement must have a value")
		}
		if err := v.validateExpression(stmt.Value, scope); err != nil {
			return fmt.Errorf("expression: %v", err)
		}

	default:
		return fmt.Errorf("unknown statement type: %s", stmt.Type)
	}

	return nil
}

// validateExpression validates an expression.
func (v *Validator) validateExpression(expr *ast.Expression, scope map[string]bool) error {
	switch expr.Type {
	case ast.ExprLiteral:
		if expr.Value == nil {
			return fmt.Errorf("literal expression must have a value")
		}

	case ast.ExprVariable:
		if expr.Name == "" {
			return fmt.Errorf("variable expression must have a name")
		}
		// Check if variable is in scope
		if !scope[expr.Name] {
			return fmt.Errorf("undefined variable: %s", expr.Name)
		}

	case ast.ExprBinary:
		if expr.Op == "" {
			return fmt.Errorf("binary expression must have an operator")
		}
		if !isValidBinaryOp(expr.Op) {
			return fmt.Errorf("invalid binary operator: %s", expr.Op)
		}
		if expr.Left == nil || expr.Right == nil {
			return fmt.Errorf("binary expression must have left and right operands")
		}
		if err := v.validateExpression(expr.Left, scope); err != nil {
			return fmt.Errorf("left operand: %v", err)
		}
		if err := v.validateExpression(expr.Right, scope); err != nil {
			return fmt.Errorf("right operand: %v", err)
		}

	case ast.ExprUnary:
		if expr.Op == "" {
			return fmt.Errorf("unary expression must have an operator")
		}
		if !isValidUnaryOp(expr.Op) {
			return fmt.Errorf("invalid unary operator: %s", expr.Op)
		}
		if expr.Right == nil {
			return fmt.Errorf("unary expression must have an operand")
		}
		if err := v.validateExpression(expr.Right, scope); err != nil {
			return fmt.Errorf("unary operand: %v", err)
		}

	case ast.ExprCall:
		if expr.Name == "" {
			return fmt.Errorf("call expression must have a function name")
		}
		// Validate arguments
		for i, arg := range expr.Args {
			if err := v.validateExpression(&arg, scope); err != nil {
				return fmt.Errorf("argument %d: %v", i, err)
			}
		}

	case ast.ExprArrayLit:
		// Validate array elements
		for i, elem := range expr.Elements {
			if err := v.validateExpression(&elem, scope); err != nil {
				return fmt.Errorf("array element %d: %v", i, err)
			}
		}

	case ast.ExprMapLit:
		// Validate map key-value pairs
		for i, pair := range expr.Pairs {
			if err := v.validateExpression(&pair.Key, scope); err != nil {
				return fmt.Errorf("map pair %d key: %v", i, err)
			}
			if err := v.validateExpression(&pair.Value, scope); err != nil {
				return fmt.Errorf("map pair %d value: %v", i, err)
			}
		}

	case ast.ExprIndex:
		if expr.Object == nil {
			return fmt.Errorf("index expression must have an object")
		}
		if expr.Index == nil {
			return fmt.Errorf("index expression must have an index")
		}
		if err := v.validateExpression(expr.Object, scope); err != nil {
			return fmt.Errorf("index object: %v", err)
		}
		if err := v.validateExpression(expr.Index, scope); err != nil {
			return fmt.Errorf("index: %v", err)
		}

	default:
		return fmt.Errorf("unknown expression type: %s", expr.Type)
	}

	return nil
}

// Helper functions

func (v *Validator) addError(format string, args ...interface{}) {
	v.errors = append(v.errors, fmt.Sprintf(format, args...))
}

func isValidType(t string) bool {
	switch t {
	case ast.TypeInt, ast.TypeFloat, ast.TypeString, ast.TypeBool,
		ast.TypeArray, ast.TypeMap, ast.TypeVoid:
		return true
	default:
		// Could be a custom type - for now accept anything that's not empty
		return t != ""
	}
}

func isValidBinaryOp(op string) bool {
	switch op {
	case ast.OpAdd, ast.OpSub, ast.OpMul, ast.OpDiv, ast.OpMod,
		ast.OpEq, ast.OpNe, ast.OpLt, ast.OpLe, ast.OpGt, ast.OpGe,
		ast.OpAnd, ast.OpOr:
		return true
	default:
		return false
	}
}

func isValidUnaryOp(op string) bool {
	switch op {
	case ast.OpNot, ast.OpNeg:
		return true
	default:
		return false
	}
}

func copyScope(scope map[string]bool) map[string]bool {
	newScope := make(map[string]bool)
	for k, v := range scope {
		newScope[k] = v
	}
	return newScope
}

// ValidateJSON validates ALaS JSON input.
func ValidateJSON(input []byte) error {
	var module ast.Module
	if err := json.Unmarshal(input, &module); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	validator := New()
	return validator.ValidateModule(&module)
}
