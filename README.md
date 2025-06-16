# ALaS - Artificial Language for Autonomous Systems

ALaS is a general-purpose, Turing-complete programming language designed exclusively for AI models to generate, manipulate, and execute. It uses structured JSON representations to enable low-error, high-speed code generation and transformation by LLMs.

**For LLM Users**: Check out our comprehensive [documentation](docs/README.md) designed specifically to help AI systems generate ALaS code effectively. The [LLM Integration Guide](docs/llm-guide.md) provides optimal prompting strategies and common patterns.

## Features

- **Machine-First Design**: Optimized for AI generation, not human readability
- **JSON-Based**: All code is represented as structured JSON following a strict schema with comprehensive validation
- **Turing-Complete**: Supports functions, conditionals, loops, and recursion
- **Type System**: Basic types including int, float, string, bool, array, and map with custom struct and enum support
- **Enhanced Validation**: Comprehensive JSON schema validation for all language constructs with detailed error reporting
- **Module System**: Import/export capabilities with dependency resolution and encapsulation
- **Standard Library**: Comprehensive runtime implementation for I/O, math, collections, strings, and more
- **Plugin System**: Dynamic extensibility with security, sandboxing, and multiple plugin types

## Project Structure

```
alas/
├── cmd/
│   ├── alas-validate/      # AST validation tool
│   ├── alas-run/           # Reference interpreter
│   ├── alas-compile/       # Single-module LLVM IR compiler
│   ├── alas-compile-multi/ # Multi-module LLVM IR compiler with linking
│   ├── alas-plugin/        # Plugin management tool
│   └── alas-stdlib/        # Standard library shared object builder
├── internal/
│   ├── ast/               # AST type definitions
│   ├── validator/         # AST validation logic
│   ├── interpreter/       # Reference interpreter
│   ├── codegen/           # LLVM IR code generator, optimizer, and multi-module system
│   ├── plugin/            # Plugin system implementation
│   ├── runtime/           # Runtime value types
│   └── stdlib/            # Standard library runtime implementation
├── stdlib/                # Standard library modules
├── examples/
│   ├── programs/          # Example ALaS programs
│   ├── modules/           # Example ALaS modules (math_utils, format_utils)
│   └── plugins/           # Example plugin implementations
├── tests/                 # Test suite with optimization and multi-module tests
└── docs/                  # Comprehensive documentation
    ├── README.md          # Documentation overview
    ├── language-spec.md   # Complete language specification
    ├── getting-started.md # Quick start guide
    ├── stdlib-reference.md # Standard library reference
    ├── plugin-system.md   # Plugin development guide
    ├── examples.md        # Code examples and patterns
    ├── llm-guide.md       # LLM integration guide
    └── troubleshooting.md # Common issues and solutions
```

## Documentation

- **[Getting Started Guide](docs/getting-started.md)** - Quick introduction to writing ALaS programs
- **[Language Specification](docs/language-spec.md)** - Complete language reference
- **[Standard Library Reference](docs/stdlib-reference.md)** - All built-in functions
- **[Plugin System Guide](docs/plugin-system.md)** - Extending ALaS
- **[Examples](docs/examples.md)** - Sample programs and patterns
- **[LLM Integration Guide](docs/llm-guide.md)** - For AI-assisted development
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

## Getting Started

### Prerequisites

- Go 1.24.4 or later

### Building

```bash
make build
```

This creates six binaries in the `bin/` directory:
- `alas-validate` - Validates ALaS JSON programs
- `alas-run` - Executes ALaS programs
- `alas-compile` - Compiles single ALaS programs to LLVM IR
- `alas-compile-multi` - Compiles multi-module ALaS programs with cross-module linking
- `alas-plugin` - Manages plugins (list, install, create, etc.)
- `alas-stdlib` - Builds standard library as shared object

### Running Examples

```bash
# Run all examples
make run-all-examples

# Run a specific example
./bin/alas-run -file examples/programs/hello.alas.json

# Run array example
./bin/alas-run -file examples/programs/simple_array.alas.json

# Run module example
./bin/alas-run -file examples/programs/module_demo.alas.json

# Run a specific function with arguments (default function is 'main')
./bin/alas-run -file examples/programs/fibonacci.alas.json -fn main
```

### Validating Programs

ALaS includes comprehensive JSON schema validation for all language constructs:

```bash
# Validate a single program
./bin/alas-validate -file examples/programs/hello.alas.json

# Validation includes:
# - Module structure and naming
# - Function definitions and parameters
# - Statement and expression syntax
# - Type definitions and usage
# - Import/export validation
# - Builtin function namespace checking
# - Custom type validation (structs/enums)
```

### Compiling to LLVM IR

```bash
# Single-module compilation
./bin/alas-compile -file examples/programs/factorial.alas.json

# Multi-module compilation with cross-module linking
./bin/alas-compile-multi -file examples/programs/module_demo.alas.json -module-path examples

# Compile with optimizations
./bin/alas-compile -file examples/programs/factorial.alas.json -O 2

# Multi-module linking modes
./bin/alas-compile-multi -file examples/programs/module_demo.alas.json -module-path examples -link all -o linked_program.ll

# Available optimization levels:
# -O 0  No optimizations (default)
# -O 1  Basic optimizations (constant folding, dead code elimination)
# -O 2  Standard optimizations (includes mem2reg, common subexpression elimination)
# -O 3  Aggressive optimizations (includes function inlining, loop optimizations)

# Compile all examples
make compile-examples

# Plugin management
make plugin-list
make validate-plugins
```

### Running Tests

```bash
make test
```

## Example Programs

ALaS programs are written in structured JSON format. Here's a simple "Hello, World!" example:

