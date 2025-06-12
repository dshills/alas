package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/interpreter"
	"github.com/dshills/alas/internal/runtime"
	"github.com/dshills/alas/internal/validator"
)

func main() {
	var input string
	var function string
	flag.StringVar(&input, "file", "", "ALaS JSON file to run (reads from stdin if not provided)")
	flag.StringVar(&function, "fn", "main", "Function to execute (default: main)")
	flag.Parse()

	// Get function arguments from remaining command line args
	args := flag.Args()

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

	// Create interpreter and load module
	interp := interpreter.New()
	if err := interp.LoadModule(&module); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading module: %v\n", err)
		os.Exit(1)
	}

	// Parse arguments into runtime values
	runtimeArgs := make([]runtime.Value, len(args))
	for i, arg := range args {
		// Try to parse as int first, then float, then string
		if val, err := strconv.ParseInt(arg, 10, 64); err == nil {
			runtimeArgs[i] = runtime.NewInt(val)
		} else if val, err := strconv.ParseFloat(arg, 64); err == nil {
			runtimeArgs[i] = runtime.NewFloat(val)
		} else if val, err := strconv.ParseBool(arg); err == nil {
			runtimeArgs[i] = runtime.NewBool(val)
		} else {
			runtimeArgs[i] = runtime.NewString(arg)
		}
	}

	// Execute the specified function
	result, err := interp.Run(function, runtimeArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}

	// Print result if not void
	if result.Type != runtime.ValueTypeVoid {
		fmt.Println(result.String())
	}
}
