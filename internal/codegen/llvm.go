package codegen

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"

	"encoding/json"
	"github.com/dshills/alas/internal/ast"
	"os"
	"path/filepath"
)

const (
	// DynamicMapType represents a dynamically-typed map variable
	DynamicMapType = "_dynamic_map"
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
	moduleLoader      ModuleResolver
	customTypes       map[string]*ast.TypeDefinition // Custom type definitions
	structTypes       map[string]types.Type          // LLVM types for custom types
	fieldIndices      map[string]map[string]int      // type name -> field name -> index
	variableTypes     map[string]string              // variable name -> ALaS type name
	currentFunction   *ast.Function                  // Current function being generated
	astFunctions      map[string]*ast.Function       // AST function definitions
	loadedModules     map[string]*ast.Module         // Cache of loaded modules
	compiledModules   map[string]*ir.Module          // Cache of compiled modules
}

// ModuleResolver interface for loading modules.
type ModuleResolver interface {
	LoadModuleByName(name string) (*ast.Module, error)
}

// FileModuleLoader loads modules from the filesystem.
type FileModuleLoader struct {
	searchPaths []string
}

// NewFileModuleLoader creates a new file-based module loader.
func NewFileModuleLoader(searchPaths []string) *FileModuleLoader {
	return &FileModuleLoader{
		searchPaths: searchPaths,
	}
}

// LoadModuleByName loads a module by name from the filesystem.
func (l *FileModuleLoader) LoadModuleByName(name string) (*ast.Module, error) {
	// Try original name first
	for _, searchPath := range l.searchPaths {
		fileName := filepath.Join(searchPath, name+".alas.json")
		if data, err := os.ReadFile(fileName); err == nil {
			var module ast.Module
			if err := json.Unmarshal(data, &module); err != nil {
				return nil, fmt.Errorf("failed to parse module %s: %v", name, err)
			}
			return &module, nil
		}
	}
	
	// For stdlib modules, try without "std." prefix
	if len(name) > 4 && name[:4] == "std." {
		simpleName := name[4:] // Remove "std." prefix
		for _, searchPath := range l.searchPaths {
			fileName := filepath.Join(searchPath, simpleName+".alas.json")
			if data, err := os.ReadFile(fileName); err == nil {
				var module ast.Module
				if err := json.Unmarshal(data, &module); err != nil {
					return nil, fmt.Errorf("failed to parse module %s: %v", name, err)
				}
				return &module, nil
			}
		}
	}
	
	return nil, fmt.Errorf("module %s not found in search paths", name)
}

// NewLLVMCodegen creates a new LLVM code generator.
func NewLLVMCodegen() *LLVMCodegen {
	// Create with default module loader
	searchPaths := []string{".", "examples/modules", "../examples/modules", "stdlib"}
	return NewLLVMCodegenWithLoader(NewFileModuleLoader(searchPaths))
}

// NewLLVMCodegenWithLoader creates a new LLVM code generator with a custom module loader.
func NewLLVMCodegenWithLoader(loader ModuleResolver) *LLVMCodegen {
	g := &LLVMCodegen{
		module:            ir.NewModule(),
		functions:         make(map[string]*ir.Func),
		variables:         make(map[string]value.Value),
		gcFunctions:       make(map[string]*ir.Func),
		externalFunctions: make(map[string]*ir.Func),
		builtinFunctions:  make(map[string]*ir.Func),
		moduleLoader:      loader,
		customTypes:       make(map[string]*ast.TypeDefinition),
		structTypes:       make(map[string]types.Type),
		fieldIndices:      make(map[string]map[string]int),
		variableTypes:     make(map[string]string),
		currentFunction:   nil,
		astFunctions:      make(map[string]*ast.Function),
		loadedModules:     make(map[string]*ast.Module),
		compiledModules:   make(map[string]*ir.Module),
	}
	g.declareGCFunctions()
	g.declareErrorHandlingFunctions()
	g.declareBuiltinFunctions()
	return g
}

// declareCustomType declares a custom type in LLVM IR.
func (g *LLVMCodegen) declareCustomType(typeDef *ast.TypeDefinition) error {
	// Skip if type definition is incomplete or uses unsupported format
	if typeDef.Definition.Kind == "" {
		// This handles old schema formats or incomplete type definitions
		// For now, we skip them to maintain backward compatibility
		return nil
	}
	
	switch typeDef.Definition.Kind {
	case ast.TypeKindStruct:
		// Create LLVM struct type
		var fieldTypes []types.Type
		fieldIndexMap := make(map[string]int)

		for i, field := range typeDef.Definition.Fields {
			fieldType, err := g.convertType(field.Type)
			if err != nil {
				return fmt.Errorf("invalid field type %s: %v", field.Type, err)
			}
			fieldTypes = append(fieldTypes, fieldType)
			fieldIndexMap[field.Name] = i
		}

		// Create named struct type
		structType := types.NewStruct(fieldTypes...)
		g.structTypes[typeDef.Name] = structType
		g.fieldIndices[typeDef.Name] = fieldIndexMap

	case ast.TypeKindEnum:
		// Enums are represented as i32 (could also use string pointers)
		// For now, we'll use i32 for enum values
		g.structTypes[typeDef.Name] = types.I32

	default:
		return fmt.Errorf("unknown type kind: %s", typeDef.Definition.Kind)
	}

	return nil
}

