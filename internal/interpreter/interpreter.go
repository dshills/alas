package interpreter

import (
	"fmt"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/runtime"
)

// Interpreter executes ALaS programs.
type Interpreter struct {
	modules   map[string]*ast.Module
	functions map[string]*ast.Function
}

// New creates a new interpreter.
func New() *Interpreter {
	return &Interpreter{
		modules:   make(map[string]*ast.Module),
		functions: make(map[string]*ast.Function),
	}
}

// LoadModule loads a module into the interpreter.
func (i *Interpreter) LoadModule(module *ast.Module) error {
	i.modules[module.Name] = module

	// Register all functions
	for idx := range module.Functions {
		fn := &module.Functions[idx]
		i.functions[fn.Name] = fn
	}

	return nil
}

// Environment represents the execution environment.
type Environment struct {
	vars   map[string]runtime.Value
	parent *Environment
}

// NewEnvironment creates a new environment.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:   make(map[string]runtime.Value),
		parent: parent,
	}
}

// Get retrieves a variable value.
func (e *Environment) Get(name string) (runtime.Value, bool) {
	if val, ok := e.vars[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return runtime.NewVoid(), false
}

// Set sets a variable value.
func (e *Environment) Set(name string, value runtime.Value) {
	e.vars[name] = value
}

// Run executes a function by name.
func (i *Interpreter) Run(functionName string, args []runtime.Value) (runtime.Value, error) {
	fn, ok := i.functions[functionName]
	if !ok {
		return runtime.NewVoid(), fmt.Errorf("function '%s' not found", functionName)
	}

	// Check argument count
	if len(args) != len(fn.Params) {
		return runtime.NewVoid(), fmt.Errorf("function '%s' expects %d arguments, got %d",
			functionName, len(fn.Params), len(args))
	}

	// Create new environment for function execution
	env := NewEnvironment(nil)

	// Bind parameters
	for idx, param := range fn.Params {
		env.Set(param.Name, args[idx])
	}

	// Execute function body
	result, _, err := i.executeStatements(fn.Body, env)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("error executing function '%s': %v", functionName, err)
	}

	return result, nil
}

// executeStatements executes a list of statements.
func (i *Interpreter) executeStatements(stmts []ast.Statement, env *Environment) (runtime.Value, bool, error) {
	var lastValue = runtime.NewVoid()

	for _, stmt := range stmts {
		val, isReturn, err := i.executeStatement(&stmt, env)
		if err != nil {
			return runtime.NewVoid(), false, err
		}
		if isReturn {
			return val, true, nil
		}
		lastValue = val
	}

	return lastValue, false, nil
}

// executeStatement executes a single statement.
func (i *Interpreter) executeStatement(stmt *ast.Statement, env *Environment) (runtime.Value, bool, error) {
	switch stmt.Type {
	case ast.StmtAssign:
		val, err := i.evaluateExpression(stmt.Value, env)
		if err != nil {
			return runtime.NewVoid(), false, err
		}
		env.Set(stmt.Target, val)
		return val, false, nil

	case ast.StmtIf:
		cond, err := i.evaluateExpression(stmt.Cond, env)
		if err != nil {
			return runtime.NewVoid(), false, err
		}

		if cond.IsTruthy() {
			return i.executeStatements(stmt.Then, env)
		} else if len(stmt.Else) > 0 {
			return i.executeStatements(stmt.Else, env)
		}
		return runtime.NewVoid(), false, nil

	case ast.StmtWhile:
		for {
			cond, err := i.evaluateExpression(stmt.Cond, env)
			if err != nil {
				return runtime.NewVoid(), false, err
			}

			if !cond.IsTruthy() {
				break
			}

			_, isReturn, err := i.executeStatements(stmt.Body, env)
			if err != nil {
				return runtime.NewVoid(), false, err
			}
			if isReturn {
				return runtime.NewVoid(), true, nil
			}
		}
		return runtime.NewVoid(), false, nil

	case ast.StmtReturn:
		if stmt.Value != nil {
			val, err := i.evaluateExpression(stmt.Value, env)
			if err != nil {
				return runtime.NewVoid(), false, err
			}
			return val, true, nil
		}
		return runtime.NewVoid(), true, nil

	case ast.StmtExpr:
		val, err := i.evaluateExpression(stmt.Value, env)
		if err != nil {
			return runtime.NewVoid(), false, err
		}
		return val, false, nil

	default:
		return runtime.NewVoid(), false, fmt.Errorf("unknown statement type: %s", stmt.Type)
	}
}

