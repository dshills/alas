# TASK.md

## Recently Completed

- [x] Fix stdlib modules with unsupported types:
  - [x] `io.print` function uses "any" type parameter - RESOLVED: Type system already supported "any", issue was in validation
  - [x] `async.Task` custom type definition not properly handled in LLVM conversion - FIXED: Corrected JSON schema structure
  - [x] Enhanced dependency resolution exposes pre-existing type system limitations - RESOLVED: Updated validator for module names with dots
- [x] Fix missing cases in switch of type runtime.ValueType (exhaustive) in tests/integration_test.go:289 - COMPLETED
- [x] Address integer overflow conversion warnings (gosec) - 5 instances, review for safety - COMPLETED: Added bounds checking
- [x] Set up CI/CD pipeline - COMPLETED: GitHub Actions workflow created
- [x] Document required dependencies (Go 1.24.4, LLVM tools) - COMPLETED: Enhanced README.md

## New Top Priority

- [ ] Fix comment formatting (godot) in internal/runtime/async.go (3 instances)
- [ ] Fix gofmt and goimports formatting in internal/codegen/llvm.go
- [ ] Address unused parameter warnings (unparam) - 2 instances in helper functions
- [ ] Review unused functions (12 functions) - these are comprehensive implementations for future features

## Linting Issues (Non-Critical)
- [ ] Additional formatting and style improvements as needed


## Setup
- [ ] Set up development environment documentation
- [x] Document required dependencies (Go 1.24.4, LLVM tools) - COMPLETED: Enhanced README.md
- [ ] Create developer onboarding guide
- [x] Set up CI/CD pipeline - COMPLETED: GitHub Actions workflow

## Core Language Implementation
- [ ] Implement remaining core language features
  - [ ] Array operations and methods
  - [ ] Map operations and methods
  - [ ] String manipulation functions
  - [ ] Type checking and inference
- [ ] Implement module import/export system
- [ ] Add support for custom type methods
- [ ] Implement plugin system architecture

## Parser and Validator
- [ ] Enhance error messages with line/column information
- [ ] Add schema version compatibility checking
- [ ] Implement strict mode vs permissive mode parsing
- [ ] Add validation for circular dependencies in imports

## Runtime and Interpreter
- [ ] Implement complete standard library
- [ ] Add runtime error handling with stack traces
- [ ] Implement memory management and garbage collection
- [ ] Add debugging support (breakpoints, step-through)
- [ ] Implement runtime type checking

## LLVM Compiler Backend
- [ ] Fix field access compilation for map objects (cannot determine type of object for field access)
- [ ] Complete LLVM codegen for all language features
  - [ ] Array and map operations
  - [ ] String operations
  - [ ] Module linking
  - [ ] Runtime library integration
- [ ] Optimize generated LLVM IR
- [ ] Add support for different target architectures
- [ ] Implement linking with C libraries

## Testing Infrastructure
- [ ] Add integration tests for compiler pipeline
- [ ] Create performance benchmarks
- [ ] Add fuzzing tests for parser robustness
- [ ] Set up test coverage reporting

## Documentation
- [ ] Complete language specification documentation
- [ ] Create user guide for writing ALaS programs
- [ ] Document standard library functions
- [ ] Create examples for common patterns
- [ ] Write contributor guidelines

## Tooling
- [ ] Create language server protocol (LSP) implementation
- [ ] Build package manager for ALaS modules
- [ ] Create code formatter
- [ ] Implement linter with configurable rules
- [ ] Add REPL for interactive development

## Security
- [ ] Implement sandboxed execution environment
- [ ] Add resource limits for runtime execution
- [ ] Validate and sanitize all JSON inputs
- [ ] Implement secure module loading
- [ ] Add cryptographic signature support for modules

## Performance
- [ ] Profile and optimize parser performance
- [ ] Implement caching for compiled modules
- [ ] Optimize runtime memory usage
- [ ] Add JIT compilation support
- [ ] Benchmark against similar languages

## Future Features
- [ ] WebAssembly target support
- [ ] Distributed execution capabilities
- [ ] Visual programming interface generator
- [ ] AI-assisted code generation tools
- [ ] Cross-language interoperability

## Completed Work
- [x] Initial project structure setup
- [x] Basic JSON parser implementation
- [x] Core validator functionality (alas-validate)
- [x] Basic interpreter (alas-run)
- [x] LLVM compiler foundation (alas-compile)
- [x] Support for basic types (int, float, string, bool)
- [x] Function definitions and calls
- [x] Control flow (if/else, while, for loops)
- [x] Basic arithmetic and comparison operations
- [x] Example programs (hello world, factorial, fibonacci)
- [x] Makefile with build/test/run commands
- [x] Custom type definitions and struct support in LLVM
- [x] Complete JSON schema validation for all language constructs
- [x] Comprehensive test suite for all language features (174+ test cases)
- [x] Enhanced validator with identifier validation and builtin namespace support
- [x] Array and map literal validation with structure checking
- [x] Module import/export validation with duplicate detection
- [x] Custom type validation for structs and enums
- [x] Builtin function validation for all namespaces (io, math, string, array, map, collections, type)
- [x] Enhanced error reporting with specific context and indices