```json
{
  "type": "module",
  "name": "hello",
  "functions": [
    {
      "name": "main",
      "params": [],
      "returns": "string",
      "body": [
        {
          "type": "return",
          "value": {
            "type": "literal",
            "value": "Hello, ALaS!"
          }
        }
      ]
    }
  ]
}
```

See **[Examples Documentation](docs/examples.md)** for more comprehensive examples including arrays, maps, functions, modules, and advanced patterns.

## Standard Library

ALaS includes a comprehensive standard library with native runtime implementations for all core functionality including I/O, math, collections, strings, type checking, error handling, and async programming.

See the **[Standard Library Reference](docs/stdlib-reference.md)** for complete documentation of all modules and functions.

## Plugin System

ALaS features a comprehensive plugin system that enables dynamic extension of the language while maintaining security and type safety. The system supports multiple plugin types including module, native, hybrid, and built-in plugins with sandboxing and capability-based security.

See the **[Plugin System Guide](docs/plugin-system.md)** for complete documentation on plugin development, management, and security features.

## LLVM Compilation

ALaS compiles to LLVM IR with multi-level optimization support (O0-O3) providing significant performance improvements. The system supports both single-module and cross-module compilation with dependency resolution and linking.


## Language Features

ALaS is a Turing-complete language with functions, conditionals, loops, recursion, arrays, maps, and a module system with import/export capabilities.

See the **[Language Specification](docs/language-spec.md)** for complete details on all language constructs, types, and syntax.

## Development Status

Current implementation includes:
- ✅ AST definition and validation
- ✅ Reference interpreter
- ✅ LLVM IR code generation with multi-level optimization
- ✅ Basic type system
- ✅ Functions and recursion
- ✅ Control flow (if/else, while loops)
- ✅ Binary and unary operations
- ✅ Arrays and maps with indexing
- ✅ Module imports/exports with dependency resolution
- ✅ Standard library runtime implementation (7 core modules)
- ✅ Plugin system with security and multi-type support
- ✅ Comprehensive test suite with optimization testing
- ✅ Runtime garbage collection for arrays/maps (reference counting)
- ✅ LLVM backend support for builtin expressions (standard library in compiled code)

Recent additions:
- ✅ **LLVM IR Optimization System** - Complete multi-level optimization framework
  - **O0**: No optimizations (baseline)
  - **O1**: Basic optimizations (constant folding, dead code elimination, mem2reg)
  - **O2**: Standard optimizations (adds common subexpression elimination, CFG simplification)
  - **O3**: Aggressive optimizations (adds function inlining, loop invariant code motion)
- ✅ **Optimization Test Suite** - Unit tests, benchmarks, and integration tests
- ✅ **Performance Improvements** - 16-63% code size reduction with optimizations
- ✅ **Cross-module LLVM Compilation and Linking** - Complete multi-module compilation system
  - **Dependency Resolution**: Topological sorting with circular dependency detection
  - **External Function Declarations**: Proper LLVM IR with qualified function names
  - **Module Loaders**: Flexible module discovery and loading system
  - **Linking Modes**: Both separate compilation and whole-program linking
  - **Multi-Module CLI**: New `alas-compile-multi` tool with comprehensive options
- ✅ **Standard Library Runtime Implementation** - Native Go implementations for all core modules
  - **std.io**: File operations and console I/O with Result types
  - **std.math**: Mathematical functions with cryptographically secure random
  - **std.collections**: Array and map utilities with slice operations
  - **std.string**: String manipulation with split/join/replace
  - **std.type**: Type checking and conversion utilities
  - **std.result**: Structured error handling pattern
  - **std.async**: Concurrent/async programming with tasks, parallel, race, and cancellation
- ✅ **Runtime Garbage Collection** - Reference counting GC for arrays and maps
  - **Reference Counting**: Automatic memory management with retain/release
  - **Nested Object Support**: Proper cleanup of nested arrays/maps
  - **Automatic Cleanup**: Variables release old values on reassignment
  - **Function Cleanup**: Local GC objects released on function return
  - **GC Threshold**: Automatic collection when object count exceeds limit
- ✅ **LLVM Builtin Support** - Standard library functions in compiled code
  - **I/O Functions**: io.print
  - **Math Functions**: math.sqrt, math.abs
  - **Collection Functions**: collections.length
  - **String Functions**: string.toUpper
  - **Type Functions**: type.typeOf

- ✅ **Enhanced LLVM Codegen and Error Handling** - Comprehensive language feature completion
  - **Dynamic Field Access**: Fixed field access compilation for dynamically-typed objects
  - **Complete Array Operations**: Array element assignment, length, slicing with bounds checking
  - **Complete Map Operations**: Map operations (get, put, contains, remove, keys, values)
  - **String Functions**: All string manipulation functions (substring, indexOf, split, join, replace, etc.)
  - **Enhanced Module System**: Module caching, dependency resolution, type imports
  - **Runtime Error Handling**: Division by zero checks, bounds checking, null pointer checks, assertions

- ✅ **std.async Module Implementation** - Full async/concurrent programming support
  - **Task System**: Spawn async tasks with context-based cancellation
  - **Synchronization**: await, awaitTimeout for task completion
  - **Concurrency Patterns**: parallel (wait for all), race (first to complete)
  - **Utilities**: sleep, timeout, cancel, isRunning, isCompleted
  - **Error Handling**: Result-based error propagation for async operations

Future work (Priority order):
- ⏳ **Plugin marketplace and hot reloading** - Dynamic plugin management
  - Hot reload capability for development
  - Remote plugin repository and marketplace
  - Automated plugin installation/updates
- ⏳ **Additional optimization passes** - Advanced compiler optimizations
  - Vectorization/auto-vectorization for SIMD operations
  - Dead store elimination (DSE)
  - Loop unrolling
  - Global value numbering
  - Instruction combining

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.