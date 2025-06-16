# ALaS Troubleshooting Guide

This guide helps you diagnose and fix common issues when working with ALaS.

## Common Errors and Solutions

### Validation Errors

#### "Module must have 'type' field set to 'module'"

**Cause:** Missing or incorrect module type declaration.

**Solution:**
```json
{
  "type": "module",  // This field is required
  "name": "my_module",
  "functions": []
}
```

#### "Function must have a body"

**Cause:** Function declared without body array.

**Solution:**
```json
{
  "type": "function",
  "name": "myFunc",
  "params": [],
  "returns": "void",
  "body": []  // Even empty functions need a body array
}
```

#### "Undefined variable: x"

**Cause:** Using a variable before it's assigned.

**Solution:**
```json
// Wrong
{
  "type": "return",
  "value": {"type": "variable", "name": "x"}
}

// Correct - assign first
{
  "type": "assign",
  "target": "x",
  "value": {"type": "literal", "value": 42}
},
{
  "type": "return",
  "value": {"type": "variable", "name": "x"}
}
```

#### "Unknown statement type: for"

**Cause:** Validator doesn't recognize for loops (may need update).

**Solution:** Ensure you're using the latest version of ALaS that supports for loops, or use while loops as a workaround.

### Runtime Errors

#### "Type mismatch in binary operation"

**Cause:** Trying to perform operations on incompatible types.

**Solution:**
```json
// Wrong - can't add string and number
{
  "type": "binary",
  "op": "+",
  "left": {"type": "literal", "value": "hello"},
  "right": {"type": "literal", "value": 42}
}

// Correct - same types
{
  "type": "binary",
  "op": "+",
  "left": {"type": "literal", "value": 10},
  "right": {"type": "literal", "value": 42}
}
```

#### "Division by zero"

**Cause:** Attempting to divide by zero.

**Solution:** Add a check before division:
```json
{
  "type": "if",
  "cond": {
    "type": "binary",
    "op": "!=",
    "left": {"type": "variable", "name": "divisor"},
    "right": {"type": "literal", "value": 0}
  },
  "then": [
    {
      "type": "assign",
      "target": "result",
      "value": {
        "type": "binary",
        "op": "/",
        "left": {"type": "variable", "name": "dividend"},
        "right": {"type": "variable", "name": "divisor"}
      }
    }
  ]
}
```

#### "Array index out of bounds"

**Cause:** Accessing array element beyond its size.

**Solution:** Check array bounds:
```json
{
  "type": "if",
  "cond": {
    "type": "binary",
    "op": "<",
    "left": {"type": "variable", "name": "index"},
    "right": {
      "type": "builtin",
      "name": "collections.length",
      "args": [{"type": "variable", "name": "array"}]
    }
  },
  "then": [
    // Safe to access array[index]
  ]
}
```

### Compilation Errors (LLVM)

#### "Failed to generate function: variable is not a pointer type"

**Cause:** LLVM backend issue with variable storage.

**Solution:** This is typically an internal compiler issue. Ensure variables are properly declared and initialized.

#### "Unknown builtin function"

**Cause:** Using a builtin function not recognized by the compiler.

**Solution:** Check the standard library reference for correct function names:
- Use `io.print` not `print`
- Use `math.sqrt` not `sqrt`
- Include the module prefix

### JSON Syntax Errors

#### "Invalid JSON"

**Common Causes:**
1. Missing commas between elements
2. Trailing commas
3. Unclosed brackets or braces
4. Incorrect string quoting

**Solution:** Use a JSON validator or editor with syntax highlighting.

```json
// Wrong - missing comma
{
  "type": "module"
  "name": "test"
}

// Wrong - trailing comma
{
  "type": "module",
  "name": "test",  // <- Remove this comma
}

// Correct
{
  "type": "module",
  "name": "test"
}
```

## Debugging Techniques

### 1. Add Print Statements

Insert print statements to trace execution:
```json
{
  "type": "expr",
  "value": {
    "type": "builtin",
    "name": "io.print",
    "args": [{"type": "literal", "value": "Debug: reached here"}]
  }
}
```

### 2. Validate Incrementally

Build your program step by step:
1. Start with minimal structure
2. Add one function at a time
3. Validate after each addition

### 3. Check Types

