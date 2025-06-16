# Getting Started with ALaS

This guide will help you write your first ALaS programs and understand the basics of the language.

## Installation

### Prerequisites
- Go 1.21 or higher
- LLVM 14+ (optional, for native compilation)

### Building from Source

```bash
git clone https://github.com/dshills/alas.git
cd alas
make build
```

This creates the following binaries in the `bin/` directory:
- `alas-validate` - Validates ALaS programs
- `alas-run` - Interprets ALaS programs
- `alas-compile` - Compiles ALaS to LLVM IR

## Your First Program

Create a file named `hello.alas.json`:

```json
{
  "type": "module",
  "name": "hello",
  "functions": [
    {
      "type": "function",
      "name": "main",
      "params": [],
      "returns": "void",
      "body": [
        {
          "type": "expr",
          "value": {
            "type": "builtin",
            "name": "io.print",
            "args": [
              {"type": "literal", "value": "Hello, ALaS!"}
            ]
          }
        }
      ]
    }
  ]
}
```

### Running Your Program

1. **Validate the syntax:**
   ```bash
   ./bin/alas-validate -file hello.alas.json
   ```

2. **Run with the interpreter:**
   ```bash
   ./bin/alas-run -file hello.alas.json
   ```

3. **Compile to LLVM IR:**
   ```bash
   ./bin/alas-compile -file hello.alas.json -o hello.ll
   ```

## Basic Concepts

### Variables and Types

ALaS is strongly typed. Here's how to work with variables:

```json
{
  "type": "function",
  "name": "variables_demo",
  "params": [],
  "returns": "int",
  "body": [
    {
      "type": "assign",
      "target": "x",
      "value": {"type": "literal", "value": 42}
    },
    {
      "type": "assign",
      "target": "message",
      "value": {"type": "literal", "value": "Hello"}
    },
    {
      "type": "assign",
      "target": "pi",
      "value": {"type": "literal", "value": 3.14159}
    },
    {
      "type": "return",
      "value": {"type": "variable", "name": "x"}
    }
  ]
}
```

### Control Flow

#### If Statements

```json
{
  "type": "if",
  "cond": {
    "type": "binary",
    "op": ">",
    "left": {"type": "variable", "name": "x"},
    "right": {"type": "literal", "value": 0}
  },
  "then": [
    {
      "type": "expr",
      "value": {
        "type": "builtin",
        "name": "io.print",
        "args": [{"type": "literal", "value": "Positive"}]
      }
    }
  ],
  "else": [
    {
      "type": "expr",
      "value": {
        "type": "builtin",
        "name": "io.print",
        "args": [{"type": "literal", "value": "Non-positive"}]
      }
    }
  ]
}
```

#### Loops

While loop example:
```json
{
  "type": "while",
  "cond": {
    "type": "binary",
    "op": "<",
    "left": {"type": "variable", "name": "i"},
    "right": {"type": "literal", "value": 10}
  },
  "body": [
    {
      "type": "assign",
      "target": "i",
      "value": {
        "type": "binary",
        "op": "+",
        "left": {"type": "variable", "name": "i"},
        "right": {"type": "literal", "value": 1}
      }
    }
  ]
}
```

### Functions

Define reusable functions:

```json
{
  "type": "function",
  "name": "add",
  "params": [
    {"name": "a", "type": "int"},
    {"name": "b", "type": "int"}
  ],
  "returns": "int",
  "body": [
    {
      "type": "return",
      "value": {
        "type": "binary",
        "op": "+",
        "left": {"type": "variable", "name": "a"},
        "right": {"type": "variable", "name": "b"}
      }
    }
  ]
}
```

Call functions:
```json
{
  "type": "call",
  "name": "add",
  "args": [
    {"type": "literal", "value": 5},
    {"type": "literal", "value": 3}
  ]
}
```

### Arrays

Create and use arrays:

```json
{
  "type": "assign",
  "target": "numbers",
  "value": {
    "type": "array_literal",
    "elements": [
      {"type": "literal", "value": 1},
      {"type": "literal", "value": 2},
      {"type": "literal", "value": 3}
    ]
  }
}
```

Access array elements:
```json
{
  "type": "index",
  "object": {"type": "variable", "name": "numbers"},
  "index": {"type": "literal", "value": 0}
}
```

## Complete Example: Factorial

Here's a complete program that calculates factorials:

```json
{
  "type": "module",
  "name": "factorial_demo",
  "functions": [
    {
      "type": "function",
      "name": "factorial",
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
            {"type": "return", "value": {"type": "literal", "value": 1}}
          ],
          "else": [
            {
              "type": "return",
              "value": {
                "type": "binary",
                "op": "*",
                "left": {"type": "variable", "name": "n"},
                "right": {
                  "type": "call",
                  "name": "factorial",
                  "args": [{
                    "type": "binary",
                    "op": "-",
                    "left": {"type": "variable", "name": "n"},
                    "right": {"type": "literal", "value": 1}
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
      "returns": "void",
      "body": [
        {
          "type": "assign",
          "target": "result",
          "value": {
            "type": "call",
            "name": "factorial",
            "args": [{"type": "literal", "value": 5}]
          }
        },
        {
          "type": "expr",
          "value": {
            "type": "builtin",
            "name": "io.print",
            "args": [{"type": "variable", "name": "result"}]
          }
        }
      ]
    }
  ]
}
```

## Next Steps

- Explore the [Standard Library Reference](stdlib-reference.md) for built-in functions
- Learn about [modules and code organization](language-spec.md#module-structure)
- Check out more [examples](examples.md)
- Read the [LLM Integration Guide](llm-guide.md) for AI-assisted development