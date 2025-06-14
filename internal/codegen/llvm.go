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
	module            *ir.Module
	builder           *ir.Block
	functions         map[string]*ir.Func
	variables         map[string]value.Value
	gcFunctions       map[string]*ir.Func
	externalFunctions map[string]*ir.Func // External functions from other modules
	builtinFunctions  map[string]*ir.Func // Builtin standard library functions
}

// NewLLVMCodegen creates a new LLVM code generator.
func NewLLVMCodegen() *LLVMCodegen {
	g := &LLVMCodegen{
		module:            ir.NewModule(),
		functions:         make(map[string]*ir.Func),
		variables:         make(map[string]value.Value),
		gcFunctions:       make(map[string]*ir.Func),
		externalFunctions: make(map[string]*ir.Func),
		builtinFunctions:  make(map[string]*ir.Func),
	}
	g.declareGCFunctions()
	g.declareBuiltinFunctions()
	return g
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
	// Convert return type
	returnType, err := g.convertType(fn.Returns)
	if err != nil {
		return fmt.Errorf("invalid return type %s: %v", fn.Returns, err)
	}

	// Create function with return type only
	llvmFunc := g.module.NewFunc(fn.Name, returnType)

	// Add parameters
	for _, param := range fn.Params {
		paramType, err := g.convertType(param.Type)
		if err != nil {
			return fmt.Errorf("invalid parameter type %s: %v", param.Type, err)
		}
		llvmParam := ir.NewParam(param.Name, paramType)
		llvmFunc.Params = append(llvmFunc.Params, llvmParam)
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
			// Create alloca for the parameter
			paramAlloca := g.builder.NewAlloca(llvmFunc.Params[i].Type())
			paramAlloca.SetName(param.Name + "_ptr")
			
			// Store the parameter value into the alloca
			g.builder.NewStore(llvmFunc.Params[i], paramAlloca)
			
			// Store the alloca in variables map
			g.variables[param.Name] = paramAlloca
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
		
		// For LLVM IR, we need to store variables properly with alloca + store
		// Allocate memory for the variable
		varAlloca := g.builder.NewAlloca(val.Type())
		varAlloca.SetName(stmt.Target + "_ptr")
		
		// Store the value
		g.builder.NewStore(val, varAlloca)
		
		// Keep track of the alloca for later loads
		g.variables[stmt.Target] = varAlloca
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

	case ast.StmtFor:
		return g.generateFor(stmt)

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
		varAlloca, ok := g.variables[expr.Name]
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", expr.Name)
		}
		
		// Load the value from the alloca
		ptrType, isPtr := varAlloca.Type().(*types.PointerType)
		if !isPtr {
			return nil, fmt.Errorf("variable %s is not a pointer type", expr.Name)
		}
		
		loadedVal := g.builder.NewLoad(ptrType.ElemType, varAlloca)
		loadedVal.SetName(expr.Name + "_val")
		return loadedVal, nil

	case ast.ExprBinary:
		return g.generateBinary(expr)

	case ast.ExprUnary:
		return g.generateUnary(expr)

	case ast.ExprCall:
		return g.generateCall(expr)

	case ast.ExprModuleCall:
		return g.generateModuleCall(expr)

	case ast.ExprArrayLit:
		return g.generateArrayLiteral(expr)

	case ast.ExprMapLit:
		return g.generateMapLiteral(expr)

	case ast.ExprIndex:
		return g.generateIndexAccess(expr)

	case ast.ExprBuiltin:
		return g.generateBuiltinCall(expr)

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

// generateFor generates LLVM IR for for loops.
// ALaS for loops are similar to while loops with a condition and body.
// Traditional for(init; cond; update) can be desugared to init + while(cond) { body; update }
func (g *LLVMCodegen) generateFor(stmt *ast.Statement) (value.Value, bool, error) {
	currentFunc := g.builder.Parent
	condBlock := currentFunc.NewBlock("for.cond")
	bodyBlock := currentFunc.NewBlock("for.body")
	endBlock := currentFunc.NewBlock("for.end")

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
func (g *LLVMCodegen) generateArrayLiteral(expr *ast.Expression) (value.Value, error) {
	// Generate all element expressions first
	elementCount := int64(len(expr.Elements))
	elements := make([]value.Value, elementCount)
	
	// Determine element type from first element (assume homogeneous arrays)
	var elemType types.Type
	if elementCount > 0 {
		firstElem, err := g.generateExpression(&expr.Elements[0])
		if err != nil {
			return nil, err
		}
		elements[0] = firstElem
		elemType = firstElem.Type()
		
		// Generate remaining elements
		for i := 1; i < int(elementCount); i++ {
			elem, err := g.generateExpression(&expr.Elements[i])
			if err != nil {
				return nil, err
			}
			elements[i] = elem
		}
	} else {
		// Empty array, default to i64
		elemType = types.I64
	}
	
	// Allocate array on stack
	arrayAlloca := g.builder.NewAlloca(types.NewArray(uint64(elementCount), elemType))
	arrayAlloca.SetName("array_literal")
	
	// Store elements
	for i, elem := range elements {
		// Get pointer to element
		elemPtr := g.builder.NewGetElementPtr(
			types.NewArray(uint64(elementCount), elemType),
			arrayAlloca,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(i)),
		)
		// Store element value
		g.builder.NewStore(elem, elemPtr)
	}
	
	// Create array struct: {data*, length}
	arrayType, _ := g.convertType(ast.TypeArray)
	structType := arrayType.(*types.StructType)
	
	// Allocate struct on stack
	structAlloca := g.builder.NewAlloca(structType)
	structAlloca.SetName("array_struct")
	
	// Store data pointer
	dataFieldPtr := g.builder.NewGetElementPtr(
		structType,
		structAlloca,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 0),
	)
	// Cast array pointer to i8*
	castedPtr := g.builder.NewBitCast(arrayAlloca, types.NewPointer(types.I8))
	g.builder.NewStore(castedPtr, dataFieldPtr)
	
	// Store length
	lengthFieldPtr := g.builder.NewGetElementPtr(
		structType,
		structAlloca,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 1),
	)
	g.builder.NewStore(constant.NewInt(types.I64, elementCount), lengthFieldPtr)
	
	// Load and return the struct
	return g.builder.NewLoad(structType, structAlloca), nil
}

