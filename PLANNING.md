# PLANNING.md

## Project Overview
ALaS (Artificial Language for Autonomous Systems) is a programming language specifically designed for machine generation and manipulation by AI/LLM systems. The language uses JSON as its primary representation format, making it optimized for programmatic generation rather than human readability. ALaS supports both interpreted execution and compilation to native binaries via LLVM IR, with a comprehensive standard library and plugin system.

## Architecture

### Core Components (API, Data, Service layers, configuration, etc)
**Executables (cmd/):**
- `alas-validate`: JSON schema validation for ALaS programs
- `alas-run`: Interpreter that executes ALaS programs with argument support
- `alas-compile`: LLVM IR code generator with optimization levels (0-3)
- `alas-plugin`: Plugin management system
- `alas-compile-multi`: Multi-module compilation support
- `alas-stdlib`: Standard library as shared library builder

**Internal Architecture (internal/):**
- **AST Layer**: JSON-based module structure with functions, types, and metadata
- **Code Generation**: LLVM IR backend with full language support and optimization pipeline
- **Interpreter**: Runtime execution engine with custom type support
- **Runtime System**: Value system with garbage collection and async execution
- **Standard Library**: Comprehensive builtin functions (IO, Math, Collections, String, etc.)
- **Plugin System**: Plugin manifest management with security controls
- **Validation**: JSON schema validation against language specification

### Data Model
ALaS uses a JSON-based data model with the following core structures:
- **Module**: Top-level container with functions, types, imports, and metadata
- **Function**: Parameters, return type, and statement body
- **Statement Types**: assignments, conditionals (if/else), loops (while/for), returns
- **Expression Types**: literals, variables, binary/unary operations, function calls
- **Type System**: Basic types (int, float, string, bool, array, map) and custom types
- **Runtime Values**: Garbage-collected value system with type safety

## API Endpoints
ALaS is a command-line language system and does not expose web API endpoints. Interaction occurs through:
- Command-line executables (`alas-validate`, `alas-run`, `alas-compile`)
- JSON file input/output for program representation
- Plugin system for extensibility
- Standard library functions for I/O operations

## Technology Stack (Language, frameworks, etc)
**Core Language:** Go 1.24.4
**Key Dependencies:**
- LLVM IR generation: `github.com/llir/llvm v0.3.6`
- Float handling: `github.com/mewmew/float`
- Error handling: `github.com/pkg/errors`
- Go toolchain: `golang.org/x/mod`, `golang.org/x/tools`

**Build System:** Make
**Compilation Target:** LLVM IR → Native binaries (via clang)
**Testing:** Go's built-in testing framework

## Project Structure
```
alas/
├── cmd/                    # Command-line executables
│   ├── alas-validate/      # JSON schema validator
│   ├── alas-run/          # Interpreter
│   ├── alas-compile/      # LLVM compiler
│   ├── alas-plugin/       # Plugin manager
│   └── alas-stdlib/       # Standard library builder
├── internal/              # Internal packages
│   ├── ast/              # Abstract syntax tree
│   ├── codegen/          # LLVM code generation
│   ├── interpreter/      # Runtime interpreter
│   ├── runtime/          # Runtime system & GC
│   ├── stdlib/           # Standard library
│   ├── plugin/           # Plugin system
│   └── validator/        # Schema validation
├── examples/             # Example ALaS programs
├── tests/               # Test suite
├── docs/                # Documentation
├── plugins/             # Plugin examples
└── Makefile            # Build system
```

## Testing Strategy
**Comprehensive Test Suite:**
- Unit tests for all internal packages
- Integration tests for interpreter execution
- LLVM code generation tests with IR validation
- Standard library function tests
- Plugin system tests with security validation
- Performance benchmarks and memory usage tracking
- Edge case and error handling tests

**Test Execution:**
- `make test`: Run full test suite with failure reporting
- Memory usage tracking during test execution
- Cleanup of generated files (.ll, executables) after tests
- Continuous integration ready (lint + test pipeline)

## Development Commands
**Build Commands:**
- `make build`: Build all executables
- `make clean`: Clean build artifacts
- `make build-stdlib`: Create shared library for standard library

**Testing Commands:**
- `make test`: Run comprehensive test suite
- `make run-all-examples`: Execute all example programs
- `make compile-examples`: Generate LLVM IR for examples

**Compilation Pipeline:**
- `make compile-to-native`: Full native compilation pipeline
- `./bin/alas-validate -file <file>`: Validate ALaS program
- `./bin/alas-run -file <file> -fn <function> [args...]`: Run program
- `./bin/alas-compile -file <file> -o <output> -format <ll|bc>`: Compile to LLVM IR

**Plugin Management:**
- `make plugin-list`: List available plugins
- `make plugin-create`: Create new plugin
- `make plugin-validate`: Validate plugin manifests

## Environment Setup
**Prerequisites:**
- Go 1.24.4 or later
- LLVM tools (llc, clang) for native compilation
- Make for build system

**Development Setup:**
1. Clone repository: `git clone github.com/dshills/alas`
2. Install dependencies: `go mod download`
3. Build tools: `make build`
4. Run tests: `make test`
5. Try examples: `make run-all-examples`

**IDE Setup:**
- Go language server support
- JSON schema validation for .alas.json files
- LLVM IR syntax highlighting recommended

## Development Guidelines
**Code Standards:**
- Follow Go best practices and idioms
- Run `golangci-lint` and fix all issues before commit
- Maintain comprehensive test coverage
- Document public APIs with Go doc comments

**Workflow:**
- Never commit directly to main branch
- Create feature branches for all changes
- Ensure all tests pass before PR submission
- Clean up generated files (*.ll, executables) after testing

**Language Design:**
- Maintain JSON schema compatibility
- Prioritize programmatic clarity over human readability
- Ensure deterministic execution behavior
- Keep components (parser, compiler, runtime) modular

## Security Considerations
**Input Validation:**
- Strict JSON schema validation for all program inputs
- Sanitization of file paths and command arguments
- Resource limits for runtime execution

**Plugin Security:**
- Capability-based access control for plugins
- Plugin manifest validation and signing
- Sandboxed plugin execution environment
- Security auditing for plugin operations

**Runtime Security:**
- Memory safety through garbage collection
- Stack overflow protection
- Controlled system resource access
- Secure module loading and dependency resolution

## Future Considerations
**Language Evolution:**
- WebAssembly compilation target
- Distributed execution capabilities
- Enhanced type system with generics
- Interactive debugging support

**Tooling Expansion:**
- Language Server Protocol (LSP) implementation
- Visual programming interface generator
- Package manager for ALaS modules
- AI-assisted code generation tools

**Performance Optimization:**
- Just-in-time (JIT) compilation
- Advanced optimization passes
- Parallel execution support
- Memory usage optimization

**Ecosystem Development:**
- Cross-language interoperability (C, Python, etc.)
- Cloud execution environment
- Standard library expansion
- Community plugin marketplace