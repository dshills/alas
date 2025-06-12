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
	flag.StringVar(&input, "file", "", "ALaS JSON file to compile (reads from stdin if not provided)")
	flag.StringVar(&output, "o", "", "Output file (default: input file with .ll extension)")
	flag.StringVar(&format, "format", "ll", "Output format: ll (LLVM IR text) or bc (LLVM bitcode)")
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

	// Generate LLVM IR
	codegen := codegen.NewLLVMCodegen()
	llvmModule, err := codegen.GenerateModule(&module)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Code generation failed: %v\n", err)
		os.Exit(1)
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
