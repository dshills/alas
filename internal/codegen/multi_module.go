package codegen

import (
	"fmt"
	"path/filepath"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"

	"github.com/dshills/alas/internal/ast"
)

// MultiModuleCodegen manages compilation of multiple interconnected ALaS modules.
type MultiModuleCodegen struct {
	modules           map[string]*ast.Module       // Module name -> AST
	compiledModules   map[string]*ir.Module        // Module name -> LLVM IR
	dependencies      map[string][]string          // Module name -> list of dependencies
	externalFunctions map[string]*ExternalFunction // Qualified name -> function info
	moduleLoaders     map[string]ModuleLoader      // Module name -> loader function
}

// ExternalFunction represents a function from another module.
type ExternalFunction struct {
	Module     string
	Name       string
	ParamTypes []types.Type
	ReturnType types.Type
	LLVMFunc   *ir.Func
}

// ModuleLoader is a function that loads a module by name.
type ModuleLoader func(name string) (*ast.Module, error)

// NewMultiModuleCodegen creates a new multi-module code generator.
func NewMultiModuleCodegen() *MultiModuleCodegen {
	return &MultiModuleCodegen{
		modules:           make(map[string]*ast.Module),
		compiledModules:   make(map[string]*ir.Module),
		dependencies:      make(map[string][]string),
		externalFunctions: make(map[string]*ExternalFunction),
		moduleLoaders:     make(map[string]ModuleLoader),
	}
}

// RegisterModuleLoader registers a loader function for a specific module.
func (m *MultiModuleCodegen) RegisterModuleLoader(moduleName string, loader ModuleLoader) {
	m.moduleLoaders[moduleName] = loader
}

// AddModule adds a module to be compiled.
func (m *MultiModuleCodegen) AddModule(module *ast.Module) error {
	if module.Name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	m.modules[module.Name] = module
	m.dependencies[module.Name] = module.Imports
	return nil
}

// LoadModule loads a module using registered loaders.
func (m *MultiModuleCodegen) LoadModule(name string) (*ast.Module, error) {
	// Check if already loaded
	if module, exists := m.modules[name]; exists {
		return module, nil
	}

	// Try to load using registered loaders
	loader, exists := m.moduleLoaders[name]
	if !exists {
		return nil, fmt.Errorf("no loader registered for module: %s", name)
	}

	module, err := loader(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load module %s: %v", name, err)
	}

	// Add the loaded module
	err = m.AddModule(module)
	if err != nil {
		return nil, fmt.Errorf("failed to add loaded module %s: %v", name, err)
	}

	return module, nil
}

// ResolveDependencies resolves all module dependencies and returns compilation order.
func (m *MultiModuleCodegen) ResolveDependencies() ([]string, error) {
	// Load all dependencies recursively
	for moduleName := range m.modules {
		if err := m.loadDependenciesRecursive(moduleName, make(map[string]bool)); err != nil {
			return nil, err
		}
	}

	// Perform topological sort to determine compilation order
	return m.topologicalSort()
}

// loadDependenciesRecursive loads all dependencies for a module recursively.
func (m *MultiModuleCodegen) loadDependenciesRecursive(moduleName string, visited map[string]bool) error {
	if visited[moduleName] {
		return fmt.Errorf("circular dependency detected involving module: %s", moduleName)
	}

	visited[moduleName] = true
	defer func() { visited[moduleName] = false }()

	module, exists := m.modules[moduleName]
	if !exists {
		// Try to load the module
		var err error
		module, err = m.LoadModule(moduleName)
		if err != nil {
			return err
		}
	}

	// Load all dependencies
	for _, dep := range module.Imports {
		if _, exists := m.modules[dep]; !exists {
			if _, err := m.LoadModule(dep); err != nil {
				return fmt.Errorf("failed to load dependency %s for module %s: %v", dep, moduleName, err)
			}
		}

		// Recursively load dependencies of dependencies
		if err := m.loadDependenciesRecursive(dep, visited); err != nil {
			return err
		}
	}

	return nil
}

// topologicalSort returns modules in dependency order (dependencies first).
func (m *MultiModuleCodegen) topologicalSort() ([]string, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	// Initialize in-degree count and build adjacency list
	for moduleName := range m.modules {
		inDegree[moduleName] = 0
		graph[moduleName] = []string{}
	}

	// Build the graph (dependency -> dependent)
	for moduleName, deps := range m.dependencies {
		for _, dep := range deps {
			if _, exists := m.modules[dep]; !exists {
				return nil, fmt.Errorf("dependency %s not found for module %s", dep, moduleName)
			}
			graph[dep] = append(graph[dep], moduleName)
			inDegree[moduleName]++
		}
	}

	// Find modules with no dependencies
	var queue []string
	for moduleName, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, moduleName)
		}
	}

	var result []string
	for len(queue) > 0 {
		// Remove module from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Update in-degrees of dependent modules
		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(m.modules) {
		return nil, fmt.Errorf("circular dependency detected in modules")
	}

	return result, nil
}

// GetQualifiedFunctionName returns the mangled name for a cross-module function call.
func (m *MultiModuleCodegen) GetQualifiedFunctionName(moduleName, functionName string) string {
	return fmt.Sprintf("%s__%s", moduleName, functionName)
}

