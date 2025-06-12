package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dshills/alas/internal/validator"
)

func main() {
	var input string
	flag.StringVar(&input, "file", "", "ALaS JSON file to validate (reads from stdin if not provided)")
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

	// Validate the JSON
	if err := validator.ValidateJSON(data); err != nil {
		fmt.Fprintf(os.Stderr, "Validation failed:\n%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Validation successful!")
}