// generateMapLiteral generates LLVM IR for map literals.
// This is a simplified implementation - a full implementation would need a proper hash table.
func (g *LLVMCodegen) generateMapLiteral(expr *ast.Expression) (value.Value, error) {
	// For a basic implementation, we'll create a simple linear array of key-value pairs
	// A real implementation would use a hash table structure
	
	pairCount := len(expr.Pairs)
	
	// Define a key-value pair struct type {i64 key, i64 value}
	// In a real implementation, this would be more generic
	kvPairType := types.NewStruct(types.I64, types.I64)
	
	// Allocate array of pairs
	pairsAlloca := g.builder.NewAlloca(types.NewArray(uint64(pairCount), kvPairType))
	pairsAlloca.SetName("map_pairs")
	
	// Store key-value pairs
	for i, pair := range expr.Pairs {
		// Generate key and value
		key, err := g.generateExpression(&pair.Key)
		if err != nil {
			return nil, err
		}
		val, err := g.generateExpression(&pair.Value)
		if err != nil {
			return nil, err
		}
		
		// Get pointer to pair
		pairPtr := g.builder.NewGetElementPtr(
			types.NewArray(uint64(pairCount), kvPairType),
			pairsAlloca,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(i)),
		)
		
		// Store key
		keyPtr := g.builder.NewGetElementPtr(
			kvPairType,
			pairPtr,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0),
		)
		g.builder.NewStore(key, keyPtr)
		
		// Store value
		valPtr := g.builder.NewGetElementPtr(
			kvPairType,
			pairPtr,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 1),
		)
		g.builder.NewStore(val, valPtr)
	}
	
	// For now, just return the pointer to the pairs array
	// A real map would have a more complex structure with hash buckets, etc.
	mapType, _ := g.convertType(ast.TypeMap)
	return g.builder.NewBitCast(pairsAlloca, mapType.(*types.PointerType)), nil
}

