# ALaS Language Specification (v0.1)

## Overview

ALaS (Artificial Language for Autonomous Systems) is a programming language designed specifically for machine generation and manipulation. Unlike traditional programming languages optimized for human readability, ALaS uses JSON and other non-human-readable representations (with potential for future binary IR formats) to maximize clarity and ease of generation for AI/LLM systems.

## Goals

1. **Machine-First Design**: Prioritize ease of generation, manipulation, and understanding by AI systems
2. **Deterministic Execution**: Ensure predictable, reproducible behavior across all execution environments
3. **Extensibility**: Support a plugin architecture for domain-specific extensions
4. **Safety**: Provide memory safety and type safety guarantees

## Module Structure

Every ALaS program is a module with the following structure:

```json
{
  "type": "module",
  "name": "module_name",
  "exports": ["function1", "function2"],  // Optional
  "imports": ["module1", "module2"],       // Optional
  "functions": [...],                      // Required
  "types": [...],                          // Optional
  "meta": {}                               // Optional metadata
}
```

## Data Types

### Basic Types

- `int` - 64-bit signed integer
- `float` - 64-bit floating point
- `string` - UTF-8 encoded text
- `bool` - Boolean (true/false)
- `void` - No value (for functions that don't return)

### Composite Types

- `array` - Ordered collection of elements
- `map` - Key-value pairs

### Type Examples

```json
// Integer literal
{"type": "literal", "value": 42}

// Float literal
{"type": "literal", "value": 3.14}

// String literal
{"type": "literal", "value": "Hello, World!"}

// Boolean literal
{"type": "literal", "value": true}

// Array literal
{
  "type": "array_literal",
  "elements": [
    {"type": "literal", "value": 1},
    {"type": "literal", "value": 2},
    {"type": "literal", "value": 3}
  ]
}

// Map literal
{
  "type": "map_literal",
  "pairs": [
    {
      "key": {"type": "literal", "value": "name"},
      "value": {"type": "literal", "value": "Alice"}
    }
  ]
}
```

## Functions

Functions are the primary building blocks of ALaS programs:

```json
{
  "type": "function",
  "name": "function_name",
  "params": [
    {"name": "param1", "type": "int"},
    {"name": "param2", "type": "string"}
  ],
  "returns": "int",
  "body": [
    // Statements
  ]
}
```

## Statements

### Assignment Statement

```json
{
  "type": "assign",
  "target": "variable_name",
  "value": {
    // Expression
  }
}
```

### If Statement

```json
{
  "type": "if",
  "cond": {
    // Condition expression
  },
  "then": [
    // Then statements
  ],
  "else": [
    // Else statements (optional)
  ]
}
```

### While Loop

```json
{
  "type": "while",
  "cond": {
    // Condition expression
  },
  "body": [
    // Loop body statements
  ]
}
```

### For Loop

```json
{
  "type": "for",
  "cond": {
    // Condition expression
  },
  "body": [
    // Loop body statements
  ]
}
```

### Return Statement

```json
{
  "type": "return",
  "value": {
    // Return expression (optional)
  }
}
```

### Expression Statement

```json
{
  "type": "expr",
  "value": {
    // Expression
  }
}
```

## Expressions

### Literals

```json
{"type": "literal", "value": 42}
```

### Variables

```json
{"type": "variable", "name": "var_name"}
```

### Binary Operations

```json
{
  "type": "binary",
  "op": "+",
  "left": {"type": "variable", "name": "a"},
  "right": {"type": "literal", "value": 1}
}
```

Supported operators:
- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`
- Logical: `&&`, `||`

### Unary Operations

```json
{
  "type": "unary",
  "op": "!",
  "operand": {"type": "variable", "name": "flag"}
}
```

Supported operators:
- `!` - Logical NOT
- `-` - Negation

### Function Calls

```json
{
  "type": "call",
  "name": "function_name",
  "args": [
    {"type": "literal", "value": 42}
  ]
}
```

### Module Function Calls

```json
{
  "type": "module_call",
  "module": "math",
  "name": "sqrt",
  "args": [
    {"type": "literal", "value": 16}
  ]
}
```

### Builtin Function Calls

```json
{
  "type": "builtin",
  "name": "io.print",
  "args": [
    {"type": "literal", "value": "Hello!"}
  ]
}
```

### Array Indexing

```json
{
  "type": "index",
  "object": {"type": "variable", "name": "array"},
  "index": {"type": "literal", "value": 0}
}
```

### Field Access

```json
{
  "type": "field",
  "object": {"type": "variable", "name": "obj"},
  "field": "field_name"
}
```

## Complete Example

```json
{
  "type": "module",
  "name": "fibonacci",
  "functions": [
    {
      "type": "function",
      "name": "fib",
      "params": [{"name": "n", "type": "int"}],
      "returns": "int",
      "body": [
        {
          "type": "if",
          "cond": {
            "type": "binary",
            "op": "<=",
            "left": {"type": "variable", "name": "n"},
            "right": {"type": "literal", "value": 1}
          },
          "then": [
            {
              "type": "return",
              "value": {"type": "variable", "name": "n"}
            }
          ],
          "else": [
            {
              "type": "return",
              "value": {
                "type": "binary",
                "op": "+",
                "left": {
                  "type": "call",
                  "name": "fib",
                  "args": [{
                    "type": "binary",
                    "op": "-",
                    "left": {"type": "variable", "name": "n"},
                    "right": {"type": "literal", "value": 1}
                  }]
                },
                "right": {
                  "type": "call",
                  "name": "fib",
                  "args": [{
                    "type": "binary",
                    "op": "-",
                    "left": {"type": "variable", "name": "n"},
                    "right": {"type": "literal", "value": 2}
                  }]
                }
              }
            }
          ]
        }
      ]
    },
    {
      "type": "function",
      "name": "main",
      "params": [],
      "returns": "int",
      "body": [
        {
          "type": "return",
          "value": {
            "type": "call",
            "name": "fib",
            "args": [{"type": "literal", "value": 10}]
          }
        }
      ]
    }
  ]
}
```

## Execution Model

ALaS programs execute in a deterministic environment with:
- Function-scoped memory model
- No global mutable state
- Explicit error handling
- Support for compilation to binary IR (via LLVM)

## Design Principles

1. **Machine-First**: Every construct is designed for easy generation and parsing by machines
2. **Explicit**: No implicit conversions or hidden behavior
3. **Structured**: Consistent JSON schema throughout
4. **Deterministic**: Same input always produces same output
5. **Extensible**: Plugin system allows language extensions

## Design Constraints

- **No Implicit Behavior**: All operations must be explicitly defined
- **Type Safety**: Strong static typing with no implicit conversions
- **Memory Safety**: No direct memory manipulation
- **Determinism**: No undefined behavior or platform-specific variations

## Appendix: Formal JSON Schema

ALaS programs conform to the following JSON Schema definition:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://alas-lang.org/schemas/v0.1/program.json",
  "title": "ALaS Program Schema",
  "type": "object",
  "required": ["type", "name", "functions"],
  "properties": {
    "type": {
      "const": "module"
    },
    "name": {
      "type": "string",
      "pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
    },
    "exports": {
      "type": "array",
      "items": {"type": "string"}
    },
    "imports": {
      "type": "array",
      "items": {"type": "string"}
    },
    "functions": {
      "type": "array",
      "items": {"$ref": "#/definitions/function"}
    },
    "types": {
      "type": "array",
      "items": {"$ref": "#/definitions/type"}
    },
    "meta": {
      "type": "object"
    }
  },
  "definitions": {
    "function": {
      "type": "object",
      "required": ["type", "name", "params", "returns", "body"],
      "properties": {
        "type": {"const": "function"},
        "name": {"type": "string"},
        "params": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "type"],
            "properties": {
              "name": {"type": "string"},
              "type": {"type": "string"}
            }
          }
        },
        "returns": {"type": "string"},
        "body": {
          "type": "array",
          "items": {"$ref": "#/definitions/statement"}
        }
      }
    },
    "statement": {
      "type": "object",
      "required": ["type"],
      "oneOf": [
        {"$ref": "#/definitions/assignStatement"},
        {"$ref": "#/definitions/ifStatement"},
        {"$ref": "#/definitions/whileStatement"},
        {"$ref": "#/definitions/forStatement"},
        {"$ref": "#/definitions/returnStatement"},
        {"$ref": "#/definitions/exprStatement"}
      ]
    },
    "expression": {
      "type": "object",
      "required": ["type"],
      "oneOf": [
        {"$ref": "#/definitions/literal"},
        {"$ref": "#/definitions/variable"},
        {"$ref": "#/definitions/binary"},
        {"$ref": "#/definitions/unary"},
        {"$ref": "#/definitions/call"},
        {"$ref": "#/definitions/arrayLiteral"},
        {"$ref": "#/definitions/mapLiteral"},
        {"$ref": "#/definitions/index"},
        {"$ref": "#/definitions/field"}
      ]
    }
  }
}
```