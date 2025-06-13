# ALaS - Artificial Language for Autonomous Systems

ALaS is a general-purpose, Turing-complete programming language designed exclusively for AI models to generate, manipulate, and execute. It uses structured JSON representations to enable low-error, high-speed code generation and transformation by LLMs.

## Features

- **Machine-First Design**: Optimized for AI generation, not human readability
- **JSON-Based**: All code is represented as structured JSON following a strict schema
- **Turing-Complete**: Supports functions, conditionals, loops, and recursion
- **Type System**: Basic types including int, float, string, bool, array, and map
- **Module System**: Import/export capabilities with dependency resolution and encapsulation
- **Standard Library**: Comprehensive set of modules for I/O, math, collections, strings, and more
- **Plugin System**: Dynamic extensibility with security, sandboxing, and multiple plugin types

## Project Structure

```
alas/
├── cmd/
│   ├── alas-validate/   # AST validation tool
│   ├── alas-run/        # Reference interpreter
│   ├── alas-compile/    # LLVM IR compiler with optimization
│   └── alas-plugin/     # Plugin management tool
├── internal/
│   ├── ast/            # AST type definitions
│   ├── validator/      # AST validation logic
│   ├── interpreter/    # Reference interpreter
│   ├── codegen/        # LLVM IR code generator and optimizer
│   ├── plugin/         # Plugin system implementation
│   └── runtime/        # Runtime value types
├── stdlib/             # Standard library modules
├── examples/
│   ├── programs/       # Example ALaS programs
│   ├── modules/        # Example ALaS modules
│   └── plugins/        # Example plugin implementations
├── tests/              # Test suite with optimization tests
└── docs/
    └── alas_lang_spec.md  # Language specification
```

## Getting Started

### Prerequisites

- Go 1.24.4 or later

### Building

```bash
make build
```

This creates four binaries in the `bin/` directory:
- `alas-validate` - Validates ALaS JSON programs
- `alas-run` - Executes ALaS programs
- `alas-compile` - Compiles ALaS programs to LLVM IR
- `alas-plugin` - Manages plugins (list, install, create, etc.)

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

```bash
./bin/alas-validate -file examples/programs/hello.alas.json
```

### Compiling to LLVM IR