// DeclareExternalFunction declares an external function from another module.
func (m *MultiModuleCodegen) DeclareExternalFunction(targetModule *ir.Module, moduleName, functionName string, paramTypes []types.Type, returnType types.Type) (*ir.Func, error) {
	qualifiedName := m.GetQualifiedFunctionName(moduleName, functionName)

	// Check if already declared
	if extFunc, exists := m.externalFunctions[qualifiedName]; exists {
		return extFunc.LLVMFunc, nil
	}

	// Declare the function as external with return type
	llvmFunc := targetModule.NewFunc(qualifiedName, returnType)
	
	// Add parameters to the function signature
	for i, paramType := range paramTypes {
		param := ir.NewParam(fmt.Sprintf("arg%d", i), paramType)
		llvmFunc.Params = append(llvmFunc.Params, param)
	}

	// Store the external function info
	m.externalFunctions[qualifiedName] = &ExternalFunction{
		Module:     moduleName,
		Name:       functionName,
		ParamTypes: paramTypes,
		ReturnType: returnType,
		LLVMFunc:   llvmFunc,
	}

	return llvmFunc, nil
}

// CompileModules compiles all modules in dependency order.
func (m *MultiModuleCodegen) CompileModules() (map[string]*ir.Module, error) {
	// Resolve compilation order
	order, err := m.ResolveDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %v", err)
	}

	// Compile modules in order
	for _, moduleName := range order {
		module := m.modules[moduleName]

		// Create enhanced LLVM codegen for this module
		codegen := NewLLVMCodegen()

		// Set up external function declarations for this module's dependencies
		if err := m.setupExternalDeclarations(codegen, module); err != nil {
			return nil, fmt.Errorf("failed to setup external declarations for module %s: %v", moduleName, err)
		}

		// Generate LLVM IR for the module
		llvmModule, err := codegen.GenerateModule(module)
		if err != nil {
			return nil, fmt.Errorf("failed to compile module %s: %v", moduleName, err)
		}

		// Store the compiled module
		m.compiledModules[moduleName] = llvmModule
	}

	return m.compiledModules, nil
}

// setupExternalDeclarations sets up external function declarations for a module's dependencies.
func (m *MultiModuleCodegen) setupExternalDeclarations(codegen *LLVMCodegen, module *ast.Module) error {
	// For each dependency, declare its exported functions as external
	for _, depName := range module.Imports {
		depModule, exists := m.modules[depName]
		if !exists {
			return fmt.Errorf("dependency module %s not found", depName)
		}

		// Declare external functions for all functions in the dependency
		for _, fn := range depModule.Functions {
			// Convert parameter types
			paramTypes := make([]types.Type, len(fn.Params))
			for i, param := range fn.Params {
				paramType, err := codegen.convertType(param.Type)
				if err != nil {
					return fmt.Errorf("invalid parameter type %s in function %s.%s: %v", param.Type, depName, fn.Name, err)
				}
				paramTypes[i] = paramType
			}

			// Convert return type
			returnType, err := codegen.convertType(fn.Returns)
			if err != nil {
				return fmt.Errorf("invalid return type %s in function %s.%s: %v", fn.Returns, depName, fn.Name, err)
			}

			// Declare the external function
			_, err = m.DeclareExternalFunction(codegen.module, depName, fn.Name, paramTypes, returnType)
			if err != nil {
				return fmt.Errorf("failed to declare external function %s.%s: %v", depName, fn.Name, err)
			}
		}
	}

	return nil
}

// LinkModules combines multiple LLVM modules into a single module.
func (m *MultiModuleCodegen) LinkModules(targetName string) (*ir.Module, error) {
	if len(m.compiledModules) == 0 {
		return nil, fmt.Errorf("no modules to link")
	}

	// Create a new module for the linked result
	linkedModule := ir.NewModule()
	linkedModule.SourceFilename = targetName

	// Copy all functions from all modules into the linked module
	for moduleName, module := range m.compiledModules {
		for _, fn := range module.Funcs {
			// Create a new function in the linked module
			_ = linkedModule.NewFunc(fn.Name(), fn.Sig)

			// Copy function body if it exists
			if len(fn.Blocks) > 0 {
				// For simplicity, we'll need to implement a proper function cloning mechanism
				// This is a placeholder that shows the structure
				_ = moduleName // Acknowledge parameter usage
				// TODO: Implement proper function body copying with value mapping
			}
		}

		// Copy global variables
		for _, global := range module.Globals {
			newGlobal := linkedModule.NewGlobalDef(global.Name(), global.Init)
			newGlobal.Immutable = global.Immutable
		}
	}

	return linkedModule, nil
}

// GetExternalFunctions returns all declared external functions.
func (m *MultiModuleCodegen) GetExternalFunctions() map[string]*ExternalFunction {
	return m.externalFunctions
}

// GetCompiledModules returns all compiled LLVM modules.
func (m *MultiModuleCodegen) GetCompiledModules() map[string]*ir.Module {
	return m.compiledModules
}

// FileSystemModuleLoader creates a module loader that loads from the file system.
func FileSystemModuleLoader(basePath string) ModuleLoader {
	return func(name string) (*ast.Module, error) {
		filename := filepath.Join(basePath, name+".alas.json")

		// This would normally use the parser to load the module
		// For now, return an error indicating the need for parser integration
		return nil, fmt.Errorf("file system module loading requires parser integration: %s", filename)
	}
}
