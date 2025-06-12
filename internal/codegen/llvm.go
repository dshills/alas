package codegen

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"github.com/dshills/alas/internal/ast"
)

// LLVMCodegen generates LLVM IR from ALaS AST.
type LLVMCodegen struct {
	module    *ir.Module
	builder   *ir.Block
	functions map[string]*ir.Func
	variables map[string]value.Value
}

// NewLLVMCodegen creates a new LLVM code generator.
func NewLLVMCodegen() *LLVMCodegen {
	return &LLVMCodegen{
		module:    ir.NewModule(),
		functions: make(map[string]*ir.Func),
		variables: make(map[string]value.Value),
	}
}

// GenerateModule generates LLVM IR for an entire ALaS module.
func (g *LLVMCodegen) GenerateModule(module *ast.Module) (*ir.Module, error) {
	g.module.SourceFilename = module.Name + ".alas"

	// First pass: declare all functions
	for _, fn := range module.Functions {
		if err := g.declareFunction(&fn); err != nil {
			return nil, fmt.Errorf("failed to declare function %s: %v", fn.Name, err)
		}
	}

	// Second pass: generate function bodies
	for _, fn := range module.Functions {
		if err := g.generateFunction(&fn); err != nil {
			return nil, fmt.Errorf("failed to generate function %s: %v", fn.Name, err)
		}
	}

	return g.module, nil
}

// declareFunction declares a function signature in LLVM IR.
func (g *LLVMCodegen) declareFunction(fn *ast.Function) error {
	// Convert parameter types
	paramTypes := make([]types.Type, len(fn.Params))
	for i, param := range fn.Params {
		paramType, err := g.convertType(param.Type)
		if err != nil {
			return fmt.Errorf("invalid parameter type %s: %v", param.Type, err)
		}
		paramTypes[i] = paramType
	}

	// Convert return type
	returnType, err := g.convertType(fn.Returns)
	if err != nil {
		return fmt.Errorf("invalid return type %s: %v", fn.Returns, err)
	}

	// Create function signature
	sig := types.NewFunc(returnType, paramTypes...)

	// Create function
	llvmFunc := g.module.NewFunc(fn.Name, sig)

	// Create parameters manually if they don't exist
	if len(llvmFunc.Params) == 0 && len(fn.Params) > 0 {
		for _, param := range fn.Params {
			paramType, _ := g.convertType(param.Type)
			llvmParam := ir.NewParam(param.Name, paramType)
			llvmFunc.Params = append(llvmFunc.Params, llvmParam)
		}
	}

	// Set parameter names
	for i, param := range fn.Params {
		if i < len(llvmFunc.Params) {
			llvmFunc.Params[i].SetName(param.Name)
		}
	}

	g.functions[fn.Name] = llvmFunc
	return nil
}

// generateFunction generates the body of a function.
func (g *LLVMCodegen) generateFunction(fn *ast.Function) error {
	llvmFunc := g.functions[fn.Name]

	// Create entry block
	entry := llvmFunc.NewBlock("entry")
	g.builder = entry

	// Create new variable scope for this function
	oldVars := g.variables
	g.variables = make(map[string]value.Value)

	// Add parameters to variable scope
	for i, param := range fn.Params {
		if i < len(llvmFunc.Params) {
			g.variables[param.Name] = llvmFunc.Params[i]
		}
	}

	// Generate function body
	var lastValue value.Value
	for _, stmt := range fn.Body {
		val, isReturn, err := g.generateStatement(&stmt)
		if err != nil {
			return err
		}
		if isReturn {
			return nil // Function already has return instruction
		}
		lastValue = val
	}

	// If no explicit return and function expects void, add return
	if fn.Returns == "void" || fn.Returns == "" {
		g.builder.NewRet(nil)
	} else if lastValue != nil {
		// Return the last expression value
		g.builder.NewRet(lastValue)
	} else {
		// Return zero value for the type
		returnType, _ := g.convertType(fn.Returns)
		zero := g.getZeroValue(returnType)
		g.builder.NewRet(zero)
	}

	// Restore previous variable scope
	g.variables = oldVars
	return nil
}