// GenerateModule generates LLVM IR for an entire ALaS module.
func (g *LLVMCodegen) GenerateModule(module *ast.Module) (*ir.Module, error) {
	g.module.SourceFilename = module.Name + ".alas"

	// Process custom types first
	for idx := range module.Types {
		typeDef := &module.Types[idx]
		g.customTypes[typeDef.Name] = typeDef
		if err := g.declareCustomType(typeDef); err != nil {
			return nil, fmt.Errorf("failed to declare type %s: %v", typeDef.Name, err)
		}
	}

	// Resolve module dependencies recursively
	visited := make(map[string]bool)
	for _, importName := range module.Imports {
		if err := g.resolveModuleDependencies(importName, visited); err != nil {
			return nil, fmt.Errorf("failed to resolve dependencies: %v", err)
		}
	}

	// Handle imports - declare external functions from imported modules
	if err := g.declareImportedFunctions(module.Imports); err != nil {
		return nil, fmt.Errorf("failed to declare imported functions: %v", err)
	}

	// First pass: declare all functions
	for i := range module.Functions {
		fn := &module.Functions[i]
		g.astFunctions[fn.Name] = fn
		if err := g.declareFunction(fn); err != nil {
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

	// Set current function
	g.currentFunction = fn

	// Create new variable scope for this function
	oldVars := g.variables
	g.variables = make(map[string]value.Value)

	// Create new type tracking scope for this function
	oldVarTypes := g.variableTypes
	g.variableTypes = make(map[string]string)

	// Add parameters to variable scope
	for i, param := range fn.Params {
		if i < len(llvmFunc.Params) {
			// Create alloca for the parameter
			paramAlloca := g.builder.NewAlloca(llvmFunc.Params[i].Type())
			paramAlloca.SetName(param.Name + "_ptr")

			// Track parameter type
			g.variableTypes[param.Name] = param.Type

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
	g.variableTypes = oldVarTypes
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

		// Check if variable already has an alloca
		varAlloca, exists := g.variables[stmt.Target]
		if !exists {
			// First assignment - allocate memory for the variable
			newAlloca := g.builder.NewAlloca(val.Type())
			newAlloca.SetName(stmt.Target + "_ptr")

			// Keep track of the alloca for later loads
			varAlloca = newAlloca
			g.variables[stmt.Target] = varAlloca
		}

		// Store the value (works for both new and existing allocas)
		g.builder.NewStore(val, varAlloca)

		// Try to infer and track variable type
		g.inferVariableType(stmt.Target, stmt.Value)

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
		// Don't set a name - let LLVM auto-generate unique names to avoid SSA conflicts
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

	case ast.ExprField:
		return g.generateFieldAccess(expr)

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
		// Add division by zero check for integer division
		if !isFloat {
			g.generateDivisionByZeroCheck(right)
		}
		if isFloat {
			return g.builder.NewFDiv(left, right), nil
		}
		return g.builder.NewSDiv(left, right), nil

	case ast.OpMod:
		// Add division by zero check for modulo operation
		if !isFloat {
			g.generateDivisionByZeroCheck(right)
		}
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
	// Support both Operand (spec-compliant) and Right (backward compatibility)
	var operandExpr *ast.Expression
	if expr.Operand != nil {
		operandExpr = expr.Operand
	} else if expr.Right != nil {
		operandExpr = expr.Right
	} else {
		return nil, fmt.Errorf("unary expression missing operand")
	}

	operand, err := g.generateExpression(operandExpr)
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

// generateLoop generates LLVM IR for loop statements (while and for).
// Both while and for loops in ALaS have the same structure: condition and body.
func (g *LLVMCodegen) generateLoop(stmt *ast.Statement, loopType string) (value.Value, bool, error) {
	currentFunc := g.builder.Parent
	condBlock := currentFunc.NewBlock(loopType + ".cond")
	bodyBlock := currentFunc.NewBlock(loopType + ".body")
	endBlock := currentFunc.NewBlock(loopType + ".end")

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

// generateWhile generates LLVM IR for while loops.
func (g *LLVMCodegen) generateWhile(stmt *ast.Statement) (value.Value, bool, error) {
	return g.generateLoop(stmt, "while")
}

// generateFor generates LLVM IR for for loops.
// ALaS for loops are similar to while loops with a condition and body.
// Traditional for(init; cond; update) can be desugared to init + while(cond) { body; update }.
func (g *LLVMCodegen) generateFor(stmt *ast.Statement) (value.Value, bool, error) {
	return g.generateLoop(stmt, "for")
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
	case "any":
		// Represent "any" type as a generic pointer - this allows stdlib functions to accept any type
		// In a real implementation, this would include type information
		return types.NewPointer(types.I8), nil
	case "function":
		// Represent function type as a function pointer - simplified implementation
		// In a real implementation, this would include proper function signatures
		return types.NewPointer(types.I8), nil
	case ast.TypeVoid, "":
		return types.Void, nil
	default:
		// Check if it's a custom type
		if structType, ok := g.structTypes[alasType]; ok {
			return structType, nil
		}
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
	// Safe conversion: elementCount is already validated to be non-negative
	if elementCount < 0 || elementCount > 0x7FFFFFFF {
		return nil, fmt.Errorf("array element count out of valid range: %d", elementCount)
	}
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
	// Check if this should be a struct construction
	if g.currentFunction != nil && g.currentFunction.Returns != "" {
		// Check if the return type is a custom type
		if typeDef, isCustomType := g.customTypes[g.currentFunction.Returns]; isCustomType {
			// Check if it's a struct type
			if typeDef.Definition.Kind == ast.TypeKindStruct {
				if structType, isStruct := g.structTypes[g.currentFunction.Returns]; isStruct {
					if _, ok := structType.(*types.StructType); ok {
						// This is a struct construction
						return g.generateStructConstruction(expr, g.currentFunction.Returns)
					}
				}
			}
		}
	}

	// Regular map literal generation
	pairCount := len(expr.Pairs)

	// Define a key-value pair struct type {i8* key, i8* value}
	// Using i8* (char*) pointers to handle both strings and boxed values
	kvPairType := types.NewStruct(types.I8Ptr, types.I8Ptr)

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

		// Convert key to string pointer if needed
		keyPtr := g.builder.NewGetElementPtr(
			kvPairType,
			pairPtr,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 0),
		)

		// Box key if needed and store it
		keyAsPtr := g.boxToI8Ptr(key, "boxed_key")
		g.builder.NewStore(keyAsPtr, keyPtr)

		// Store value
		valPtr := g.builder.NewGetElementPtr(
			kvPairType,
			pairPtr,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, 1),
		)

		// Box value if needed and store it
		valAsPtr := g.boxToI8Ptr(val, "boxed_value")
		g.builder.NewStore(valAsPtr, valPtr)
	}

	// Create a proper map by calling the runtime map creation function
	if g.builtinFunctions["alas_runtime_map_create"] == nil {
		// Declare map creation function
		mapCreateFunc := g.module.NewFunc("alas_runtime_map_create", types.NewPointer(types.I8))
		mapCreateFunc.Params = append(mapCreateFunc.Params,
			ir.NewParam("pairs", types.NewPointer(types.I8)),
			ir.NewParam("count", types.I64))
		g.builtinFunctions["alas_runtime_map_create"] = mapCreateFunc
	}

	// Cast pairs array to i8* and call runtime function
	pairsPtr := g.builder.NewBitCast(pairsAlloca, types.NewPointer(types.I8))
	mapResult := g.builder.NewCall(g.builtinFunctions["alas_runtime_map_create"],
		pairsPtr, constant.NewInt(types.I64, int64(pairCount)))

	return mapResult, nil
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
	if structType, ok := objType.(*types.StructType); ok && g.isArrayStructType(structType) {
		// This is explicitly identified as our array struct
		// Extract data pointer
		dataPtr := g.builder.NewExtractValue(obj, 0)
		dataPtr.SetName("array_data_ptr")

		// Add bounds checking using the length field
		length := g.builder.NewExtractValue(obj, 1)
		g.generateBoundsCheck(index, length)

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

	// For maps, implement proper map indexing
	if obj.Type().Equal(types.NewPointer(types.I8)) {
		// This could be a map or string (i8* pointer) - determine which based on context
		// For now, we'll assume it's a map. String indexing would need runtime type detection
		return g.generateMapIndexAccess(obj, index)
	}

	// For other types, return placeholder for now
	return constant.NewInt(types.I64, 0), nil
}

// generateStructConstruction generates LLVM IR for constructing a struct from a map literal.
func (g *LLVMCodegen) generateStructConstruction(expr *ast.Expression, typeName string) (value.Value, error) {
	structType, ok := g.structTypes[typeName].(*types.StructType)
	if !ok {
		return nil, fmt.Errorf("type %s is not a struct", typeName)
	}

	fieldIndices := g.fieldIndices[typeName]
	if fieldIndices == nil {
		return nil, fmt.Errorf("no field indices found for struct %s", typeName)
	}

	// Allocate struct on stack
	structAlloca := g.builder.NewAlloca(structType)
	structAlloca.SetName(typeName + "_struct")

	// Initialize all fields to zero first
	for i, fieldType := range structType.Fields {
		fieldPtr := g.builder.NewGetElementPtr(
			structType,
			structAlloca,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(i)),
		)
		fieldPtr.SetName(fmt.Sprintf("%s_init_%d", typeName, i))
		zeroVal := g.getZeroValue(fieldType)
		initStore := g.builder.NewStore(zeroVal, fieldPtr)
		if initStore == nil {
			return nil, fmt.Errorf("failed to create init store for field %d", i)
		}
	}

	// Process each field from the map literal
	for _, pair := range expr.Pairs {
		// Get field name from key
		keyLit, ok := pair.Key.Value.(string)
		if !ok {
			return nil, fmt.Errorf("struct field key must be a string literal")
		}

		// Find field index
		fieldIdx, ok := fieldIndices[keyLit]
		if !ok {
			return nil, fmt.Errorf("unknown field %s in struct %s", keyLit, typeName)
		}

		// Generate value
		fieldVal, err := g.generateExpression(&pair.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to generate value for field %s: %v", keyLit, err)
		}

		// Get pointer to field
		fieldPtr := g.builder.NewGetElementPtr(
			structType,
			structAlloca,
			constant.NewInt(types.I32, 0),
			constant.NewInt(types.I32, int64(fieldIdx)),
		)
		fieldPtr.SetName(fmt.Sprintf("%s.%s_ptr", typeName, keyLit))

		// Store value in field
		store := g.builder.NewStore(fieldVal, fieldPtr)
		if store == nil {
			return nil, fmt.Errorf("failed to create store instruction for field %s", keyLit)
		}
	}

	// Load and return the struct
	loadedStruct := g.builder.NewLoad(structType, structAlloca)
	return loadedStruct, nil
}

// inferVariableType tries to infer the ALaS type of a variable from its value expression.
func (g *LLVMCodegen) inferVariableType(varName string, valueExpr *ast.Expression) {
	switch valueExpr.Type {
	case ast.ExprCall:
		// Check if the called function returns a custom type
		if astFn, ok := g.astFunctions[valueExpr.Name]; ok {
			if _, isCustomType := g.customTypes[astFn.Returns]; isCustomType {
				g.variableTypes[varName] = astFn.Returns
			}
		}
	case ast.ExprMapLit:
		// Try to infer struct type from map literal structure
		// Look for custom types that match the field pattern
		if len(valueExpr.Pairs) > 0 {
			mapFields := make(map[string]bool)
			for _, pair := range valueExpr.Pairs {
				if keyLit, ok := pair.Key.Value.(string); ok {
					mapFields[keyLit] = true
				}
			}

			// Check if any custom struct type matches this field pattern
			for typeName, typeDef := range g.customTypes {
				if typeDef.Definition.Kind == ast.TypeKindStruct {
					allFieldsMatch := len(typeDef.Definition.Fields) == len(mapFields)
					if allFieldsMatch {
						for _, field := range typeDef.Definition.Fields {
							if !mapFields[field.Name] {
								allFieldsMatch = false
								break
							}
						}
						if allFieldsMatch {
							g.variableTypes[varName] = typeName
							return
						}
					}
				}
			}
		}
		// If no perfect match, mark as dynamic map type for field access
		g.variableTypes[varName] = DynamicMapType
	}
}

// generateFieldAccess generates LLVM IR for field access (struct.field).
func (g *LLVMCodegen) generateFieldAccess(expr *ast.Expression) (value.Value, error) {
	// Generate object expression
	obj, err := g.generateExpression(expr.Object)
	if err != nil {
		return nil, err
	}

	// Get the type of the object from variable tracking
	var objTypeName string
	if expr.Object != nil && expr.Object.Type == ast.ExprVariable {
		objTypeName = g.variableTypes[expr.Object.Name]
	}

	// Try to determine if this is a proper struct type
	if objTypeName != "" && objTypeName != DynamicMapType {
		// We know the exact type - handle as struct field access
		fieldIndices, ok := g.fieldIndices[objTypeName]
		if ok {
			fieldIdx, ok := fieldIndices[expr.Field]
			if ok {
				// Extract field value from struct
				if fieldIdx < 0 || fieldIdx > 0xFFFFFFFF {
					return nil, fmt.Errorf("field index out of valid range: %d", fieldIdx)
				}
				return g.builder.NewExtractValue(obj, uint64(fieldIdx)), nil
			}
		}
	}

	// Handle dynamic map field access
	if objTypeName == DynamicMapType {
		return g.generateDynamicFieldAccess(obj, expr.Field)
	}

	// Try to infer from the object's LLVM type
	if structType, ok := obj.Type().(*types.StructType); ok {
		// Look for matching struct type
		for typeName, llvmType := range g.structTypes {
			if llvmType == structType {
				objTypeName = typeName
				fieldIndices, ok := g.fieldIndices[objTypeName]
				if ok {
					fieldIdx, ok := fieldIndices[expr.Field]
					if ok {
						// Extract field value from struct
						if fieldIdx < 0 || fieldIdx > 0xFFFFFFFF {
							return nil, fmt.Errorf("field index out of valid range: %d", fieldIdx)
						}
						return g.builder.NewExtractValue(obj, uint64(fieldIdx)), nil
					}
				}
				break
			}
		}
	}

	// Handle dynamic field access on map-like objects
	// This is for cases where we're accessing fields on map literals as if they were objects
	if obj.Type().Equal(types.NewPointer(types.I8)) {
		// Object is a map (i8* pointer) - generate dynamic field access
		return g.generateDynamicFieldAccess(obj, expr.Field)
	}

	// Check if object is an array struct with dynamic map-like behavior
	if structType, ok := obj.Type().(*types.StructType); ok && len(structType.Fields) == 2 {
		// Check if this looks like our map representation (key-value pairs)
		// For now, implement a simplified lookup that returns a default value
		// In a full implementation, this would search through the key-value pairs
		return g.generateMapFieldLookup(obj, expr.Field)
	}

	return nil, fmt.Errorf("cannot determine type of object for field access on %T", obj.Type())
}

// generateDynamicFieldAccess generates LLVM IR for dynamic field access on map-like objects.
func (g *LLVMCodegen) generateDynamicFieldAccess(mapObj value.Value, fieldName string) (value.Value, error) {
	// For now, we'll implement a simplified version that assumes the map contains
	// the field and returns a placeholder value. In a full implementation, this would:
	// 1. Call a runtime function to lookup the field in the map
	// 2. Handle type conversion between the stored value and expected type
	// 3. Return appropriate error handling for missing fields

	// Declare runtime map field access function if not already declared
	mapGetFieldFunc, exists := g.builtinFunctions["map_get_field"]
	if !exists {
		// Create function signature: map_get_field(map i8*, field i8*) -> i8*
		fieldNameType := types.NewPointer(types.I8) // string
		mapGetFieldFunc = g.module.NewFunc("alas_runtime_map_get_field", types.NewPointer(types.I8))
		mapGetFieldFunc.Params = append(mapGetFieldFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)),
			ir.NewParam("field", fieldNameType))
		g.builtinFunctions["map_get_field"] = mapGetFieldFunc
	}

	// Create string literal for field name
	fieldNameLiteral := g.createStringLiteral(fieldName)

	// Call runtime function to get field value
	result := g.builder.NewCall(mapGetFieldFunc, mapObj, fieldNameLiteral)

	// For now, assume the result needs to be converted to the expected type
	// In this simplified implementation, we'll return the raw pointer
	// and let the caller handle type conversion
	return result, nil
}

// generateMapFieldLookup generates LLVM IR for field lookup in map structures.
func (g *LLVMCodegen) generateMapFieldLookup(mapStruct value.Value, fieldName string) (value.Value, error) {
	// This is a simplified implementation for map field access
	// In a real implementation, this would iterate through key-value pairs
	// and compare keys to find the matching field

	// For now, return a placeholder that indicates the field was "found"
	// This allows the compilation to proceed while we develop the full implementation
	return constant.NewInt(types.I64, 42), nil // Placeholder return value
}

// createStringLiteral creates a string literal constant.
func (g *LLVMCodegen) createStringLiteral(str string) value.Value {
	// Create a global string constant
	charArray := constant.NewCharArrayFromString(str + "\x00")
	globalStr := g.module.NewGlobalDef("", charArray)
	globalStr.Immutable = true

	// Return pointer to the first character of the string
	return g.builder.NewGetElementPtr(charArray.Type(), globalStr,
		constant.NewInt(types.I64, 0), constant.NewInt(types.I64, 0))
}

// isArrayStructType checks if a struct type represents our array structure.
func (g *LLVMCodegen) isArrayStructType(structType *types.StructType) bool {
	// Our array struct has exactly 2 fields: {i8* data, i64 length}
	if len(structType.Fields) != 2 {
		return false
	}

	// First field should be i8* (data pointer)
	firstField := structType.Fields[0]
	if ptrType, ok := firstField.(*types.PointerType); !ok || !ptrType.ElemType.Equal(types.I8) {
		return false
	}

	// Second field should be i64 (length)
	secondField := structType.Fields[1]
	return secondField.Equal(types.I64)
}

// generateBoundsCheck generates LLVM IR for array bounds checking.
func (g *LLVMCodegen) generateBoundsCheck(index, length value.Value) {
	// Use the enhanced bounds checking with error reporting
	g.generateBoundsCheckWithError(index, length, "array")
}

// generateArrayElementAssignment generates LLVM IR for array element assignment.
func (g *LLVMCodegen) generateArrayElementAssignment(arrayObj, index, value value.Value) error {
	// Check if object is an array struct
	objType := arrayObj.Type()
	if structType, ok := objType.(*types.StructType); ok && g.isArrayStructType(structType) {
		// Extract data pointer and length
		dataPtr := g.builder.NewExtractValue(arrayObj, 0)
		length := g.builder.NewExtractValue(arrayObj, 1)

		// Bounds check
		g.generateBoundsCheck(index, length)

		// Determine element type - for now assume i64
		elemType := types.I64
		typedPtr := g.builder.NewBitCast(dataPtr, types.NewPointer(elemType))

		// Calculate element address and store value
		elemPtr := g.builder.NewGetElementPtr(elemType, typedPtr, index)
		g.builder.NewStore(value, elemPtr)

		return nil
	}

	return fmt.Errorf("cannot assign to non-array object")
}

// generateArrayLength generates LLVM IR for getting array length.
func (g *LLVMCodegen) generateArrayLength(arrayObj value.Value) (value.Value, error) {
	// Check if object is an array struct
	objType := arrayObj.Type()
	if structType, ok := objType.(*types.StructType); ok && g.isArrayStructType(structType) {
		// Extract and return length field
		return g.builder.NewExtractValue(arrayObj, 1), nil
	}

	return nil, fmt.Errorf("cannot get length of non-array object")
}

// generateArraySlice generates LLVM IR for array slicing.
func (g *LLVMCodegen) generateArraySlice(arrayObj, start, end value.Value) (value.Value, error) {
	// Check if object is an array struct
	objType := arrayObj.Type()
	if structType, ok := objType.(*types.StructType); ok && g.isArrayStructType(structType) {
		// Extract data pointer and length
		dataPtr := g.builder.NewExtractValue(arrayObj, 0)
		length := g.builder.NewExtractValue(arrayObj, 1)

		// Bounds check for start and end indices
		g.generateBoundsCheck(start, length)
		g.generateBoundsCheck(end, length)

		// Calculate new length
		newLength := g.builder.NewSub(end, start)

		// Calculate offset pointer
		elemType := types.I64
		typedPtr := g.builder.NewBitCast(dataPtr, types.NewPointer(elemType))
		offsetPtr := g.builder.NewGetElementPtr(elemType, typedPtr, start)

		// Create new array struct for the slice
		arrayType, _ := g.convertType(ast.TypeArray)
		structType := arrayType.(*types.StructType)
		structAlloca := g.builder.NewAlloca(structType)

		// Store sliced data pointer (cast back to i8*)
		dataFieldPtr := g.builder.NewGetElementPtr(structType, structAlloca,
			constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0))
		slicedPtr := g.builder.NewBitCast(offsetPtr, types.NewPointer(types.I8))
		g.builder.NewStore(slicedPtr, dataFieldPtr)

		// Store new length
		lengthFieldPtr := g.builder.NewGetElementPtr(structType, structAlloca,
			constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 1))
		g.builder.NewStore(newLength, lengthFieldPtr)

		// Load and return the slice struct
		return g.builder.NewLoad(structType, structAlloca), nil
	}

	return nil, fmt.Errorf("cannot slice non-array object")
}