// generateIndexAccess generates LLVM IR for array/map indexing.
func (g *LLVMCodegen) generateIndexAccess(expr *ast.Expression) (value.Value, error) {
	// Generate object expression
	obj, err := g.generateExpression(expr.Object)
	if err != nil {
		return nil, err
	}

	// Generate index expression
	index, err := g.generateExpression(expr.Index)
	if err != nil {
		return nil, err
	}

	// Check if object is an array struct
	objType := obj.Type()
	if structType, ok := objType.(*types.StructType); ok && structType.Name() == "array_struct" {
		// This is explicitly identified as our array struct
		// Extract data pointer
		dataPtr := g.builder.NewExtractValue(obj, 0)
		dataPtr.SetName("array_data_ptr")
		
		// TODO: Add bounds checking here using the length field
		// length := g.builder.NewExtractValue(obj, 1)
		
		// For now, assume the array contains i64 elements (should be determined from context)
		// Cast i8* back to proper element type pointer
		elemType := types.I64
		typedPtr := g.builder.NewBitCast(dataPtr, types.NewPointer(elemType))
		
		// Calculate element address
		elemPtr := g.builder.NewGetElementPtr(elemType, typedPtr, index)
		elemPtr.SetName("elem_ptr")
		
		// Load and return element value
		return g.builder.NewLoad(elemType, elemPtr), nil
	}
	
	// For maps or other types, return placeholder for now
	return constant.NewInt(types.I64, 0), nil
}

// generateModuleCall generates LLVM IR for module function calls.
func (g *LLVMCodegen) generateModuleCall(expr *ast.Expression) (value.Value, error) {
	// Create qualified function name: module_name__function_name
	qualifiedName := fmt.Sprintf("%s__%s", expr.Module, expr.Name)

	// Look up the external function
	externalFunc, exists := g.externalFunctions[qualifiedName]
	if !exists {
		return nil, fmt.Errorf("external function %s not declared", qualifiedName)
	}

	// Generate arguments
	args := make([]value.Value, len(expr.Args))
	for i, arg := range expr.Args {
		argVal, err := g.generateExpression(&arg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate argument %d for %s: %v", i, qualifiedName, err)
		}
		args[i] = argVal
	}

	// Generate the function call
	return g.builder.NewCall(externalFunc, args...), nil
}

// DeclareExternalFunction declares an external function from another module.
func (g *LLVMCodegen) DeclareExternalFunction(moduleName, functionName string, paramTypes []types.Type, returnType types.Type) (*ir.Func, error) {
	qualifiedName := fmt.Sprintf("%s__%s", moduleName, functionName)

	// Check if already declared
	if existing, exists := g.externalFunctions[qualifiedName]; exists {
		return existing, nil
	}

	// Create function signature
	sig := types.NewFunc(returnType, paramTypes...)

	// Declare the function as external in this module
	externalFunc := g.module.NewFunc(qualifiedName, sig)

	// Store the external function
	g.externalFunctions[qualifiedName] = externalFunc

	return externalFunc, nil
}

