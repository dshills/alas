# ALaS - Artificial Language for Autonomous Systems

ALaS is a general-purpose, Turing-complete programming language designed exclusively for AI models to generate, manipulate, and execute. It uses structured JSON representations to enable low-error, high-speed code generation and transformation by LLMs.

## Features

- **Machine-First Design**: Optimized for AI generation, not human readability
- **JSON-Based**: All code is represented as structured JSON following a strict schema
- **Turing-Complete**: Supports functions, conditionals, loops, and recursion
- **Type System**: Basic types including int, float, string, bool, array, and map
- **Module System**: Import/export capabilities with dependency resolution and encapsulation

## Project Structure

```
alas/
├── cmd/
│   ├── alas-validate/   # AST validation tool
│   ├── alas-run/        # Reference interpreter
│   └── alas-compile/    # LLVM IR compiler
├── internal/
│   ├── ast/            # AST type definitions
│   ├── validator/      # AST validation logic
│   ├── interpreter/    # Reference interpreter
│   ├── codegen/        # LLVM IR code generator
│   └── runtime/        # Runtime value types
├── examples/
│   ├── programs/       # Example ALaS programs
│   └── modules/        # Example ALaS modules
├── tests/              # Test suite
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

This creates three binaries in the `bin/` directory:
- `alas-validate` - Validates ALaS JSON programs
- `alas-run` - Executes ALaS programs
- `alas-compile` - Compiles ALaS programs to LLVM IR

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
# Compile to LLVM IR
./bin/alas-compile -file examples/programs/factorial.alas.json

# Compile all examples
make compile-examples
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
- ✅ LLVM IR code generation
- ✅ Basic type system
- ✅ Functions and recursion
- ✅ Control flow (if/else, while loops)
- ✅ Binary and unary operations
- ✅ Arrays and maps with indexing
- ✅ Module imports/exports with dependency resolution
- ✅ Comprehensive test suite

Future work:
- ⏳ Standard library
- ⏳ Plugin system
- ⏳ Advanced LLVM IR optimizations
- ⏳ Runtime garbage collection for arrays/maps
- ⏳ Cross-module LLVM compilation and linking

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.