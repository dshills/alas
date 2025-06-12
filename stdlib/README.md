# ALaS Standard Library

This directory contains the standard library modules for the ALaS programming language. Each module is defined as a JSON file following the ALaS language specification.

## Modules

### Core Modules

- **`std.io`** - Input/Output operations (file read/write, console I/O)
- **`std.math`** - Mathematical functions and constants
- **`std.collections`** - Array and map utilities
- **`std.string`** - String manipulation functions
- **`std.type`** - Type conversion and reflection
- **`std.result`** - Error handling with Result types

### Advanced Modules

- **`std.async`** - Concurrent/async execution primitives

## Usage

Standard library modules can be imported in ALaS programs using the `imports` field:

```json
{
  "type": "module",
  "name": "myProgram",
  "imports": ["std.io", "std.math"],
  "functions": [...]
}
```

## Design Principles

1. **Machine-First**: All functions have clear JSON schemas with predictable inputs/outputs
2. **Deterministic**: Operations are pure when possible, side effects are explicit
3. **Error Handling**: Uses Result types for operations that can fail
4. **Type Safety**: All functions have complete type signatures
5. **Modularity**: Each module is self-contained and independently importable

## Implementation Status

All modules are currently defined as ALaS JSON specifications. The runtime implementation of built-in functions (prefixed with module names like `io.readFile`, `math.sin`, etc.) would be implemented in the ALaS runtime/interpreter.

## Built-in Function Naming Convention

Standard library functions delegate to built-in runtime functions using the pattern:
- `io.readFile` → calls built-in `io.readFile`
- `math.sin` → calls built-in `math.sin`
- `collections.arrayLength` → calls built-in `collections.arrayLength`

This allows the standard library to serve as both documentation and a stable API layer over the runtime implementation.