// declareGCFunctions declares external GC runtime functions for LLVM IR.
func (g *LLVMCodegen) declareGCFunctions() {
	// GC object pointer type - representing *GCObject
	gcObjectPtrType := types.NewPointer(types.I8)

	// Object ID type - representing ObjectID (int64)
	objectIDType := types.I64

	// Value pointer type - representing *Value
	valuePtrType := types.NewPointer(types.I8)

	// Array allocation: alas_gc_alloc_array(values *Value, count i64) -> *GCObject
	arrayAllocFunc := g.module.NewFunc("alas_gc_alloc_array", gcObjectPtrType)
	arrayAllocFunc.Params = append(arrayAllocFunc.Params, 
		ir.NewParam("", valuePtrType),
		ir.NewParam("", types.I64))
	g.gcFunctions["alas_gc_alloc_array"] = arrayAllocFunc

	// Map allocation: alas_gc_alloc_map(pairs *MapPair, count i64) -> *GCObject
	mapAllocFunc := g.module.NewFunc("alas_gc_alloc_map", gcObjectPtrType)
	mapAllocFunc.Params = append(mapAllocFunc.Params,
		ir.NewParam("", valuePtrType),
		ir.NewParam("", types.I64))
	g.gcFunctions["alas_gc_alloc_map"] = mapAllocFunc

	// Reference counting: alas_gc_retain(id ObjectID) -> void
	retainFunc := g.module.NewFunc("alas_gc_retain", types.Void)
	retainFunc.Params = append(retainFunc.Params, ir.NewParam("", objectIDType))
	g.gcFunctions["alas_gc_retain"] = retainFunc

	// Reference counting: alas_gc_release(id ObjectID) -> void
	releaseFunc := g.module.NewFunc("alas_gc_release", types.Void)
	releaseFunc.Params = append(releaseFunc.Params, ir.NewParam("", objectIDType))
	g.gcFunctions["alas_gc_release"] = releaseFunc

	// Array access: alas_gc_array_get(obj *GCObject, index i64) -> *Value
	arrayGetFunc := g.module.NewFunc("alas_gc_array_get", valuePtrType)
	arrayGetFunc.Params = append(arrayGetFunc.Params,
		ir.NewParam("", gcObjectPtrType),
		ir.NewParam("", types.I64))
	g.gcFunctions["alas_gc_array_get"] = arrayGetFunc

	// Map access: alas_gc_map_get(obj *GCObject, key *Value) -> *Value
	mapGetFunc := g.module.NewFunc("alas_gc_map_get", valuePtrType)
	mapGetFunc.Params = append(mapGetFunc.Params,
		ir.NewParam("", gcObjectPtrType),
		ir.NewParam("", valuePtrType))
	g.gcFunctions["alas_gc_map_get"] = mapGetFunc

	// Force GC: alas_gc_run() -> void
	runGCFunc := g.module.NewFunc("alas_gc_run", types.Void)
	g.gcFunctions["alas_gc_run"] = runGCFunc
}