// generateMapIndexAccess generates LLVM IR for map indexing operations.
func (g *LLVMCodegen) generateMapIndexAccess(mapObj, key value.Value) (value.Value, error) {
	// Declare runtime map get function if not already declared
	mapGetFunc, exists := g.builtinFunctions["alas_runtime_map_get"]
	if !exists {
		// Create function signature: map_get(map i8*, key i8*) -> i8*
		mapGetFunc = g.module.NewFunc("alas_runtime_map_get", types.NewPointer(types.I8))
		mapGetFunc.Params = append(mapGetFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)),
			ir.NewParam("key", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_get"] = mapGetFunc
	}

	// Convert key to i8* if needed
	keyPtr := g.boxToI8Ptr(key, "map_key")

	// Call runtime function to get value
	result := g.builder.NewCall(mapGetFunc, mapObj, keyPtr)

	// The result is an i8* pointer that may need type conversion
	// For now, we'll return it as-is and let the caller handle conversion
	return result, nil
}

// generateMapElementAssignment generates LLVM IR for map element assignment.
func (g *LLVMCodegen) generateMapElementAssignment(mapObj, key, value value.Value) error {
	// Declare runtime map put function if not already declared
	mapPutFunc, exists := g.builtinFunctions["alas_runtime_map_put"]
	if !exists {
		// Create function signature: map_put(map i8*, key i8*, value i8*) -> void
		mapPutFunc = g.module.NewFunc("alas_runtime_map_put", types.Void)
		mapPutFunc.Params = append(mapPutFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)),
			ir.NewParam("key", types.NewPointer(types.I8)),
			ir.NewParam("value", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_put"] = mapPutFunc
	}

	// Convert key and value to i8* if needed
	keyPtr := g.boxToI8Ptr(key, "map_key")
	valuePtr := g.boxToI8Ptr(value, "map_value")

	// Call runtime function to set value
	g.builder.NewCall(mapPutFunc, mapObj, keyPtr, valuePtr)

	return nil
}