// evaluateExpression evaluates an expression.
func (i *Interpreter) evaluateExpression(expr *ast.Expression, env *Environment) (runtime.Value, error) {
	switch expr.Type {
	case ast.ExprLiteral:
		return i.evaluateLiteral(expr.Value)

	case ast.ExprVariable:
		val, ok := env.Get(expr.Name)
		if !ok {
			return runtime.NewVoid(), fmt.Errorf("undefined variable: %s", expr.Name)
		}
		return val, nil

	case ast.ExprBinary:
		left, err := i.evaluateExpression(expr.Left, env)
		if err != nil {
			return runtime.NewVoid(), err
		}
		right, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return runtime.NewVoid(), err
		}
		return i.evaluateBinaryOp(expr.Op, left, right)

	case ast.ExprUnary:
		operand, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return runtime.NewVoid(), err
		}
		return i.evaluateUnaryOp(expr.Op, operand)

	case ast.ExprCall:
		// Evaluate arguments
		args := make([]runtime.Value, len(expr.Args))
		for idx, arg := range expr.Args {
			val, err := i.evaluateExpression(&arg, env)
			if err != nil {
				return runtime.NewVoid(), err
			}
			args[idx] = val
		}
		return i.Run(expr.Name, args)

	case ast.ExprArrayLit:
		// Evaluate array literal
		elements := make([]runtime.Value, len(expr.Elements))
		for idx, elem := range expr.Elements {
			val, err := i.evaluateExpression(&elem, env)
			if err != nil {
				return runtime.NewVoid(), err
			}
			elements[idx] = val
		}
		return runtime.NewArray(elements), nil

	case ast.ExprMapLit:
		// Evaluate map literal
		mapValue := make(map[string]runtime.Value)
		for _, pair := range expr.Pairs {
			key, err := i.evaluateExpression(&pair.Key, env)
			if err != nil {
				return runtime.NewVoid(), err
			}
			value, err := i.evaluateExpression(&pair.Value, env)
			if err != nil {
				return runtime.NewVoid(), err
			}

			keyStr, err := key.AsString()
			if err != nil {
				return runtime.NewVoid(), fmt.Errorf("map key must be a string: %v", err)
			}
			mapValue[keyStr] = value
		}
		return runtime.NewMap(mapValue), nil

	case ast.ExprIndex:
		// Evaluate indexing (array[index] or map[key])
		object, err := i.evaluateExpression(expr.Object, env)
		if err != nil {
			return runtime.NewVoid(), err
		}
		index, err := i.evaluateExpression(expr.Index, env)
		if err != nil {
			return runtime.NewVoid(), err
		}
		return i.evaluateIndexAccess(object, index)

	default:
		return runtime.NewVoid(), fmt.Errorf("unknown expression type: %s", expr.Type)
	}
}

// evaluateLiteral evaluates a literal value.
func (i *Interpreter) evaluateLiteral(value interface{}) (runtime.Value, error) {
	switch v := value.(type) {
	case float64:
		// JSON numbers are always float64
		if float64(int64(v)) == v {
			return runtime.NewInt(int64(v)), nil
		}
		return runtime.NewFloat(v), nil
	case string:
		return runtime.NewString(v), nil
	case bool:
		return runtime.NewBool(v), nil
	case nil:
		return runtime.NewVoid(), nil
	default:
		return runtime.NewVoid(), fmt.Errorf("unsupported literal type: %T", value)
	}
}

// evaluateBinaryOp evaluates a binary operation.
func (i *Interpreter) evaluateBinaryOp(op string, left, right runtime.Value) (runtime.Value, error) {
	switch op {
	case ast.OpAdd:
		if left.Type == runtime.ValueTypeString || right.Type == runtime.ValueTypeString {
			// String concatenation
			return runtime.NewString(left.String() + right.String()), nil
		}
		// Numeric addition
		if left.Type == runtime.ValueTypeFloat || right.Type == runtime.ValueTypeFloat {
			l, _ := left.AsFloat()
			r, _ := right.AsFloat()
			return runtime.NewFloat(l + r), nil
		}
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		return runtime.NewInt(l + r), nil

	case ast.OpSub:
		if left.Type == runtime.ValueTypeFloat || right.Type == runtime.ValueTypeFloat {
			l, _ := left.AsFloat()
			r, _ := right.AsFloat()
			return runtime.NewFloat(l - r), nil
		}
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		return runtime.NewInt(l - r), nil

	case ast.OpMul:
		if left.Type == runtime.ValueTypeFloat || right.Type == runtime.ValueTypeFloat {
			l, _ := left.AsFloat()
			r, _ := right.AsFloat()
			return runtime.NewFloat(l * r), nil
		}
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		return runtime.NewInt(l * r), nil

	case ast.OpDiv:
		if left.Type == runtime.ValueTypeFloat || right.Type == runtime.ValueTypeFloat {
			l, _ := left.AsFloat()
			r, _ := right.AsFloat()
			if r == 0 {
				return runtime.NewVoid(), fmt.Errorf("division by zero")
			}
			return runtime.NewFloat(l / r), nil
		}
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		if r == 0 {
			return runtime.NewVoid(), fmt.Errorf("division by zero")
		}
		return runtime.NewInt(l / r), nil

	case ast.OpMod:
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		if r == 0 {
			return runtime.NewVoid(), fmt.Errorf("modulo by zero")
		}
		return runtime.NewInt(l % r), nil

	case ast.OpEq:
		return runtime.NewBool(i.valuesEqual(left, right)), nil

	case ast.OpNe:
		return runtime.NewBool(!i.valuesEqual(left, right)), nil

	case ast.OpLt:
		result := i.compareValues(left, right)
		return runtime.NewBool(result < 0), nil

	case ast.OpLe:
		result := i.compareValues(left, right)
		return runtime.NewBool(result <= 0), nil

	case ast.OpGt:
		result := i.compareValues(left, right)
		return runtime.NewBool(result > 0), nil

	case ast.OpGe:
		result := i.compareValues(left, right)
		return runtime.NewBool(result >= 0), nil

	case ast.OpAnd:
		return runtime.NewBool(left.IsTruthy() && right.IsTruthy()), nil

	case ast.OpOr:
		return runtime.NewBool(left.IsTruthy() || right.IsTruthy()), nil

	default:
		return runtime.NewVoid(), fmt.Errorf("unknown binary operator: %s", op)
	}
}

