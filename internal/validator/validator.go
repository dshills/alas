package validator

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	} else if !isValidModuleName(m.Name) {
		v.addError("invalid module name '%s', must be valid module name", m.Name)
	}

	// Validate custom types
	typeNames := make(map[string]bool)
	for i, typeDef := range m.Types {
		if err := v.validateTypeDefinition(&typeDef); err != nil {
			v.addError("type %d: %v", i, err)
		}
		if typeNames[typeDef.Name] {
			v.addError("duplicate type name: %s", typeDef.Name)
		}
		typeNames[typeDef.Name] = true
	}

	// Validate functions
	if len(m.Functions) == 0 {
		v.addError("module must contain at least one function")
	}

	functionNames := make(map[string]bool)
	for i, fn := range m.Functions {
		if err := v.validateFunction(&fn, typeNames); err != nil {
			v.addError("function %d: %v", i, err)
		}
		if functionNames[fn.Name] {
			v.addError("duplicate function name: %s", fn.Name)
		}
		functionNames[fn.Name] = true
	}

	// Validate exports reference actual functions
	for i, export := range m.Exports {
		if export == "" {
			v.addError("export %d: name cannot be empty", i)
		} else if !isValidIdentifier(export) {
			v.addError("export %d: invalid name '%s'", i, export)
		} else if !functionNames[export] {
			v.addError("exported function '%s' not found in module", export)
		}
	}

	// Validate imports are non-empty strings and don't include self
	for i, importName := range m.Imports {
		if importName == "" {
			v.addError("import %d: name cannot be empty", i)
		} else if !isValidModuleName(importName) {
			v.addError("import %d: invalid name '%s'", i, importName)
		} else if importName == m.Name {
			v.addError("module cannot import itself")
		}
	}

	// Check for duplicate imports
	importSet := make(map[string]bool)
	for i, importName := range m.Imports {
		if importSet[importName] {
			v.addError("import %d: duplicate import '%s'", i, importName)
		}
		importSet[importName] = true
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("validation errors:\n%s", strings.Join(v.errors, "\n"))
	}

	return nil
}

// validateTypeDefinition validates a custom type definition.
func (v *Validator) validateTypeDefinition(typeDef *ast.TypeDefinition) error {
	if typeDef.Name == "" {
		return fmt.Errorf("type name cannot be empty")
	}

	switch typeDef.Definition.Kind {
	case ast.TypeKindStruct:
		if len(typeDef.Definition.Fields) == 0 {
			return fmt.Errorf("struct type '%s' must have at least one field", typeDef.Name)
		}
		fieldNames := make(map[string]bool)
		for i, field := range typeDef.Definition.Fields {
			if field.Name == "" {
				return fmt.Errorf("field %d: name cannot be empty", i)
			}
			if fieldNames[field.Name] {
				return fmt.Errorf("duplicate field name: %s", field.Name)
			}
			fieldNames[field.Name] = true
			if !isValidType(field.Type, nil) {
				return fmt.Errorf("field %s: invalid type '%s'", field.Name, field.Type)
			}
		}
	case ast.TypeKindEnum:
		if len(typeDef.Definition.Values) == 0 {
			return fmt.Errorf("enum type '%s' must have at least one value", typeDef.Name)
		}
		valueNames := make(map[string]bool)
		for _, value := range typeDef.Definition.Values {
			if value == "" {
				return fmt.Errorf("enum value cannot be empty")
			}
			if valueNames[value] {
				return fmt.Errorf("duplicate enum value: %s", value)
			}
			valueNames[value] = true
		}
	default:
		return fmt.Errorf("unknown type kind: %s", typeDef.Definition.Kind)
	}

	return nil
}

// validateFunction validates a function definition.
func (v *Validator) validateFunction(fn *ast.Function, typeNames map[string]bool) error {
	if fn.Type != "function" {
		return fmt.Errorf("type must be 'function', got '%s'", fn.Type)
	}

	if fn.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !isValidIdentifier(fn.Name) {
		return fmt.Errorf("invalid function name '%s'", fn.Name)
	}

	// Validate parameters
	paramNames := make(map[string]bool)
	for i, param := range fn.Params {
		if param.Name == "" {
			return fmt.Errorf("parameter %d: name cannot be empty", i)
		}
		if !isValidIdentifier(param.Name) {
			return fmt.Errorf("parameter %d: invalid name '%s'", i, param.Name)
		}
		if paramNames[param.Name] {
			return fmt.Errorf("duplicate parameter name: %s", param.Name)
		}
		paramNames[param.Name] = true

		if !isValidType(param.Type, typeNames) {
			return fmt.Errorf("parameter %s: invalid type '%s'", param.Name, param.Type)
		}
	}

	// Validate return type
	if fn.Returns != "" && !isValidType(fn.Returns, typeNames) {
		return fmt.Errorf("invalid return type '%s'", fn.Returns)
	}

	// Validate body exists
	if fn.Body == nil {
		return fmt.Errorf("function body cannot be null")
	}

	// Create scope with parameters and type names
	scope := make(map[string]bool)
	for name := range paramNames {
		scope[name] = true
	}

	// Validate body statements
	for i, stmt := range fn.Body {
		if err := v.validateStatement(&stmt, scope, typeNames); err != nil {
			return fmt.Errorf("statement %d: %v", i, err)
		}
	}

	return nil
}