Use type checking builtins:
```json
{
  "type": "assign",
  "target": "valueType",
  "value": {
    "type": "builtin",
    "name": "type.typeOf",
    "args": [{"type": "variable", "name": "myVar"}]
  }
},
{
  "type": "expr",
  "value": {
    "type": "builtin",
    "name": "io.print",
    "args": [{"type": "variable", "name": "valueType"}]
  }
}
```

### 4. Simplify Complex Expressions

Break down complex expressions:
```json
// Instead of one complex expression
// Break it into steps:
{
  "type": "assign",
  "target": "temp1",
  "value": {/* first part */}
},
{
  "type": "assign",
  "target": "temp2",
  "value": {/* second part */}
},
{
  "type": "assign",
  "target": "result",
  "value": {/* combine temp1 and temp2 */}
}
```

## Performance Issues

### Slow Recursive Functions

**Problem:** Deep recursion causing stack overflow or slow performance.

**Solution:** Consider iterative approach:
```json
// Convert recursive factorial to iterative
{
  "type": "function",
  "name": "factorial_iterative",
  "params": [{"name": "n", "type": "int"}],
  "returns": "int",
  "body": [
    {
      "type": "assign",
      "target": "result",
      "value": {"type": "literal", "value": 1}
    },
    {
      "type": "assign",
      "target": "i",
      "value": {"type": "literal", "value": 2}
    },
    {
      "type": "while",
      "cond": {
        "type": "binary",
        "op": "<=",
        "left": {"type": "variable", "name": "i"},
        "right": {"type": "variable", "name": "n"}
      },
      "body": [
        {
          "type": "assign",
          "target": "result",
          "value": {
            "type": "binary",
            "op": "*",
            "left": {"type": "variable", "name": "result"},
            "right": {"type": "variable", "name": "i"}
          }
        },
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
    },
    {
      "type": "return",
      "value": {"type": "variable", "name": "result"}
    }
  ]
}
```

### Large Array Operations

**Problem:** Operations on large arrays are slow.

**Solution:** 
- Process arrays in chunks
- Use appropriate algorithms (e.g., binary search instead of linear search)
- Consider using maps for lookups

## Tool-Specific Issues

### alas-validate

**"File not found"**
- Check file path is correct
- Use absolute paths if relative paths aren't working

### alas-run

**"Function 'main' not found"**
- Ensure your module has a main function
- Specify function name with `-fn` flag if using different entry point

### alas-compile

**"LLVM IR generation failed"**
- Check for unsupported language features
- Ensure all types are properly defined
- Try with simpler program to isolate issue

## Getting Help

If you encounter issues not covered here:

1. **Check Examples**: Review the examples directory for working code patterns
2. **Validate First**: Always run the validator before other tools
3. **Minimal Reproduction**: Create the smallest program that shows the issue
4. **Error Messages**: Include complete error messages when seeking help
5. **Version Info**: Note which version of ALaS you're using

## FAQ

**Q: Can I use relative paths for imports?**
A: Module imports use module names, not file paths. Ensure modules are in the module search path.

**Q: Why does my array index return wrong values?**
A: Check that you're using 0-based indexing. Arrays in ALaS start at index 0.

**Q: How do I debug infinite loops?**
A: Add a counter and print statements inside the loop to track iterations:
```json
{
  "type": "assign",
  "target": "debug_counter",
  "value": {"type": "literal", "value": 0}
},
{
  "type": "while",
  "cond": {/* your condition */},
  "body": [
    {
      "type": "expr",
      "value": {
        "type": "builtin",
        "name": "io.print",
        "args": [{"type": "variable", "name": "debug_counter"}]
      }
    },
    {
      "type": "assign",
      "target": "debug_counter",
      "value": {
        "type": "binary",
        "op": "+",
        "left": {"type": "variable", "name": "debug_counter"},
        "right": {"type": "literal", "value": 1}
      }
    },
    // Rest of loop body
  ]
}
```

**Q: Can I mix types in arrays?**
A: No, ALaS arrays are homogeneous. All elements must be the same type.

**Q: How do I handle errors in ALaS?**
A: Currently, ALaS doesn't have exception handling. Use return values to indicate errors:
```json
// Return -1 for error, positive value for success
{
  "type": "if",
  "cond": {/* error condition */},
  "then": [
    {"type": "return", "value": {"type": "literal", "value": -1}}
  ]
}
```