// generateMapLength generates LLVM IR for getting map length.
func (g *LLVMCodegen) generateMapLength(mapObj value.Value) (value.Value, error) {
	// Declare runtime map size function if not already declared
	mapSizeFunc, exists := g.builtinFunctions["alas_runtime_map_size"]
	if !exists {
		// Create function signature: map_size(map i8*) -> i64
		mapSizeFunc = g.module.NewFunc("alas_runtime_map_size", types.I64)
		mapSizeFunc.Params = append(mapSizeFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_size"] = mapSizeFunc
	}

	// Call runtime function to get size
	return g.builder.NewCall(mapSizeFunc, mapObj), nil
}

// generateMapContains generates LLVM IR for checking if map contains a key.
func (g *LLVMCodegen) generateMapContains(mapObj, key value.Value) (value.Value, error) {
	// Declare runtime map contains function if not already declared
	mapContainsFunc, exists := g.builtinFunctions["alas_runtime_map_contains"]
	if !exists {
		// Create function signature: map_contains(map i8*, key i8*) -> i1
		mapContainsFunc = g.module.NewFunc("alas_runtime_map_contains", types.I1)
		mapContainsFunc.Params = append(mapContainsFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)),
			ir.NewParam("key", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_contains"] = mapContainsFunc
	}

	// Convert key to i8* if needed
	keyPtr := g.boxToI8Ptr(key, "map_key")

	// Call runtime function to check containment
	return g.builder.NewCall(mapContainsFunc, mapObj, keyPtr), nil
}

// generateMapRemove generates LLVM IR for removing a key from map.
func (g *LLVMCodegen) generateMapRemove(mapObj, key value.Value) error {
	// Declare runtime map remove function if not already declared
	mapRemoveFunc, exists := g.builtinFunctions["alas_runtime_map_remove"]
	if !exists {
		// Create function signature: map_remove(map i8*, key i8*) -> void
		mapRemoveFunc = g.module.NewFunc("alas_runtime_map_remove", types.Void)
		mapRemoveFunc.Params = append(mapRemoveFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)),
			ir.NewParam("key", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_remove"] = mapRemoveFunc
	}

	// Convert key to i8* if needed
	keyPtr := g.boxToI8Ptr(key, "map_key")

	// Call runtime function to remove key
	g.builder.NewCall(mapRemoveFunc, mapObj, keyPtr)

	return nil
}

// generateMapKeys generates LLVM IR for getting all keys from a map.
func (g *LLVMCodegen) generateMapKeys(mapObj value.Value) (value.Value, error) {
	// Declare runtime map keys function if not already declared
	mapKeysFunc, exists := g.builtinFunctions["alas_runtime_map_keys"]
	if !exists {
		// Create function signature: map_keys(map i8*) -> array struct
		arrayType, _ := g.convertType(ast.TypeArray)
		mapKeysFunc = g.module.NewFunc("alas_runtime_map_keys", arrayType)
		mapKeysFunc.Params = append(mapKeysFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_keys"] = mapKeysFunc
	}

	// Call runtime function to get keys array
	return g.builder.NewCall(mapKeysFunc, mapObj), nil
}

// generateMapValues generates LLVM IR for getting all values from a map.
func (g *LLVMCodegen) generateMapValues(mapObj value.Value) (value.Value, error) {
	// Declare runtime map values function if not already declared
	mapValuesFunc, exists := g.builtinFunctions["alas_runtime_map_values"]
	if !exists {
		// Create function signature: map_values(map i8*) -> array struct
		arrayType, _ := g.convertType(ast.TypeArray)
		mapValuesFunc = g.module.NewFunc("alas_runtime_map_values", arrayType)
		mapValuesFunc.Params = append(mapValuesFunc.Params,
			ir.NewParam("map", types.NewPointer(types.I8)))
		g.builtinFunctions["alas_runtime_map_values"] = mapValuesFunc
	}

	// Call runtime function to get values array
	return g.builder.NewCall(mapValuesFunc, mapObj), nil
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

// declareErrorHandlingFunctions declares runtime error handling functions.
func (g *LLVMCodegen) declareErrorHandlingFunctions() {
	// String pointer type for error messages
	stringPtrType := types.NewPointer(types.I8)

	// Error reporting: alas_runtime_error(message *i8, file *i8, line i32, column i32) -> void
	runtimeErrorFunc := g.module.NewFunc("alas_runtime_error", types.Void)
	runtimeErrorFunc.Params = append(runtimeErrorFunc.Params,
		ir.NewParam("message", stringPtrType),
		ir.NewParam("file", stringPtrType),
		ir.NewParam("line", types.I32),
		ir.NewParam("column", types.I32))
	g.builtinFunctions["alas_runtime_error"] = runtimeErrorFunc

	// Stack trace: alas_runtime_stack_trace() -> void
	stackTraceFunc := g.module.NewFunc("alas_runtime_stack_trace", types.Void)
	g.builtinFunctions["alas_runtime_stack_trace"] = stackTraceFunc

	// Panic with message: alas_runtime_panic(message *i8) -> void (noreturn)
	panicFunc := g.module.NewFunc("alas_runtime_panic", types.Void)
	panicFunc.Params = append(panicFunc.Params, ir.NewParam("message", stringPtrType))
	g.builtinFunctions["alas_runtime_panic"] = panicFunc

	// Assert: alas_runtime_assert(condition i1, message *i8, file *i8, line i32) -> void
	assertFunc := g.module.NewFunc("alas_runtime_assert", types.Void)
	assertFunc.Params = append(assertFunc.Params,
		ir.NewParam("condition", types.I1),
		ir.NewParam("message", stringPtrType),
		ir.NewParam("file", stringPtrType),
		ir.NewParam("line", types.I32))
	g.builtinFunctions["alas_runtime_assert"] = assertFunc

	// Division by zero check: alas_runtime_check_div_zero(divisor i64, file *i8, line i32) -> void
	checkDivZeroFunc := g.module.NewFunc("alas_runtime_check_div_zero", types.Void)
	checkDivZeroFunc.Params = append(checkDivZeroFunc.Params,
		ir.NewParam("divisor", types.I64),
		ir.NewParam("file", stringPtrType),
		ir.NewParam("line", types.I32))
	g.builtinFunctions["alas_runtime_check_div_zero"] = checkDivZeroFunc

	// Array bounds check: alas_runtime_check_bounds(index i64, length i64, file *i8, line i32) -> void
	checkBoundsFunc := g.module.NewFunc("alas_runtime_check_bounds", types.Void)
	checkBoundsFunc.Params = append(checkBoundsFunc.Params,
		ir.NewParam("index", types.I64),
		ir.NewParam("length", types.I64),
		ir.NewParam("file", stringPtrType),
		ir.NewParam("line", types.I32))
	g.builtinFunctions["alas_runtime_check_bounds"] = checkBoundsFunc

	// Null pointer check: alas_runtime_check_null(ptr i8*, file *i8, line i32) -> void
	checkNullFunc := g.module.NewFunc("alas_runtime_check_null", types.Void)
	checkNullFunc.Params = append(checkNullFunc.Params,
		ir.NewParam("ptr", stringPtrType),
		ir.NewParam("file", stringPtrType),
		ir.NewParam("line", types.I32))
	g.builtinFunctions["alas_runtime_check_null"] = checkNullFunc
}

// generateDivisionByZeroCheck generates runtime division by zero checking.
func (g *LLVMCodegen) generateDivisionByZeroCheck(divisor value.Value) {
	// Get the division by zero check function
	checkFunc, exists := g.builtinFunctions["alas_runtime_check_div_zero"]
	if !exists {
		// Force declare if not found
		stringPtrType := types.NewPointer(types.I8)
		checkDivZeroFunc := g.module.NewFunc("alas_runtime_check_div_zero", types.Void)
		checkDivZeroFunc.Params = append(checkDivZeroFunc.Params,
			ir.NewParam("divisor", types.I64),
			ir.NewParam("file", stringPtrType),
			ir.NewParam("line", types.I32))
		g.builtinFunctions["alas_runtime_check_div_zero"] = checkDivZeroFunc
		checkFunc = checkDivZeroFunc
	}

	// Convert divisor to i64 if needed
	var divisorI64 value.Value
	if divisor.Type().Equal(types.I64) {
		divisorI64 = divisor
	} else if divisor.Type().Equal(types.I32) {
		divisorI64 = g.builder.NewSExt(divisor, types.I64)
	} else {
		// For other types, assume it's already compatible or skip check
		return
	}

	// Create filename and line number literals (for now use placeholder values)
	fileName := g.createStringLiteral("unknown.alas")
	lineNumber := constant.NewInt(types.I32, 0)

	// Call the runtime check function
	call := g.builder.NewCall(checkFunc, divisorI64, fileName, lineNumber)
	_ = call // Use the call to prevent optimization
}

// generateBoundsCheckWithError generates enhanced bounds checking with error reporting.
func (g *LLVMCodegen) generateBoundsCheckWithError(index, length value.Value, arrayName string) {
	// Get the bounds check function
	checkFunc, exists := g.builtinFunctions["alas_runtime_check_bounds"]
	if !exists {
		return // Function not declared, skip check
	}

	// Convert index and length to i64 if needed
	var indexI64, lengthI64 value.Value

	if index.Type().Equal(types.I64) {
		indexI64 = index
	} else if index.Type().Equal(types.I32) {
		indexI64 = g.builder.NewSExt(index, types.I64)
	} else {
		return
	}

	if length.Type().Equal(types.I64) {
		lengthI64 = length
	} else if length.Type().Equal(types.I32) {
		lengthI64 = g.builder.NewSExt(length, types.I64)
	} else {
		return
	}

	// Create filename and line number literals
	fileName := g.createStringLiteral("unknown.alas")
	lineNumber := constant.NewInt(types.I32, 0)

	// Call the runtime check function
	g.builder.NewCall(checkFunc, indexI64, lengthI64, fileName, lineNumber)
}

// generateNullPointerCheck generates null pointer checking.
func (g *LLVMCodegen) generateNullPointerCheck(ptr value.Value, context string) {
	// Get the null check function
	checkFunc, exists := g.builtinFunctions["alas_runtime_check_null"]
	if !exists {
		return // Function not declared, skip check
	}

	// Only check pointer types
	if _, isPtr := ptr.Type().(*types.PointerType); !isPtr {
		return
	}

	// Create filename and line number literals
	fileName := g.createStringLiteral("unknown.alas")
	lineNumber := constant.NewInt(types.I32, 0)

	// Call the runtime check function
	g.builder.NewCall(checkFunc, ptr, fileName, lineNumber)
}

// generateAssert generates runtime assertion checking.
func (g *LLVMCodegen) generateAssert(condition value.Value, message string) {
	// Get the assert function
	assertFunc, exists := g.builtinFunctions["alas_runtime_assert"]
	if !exists {
		return // Function not declared, skip check
	}

	// Create message literal
	messageLiteral := g.createStringLiteral(message)

	// Create filename and line number literals
	fileName := g.createStringLiteral("unknown.alas")
	lineNumber := constant.NewInt(types.I32, 0)

	// Call the runtime assert function
	g.builder.NewCall(assertFunc, condition, messageLiteral, fileName, lineNumber)
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

	// math.max and math.min take two arguments
	maxFunc := g.module.NewFunc("alas_builtin_math_max", cvalueReturnType)
	maxFunc.Params = append(maxFunc.Params, ir.NewParam("", cvalueArgType))
	maxFunc.Params = append(maxFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["math.max"] = maxFunc

	minFunc := g.module.NewFunc("alas_builtin_math_min", cvalueReturnType)
	minFunc.Params = append(minFunc.Params, ir.NewParam("", cvalueArgType))
	minFunc.Params = append(minFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["math.min"] = minFunc

	// Collections functions
	// void* alas_builtin_collections_length(void* val)
	lengthFunc := g.module.NewFunc("alas_builtin_collections_length", cvalueReturnType)
	lengthFunc.Params = append(lengthFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["collections.length"] = lengthFunc

	// void* alas_builtin_collections_contains(void* collection, void* item)
	containsFunc := g.module.NewFunc("alas_builtin_collections_contains", cvalueReturnType)
	containsFunc.Params = append(containsFunc.Params, ir.NewParam("", cvalueArgType))
	containsFunc.Params = append(containsFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["collections.contains"] = containsFunc

	// Array functions
	// void* alas_builtin_array_length(void* array)
	arrayLengthFunc := g.module.NewFunc("alas_builtin_array_length", cvalueReturnType)
	arrayLengthFunc.Params = append(arrayLengthFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["array.length"] = arrayLengthFunc

	// void* alas_builtin_array_push(void* array, void* element)
	arrayPushFunc := g.module.NewFunc("alas_builtin_array_push", cvalueReturnType)
	arrayPushFunc.Params = append(arrayPushFunc.Params, ir.NewParam("", cvalueArgType))
	arrayPushFunc.Params = append(arrayPushFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["array.push"] = arrayPushFunc

	// void* alas_builtin_array_pop(void* array)
	arrayPopFunc := g.module.NewFunc("alas_builtin_array_pop", cvalueReturnType)
	arrayPopFunc.Params = append(arrayPopFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["array.pop"] = arrayPopFunc

	// void* alas_builtin_array_slice(void* array, void* start, void* end)
	arraySliceFunc := g.module.NewFunc("alas_builtin_array_slice", cvalueReturnType)
	arraySliceFunc.Params = append(arraySliceFunc.Params, ir.NewParam("", cvalueArgType))
	arraySliceFunc.Params = append(arraySliceFunc.Params, ir.NewParam("", cvalueArgType))
	arraySliceFunc.Params = append(arraySliceFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["array.slice"] = arraySliceFunc

	// Map functions
	// void* alas_builtin_map_get(void* map, void* key)
	mapGetBuiltinFunc := g.module.NewFunc("alas_builtin_map_get", cvalueReturnType)
	mapGetBuiltinFunc.Params = append(mapGetBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	mapGetBuiltinFunc.Params = append(mapGetBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.get"] = mapGetBuiltinFunc

	// void alas_builtin_map_put(void* map, void* key, void* value)
	mapPutBuiltinFunc := g.module.NewFunc("alas_builtin_map_put", types.Void)
	mapPutBuiltinFunc.Params = append(mapPutBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	mapPutBuiltinFunc.Params = append(mapPutBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	mapPutBuiltinFunc.Params = append(mapPutBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.put"] = mapPutBuiltinFunc

	// void* alas_builtin_map_size(void* map)
	mapSizeBuiltinFunc := g.module.NewFunc("alas_builtin_map_size", cvalueReturnType)
	mapSizeBuiltinFunc.Params = append(mapSizeBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.size"] = mapSizeBuiltinFunc

	// void* alas_builtin_map_contains(void* map, void* key)
	mapContainsBuiltinFunc := g.module.NewFunc("alas_builtin_map_contains", cvalueReturnType)
	mapContainsBuiltinFunc.Params = append(mapContainsBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	mapContainsBuiltinFunc.Params = append(mapContainsBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.contains"] = mapContainsBuiltinFunc

	// void alas_builtin_map_remove(void* map, void* key)
	mapRemoveBuiltinFunc := g.module.NewFunc("alas_builtin_map_remove", types.Void)
	mapRemoveBuiltinFunc.Params = append(mapRemoveBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	mapRemoveBuiltinFunc.Params = append(mapRemoveBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.remove"] = mapRemoveBuiltinFunc

	// void* alas_builtin_map_keys(void* map)
	mapKeysBuiltinFunc := g.module.NewFunc("alas_builtin_map_keys", cvalueReturnType)
	mapKeysBuiltinFunc.Params = append(mapKeysBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.keys"] = mapKeysBuiltinFunc

	// void* alas_builtin_map_values(void* map)
	mapValuesBuiltinFunc := g.module.NewFunc("alas_builtin_map_values", cvalueReturnType)
	mapValuesBuiltinFunc.Params = append(mapValuesBuiltinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["map.values"] = mapValuesBuiltinFunc

	// String functions
	// void* alas_builtin_string_toUpper(void* val)
	toUpperFunc := g.module.NewFunc("alas_builtin_string_toUpper", cvalueReturnType)
	toUpperFunc.Params = append(toUpperFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.toUpper"] = toUpperFunc

	toLowerFunc := g.module.NewFunc("alas_builtin_string_toLower", cvalueReturnType)
	toLowerFunc.Params = append(toLowerFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.toLower"] = toLowerFunc

	lengthStrFunc := g.module.NewFunc("alas_builtin_string_length", cvalueReturnType)
	lengthStrFunc.Params = append(lengthStrFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.length"] = lengthStrFunc

	// Additional string functions
	// void* alas_builtin_string_substring(void* str, void* start, void* end)
	substringFunc := g.module.NewFunc("alas_builtin_string_substring", cvalueReturnType)
	substringFunc.Params = append(substringFunc.Params, ir.NewParam("", cvalueArgType))
	substringFunc.Params = append(substringFunc.Params, ir.NewParam("", cvalueArgType))
	substringFunc.Params = append(substringFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.substring"] = substringFunc

	// void* alas_builtin_string_indexOf(void* str, void* search)
	indexOfFunc := g.module.NewFunc("alas_builtin_string_indexOf", cvalueReturnType)
	indexOfFunc.Params = append(indexOfFunc.Params, ir.NewParam("", cvalueArgType))
	indexOfFunc.Params = append(indexOfFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.indexOf"] = indexOfFunc

	// void* alas_builtin_string_split(void* str, void* delimiter)
	splitFunc := g.module.NewFunc("alas_builtin_string_split", cvalueReturnType)
	splitFunc.Params = append(splitFunc.Params, ir.NewParam("", cvalueArgType))
	splitFunc.Params = append(splitFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.split"] = splitFunc

	// void* alas_builtin_string_join(void* array, void* separator)
	joinFunc := g.module.NewFunc("alas_builtin_string_join", cvalueReturnType)
	joinFunc.Params = append(joinFunc.Params, ir.NewParam("", cvalueArgType))
	joinFunc.Params = append(joinFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.join"] = joinFunc

	// void* alas_builtin_string_replace(void* str, void* search, void* replacement)
	replaceFunc := g.module.NewFunc("alas_builtin_string_replace", cvalueReturnType)
	replaceFunc.Params = append(replaceFunc.Params, ir.NewParam("", cvalueArgType))
	replaceFunc.Params = append(replaceFunc.Params, ir.NewParam("", cvalueArgType))
	replaceFunc.Params = append(replaceFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.replace"] = replaceFunc

	// void* alas_builtin_string_trim(void* str)
	trimFunc := g.module.NewFunc("alas_builtin_string_trim", cvalueReturnType)
	trimFunc.Params = append(trimFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.trim"] = trimFunc

	// void* alas_builtin_string_startsWith(void* str, void* prefix)
	startsWithFunc := g.module.NewFunc("alas_builtin_string_startsWith", cvalueReturnType)
	startsWithFunc.Params = append(startsWithFunc.Params, ir.NewParam("", cvalueArgType))
	startsWithFunc.Params = append(startsWithFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.startsWith"] = startsWithFunc

	// void* alas_builtin_string_endsWith(void* str, void* suffix)
	endsWithFunc := g.module.NewFunc("alas_builtin_string_endsWith", cvalueReturnType)
	endsWithFunc.Params = append(endsWithFunc.Params, ir.NewParam("", cvalueArgType))
	endsWithFunc.Params = append(endsWithFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.endsWith"] = endsWithFunc

	// void* alas_builtin_string_format(void* template, void* args)
	formatFunc := g.module.NewFunc("alas_builtin_string_format", cvalueReturnType)
	formatFunc.Params = append(formatFunc.Params, ir.NewParam("", cvalueArgType))
	formatFunc.Params = append(formatFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.format"] = formatFunc

	// void* alas_builtin_string_charAt(void* str, void* index)
	charAtFunc := g.module.NewFunc("alas_builtin_string_charAt", cvalueReturnType)
	charAtFunc.Params = append(charAtFunc.Params, ir.NewParam("", cvalueArgType))
	charAtFunc.Params = append(charAtFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.charAt"] = charAtFunc

	// void* alas_builtin_string_charCodeAt(void* str, void* index)
	charCodeAtFunc := g.module.NewFunc("alas_builtin_string_charCodeAt", cvalueReturnType)
	charCodeAtFunc.Params = append(charCodeAtFunc.Params, ir.NewParam("", cvalueArgType))
	charCodeAtFunc.Params = append(charCodeAtFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.charCodeAt"] = charCodeAtFunc

	// void* alas_builtin_string_fromCharCode(void* code)
	fromCharCodeFunc := g.module.NewFunc("alas_builtin_string_fromCharCode", cvalueReturnType)
	fromCharCodeFunc.Params = append(fromCharCodeFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.fromCharCode"] = fromCharCodeFunc

	// void* alas_builtin_string_repeat(void* str, void* count)
	repeatFunc := g.module.NewFunc("alas_builtin_string_repeat", cvalueReturnType)
	repeatFunc.Params = append(repeatFunc.Params, ir.NewParam("", cvalueArgType))
	repeatFunc.Params = append(repeatFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.repeat"] = repeatFunc

	// void* alas_builtin_string_padStart(void* str, void* length, void* padString)
	padStartFunc := g.module.NewFunc("alas_builtin_string_padStart", cvalueReturnType)
	padStartFunc.Params = append(padStartFunc.Params, ir.NewParam("", cvalueArgType))
	padStartFunc.Params = append(padStartFunc.Params, ir.NewParam("", cvalueArgType))
	padStartFunc.Params = append(padStartFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.padStart"] = padStartFunc

	// void* alas_builtin_string_padEnd(void* str, void* length, void* padString)
	padEndFunc := g.module.NewFunc("alas_builtin_string_padEnd", cvalueReturnType)
	padEndFunc.Params = append(padEndFunc.Params, ir.NewParam("", cvalueArgType))
	padEndFunc.Params = append(padEndFunc.Params, ir.NewParam("", cvalueArgType))
	padEndFunc.Params = append(padEndFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.padEnd"] = padEndFunc

	// void* alas_builtin_string_contains(void* str, void* search)
	containsStrFunc := g.module.NewFunc("alas_builtin_string_contains", cvalueReturnType)
	containsStrFunc.Params = append(containsStrFunc.Params, ir.NewParam("", cvalueArgType))
	containsStrFunc.Params = append(containsStrFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.contains"] = containsStrFunc

	// void* alas_builtin_string_concat(void* str1, void* str2)
	concatFunc := g.module.NewFunc("alas_builtin_string_concat", cvalueReturnType)
	concatFunc.Params = append(concatFunc.Params, ir.NewParam("", cvalueArgType))
	concatFunc.Params = append(concatFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["string.concat"] = concatFunc

	// Type functions
	// void* alas_builtin_type_typeOf(void* val)
	typeOfFunc := g.module.NewFunc("alas_builtin_type_typeOf", cvalueReturnType)
	typeOfFunc.Params = append(typeOfFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["type.typeOf"] = typeOfFunc

	isIntFunc := g.module.NewFunc("alas_builtin_type_isInt", cvalueReturnType)
	isIntFunc.Params = append(isIntFunc.Params, ir.NewParam("", cvalueArgType))
	g.builtinFunctions["type.isInt"] = isIntFunc

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
		// Return a dummy value for void functions
		return constant.NewInt(types.I32, 0), nil
	}

	// Handle functions that take multiple arguments (2 args)
	if expr.Name == "math.max" || expr.Name == "math.min" || expr.Name == "collections.contains" ||
		expr.Name == "array.push" || expr.Name == "map.get" || expr.Name == "map.contains" ||
		expr.Name == "map.remove" || expr.Name == "string.indexOf" || expr.Name == "string.split" ||
		expr.Name == "string.join" || expr.Name == "string.startsWith" || expr.Name == "string.endsWith" ||
		expr.Name == "string.format" || expr.Name == "string.charAt" || expr.Name == "string.charCodeAt" ||
		expr.Name == "string.repeat" || expr.Name == "string.contains" || expr.Name == "string.concat" {
		// These functions take 2 arguments
		expectedArgs := 2
		if len(expr.Args) != expectedArgs {
			return nil, fmt.Errorf("%s expects %d arguments, got %d", expr.Name, expectedArgs, len(expr.Args))
		}

		// Generate and convert both arguments
		var args []value.Value
		for i := 0; i < expectedArgs; i++ {
			argVal, err := g.generateExpression(&expr.Args[i])
			if err != nil {
				return nil, err
			}
			args = append(args, g.convertToCValue(argVal))
		}

		// Call the function with both arguments
		result := g.builder.NewCall(builtinFunc, args...)

		// Convert result from CValue
		return g.convertFromCValue(result)
	}

	// Handle functions that take three arguments
	if expr.Name == "array.slice" || expr.Name == "map.put" || expr.Name == "string.substring" ||
		expr.Name == "string.replace" || expr.Name == "string.padStart" || expr.Name == "string.padEnd" {
		// These functions take 3 arguments
		expectedArgs := 3
		if len(expr.Args) != expectedArgs {
			return nil, fmt.Errorf("%s expects %d arguments, got %d", expr.Name, expectedArgs, len(expr.Args))
		}

		// Generate and convert all arguments
		var args []value.Value
		for i := 0; i < expectedArgs; i++ {
			argVal, err := g.generateExpression(&expr.Args[i])
			if err != nil {
				return nil, err
			}
			args = append(args, g.convertToCValue(argVal))
		}

		// Call the function with all arguments
		result := g.builder.NewCall(builtinFunc, args...)

		// Convert result from CValue
		return g.convertFromCValue(result)
	}

	// For functions that return values with single argument
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
			types.I64,                  // int_val
			types.Double,               // float_val
			types.NewPointer(types.I8), // string_val
			types.NewPointer(types.I8), // array_val
			types.NewPointer(types.I8), // map_val
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
			types.I64,                  // int_val
			types.Double,               // float_val
			types.NewPointer(types.I8), // string_val
			types.NewPointer(types.I8), // array_val
			types.NewPointer(types.I8), // map_val
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

// declareImportedFunctions declares external functions from imported modules.
func (g *LLVMCodegen) declareImportedFunctions(imports []string) error {
	// If no module loader is set, we can't load imports
	// This is okay for single-module compilation
	if g.moduleLoader == nil {
		return nil
	}

	for _, importName := range imports {
		// Check if module is already loaded (caching)
		var importedModule *ast.Module
		var err error

		if cachedModule, exists := g.loadedModules[importName]; exists {
			importedModule = cachedModule
		} else {
			// Load the imported module
			importedModule, err = g.moduleLoader.LoadModuleByName(importName)
			if err != nil {
				// Skip if module not found - this allows testing without full module resolution
				continue
			}
			// Cache the loaded module
			g.loadedModules[importName] = importedModule
		}

		// Import custom types from the module
		for _, typeDef := range importedModule.Types {
			// Check if type is exported (assume all types are exported for now)
			qualifiedTypeName := fmt.Sprintf("%s__%s", importName, typeDef.Name)
			g.customTypes[qualifiedTypeName] = &typeDef

			// Generate LLVM struct type for custom types
			if err := g.declareCustomType(&typeDef); err != nil {
				return fmt.Errorf("failed to generate type %s: %v", qualifiedTypeName, err)
			}
		}

		// Declare all exported functions from the imported module
		for _, fn := range importedModule.Functions {
			// Check if function is exported
			isExported := false
			for _, exportName := range importedModule.Exports {
				if exportName == fn.Name {
					isExported = true
					break
				}
			}

			if isExported {
				// Create qualified name: module__function
				qualifiedName := fmt.Sprintf("%s__%s", importName, fn.Name)

				// Check if function is already declared
				if _, exists := g.externalFunctions[qualifiedName]; exists {
					continue
				}

				// Convert return type
				retType, err := g.convertType(fn.Returns)
				if err != nil {
					return fmt.Errorf("failed to convert return type for %s: %v", qualifiedName, err)
				}

				// Convert parameter types
				var paramTypes []types.Type
				for _, param := range fn.Params {
					paramType, err := g.convertType(param.Type)
					if err != nil {
						return fmt.Errorf("failed to convert parameter type for %s: %v", qualifiedName, err)
					}
					paramTypes = append(paramTypes, paramType)
				}

				// Declare the external function
				var params []*ir.Param
				for i := 0; i < len(paramTypes); i++ {
					params = append(params, ir.NewParam("", paramTypes[i]))
				}
				externalFunc := g.module.NewFunc(qualifiedName, retType, params...)

				// Mark as external (no body)
				g.externalFunctions[qualifiedName] = externalFunc

				// Cache the function AST for potential inlining or analysis
				g.astFunctions[qualifiedName] = &fn
			}
		}
	}

	return nil
}

// resolveModuleDependencies resolves all module dependencies recursively.
func (g *LLVMCodegen) resolveModuleDependencies(moduleName string, visited map[string]bool) error {
	// Prevent circular dependencies
	if visited[moduleName] {
		return fmt.Errorf("circular dependency detected: %s", moduleName)
	}
	visited[moduleName] = true
	defer delete(visited, moduleName) // Allow re-entry from different paths

	// Check if module is already resolved
	if _, exists := g.loadedModules[moduleName]; exists {
		return nil
	}

	// Load the module
	module, err := g.moduleLoader.LoadModuleByName(moduleName)
	if err != nil {
		return fmt.Errorf("failed to load module %s: %v", moduleName, err)
	}

	// Recursively resolve dependencies
	for _, importName := range module.Imports {
		if err := g.resolveModuleDependencies(importName, visited); err != nil {
			return err
		}
	}

	// Cache the resolved module
	g.loadedModules[moduleName] = module

	return nil
}

// compileModule compiles a module and caches the result.
func (g *LLVMCodegen) compileModule(moduleName string) (*ir.Module, error) {
	// Check if module is already compiled
	if compiledModule, exists := g.compiledModules[moduleName]; exists {
		return compiledModule, nil
	}

	// Load the module AST
	module, exists := g.loadedModules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module %s not loaded", moduleName)
	}

	// Create a new codegen instance for this module
	moduleCodegen := NewLLVMCodegenWithLoader(g.moduleLoader)

	// Copy shared state (types, etc.)
	for name, typeDef := range g.customTypes {
		moduleCodegen.customTypes[name] = typeDef
	}
	for name, structType := range g.structTypes {
		moduleCodegen.structTypes[name] = structType
	}
	for name, fieldMap := range g.fieldIndices {
		moduleCodegen.fieldIndices[name] = fieldMap
	}

	// Generate the module
	compiledModule, err := moduleCodegen.GenerateModule(module)
	if err != nil {
		return nil, fmt.Errorf("failed to compile module %s: %v", moduleName, err)
	}

	// Cache the compiled module
	g.compiledModules[moduleName] = compiledModule

	return compiledModule, nil
}

// getTypeSize returns the size in bytes for a given LLVM type.
func (g *LLVMCodegen) getTypeSize(t types.Type) int64 {
	switch typ := t.(type) {
	case *types.IntType:
		// Round up to the nearest byte for non-byte-aligned types
		// e.g., i1 needs 1 byte, i7 needs 1 byte, i9 needs 2 bytes
		// Safe conversion: BitSize is always positive
		// Round up to the nearest byte for non-byte-aligned types
		if typ.BitSize > 0x7FFFFFFF {
			return 8 // Default to 8 bytes for invalid bit sizes
		}
		// Safe conversion: checked above
		bitSize := int64(typ.BitSize)
		return (bitSize + 7) / 8
	case *types.FloatType:
		switch typ.Kind {
		case types.FloatKindHalf:
			return 2
		case types.FloatKindFloat:
			return 4
		case types.FloatKindDouble:
			return 8
		case types.FloatKindFP128:
			return 16
		case types.FloatKindX86_FP80:
			return 10 // 80 bits = 10 bytes
		case types.FloatKindPPC_FP128:
			return 16
		default:
			return 8
		}
	case *types.PointerType:
		return 8 // 64-bit pointers
	default:
		// Default to 8 bytes for unknown types
		return 8
	}
}

// boxToI8Ptr boxes a value into heap memory and returns it as an i8* pointer.
// If the value is already an i8* pointer, it returns it unchanged.
func (g *LLVMCodegen) boxToI8Ptr(val value.Value, name string) value.Value {
	if val.Type() == types.I8Ptr {
		return val
	}

	// Ensure malloc is declared
	mallocFunc, exists := g.builtinFunctions["malloc"]
	if !exists {
		mallocFunc = g.module.NewFunc("malloc", types.I8Ptr, ir.NewParam("size", types.I64))
		g.builtinFunctions["malloc"] = mallocFunc
	}

	// Calculate size and allocate heap memory
	size := constant.NewInt(types.I64, g.getTypeSize(val.Type()))
	heapPtr := g.builder.NewCall(mallocFunc, size)
	heapPtr.SetName(name)

	// Note: We're not checking for malloc failure here as it would require
	// complex control flow manipulation. In a production system, consider:
	// 1. Using a runtime that guarantees allocation or aborts
	// 2. Implementing a separate allocation wrapper function
	// 3. Using LLVM's garbage collection infrastructure

	// Cast to proper type and store value
	typedPtr := g.builder.NewBitCast(heapPtr, types.NewPointer(val.Type()))
	g.builder.NewStore(val, typedPtr)

	// Return as i8*
	return heapPtr
}