```bash
# Compile to LLVM IR (unoptimized)
./bin/alas-compile -file examples/programs/factorial.alas.json

# Compile with optimizations
./bin/alas-compile -file examples/programs/factorial.alas.json -O 2

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

### Hello World

Here's a simple "Hello, World!" program in ALaS:

```json
{
  "type": "module",
  "name": "hello",
  "functions": [
    {
      "type": "function",
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

### Arrays and Maps

Here's an example demonstrating array and map operations:

```json
{
  "type": "module",
  "name": "arrays_demo",
  "functions": [
    {
      "type": "function",
      "name": "main",
      "params": [],
      "returns": "int",
      "body": [
        {
          "type": "assign",
          "target": "numbers",
          "value": {
            "type": "array_literal",
            "elements": [
              {"type": "literal", "value": 10},
              {"type": "literal", "value": 20},
              {"type": "literal", "value": 30}
            ]
          }
        },
        {
          "type": "assign",
          "target": "person",
          "value": {
            "type": "map_literal",
            "pairs": [
              {
                "key": {"type": "literal", "value": "name"},
                "value": {"type": "literal", "value": "Alice"}
              },
              {
                "key": {"type": "literal", "value": "age"},
                "value": {"type": "literal", "value": 30}
              }
            ]
          }
        },
        {
          "type": "return",
          "value": {
            "type": "index",
            "object": {"type": "variable", "name": "numbers"},
            "index": {"type": "literal", "value": 1}
          }
        }
      ]
    }
  ]
}
```

## Standard Library

ALaS includes a comprehensive standard library with the following modules:

- **`std.io`** - File operations and console I/O (readFile, writeFile, print, readLine)
- **`std.math`** - Mathematical functions and constants (sin, cos, sqrt, PI, E, etc.)
- **`std.collections`** - Array and map utilities (filter, map, reduce, sort, etc.)
- **`std.string`** - String manipulation (split, join, replace, format, etc.)
- **`std.type`** - Type checking and conversion (typeOf, parseInt, toString, etc.)
- **`std.result`** - Structured error handling with Result types
- **`std.async`** - Concurrent execution primitives (spawn, await, parallel, etc.)

Standard library modules can be imported like any other module:

```json
{
  "type": "module",
  "name": "myProgram",
  "imports": ["std.io", "std.math"],
  "functions": [
    {
      "name": "main",
      "body": [
        {
          "type": "assign",
          "target": "data",
          "value": {
            "type": "module_call",
            "module": "std.io",
            "name": "readFile",
            "args": [{"type": "literal", "value": "data.txt"}]
          }
        }
      ]
    }
  ]
}
```

See the [stdlib/README.md](stdlib/README.md) for complete documentation.

## Plugin System

ALaS features a comprehensive plugin system that enables dynamic extension of the language while maintaining security and type safety. The plugin system supports multiple plugin types and provides a rich development and management experience.

### Plugin Types

- **Module Plugins** - Pure ALaS implementations for business logic and algorithms
- **Native Plugins** - Compiled shared libraries for performance-critical operations  
- **Hybrid Plugins** - Combination of ALaS modules with native function implementations
- **Built-in Plugins** - Runtime-integrated plugins for core system functionality

### Plugin Management

```bash
# List available plugins
./bin/alas-plugin list -path examples/plugins

# Get detailed plugin information
./bin/alas-plugin info -path examples/plugins hello-world

# Create a new plugin from template
./bin/alas-plugin create my-plugin

# Validate plugin manifest
./bin/alas-plugin validate plugin.json

# Load/unload plugins at runtime
./bin/alas-plugin load my-plugin
./bin/alas-plugin unload my-plugin
```

### Example Plugin Usage

```json
{
  "type": "module",
  "name": "my_program", 
  "imports": ["hello"],
  "functions": [
    {
      "name": "main",
      "body": [
        {
          "type": "assign",
          "target": "greeting",
          "value": {
            "type": "module_call",
            "module": "hello",
            "name": "greet",
            "args": [{"type": "literal", "value": "World"}]
          }
        }
      ]
    }
  ]
}
```

### Security Features

- **Sandboxing** - Isolated execution environments with resource limits
- **Capability System** - Fine-grained permission control
- **Resource Monitoring** - Memory, CPU, and timeout limits
- **Validation** - Comprehensive manifest and dependency validation

See the [docs/plugin_system.md](docs/plugin_system.md) for complete plugin development guide.

## LLVM IR Optimization

ALaS includes a sophisticated LLVM IR optimization system that provides multiple optimization levels to balance compilation speed and code performance. The optimizer applies various passes to reduce code size and improve execution speed.

### Optimization Levels

| Level | Description | Optimizations Applied |
|-------|-------------|----------------------|
| **O0** | No optimizations | Baseline compilation for debugging |
| **O1** | Basic optimizations | • Constant folding<br>• Dead code elimination<br>• mem2reg (promote memory to registers) |
| **O2** | Standard optimizations | O1 optimizations plus:<br>• Common subexpression elimination<br>• Control flow graph simplification |
| **O3** | Aggressive optimizations | O2 optimizations plus:<br>• Function inlining<br>• Loop invariant code motion |

### Optimization Performance

The optimizer achieves significant code size reductions:
- **Constant-heavy code**: 10-25% reduction
- **Dead code scenarios**: 16-30% reduction  
- **Function call patterns**: 5-15% reduction with inlining
- **Complex algorithms**: 20-63% overall reduction

### Usage Examples

```bash
# Compile with different optimization levels
./bin/alas-compile -file program.alas.json -O 0  # No optimization
./bin/alas-compile -file program.alas.json -O 1  # Basic optimization
./bin/alas-compile -file program.alas.json -O 2  # Standard optimization
./bin/alas-compile -file program.alas.json -O 3  # Aggressive optimization

# Generate LLVM bitcode instead of text IR
./bin/alas-compile -file program.alas.json -O 2 -format bc -o program.bc

# Compile to native executable (requires LLVM tools)
./bin/alas-compile -file program.alas.json -O 2 -o program.ll
llc program.ll -o program.o
clang program.o -o program
```

### Testing and Validation

The optimization system includes comprehensive testing:
- **Unit tests**: Verify individual optimization passes
- **Integration tests**: Test full compilation pipeline with example programs
- **Benchmark tests**: Measure optimization effectiveness
- **Regression tests**: Ensure optimizations don't break functionality

Run optimization tests:
```bash
# Run all optimizer tests
go test ./tests -run TestOptimizer

# Run optimization benchmarks
go test ./tests -bench=BenchmarkOptimizer

# Test optimization effectiveness
go test ./tests -run TestOptimizationEffectiveness -v
```

## Language Features

### Statements
- `assign` - Variable assignment
- `if` - Conditional execution with optional else
- `while` - Loop while condition is true
- `return` - Return from function
- `expr` - Expression statement

### Expressions
- `literal` - Literal values (int, float, string, bool)
- `variable` - Variable reference
- `binary` - Binary operations (+, -, *, /, %, ==, !=, <, <=, >, >=, &&, ||)
- `unary` - Unary operations (!, -)
- `call` - Function calls
- `module_call` - Cross-module function calls (module.function)
- `array_literal` - Array literals with elements
- `map_literal` - Map literals with key-value pairs
- `index` - Array/map indexing operations

### Types
- `int` - Integer numbers
- `float` - Floating-point numbers
- `string` - Text strings
- `bool` - Boolean values (true/false)
- `array` - Arrays of values with integer indexing
- `map` - Key-value maps with string keys
- `void` - No return value

### Module System

ALaS supports modular programming with import/export capabilities:

- **Imports**: Declare dependencies on other modules using the `imports` array
- **Exports**: Specify which functions are accessible from other modules using the `exports` array
- **Module Calls**: Call exported functions using `module.function` syntax
- **Dependency Resolution**: Modules are automatically loaded when imported
- **Encapsulation**: Non-exported functions remain private to the module

Example:
```json
{
  "type": "module",
  "name": "main",
  "imports": ["math_utils"],
  "functions": [{
    "name": "calculate",
    "body": [{
      "type": "assign",
      "target": "result",
      "value": {
        "type": "module_call",
        "module": "math_utils",
        "name": "add",
        "args": [
          {"type": "literal", "value": 10},
          {"type": "literal", "value": 5}
        ]
      }
    }]
  }]
}
```

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
- ✅ Standard library specification (8 core modules)
- ✅ Plugin system with security and multi-type support
- ✅ Comprehensive test suite with optimization testing

Recent additions:
- ✅ **LLVM IR Optimization System** - Complete multi-level optimization framework
  - **O0**: No optimizations (baseline)
  - **O1**: Basic optimizations (constant folding, dead code elimination, mem2reg)
  - **O2**: Standard optimizations (adds common subexpression elimination, CFG simplification)
  - **O3**: Aggressive optimizations (adds function inlining, loop invariant code motion)
- ✅ **Optimization Test Suite** - Unit tests, benchmarks, and integration tests
- ✅ **Performance Improvements** - 16-63% code size reduction with optimizations

Future work:
- ⏳ Runtime garbage collection for arrays/maps
- ⏳ Cross-module LLVM compilation and linking
- ⏳ Standard library runtime implementation
- ⏳ Plugin marketplace and hot reloading
- ⏳ Additional optimization passes (vectorization, dead store elimination)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.