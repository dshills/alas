# ALaS Plugin System

The ALaS plugin system allows you to extend the language with custom functionality. Plugins can add new functions, types, and capabilities to ALaS programs.

## Overview

Plugins in ALaS are:
- JSON-based configuration files
- Loaded at runtime
- Can provide new builtin functions
- Support versioning and dependencies

## Plugin Structure

A plugin is defined by a `plugin.json` file:

```json
{
  "name": "example-plugin",
  "version": "1.0.0",
  "description": "An example ALaS plugin",
  "author": "Your Name",
  "license": "MIT",
  "main": "plugin.alas.json",
  "exports": {
    "functions": [
      {
        "name": "example.hello",
        "signature": {
          "params": [{"name": "name", "type": "string"}],
          "returns": "string"
        },
        "description": "Returns a greeting message"
      }
    ]
  },
  "dependencies": {
    "alas": ">=0.1.0"
  },
  "metadata": {
    "homepage": "https://example.com",
    "repository": "https://github.com/example/plugin"
  }
}
```

## Creating a Plugin

### Step 1: Create Plugin Directory

```bash
mkdir my-plugin
cd my-plugin
```

### Step 2: Create plugin.json

Define your plugin's metadata and exported functions:

```json
{
  "name": "math-extra",
  "version": "1.0.0",
  "description": "Extended math functions for ALaS",
  "author": "Your Name",
  "exports": {
    "functions": [
      {
        "name": "mathx.factorial",
        "signature": {
          "params": [{"name": "n", "type": "int"}],
          "returns": "int"
        },
        "description": "Calculates factorial of n"
      },
      {
        "name": "mathx.gcd",
        "signature": {
          "params": [
            {"name": "a", "type": "int"},
            {"name": "b", "type": "int"}
          ],
          "returns": "int"
        },
        "description": "Greatest common divisor"
      }
    ]
  }
}
```

### Step 3: Implement Plugin Functions

Create `plugin.alas.json` with the implementation:

```json
{
  "type": "module",
  "name": "math_extra_plugin",
  "exports": ["factorial", "gcd"],
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
      "name": "gcd",
      "params": [
        {"name": "a", "type": "int"},
        {"name": "b", "type": "int"}
      ],
      "returns": "int",
      "body": [
        {
          "type": "if",
          "cond": {
            "type": "binary",
            "op": "==",
            "left": {"type": "variable", "name": "b"},
            "right": {"type": "literal", "value": 0}
          },
          "then": [
            {"type": "return", "value": {"type": "variable", "name": "a"}}
          ],
          "else": [
            {
              "type": "return",
              "value": {
                "type": "call",
                "name": "gcd",
                "args": [
                  {"type": "variable", "name": "b"},
                  {
                    "type": "binary",
                    "op": "%",
                    "left": {"type": "variable", "name": "a"},
                    "right": {"type": "variable", "name": "b"}
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## Plugin Management

### Listing Plugins

```bash
./bin/alas-plugin list -path plugins/
```

### Validating a Plugin

```bash
./bin/alas-plugin validate my-plugin/plugin.json
```

### Creating a New Plugin

```bash
./bin/alas-plugin create my-new-plugin
```

## Using Plugins in ALaS Programs

Once a plugin is installed, you can use its functions:

```json
{
  "type": "module",
  "name": "plugin_usage",
  "imports": ["math-extra"],
  "functions": [
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
            "type": "builtin",
            "name": "mathx.factorial",
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

## Advanced Plugin Features

### Plugin Dependencies

Plugins can depend on other plugins:

```json
{
  "dependencies": {
    "alas": ">=0.1.0",
    "another-plugin": "^1.0.0"
  }
}
```

### Plugin Configuration

Plugins can accept configuration:

```json
{
  "config": {
    "precision": {
      "type": "int",
      "default": 10,
      "description": "Decimal precision for calculations"
    }
  }
}
```

### Native Extensions

For performance-critical operations, plugins can reference native implementations:

```json
{
  "native": {
    "library": "libmathx.so",
    "functions": {
      "mathx.factorial": "mathx_factorial"
    }
  }
}
```

## Example Plugins

### String Utilities Plugin

```json
{
  "name": "string-utils",
  "version": "1.0.0",
  "exports": {
    "functions": [
      {
        "name": "strutil.reverse",
        "signature": {
          "params": [{"name": "str", "type": "string"}],
          "returns": "string"
        }
      },
      {
        "name": "strutil.repeat",
        "signature": {
          "params": [
            {"name": "str", "type": "string"},
            {"name": "count", "type": "int"}
          ],
          "returns": "string"
        }
      }
    ]
  }
}
```

### Date/Time Plugin

```json
{
  "name": "datetime",
  "version": "1.0.0",
  "exports": {
    "functions": [
      {
        "name": "datetime.now",
        "signature": {
          "params": [],
          "returns": "int"
        },
        "description": "Returns current Unix timestamp"
      },
      {
        "name": "datetime.format",
        "signature": {
          "params": [
            {"name": "timestamp", "type": "int"},
            {"name": "format", "type": "string"}
          ],
          "returns": "string"
        }
      }
    ]
  }
}
```

## Best Practices

1. **Namespace Functions**: Use a unique prefix for your plugin functions (e.g., `myplugin.function`)
2. **Version Carefully**: Follow semantic versioning for your plugins
3. **Document Thoroughly**: Provide clear descriptions for all exported functions
4. **Test Extensively**: Include test cases with your plugin
5. **Handle Errors**: Ensure your plugin functions handle edge cases gracefully

## Plugin Development Tips

- Start with simple, pure functions
- Test your plugin with the validator before distribution
- Consider performance implications for complex operations
- Provide examples of plugin usage
- Keep plugin size reasonable - split large plugins into multiple modules

## Distribution

Plugins can be distributed through:
- Git repositories
- Package registries
- Direct file sharing
- Plugin marketplace (future feature)

## Security Considerations

- Plugins run with the same permissions as the ALaS runtime
- Always validate and review third-party plugins before use
- Consider sandboxing for untrusted plugins
- Check plugin signatures when available