// declareBuiltinFunctions declares external builtin standard library functions.
func (g *LLVMCodegen) declareBuiltinFunctions() {
	// For C compatibility, use simple i8* (void*) for CValue parameters
	// This matches the actual C function signatures generated by CGO
	cvalueArgType := types.NewPointer(types.I8) // void* for CValue*
	
	// For functions that return CValue, we'll also use a simple i8* for now
	// In a complete implementation, we'd use the actual struct type
	cvalueReturnType := types.NewPointer(types.I8) // void* for CValue return

	// I/O functions
	// void alas_builtin_io_print(void* val)
	printFunc := g.module.NewFunc("alas_builtin_io_print", types.Void)
	printFunc.Params = append(printFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["io.print"] = printFunc

	// Math functions  
	// void* alas_builtin_math_sqrt(void* val) - simplified for C compatibility
	sqrtFunc := g.module.NewFunc("alas_builtin_math_sqrt", cvalueReturnType)
	sqrtFunc.Params = append(sqrtFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["math.sqrt"] = sqrtFunc
	
	absFunc := g.module.NewFunc("alas_builtin_math_abs", cvalueReturnType)
	absFunc.Params = append(absFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["math.abs"] = absFunc

	// Collections functions
	// void* alas_builtin_collections_length(void* val)
	lengthFunc := g.module.NewFunc("alas_builtin_collections_length", cvalueReturnType)
	lengthFunc.Params = append(lengthFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["collections.length"] = lengthFunc

	// String functions
	// void* alas_builtin_string_toUpper(void* val)
	toUpperFunc := g.module.NewFunc("alas_builtin_string_toUpper", cvalueReturnType)
	toUpperFunc.Params = append(toUpperFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.toUpper"] = toUpperFunc

	// Type functions
	// void* alas_builtin_type_typeOf(void* val)
	typeOfFunc := g.module.NewFunc("alas_builtin_type_typeOf", cvalueReturnType)
	typeOfFunc.Params = append(typeOfFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["type.typeOf"] = typeOfFunc

	// TODO: Add more builtin functions as needed
}

// generateBuiltinCall generates LLVM IR for builtin function calls.
func (g *LLVMCodegen) generateBuiltinCall(expr *ast.Expression) (value.Value, error) {
	// Look up the builtin function
	builtinFunc, exists := g.builtinFunctions[expr.Name]
	if !exists {
		return nil, fmt.Errorf("unknown builtin function: %s", expr.Name)
	}

	// For now, we'll handle a simplified case with single arguments
	// A full implementation would handle multiple arguments and complex types
	
	if expr.Name == "io.print" {
		// Special case for io.print which returns void
		if len(expr.Args) != 1 {
			return nil, fmt.Errorf("io.print expects 1 argument, got %d", len(expr.Args))
		}

		// Generate the argument
		argVal, err := g.generateExpression(&expr.Args[0])
		if err != nil {
			return nil, err
		}

		// Convert to CValue
		cval := g.convertToCValue(argVal)
		
		// Call the function
		g.builder.NewCall(builtinFunc, cval)
		return nil, nil // io.print returns void
	}

	// For functions that return values
	if len(expr.Args) != 1 {
		return nil, fmt.Errorf("%s expects 1 argument, got %d", expr.Name, len(expr.Args))
	}

	// Generate the argument
	argVal, err := g.generateExpression(&expr.Args[0])
	if err != nil {
		return nil, err
	}

	// Convert to CValue - check if it's already a CValue* (i8*)
	var cval value.Value
	
	// Check if this is already an i8* (CValue pointer)
	if ptrType, isPtr := argVal.Type().(*types.PointerType); isPtr {
		if ptrType.ElemType.Equal(types.I8) {
			// Already a CValue* (i8*), use directly
			cval = argVal
		} else {
			// Other pointer type, convert to CValue
			cval = g.convertToCValue(argVal)
		}
	} else {
		// Not a pointer type, convert to CValue
		cval = g.convertToCValue(argVal)
	}
	
	// Debug: Check the types involved in the call
	// builtinFunc should be a function, result should be a struct
	
	// Call the function and get result
	result := g.builder.NewCall(builtinFunc, cval)
	
	// Known issue: The LLVM Go library has a type handling issue where
	// NewCall().Type() returns *types.FuncType instead of the expected return type
	// However, the generated LLVM IR is correct and shows proper function calls
	// For now, we work around this by using a context-aware placeholder
	
	// For functions that return values, return the raw CValue* so it can be reused
	// For io.print (void), we don't need to return anything meaningful
	if expr.Name == "io.print" {
		return constant.NewInt(types.I32, 0), nil // Dummy return for void functions
	}
	
	// The LLVM Go library has a type issue where result.Type() returns the wrong type
	// But the actual LLVM IR is correct (i8*). We need to cast to the correct type.
	
	// Cast the result to i8* to fix the type issue
	if _, isFuncType := result.Type().(*types.FuncType); isFuncType {
		// The LLVM IR is correct but the Go type is wrong, create a bitcast to fix it
		correctResult := g.builder.NewBitCast(result, types.NewPointer(types.I8))
		return correctResult, nil
	}
	
	// Return the raw CValue* result for reuse in variables and other calls
	return result, nil
}

// convertToCValue converts an LLVM value to a CValue pointer.
func (g *LLVMCodegen) convertToCValue(val value.Value) value.Value {
	// Check if this is already a CValue* (i8*) from a previous builtin function call
	if ptrType, isPtr := val.Type().(*types.PointerType); isPtr {
		if ptrType.ElemType.Equal(types.I8) {
			// This is already a CValue*, return it directly
			return val
		}
	}
	
	// Create CValue type directly to match our CGO definition
	cvalueType := types.NewStruct(
		types.I32, // type field
		types.NewStruct( // data union (simplified as struct with all fields)
			types.I64,                   // int_val
			types.Double,                // float_val
			types.NewPointer(types.I8),  // string_val
			types.NewPointer(types.I8),  // array_val
			types.NewPointer(types.I8),  // map_val
		),
	)
	
	// Allocate space for CValue on stack
	cval := g.builder.NewAlloca(cvalueType)
	
	// Get pointers to type field and data union
	typeField := g.builder.NewGetElementPtr(cvalueType, cval, 
		constant.NewInt(types.I32, 0), 
		constant.NewInt(types.I32, 0))
	dataField := g.builder.NewGetElementPtr(cvalueType, cval,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 1))
	
	// Determine value type and store
	valType := val.Type()
	switch {
	case valType.Equal(types.I64):
		// Integer
		g.builder.NewStore(constant.NewInt(types.I32, 0), typeField) // CValueTypeInt
		intField := g.builder.NewGetElementPtr(dataField.Type().(*types.PointerType).ElemType, dataField,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0))
		g.builder.NewStore(val, intField)
		
	case valType.Equal(types.Double):
		// Float
		g.builder.NewStore(constant.NewInt(types.I32, 1), typeField) // CValueTypeFloat
		floatField := g.builder.NewGetElementPtr(dataField.Type().(*types.PointerType).ElemType, dataField,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 1))
		g.builder.NewStore(val, floatField)
		
	case valType.Equal(types.I1):
		// Boolean
		g.builder.NewStore(constant.NewInt(types.I32, 3), typeField) // CValueTypeBool
		intField := g.builder.NewGetElementPtr(dataField.Type().(*types.PointerType).ElemType, dataField,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0))
		// Extend bool to i64
		extended := g.builder.NewZExt(val, types.I64)
		g.builder.NewStore(extended, intField)
		
	case valType.Equal(types.NewPointer(types.I8)):
		// String
		g.builder.NewStore(constant.NewInt(types.I32, 2), typeField) // CValueTypeString
		stringField := g.builder.NewGetElementPtr(dataField.Type().(*types.PointerType).ElemType, dataField,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 2))
		g.builder.NewStore(val, stringField)
		
	default:
		// Void or unsupported
		g.builder.NewStore(constant.NewInt(types.I32, 6), typeField) // CValueTypeVoid
	}
	
	// Cast to i8* for C compatibility
	return g.builder.NewBitCast(cval, types.NewPointer(types.I8))
}

// convertFromCValue converts a CValue to an LLVM value.
func (g *LLVMCodegen) convertFromCValue(cval value.Value) (value.Value, error) {
	// This function handles extracting values from CValue structs returned by builtin functions
	
	// Define CValue struct type to match our CGO definition
	cvalueType := types.NewStruct(
		types.I32, // type field
		types.NewStruct( // data union
			types.I64,                   // int_val
			types.Double,                // float_val
			types.NewPointer(types.I8),  // string_val
			types.NewPointer(types.I8),  // array_val
			types.NewPointer(types.I8),  // map_val
		),
	)
	
	// Check if we have a pointer type (i8* from builtin function call)
	if ptrType, isPtr := cval.Type().(*types.PointerType); isPtr {
		if ptrType.ElemType.Equal(types.I8) {
			// This is an i8* pointer to CValue, cast it to CValue* and load
			cvaluePtr := g.builder.NewBitCast(cval, types.NewPointer(cvalueType))
			cvalueStruct := g.builder.NewLoad(cvalueType, cvaluePtr)
			
			// Extract the type field to determine what kind of value this is
			_ = g.builder.NewExtractValue(cvalueStruct, 0) // typeField for future type switching
			dataUnion := g.builder.NewExtractValue(cvalueStruct, 1)
			
			// For now, assume it's a float value (type 1) and extract the float field
			// TODO: Add proper type switching based on typeField
			floatVal := g.builder.NewExtractValue(dataUnion, 1)
			return floatVal, nil
		}
	}
	
	// Check if we have the expected struct type directly
	if _, isStruct := cval.Type().(*types.StructType); isStruct {
		// We have a proper struct! Extract the value based on the type field
		dataUnion := g.builder.NewExtractValue(cval, 1)
		floatVal := g.builder.NewExtractValue(dataUnion, 1)
		return floatVal, nil
	}
	
	// Fallback for unexpected types
	return constant.NewFloat(types.Double, 0.0), nil
}
