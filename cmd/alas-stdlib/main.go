// Package main provides the CGO exports for the ALaS standard library.
// This is built as a shared library for use with LLVM-compiled programs.
package main

import (
	_ "github.com/dshills/alas/internal/stdlib" // Import for CGO exports
)

func main() {
	// This is required for building a shared library
	// The actual exports are in the stdlib package
}
