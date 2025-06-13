package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/codegen"
	"github.com/dshills/alas/internal/validator"
)

func main() {
	var input string
	var output string
	var format string
	var optLevel string
	flag.StringVar(&input, "file", "", "ALaS JSON file to compile (reads from stdin if not provided)")
	flag.StringVar(&output, "o", "", "Output file (default: input file with .ll extension)")
	flag.StringVar(&format, "format", "ll", "Output format: ll (LLVM IR text) or bc (LLVM bitcode)")
	flag.StringVar(&optLevel, "O", "1", "Optimization level: 0 (none), 1 (basic), 2 (standard), 3 (aggressive)")
	flag.Parse()

	var data []byte
	var err error

	if input == "" {
		// Read from stdin
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Read from file
		data, err = os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", input, err)
			os.Exit(1)
		}
	}

	// Validate the JSON first
	if err := validator.ValidateJSON(data); err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed:\n%v\n", err)
		os.Exit(1)
	}

	// Parse the module
	var module ast.Module
	if err := json.Unmarshal(data, &module); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
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

	// Generate LLVM IR
	codegenInstance := codegen.NewLLVMCodegen()
	llvmModule, err := codegenInstance.GenerateModule(&module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation failed: %v\n", err)
		os.Exit(1)
	}

	// Apply optimizations
	if optimizationLevel > codegen.OptNone {
		optimizer := codegen.NewOptimizer(optimizationLevel)
		if err := optimizer.OptimizeModule(llvmModule); err != nil {
			fmt.Fprintf(os.Stderr, "Optimization failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Determine output filename
	if output == "" {
		if input == "" {
			output = "output." + format
		} else {
			base := strings.TrimSuffix(input, filepath.Ext(input))
			output = base + "." + format
		}
	}

	// Write output
	switch format {
	case "ll":
		err = os.WriteFile(output, []byte(llvmModule.String()), 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing LLVM IR: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("LLVM IR written to %s\n", output)

	case "bc":
		// For bitcode, we would need to use LLVM tools
		// For now, just output the IR and suggest using llvm-as
		llFile := strings.TrimSuffix(output, ".bc") + ".ll"
		err = os.WriteFile(llFile, []byte(llvmModule.String()), 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing LLVM IR: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("LLVM IR written to %s\n", llFile)
		fmt.Printf("To generate bitcode, run: llvm-as %s -o %s\n", llFile, output)

	default:
		fmt.Fprintf(os.Stderr, "Unsupported format: %s\n", format)
		os.Exit(1)
	}
}