// validateStatement validates a statement.
func (v *Validator) validateStatement(stmt *ast.Statement, scope map[string]bool, typeNames map[string]bool) error {
	switch stmt.Type {
	case ast.StmtAssign:
		if stmt.Target == "" {
			return fmt.Errorf("assign statement must have a target")
		}
		if !isValidIdentifier(stmt.Target) {
			return fmt.Errorf("invalid assignment target '%s'", stmt.Target)
		}
		if stmt.Value == nil {
			return fmt.Errorf("assign statement must have a value")
		}
		if err := v.validateExpression(stmt.Value, scope, typeNames); err != nil {
			return fmt.Errorf("assign value: %v", err)
		}
		// Add target to scope
		scope[stmt.Target] = true

	case ast.StmtIf:
		if stmt.Cond == nil {
			return fmt.Errorf("if statement must have a condition")
		}
		if err := v.validateExpression(stmt.Cond, scope, typeNames); err != nil {
			return fmt.Errorf("if condition: %v", err)
		}
		if len(stmt.Then) == 0 {
			return fmt.Errorf("if statement must have a then block")
		}
		// Validate then block
		thenScope := copyScope(scope)
		for i, s := range stmt.Then {
			if err := v.validateStatement(&s, thenScope, typeNames); err != nil {
				return fmt.Errorf("then block statement %d: %v", i, err)
			}
		}
		// Validate else block if present
		if len(stmt.Else) > 0 {
			elseScope := copyScope(scope)
			for i, s := range stmt.Else {
				if err := v.validateStatement(&s, elseScope, typeNames); err != nil {
					return fmt.Errorf("else block statement %d: %v", i, err)
				}
			}
		}

	case ast.StmtWhile:
		if stmt.Cond == nil {
			return fmt.Errorf("while statement must have a condition")
		}
		if err := v.validateExpression(stmt.Cond, scope, typeNames); err != nil {
			return fmt.Errorf("while condition: %v", err)
		}
		if len(stmt.Body) == 0 {
			return fmt.Errorf("while statement must have a body")
		}
		// Validate body
		bodyScope := copyScope(scope)
		for i, s := range stmt.Body {
			if err := v.validateStatement(&s, bodyScope, typeNames); err != nil {
				return fmt.Errorf("while body statement %d: %v", i, err)
			}
		}

	case ast.StmtFor:
		if stmt.Cond == nil {
			return fmt.Errorf("for statement must have a condition")
		}
		if err := v.validateExpression(stmt.Cond, scope, typeNames); err != nil {
			return fmt.Errorf("for condition: %v", err)
		}
		if len(stmt.Body) == 0 {
			return fmt.Errorf("for statement must have a body")
		}
		// Validate body
		bodyScope := copyScope(scope)
		for i, s := range stmt.Body {
			if err := v.validateStatement(&s, bodyScope, typeNames); err != nil {
				return fmt.Errorf("for body statement %d: %v", i, err)
			}
		}

	case ast.StmtReturn:
		if stmt.Value != nil {
			if err := v.validateExpression(stmt.Value, scope, typeNames); err != nil {
				return fmt.Errorf("return value: %v", err)
			}
		}

	case ast.StmtExpr:
		if stmt.Value == nil {
			return fmt.Errorf("expression statement must have a value")
		}
		if err := v.validateExpression(stmt.Value, scope, typeNames); err != nil {
			return fmt.Errorf("expression: %v", err)
		}

	default:
		return fmt.Errorf("unknown statement type: %s", stmt.Type)
	}

	return nil
}

