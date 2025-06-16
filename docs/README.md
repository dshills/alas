# ALaS Documentation

Welcome to the ALaS (Artificial Language for Autonomous Systems) documentation. ALaS is a programming language specifically designed for machine generation and manipulation by AI/LLM systems.

## Documentation Overview

- **[Language Specification](language-spec.md)** - Complete ALaS language reference
- **[Getting Started Guide](getting-started.md)** - Quick introduction to writing ALaS programs
- **[Standard Library Reference](stdlib-reference.md)** - Built-in functions and modules
- **[Plugin System](plugin-system.md)** - Extending ALaS with custom functionality
- **[Examples](examples.md)** - Sample programs and patterns
- **[LLM Integration Guide](llm-guide.md)** - Best practices for generating ALaS with AI

## Key Features

- **JSON-based syntax** - Machine-readable format optimized for AI generation
- **Type safety** - Strong typing with inference capabilities
- **Module system** - Organize code into reusable modules
- **Standard library** - Rich set of built-in functions
- **Plugin architecture** - Extend the language with custom functionality
- **Multiple backends** - Interpreter and LLVM compiler support

## Quick Example

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

## Getting Help

- Check the [troubleshooting guide](troubleshooting.md) for common issues
- View [examples](examples.md) for practical code patterns
- Refer to the [language specification](language-spec.md) for detailed syntax

## For AI/LLM Users

If you're using an AI system to generate ALaS code, see the [LLM Integration Guide](llm-guide.md) for:
- Optimal prompting strategies
- Common patterns and idioms
- Error handling and debugging tips
- Performance considerations