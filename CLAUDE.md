# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ALaS (Artificial Language for Autonomous Systems) is a programming language specifically designed for machine generation and manipulation by AI/LLM systems. The language uses JSON as its primary representation format, making it optimized for programmatic generation rather than human readability.

## Project Structure

This is an early-stage Go project (module: `github.com/dshills/alas`) that will implement the ALaS language specification.

Key files:
- `alas_lang_spec.md` - Complete language specification (v0.1) defining the JSON schema and language design
- `go.mod` - Go module definition (Go 1.24.4)

## Development Commands

- `make build` - Build all binaries (alas-validate, alas-run, alas-compile)
- `make test` - Run the test suite
- `make clean` - Clean build artifacts
- `make run-all-examples` - Run all example programs
- `make compile-examples` - Compile all examples to LLVM IR
- `make validate-example` - Validate the hello.alas.json example
- `make run-example` - Run the hello.alas.json example
- `./bin/alas-validate -file <file>` - Validate an ALaS program
- `./bin/alas-run -file <file> -fn <function> [args...]` - Run an ALaS program with optional arguments (default function: main)
- `./bin/alas-compile -file <file> -o <output> -format <ll|bc>` - Compile ALaS to LLVM IR

## Architecture Overview

ALaS is designed around these core concepts:

1. **Module System**: Programs are organized as modules containing functions, types, and metadata
2. **JSON Representation**: All code is represented as structured JSON following a strict schema
3. **Core Language Elements**:
   - Modules with import/export capabilities
   - Functions with parameters, return types, and bodies
   - Statements: assignments, conditionals (`if`/`else`), loops (`while`, `for`)
   - Expressions: literals, variables, operations, function calls
   - Types: basic types (int, float, string, bool, array, map) and custom types

4. **Execution Model**:
   - Deterministic execution
   - Function-scoped memory model
   - No global mutable state
   - Eventually compilable to binary IR

## Implementation Considerations

When implementing features for ALaS:

1. **Follow the JSON Schema**: All language constructs must conform to the schema defined in `alas_lang_spec.md`
2. **Machine-First Design**: Prioritize programmatic clarity over human readability
3. **Modularity**: Keep components (parser, compiler, runtime) separate and well-defined
4. **Testing**: Given the language's deterministic nature, comprehensive testing of all language constructs is essential
5. **Error Handling**: Provide structured error messages that can be easily parsed by AI systems

## LLVM Backend

The project includes an LLVM IR code generator:

- **Location**: `internal/codegen/llvm.go`
- **Features**: Functions, recursion, arithmetic, conditionals, loops
- **Output**: LLVM IR (.ll files) that can be compiled to native binaries
- **Usage**: `./bin/alas-compile -file program.alas.json -o output.ll`

To compile LLVM IR to native code:
```bash
# Generate LLVM IR
./bin/alas-compile -file examples/programs/factorial.alas.json

# Compile to object file
llc factorial.ll -o factorial.o

# Link to executable (may need runtime library)
clang factorial.o -o factorial
```

## Key Design Principles

- **Not Human-Readable**: The language intentionally uses JSON/binary formats rather than text syntax
- **Schema-Defined**: Every language construct has a precise JSON schema definition
- **AI-Optimized**: Designed for easy generation and manipulation by language models
- **Extensible**: Plugin system planned for future extensions
- **Introspectable**: Programs can analyze and modify themselves

## Development Best Practices

- Always run golangci-lint and fix issues