// generateStatement generates LLVM IR for a statement.
func (g *LLVMCodegen) generateStatement(stmt *ast.Statement) (value.Value, bool, error) {
	switch stmt.Type {
	case ast.StmtAssign:
		val, err := g.generateExpression(stmt.Value)
		if err != nil {
			return nil, false, err
		}
		g.variables[stmt.Target] = val
		return val, false, nil

	case ast.StmtReturn:
		if stmt.Value != nil {
			val, err := g.generateExpression(stmt.Value)
			if err != nil {
				return nil, false, err
			}
			g.builder.NewRet(val)
		} else {
			g.builder.NewRet(nil)
		}
		return nil, true, nil

	case ast.StmtExpr:
		val, err := g.generateExpression(stmt.Value)
		if err != nil {
			return nil, false, err
		}
		return val, false, nil

	case ast.StmtIf:
		return g.generateIf(stmt)

	case ast.StmtWhile:
		return g.generateWhile(stmt)

	default:
		return nil, false, fmt.Errorf("unsupported statement type: %s", stmt.Type)
	}
}

// generateExpression generates LLVM IR for an expression.
func (g *LLVMCodegen) generateExpression(expr *ast.Expression) (value.Value, error) {
	switch expr.Type {
	case ast.ExprLiteral:
		return g.generateLiteral(expr.Value)

	case ast.ExprVariable:
		val, ok := g.variables[expr.Name]
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", expr.Name)
		}
		return val, nil

	case ast.ExprBinary:
		return g.generateBinary(expr)

	case ast.ExprUnary:
		return g.generateUnary(expr)

	case ast.ExprCall:
		return g.generateCall(expr)

	case ast.ExprArrayLit:
		return g.generateArrayLiteral(expr)

	case ast.ExprMapLit:
		return g.generateMapLiteral(expr)

	case ast.ExprIndex:
		return g.generateIndexAccess(expr)

	default:
		return nil, fmt.Errorf("unsupported expression type: %s", expr.Type)
	}
}

// generateLiteral generates LLVM IR for a literal value.
func (g *LLVMCodegen) generateLiteral(value interface{}) (value.Value, error) {
	switch v := value.(type) {
	case float64:
		// JSON numbers are always float64 - check if it's actually an int
		if float64(int64(v)) == v {
			return constant.NewInt(types.I64, int64(v)), nil
		}
		return constant.NewFloat(types.Double, v), nil
	case string:
		// Create a global string constant
		charArray := constant.NewCharArrayFromString(v + "\x00")
		str := g.module.NewGlobalDef("", charArray)
		str.Immutable = true
		// Return pointer to the first character of the string
		return g.builder.NewGetElementPtr(charArray.Type(), str, constant.NewInt(types.I64, 0), constant.NewInt(types.I64, 0)), nil
	case bool:
		if v {
			return constant.NewInt(types.I1, 1), nil
		}
		return constant.NewInt(types.I1, 0), nil
	default:
		return nil, fmt.Errorf("unsupported literal type: %T", value)
	}
}

// generateBinary generates LLVM IR for binary operations.
func (g *LLVMCodegen) generateBinary(expr *ast.Expression) (value.Value, error) {
	left, err := g.generateExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := g.generateExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	// Type promotion: if either operand is float, promote both to float
	leftType := left.Type()
	rightType := right.Type()

	isFloat := (leftType.Equal(types.Double) || rightType.Equal(types.Double))

	if isFloat {
		// Promote to float if needed
		if !leftType.Equal(types.Double) {
			left = g.builder.NewSIToFP(left, types.Double)
		}
		if !rightType.Equal(types.Double) {
			right = g.builder.NewSIToFP(right, types.Double)
		}
	}

	switch expr.Op {
	case ast.OpAdd:
		if isFloat {
			return g.builder.NewFAdd(left, right), nil
		}
		return g.builder.NewAdd(left, right), nil

	case ast.OpSub:
		if isFloat {
			return g.builder.NewFSub(left, right), nil
		}
		return g.builder.NewSub(left, right), nil

	case ast.OpMul:
		if isFloat {
			return g.builder.NewFMul(left, right), nil
		}
		return g.builder.NewMul(left, right), nil

	case ast.OpDiv:
		if isFloat {
			return g.builder.NewFDiv(left, right), nil
		}
		return g.builder.NewSDiv(left, right), nil

	case ast.OpMod:
		if isFloat {
			return g.builder.NewFRem(left, right), nil
		}
		return g.builder.NewSRem(left, right), nil

	case ast.OpEq:
		if isFloat {
			return g.builder.NewFCmp(enum.FPredOEQ, left, right), nil
		}
		return g.builder.NewICmp(enum.IPredEQ, left, right), nil

	case ast.OpNe:
		if isFloat {
			return g.builder.NewFCmp(enum.FPredONE, left, right), nil
		}
		return g.builder.NewICmp(enum.IPredNE, left, right), nil

	case ast.OpLt:
		if isFloat {
			return g.builder.NewFCmp(enum.FPredOLT, left, right), nil
		}
		return g.builder.NewICmp(enum.IPredSLT, left, right), nil

	case ast.OpLe:
		if isFloat {
			return g.builder.NewFCmp(enum.FPredOLE, left, right), nil
		}
		return g.builder.NewICmp(enum.IPredSLE, left, right), nil

	case ast.OpGt:
		if isFloat {
			return g.builder.NewFCmp(enum.FPredOGT, left, right), nil
		}
		return g.builder.NewICmp(enum.IPredSGT, left, right), nil

	case ast.OpGe:
		if isFloat {
			return g.builder.NewFCmp(enum.FPredOGE, left, right), nil
		}
		return g.builder.NewICmp(enum.IPredSGE, left, right), nil

	case ast.OpAnd:
		return g.builder.NewAnd(left, right), nil

	case ast.OpOr:
		return g.builder.NewOr(left, right), nil

	default:
		return nil, fmt.Errorf("unsupported binary operator: %s", expr.Op)
	}
}