// evaluateUnaryOp evaluates a unary operation.
func (i *Interpreter) evaluateUnaryOp(op string, operand runtime.Value) (runtime.Value, error) {
	switch op {
	case ast.OpNot:
		return runtime.NewBool(!operand.IsTruthy()), nil

	case ast.OpNeg:
		if operand.Type == runtime.ValueTypeFloat {
			v, _ := operand.AsFloat()
			return runtime.NewFloat(-v), nil
		}
		v, _ := operand.AsInt()
		return runtime.NewInt(-v), nil

	default:
		return runtime.NewVoid(), fmt.Errorf("unknown unary operator: %s", op)
	}
}

// valuesEqual checks if two values are equal.
func (i *Interpreter) valuesEqual(left, right runtime.Value) bool {
	if left.Type != right.Type {
		return false
	}

	switch left.Type {
	case runtime.ValueTypeInt:
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		return l == r
	case runtime.ValueTypeFloat:
		l, _ := left.AsFloat()
		r, _ := right.AsFloat()
		return l == r
	case runtime.ValueTypeString:
		l, _ := left.AsString()
		r, _ := right.AsString()
		return l == r
	case runtime.ValueTypeBool:
		l, _ := left.AsBool()
		r, _ := right.AsBool()
		return l == r
	case runtime.ValueTypeVoid:
		return true
	case runtime.ValueTypeArray:
		l, _ := left.AsArray()
		r, _ := right.AsArray()
		if len(l) != len(r) {
			return false
		}
		for idx := range l {
			if !i.valuesEqual(l[idx], r[idx]) {
				return false
			}
		}
		return true
	case runtime.ValueTypeMap:
		l, _ := left.AsMap()
		r, _ := right.AsMap()
		if len(l) != len(r) {
			return false
		}
		for k, v := range l {
			if rv, ok := r[k]; !ok || !i.valuesEqual(v, rv) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// compareValues compares two values.
func (i *Interpreter) compareValues(left, right runtime.Value) int {
	if left.Type == runtime.ValueTypeString && right.Type == runtime.ValueTypeString {
		l, _ := left.AsString()
		r, _ := right.AsString()
		if l < r {
			return -1
		} else if l > r {
			return 1
		}
		return 0
	}

	// Numeric comparison
	var l, r float64
	if left.Type == runtime.ValueTypeFloat || right.Type == runtime.ValueTypeFloat {
		l, _ = left.AsFloat()
		r, _ = right.AsFloat()
	} else {
		li, _ := left.AsInt()
		ri, _ := right.AsInt()
		l = float64(li)
		r = float64(ri)
	}

	if l < r {
		return -1
	} else if l > r {
		return 1
	}
	return 0
}

// evaluateIndexAccess handles array and map indexing.
func (i *Interpreter) evaluateIndexAccess(object, index runtime.Value) (runtime.Value, error) {
	switch object.Type {
	case runtime.ValueTypeArray:
		arr, err := object.AsArray()
		if err != nil {
			return runtime.NewVoid(), err
		}

		idx, err := index.AsInt()
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("array index must be an integer: %v", err)
		}

		if idx < 0 || idx >= int64(len(arr)) {
			return runtime.NewVoid(), fmt.Errorf("array index out of bounds: %d", idx)
		}

		return arr[idx], nil

	case runtime.ValueTypeMap:
		m, err := object.AsMap()
		if err != nil {
			return runtime.NewVoid(), err
		}

		key, err := index.AsString()
		if err != nil {
			return runtime.NewVoid(), fmt.Errorf("map key must be a string: %v", err)
		}

		if val, ok := m[key]; ok {
			return val, nil
		}

		return runtime.NewVoid(), fmt.Errorf("map key not found: %s", key)

	case runtime.ValueTypeInt, runtime.ValueTypeFloat, runtime.ValueTypeString, runtime.ValueTypeBool, runtime.ValueTypeVoid:
		return runtime.NewVoid(), fmt.Errorf("cannot index into %v", object.Type)
	default:
		return runtime.NewVoid(), fmt.Errorf("cannot index into %v", object.Type)
	}
}