// validateExpression validates an expression with comprehensive schema checking.
// The typeNames parameter is currently unused but kept for future type checking enhancements.
//
//nolint:unparam // typeNames will be used for type inference in future
func (v *Validator) validateExpression(expr *ast.Expression, scope map[string]bool, typeNames map[string]bool) error {
	switch expr.Type {
	case ast.ExprLiteral:
		if expr.Value == nil {
			return fmt.Errorf("literal expression must have a value")
		}
		// Enhanced literal validation based on value type
		switch expr.Value.(type) {
		case string:
			if err := v.validateStringLiteral(expr.Value); err != nil {
				return fmt.Errorf("string literal: %v", err)
			}
		case bool:
			if err := v.validateBooleanLiteral(expr.Value); err != nil {
				return fmt.Errorf("boolean literal: %v", err)
			}
		default:
			// Numeric literals (int, float)
			if err := v.validateNumericLiteral(expr.Value); err != nil {
				return fmt.Errorf("numeric literal: %v", err)
			}
		}

	case ast.ExprVariable:
		if expr.Name == "" {
			return fmt.Errorf("variable expression must have a name")
		}
		// Validate variable name format
		if !isValidIdentifier(expr.Name) {
			return fmt.Errorf("invalid variable name '%s'", expr.Name)
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
		if err := v.validateExpression(expr.Left, scope, typeNames); err != nil {
			return fmt.Errorf("left operand: %v", err)
		}
		if err := v.validateExpression(expr.Right, scope, typeNames); err != nil {
			return fmt.Errorf("right operand: %v", err)
		}

	case ast.ExprUnary:
		if expr.Op == "" {
			return fmt.Errorf("unary expression must have an operator")
		}
		if !isValidUnaryOp(expr.Op) {
			return fmt.Errorf("invalid unary operator: %s", expr.Op)
		}
		// Support both Operand (spec-compliant) and Right (backward compatibility)
		var operandExpr *ast.Expression
		if expr.Operand != nil {
			operandExpr = expr.Operand
		} else if expr.Right != nil {
			operandExpr = expr.Right
		} else {
			return fmt.Errorf("unary expression must have an operand")
		}
		if err := v.validateExpression(operandExpr, scope, typeNames); err != nil {
			return fmt.Errorf("unary operand: %v", err)
		}

	case ast.ExprCall:
		if expr.Name == "" {
			return fmt.Errorf("call expression must have a function name")
		}
		// Validate function name format
		if !isValidIdentifier(expr.Name) {
			return fmt.Errorf("invalid function name '%s'", expr.Name)
		}
		// Validate arguments structure
		if expr.Args == nil {
			return fmt.Errorf("function call must have args field (can be empty)")
		}
		// Validate arguments
		for i, arg := range expr.Args {
			if err := v.validateExpression(&arg, scope, typeNames); err != nil {
				return fmt.Errorf("argument %d: %v", i, err)
			}
		}

	case ast.ExprArrayLit:
		// Validate array literal structure
		if expr.Elements == nil {
			return fmt.Errorf("array literal must have elements field (can be empty)")
		}
		// Validate array elements
		for i, elem := range expr.Elements {
			if err := v.validateExpression(&elem, scope, typeNames); err != nil {
				return fmt.Errorf("array element %d: %v", i, err)
			}
			// Validate array element has proper structure
			if elem.Type == "" {
				return fmt.Errorf("array element %d: missing type field", i)
			}
		}

	case ast.ExprMapLit:
		// Validate map literal structure
		if expr.Pairs == nil {
			return fmt.Errorf("map literal must have pairs field (can be empty)")
		}
		// Validate map key-value pairs
		for i, pair := range expr.Pairs {
			if err := v.validateExpression(&pair.Key, scope, typeNames); err != nil {
				return fmt.Errorf("map pair %d key: %v", i, err)
			}
			if err := v.validateExpression(&pair.Value, scope, typeNames); err != nil {
				return fmt.Errorf("map pair %d value: %v", i, err)
			}
			// Validate key and value have proper structure
			if pair.Key.Type == "" {
				return fmt.Errorf("map pair %d key: missing type field", i)
			}
			if pair.Value.Type == "" {
				return fmt.Errorf("map pair %d value: missing type field", i)
			}
		}

	case ast.ExprIndex:
		if expr.Object == nil {
			return fmt.Errorf("index expression must have an object")
		}
		if expr.Index == nil {
			return fmt.Errorf("index expression must have an index")
		}
		if err := v.validateExpression(expr.Object, scope, typeNames); err != nil {
			return fmt.Errorf("index object: %v", err)
		}
		if err := v.validateExpression(expr.Index, scope, typeNames); err != nil {
			return fmt.Errorf("index: %v", err)
		}

	case ast.ExprModuleCall:
		if expr.Module == "" {
			return fmt.Errorf("module call expression must have a module name")
		}
		if expr.Name == "" {
			return fmt.Errorf("module call expression must have a function name")
		}
		// Validate identifiers
		if !isValidModuleName(expr.Module) {
			return fmt.Errorf("invalid module name '%s'", expr.Module)
		}
		if !isValidIdentifier(expr.Name) {
			return fmt.Errorf("invalid function name '%s'", expr.Name)
		}
		// Validate arguments structure
		if expr.Args == nil {
			return fmt.Errorf("module call must have args field (can be empty)")
		}
		// Validate arguments
		for i, arg := range expr.Args {
			if err := v.validateExpression(&arg, scope, typeNames); err != nil {
				return fmt.Errorf("module call argument %d: %v", i, err)
			}
		}

	case ast.ExprBuiltin:
		if expr.Name == "" {
			return fmt.Errorf("builtin call expression must have a function name")
		}
		// Validate builtin function name format
		if err := v.validateBuiltinName(expr.Name); err != nil {
			return fmt.Errorf("invalid builtin name: %v", err)
		}
		// Validate arguments structure
		if expr.Args == nil {
			return fmt.Errorf("builtin call must have args field (can be empty)")
		}
		// Validate arguments
		for i, arg := range expr.Args {
			if err := v.validateExpression(&arg, scope, typeNames); err != nil {
				return fmt.Errorf("builtin call argument %d: %v", i, err)
			}
		}

	case ast.ExprField:
		if expr.Object == nil {
			return fmt.Errorf("field expression must have an object")
		}
		if expr.Field == "" {
			return fmt.Errorf("field expression must have a field name")
		}
		if err := v.validateExpression(expr.Object, scope, typeNames); err != nil {
			return fmt.Errorf("field object: %v", err)
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

func isValidType(t string, typeNames map[string]bool) bool {
	switch t {
	case ast.TypeInt, ast.TypeFloat, ast.TypeString, ast.TypeBool,
		ast.TypeArray, ast.TypeMap, ast.TypeVoid:
		return true
	default:
		// Check if it's a custom type
		if typeNames != nil && typeNames[t] {
			return true
		}
		// For backward compatibility, accept any non-empty string as a type
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

// validateBuiltinName validates builtin function names follow expected format.
func (v *Validator) validateBuiltinName(name string) error {
	// Builtin names should follow format: namespace.function
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return fmt.Errorf("builtin name must be in format 'namespace.function', got '%s'", name)
	}
	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("builtin namespace and function cannot be empty in '%s'", name)
	}
	// Validate known builtin namespaces
	knownNamespaces := map[string]bool{
		"io":          true,
		"math":        true,
		"string":      true,
		"array":       true,
		"map":         true,
		"collections": true,
		"type":        true,
		"async":       true,
	}
	if !knownNamespaces[parts[0]] {
		return fmt.Errorf("unknown builtin namespace '%s', expected one of: io, math, string, array, map, collections, type, async", parts[0])
	}
	return nil
}

// isValidIdentifier validates that a string is a valid identifier.
func isValidIdentifier(name string) bool {
	// ALaS identifiers must start with letter or underscore, followed by letters, digits, or underscores
	identifierPattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return identifierPattern.MatchString(name)
}

// isValidModuleName validates that a string is a valid module name (allows dots for namespacing).
func isValidModuleName(name string) bool {
	// Module names can have dots for namespacing, like "std.io"
	modulePattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*[a-zA-Z0-9_]$|^[a-zA-Z_]$`)
	return modulePattern.MatchString(name)
}

// validateStringLiteral validates string literal values.
func (v *Validator) validateStringLiteral(value interface{}) error {
	if value == nil {
		return fmt.Errorf("string literal cannot be null")
	}
	if _, ok := value.(string); !ok {
		return fmt.Errorf("string literal value must be a string, got %T", value)
	}
	return nil
}

// validateNumericLiteral validates numeric literal values.
func (v *Validator) validateNumericLiteral(value interface{}) error {
	if value == nil {
		return fmt.Errorf("numeric literal cannot be null")
	}
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return nil
	case float32, float64:
		return nil
	default:
		return fmt.Errorf("numeric literal value must be int or float, got %T", value)
	}
}

// validateBooleanLiteral validates boolean literal values.
func (v *Validator) validateBooleanLiteral(value interface{}) error {
	if value == nil {
		return fmt.Errorf("boolean literal cannot be null")
	}
	if _, ok := value.(bool); !ok {
		return fmt.Errorf("boolean literal value must be a boolean, got %T", value)
	}
	return nil
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