// generateUnary generates LLVM IR for unary operations.
func (g *LLVMCodegen) generateUnary(expr *ast.Expression) (value.Value, error) {
	operand, err := g.generateExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	case ast.OpNot:
		// XOR with 1 for boolean not
		one := constant.NewInt(operand.Type().(*types.IntType), 1)
		return g.builder.NewXor(operand, one), nil

	case ast.OpNeg:
		if operand.Type().Equal(types.Double) {
			zero := constant.NewFloat(types.Double, 0.0)
			return g.builder.NewFSub(zero, operand), nil
		}
		zero := constant.NewInt(operand.Type().(*types.IntType), 0)
		return g.builder.NewSub(zero, operand), nil

	default:
		return nil, fmt.Errorf("unsupported unary operator: %s", expr.Op)
	}
}

// generateCall generates LLVM IR for function calls.
func (g *LLVMCodegen) generateCall(expr *ast.Expression) (value.Value, error) {
	fn, ok := g.functions[expr.Name]
	if !ok {
		return nil, fmt.Errorf("undefined function: %s", expr.Name)
	}

	// Generate arguments
	args := make([]value.Value, len(expr.Args))
	for i, arg := range expr.Args {
		val, err := g.generateExpression(&arg)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return g.builder.NewCall(fn, args...), nil
}

// generateIf generates LLVM IR for if statements.
func (g *LLVMCodegen) generateIf(stmt *ast.Statement) (value.Value, bool, error) {
	// Generate condition
	cond, err := g.generateExpression(stmt.Cond)
	if err != nil {
		return nil, false, err
	}

	// Create basic blocks
	currentFunc := g.builder.Parent
	thenBlock := currentFunc.NewBlock("if.then")
	elseBlock := currentFunc.NewBlock("if.else")
	endBlock := currentFunc.NewBlock("if.end")

	// Branch based on condition
	g.builder.NewCondBr(cond, thenBlock, elseBlock)

	// Generate then block
	g.builder = thenBlock
	var thenValue value.Value
	var thenReturn bool
	for _, s := range stmt.Then {
		val, isReturn, err := g.generateStatement(&s)
		if err != nil {
			return nil, false, err
		}
		if isReturn {
			thenReturn = true
			break
		}
		thenValue = val
	}
	if !thenReturn {
		g.builder.NewBr(endBlock)
	}

	// Generate else block
	g.builder = elseBlock
	var elseValue value.Value
	var elseReturn bool
	if len(stmt.Else) > 0 {
		for _, s := range stmt.Else {
			val, isReturn, err := g.generateStatement(&s)
			if err != nil {
				return nil, false, err
			}
			if isReturn {
				elseReturn = true
				break
			}
			elseValue = val
		}
	}
	if !elseReturn {
		g.builder.NewBr(endBlock)
	}

	// Continue with end block
	g.builder = endBlock

	// If both branches returned, this block is unreachable, but still needs a terminator
	if thenReturn && elseReturn {
		g.builder.NewUnreachable()
		return nil, true, nil
	}

	// TODO: Handle phi nodes for values from different branches
	if thenValue != nil || elseValue != nil {
		// For now, just return the then value if available
		if thenValue != nil {
			return thenValue, false, nil
		}
		return elseValue, false, nil
	}

	return nil, false, nil
}

// generateWhile generates LLVM IR for while loops.
func (g *LLVMCodegen) generateWhile(stmt *ast.Statement) (value.Value, bool, error) {
	currentFunc := g.builder.Parent
	condBlock := currentFunc.NewBlock("while.cond")
	bodyBlock := currentFunc.NewBlock("while.body")
	endBlock := currentFunc.NewBlock("while.end")

	// Jump to condition block
	g.builder.NewBr(condBlock)

	// Generate condition block
	g.builder = condBlock
	cond, err := g.generateExpression(stmt.Cond)
	if err != nil {
		return nil, false, err
	}
	g.builder.NewCondBr(cond, bodyBlock, endBlock)

	// Generate body block
	g.builder = bodyBlock
	for _, s := range stmt.Body {
		_, isReturn, err := g.generateStatement(&s)
		if err != nil {
			return nil, false, err
		}
		if isReturn {
			return nil, true, nil
		}
	}
	g.builder.NewBr(condBlock) // Loop back to condition

	// Continue with end block
	g.builder = endBlock
	return nil, false, nil
}

// convertType converts ALaS type to LLVM type.
func (g *LLVMCodegen) convertType(alasType string) (types.Type, error) {
	switch alasType {
	case ast.TypeInt:
		return types.I64, nil
	case ast.TypeFloat:
		return types.Double, nil
	case ast.TypeBool:
		return types.I1, nil
	case ast.TypeString:
		// For now, represent strings as i8* (simplified)
		return types.NewPointer(types.I8), nil
	case ast.TypeArray:
		// Represent arrays as a struct with pointer and length
		// struct { i8* data, i64 length }
		return types.NewStruct(types.NewPointer(types.I8), types.I64), nil
	case ast.TypeMap:
		// Represent maps as a simple pointer (simplified implementation)
		// In a real implementation, this would be a hash table structure
		return types.NewPointer(types.I8), nil
	case ast.TypeVoid, "":
		return types.Void, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", alasType)
	}
}

// getZeroValue returns the zero value for a given LLVM type.
func (g *LLVMCodegen) getZeroValue(t types.Type) value.Value {
	switch t {
	case types.I1:
		return constant.NewInt(types.I1, 0)
	case types.I64:
		return constant.NewInt(types.I64, 0)
	case types.Double:
		return constant.NewFloat(types.Double, 0.0)
	default:
		// For pointer types, use null
		if ptr, ok := t.(*types.PointerType); ok {
			return constant.NewNull(ptr)
		}
		// For struct types (arrays), create zero struct
		if structType, ok := t.(*types.StructType); ok {
			fields := make([]constant.Constant, len(structType.Fields))
			for i, fieldType := range structType.Fields {
				fields[i] = g.getZeroValue(fieldType).(constant.Constant)
			}
			return constant.NewStruct(structType, fields...)
		}
		return constant.NewInt(types.I64, 0) // Default fallback
	}
}

// generateArrayLiteral generates LLVM IR for array literals.
// For simplicity, this creates a simplified array representation
func (g *LLVMCodegen) generateArrayLiteral(expr *ast.Expression) (value.Value, error) {
	// For now, create a simplified array structure
	// This is a basic implementation - a full implementation would allocate memory
	// and properly initialize the array elements
	
	arrayType, _ := g.convertType(ast.TypeArray)
	structType := arrayType.(*types.StructType)
	
	// Create null pointer for data (simplified)
	dataPtr := constant.NewNull(types.NewPointer(types.I8))
	// Set length
	length := constant.NewInt(types.I64, int64(len(expr.Elements)))
	
	// Create struct with data pointer and length
	fields := []constant.Constant{dataPtr, length}
	return constant.NewStruct(structType, fields...), nil
}

// generateMapLiteral generates LLVM IR for map literals.
// For simplicity, this creates a null pointer (maps would need complex runtime support)
func (g *LLVMCodegen) generateMapLiteral(expr *ast.Expression) (value.Value, error) {
	// For now, just return a null pointer
	// A full implementation would need runtime support for hash tables
	mapType, _ := g.convertType(ast.TypeMap)
	return constant.NewNull(mapType.(*types.PointerType)), nil
}

// generateIndexAccess generates LLVM IR for array/map indexing.
// This is a simplified implementation that would need runtime support
func (g *LLVMCodegen) generateIndexAccess(expr *ast.Expression) (value.Value, error) {
	// Generate object and index expressions
	_, err := g.generateExpression(expr.Object)
	if err != nil {
		return nil, err
	}
	
	_, err = g.generateExpression(expr.Index)
	if err != nil {
		return nil, err
	}
	
	// For now, return a placeholder zero value
	// A full implementation would need to:
	// 1. Check if object is array or map
	// 2. Generate appropriate indexing logic
	// 3. Handle bounds checking for arrays
	// 4. Handle key lookup for maps
	return constant.NewInt(types.I64, 0), nil
}
