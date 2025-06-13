package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/llir/llvm/ir"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/validator"
)

func main() {
	var input string
	var output string
	var format string
	var optLevel string
	var modulePath string
	var linkMode string
	var mainModule string

	flag.StringVar(&input, "file", "", "ALaS JSON file to compile")
	flag.StringVar(&output, "o", "", "Output file (default: input file with .ll extension)")
	flag.StringVar(&format, "format", "ll", "Output format: ll (LLVM IR text) or bc (LLVM bitcode)")
	flag.StringVar(&optLevel, "O", "1", "Optimization level: 0 (none), 1 (basic), 2 (standard), 3 (aggressive)")
	flag.StringVar(&modulePath, "module-path", ".", "Path to search for module dependencies")
	flag.StringVar(&linkMode, "link", "none", "Linking mode: none (separate modules), all (link all modules)")
	flag.StringVar(&mainModule, "main", "", "Main module name for whole-program compilation")
	flag.Parse()

	if input == "" {
		fmt.Fprintf(os.Stderr, "Error: -file parameter is required for multi-module compilation\n")
		os.Exit(1)
	}

	// Parse optimization level
	var optimizationLevel codegen.OptimizationLevel
	switch optLevel {
	case "0":
		optimizationLevel = codegen.OptNone
	case "1":
		optimizationLevel = codegen.OptBasic
	case "2":
		optimizationLevel = codegen.OptStandard
	case "3":
		optimizationLevel = codegen.OptAggressive
	default:
		fmt.Fprintf(os.Stderr, "Invalid optimization level: %s (use 0, 1, 2, or 3)\n", optLevel)
		os.Exit(1)
	}

	// Create multi-module code generator
	multiCodegen := codegen.NewMultiModuleCodegen()

	// Register file system module loader
	moduleLoader := createFileSystemModuleLoader(modulePath)

	// Load the main module
	mainModuleAST, err := loadModuleFromFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading main module: %v\n", err)
		os.Exit(1)
	}

	// Add main module to the multi-module codegen
	if err := multiCodegen.AddModule(mainModuleAST); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding main module: %v\n", err)
		os.Exit(1)
	}

	// Register module loader for dependencies
	multiCodegen.RegisterModuleLoader("math_utils", moduleLoader)
	multiCodegen.RegisterModuleLoader("format_utils", moduleLoader)
	// Add more module loaders as needed based on common dependencies

	if linkMode == "all" || mainModule != "" {
		// Whole-program compilation mode
		err = compileLinkedProgram(multiCodegen, mainModuleAST.Name, output, format, optimizationLevel)
	} else {
		// Separate compilation mode
		err = compileSeparateModules(multiCodegen, input, output, format, optimizationLevel)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation failed: %v\n", err)
		os.Exit(1)
	}
}

// loadModuleFromFile loads an ALaS module from a JSON file.
func loadModuleFromFile(filename string) (*ast.Module, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	// Validate the JSON first
	if err := validator.ValidateJSON(data); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	// Parse the module
	var module ast.Module
	if err := json.Unmarshal(data, &module); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &module, nil
}

// createFileSystemModuleLoader creates a module loader that searches the file system.
func createFileSystemModuleLoader(basePath string) codegen.ModuleLoader {
	return func(name string) (*ast.Module, error) {
		// Try different possible file locations
		possiblePaths := []string{
			filepath.Join(basePath, name+".alas.json"),
			filepath.Join(basePath, "modules", name+".alas.json"),
			filepath.Join(basePath, "lib", name+".alas.json"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				return loadModuleFromFile(path)
			}
		}

		return nil, fmt.Errorf("module %s not found in any of the search paths", name)
	}
}

// compileLinkedProgram compiles all modules and links them into a single output.
func compileLinkedProgram(multiCodegen *codegen.MultiModuleCodegen, mainModuleName, output, format string, optLevel codegen.OptimizationLevel) error {
	// Compile all modules
	compiledModules, err := multiCodegen.CompileModules()
	if err != nil {
		return fmt.Errorf("failed to compile modules: %v", err)
	}

	fmt.Printf("Compiled %d modules successfully\n", len(compiledModules))

	// Link all modules
	linkedModule, err := multiCodegen.LinkModules(mainModuleName + "_linked")
	if err != nil {
		return fmt.Errorf("failed to link modules: %v", err)
	}

	// Apply optimizations to the linked module
	if optLevel > codegen.OptNone {
		optimizer := codegen.NewOptimizer(optLevel)
		if err := optimizer.OptimizeModule(linkedModule); err != nil {
			return fmt.Errorf("optimization failed: %v", err)
		}
	}

	// Determine output filename
	if output == "" {
		output = mainModuleName + "_linked." + format
	}

	// Write the linked output
	return writeOutput(linkedModule, output, format)
}

// compileSeparateModules compiles each module separately.
func compileSeparateModules(multiCodegen *codegen.MultiModuleCodegen, input, output, format string, optLevel codegen.OptimizationLevel) error {
	// Compile all modules
	compiledModules, err := multiCodegen.CompileModules()
	if err != nil {
		return fmt.Errorf("failed to compile modules: %v", err)
	}

	fmt.Printf("Compiled %d modules successfully\n", len(compiledModules))

	// Write each module separately
	for moduleName, llvmModule := range compiledModules {
		// Apply optimizations
		if optLevel > codegen.OptNone {
			optimizer := codegen.NewOptimizer(optLevel)
			if err := optimizer.OptimizeModule(llvmModule); err != nil {
				return fmt.Errorf("optimization failed for module %s: %v", moduleName, err)
			}
		}

		// Determine output filename for this module
		var moduleOutput string
		if output != "" {
			// Use provided output as base, append module name
			base := strings.TrimSuffix(output, filepath.Ext(output))
			moduleOutput = fmt.Sprintf("%s_%s.%s", base, moduleName, format)
		} else {
			// Use input file as base
			base := strings.TrimSuffix(input, filepath.Ext(input))
			moduleOutput = fmt.Sprintf("%s_%s.%s", base, moduleName, format)
		}

		// Write this module's output
		if err := writeOutput(llvmModule, moduleOutput, format); err != nil {
			return fmt.Errorf("failed to write output for module %s: %v", moduleName, err)
		}

		fmt.Printf("Module %s written to %s\n", moduleName, moduleOutput)
	}

	return nil
}

// writeOutput writes LLVM IR to a file in the specified format.
func writeOutput(llvmModule *ir.Module, output, format string) error {
	moduleStr := llvmModule.String()

	switch format {
	case "ll":
		err := os.WriteFile(output, []byte(moduleStr), 0600)
		if err != nil {
			return fmt.Errorf("error writing LLVM IR: %v", err)
		}
		fmt.Printf("LLVM IR written to %s\n", output)

	case "bc":
		// For bitcode, write IR first and suggest using llvm-as
		llFile := strings.TrimSuffix(output, ".bc") + ".ll"
		err := os.WriteFile(llFile, []byte(moduleStr), 0600)
		if err != nil {
			return fmt.Errorf("error writing LLVM IR: %v", err)
		}
		fmt.Printf("LLVM IR written to %s\n", llFile)
		fmt.Printf("To generate bitcode, run: llvm-as %s -o %s\n", llFile, output)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}
