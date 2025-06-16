# ALaS LLM Integration Guide

This guide provides best practices for using AI/LLM systems to generate ALaS code effectively.

## Overview

ALaS is specifically designed for machine generation. Its JSON-based syntax and explicit structure make it ideal for LLMs to generate correct, working code consistently.

## Key Advantages for LLM Generation

1. **Structured Format**: JSON syntax eliminates parsing ambiguities
2. **Explicit Types**: No type inference reduces generation errors
3. **No Implicit Behavior**: Every operation is explicitly defined
4. **Consistent Schema**: Predictable structure across all constructs
5. **Machine-Readable Errors**: Error messages designed for AI interpretation

## Prompting Strategies

### Basic Program Generation

**Effective Prompt:**
```
Generate an ALaS program that calculates the factorial of a number.
The program should:
- Have a recursive factorial function
- Have a main function that calls factorial(5)
- Print the result
```

**LLM Response Pattern:**
The LLM should generate a complete module with proper JSON structure, including all required fields.

### Specific Function Generation

**Effective Prompt:**
```
Create an ALaS function that:
- Name: "isPalindrome"
- Parameter: string called "str"
- Returns: bool
- Logic: Check if the string reads the same forwards and backwards
```

### Complex Program Generation

**Effective Prompt:**
```
Generate an ALaS program with multiple functions that:
1. Implements bubble sort for an array of integers
2. Has a helper function to swap two elements
3. Has a main function that sorts [5, 2, 8, 1, 9] and prints each step
```

## Common Patterns for LLMs

### Function Template

When generating functions, LLMs should follow this pattern:

```json
{
  "type": "function",
  "name": "<function_name>",
  "params": [
    {"name": "<param1>", "type": "<type1>"},
    {"name": "<param2>", "type": "<type2>"}
  ],
  "returns": "<return_type>",
  "body": [
    // Statements
  ]
}
```

### Expression Building

For complex expressions, build them hierarchically:

```json
// Instead of trying to generate: a + b * c - d
// Build it as: (a + (b * c)) - d
{
  "type": "binary",
  "op": "-",
  "left": {
    "type": "binary",
    "op": "+",
    "left": {"type": "variable", "name": "a"},
    "right": {
      "type": "binary",
      "op": "*",
      "left": {"type": "variable", "name": "b"},
      "right": {"type": "variable", "name": "c"}
    }
  },
  "right": {"type": "variable", "name": "d"}
}
```

### Control Flow Generation

For if-else statements:
```json
{
  "type": "if",
  "cond": {/* condition expression */},
  "then": [/* list of statements */],
  "else": [/* list of statements */]  // Optional
}
```

## Critical Implementation Notes

### Function Declaration Patterns

When generating functions, avoid using internal type constructors. Always use explicit parameter lists:

```json
{
  "type": "function",
  "name": "myFunc",
  "params": [
    {"name": "x", "type": "int"},
    {"name": "y", "type": "int"}
  ],
  "returns": "int",
  "body": [/* ... */]
}
```

### For Loop Support

For loops are now fully supported:

```json
{
  "type": "for",
  "init": {
    "type": "assign",
    "target": "i",
    "value": {"type": "literal", "value": 0}
  },
  "cond": {
    "type": "binary",
    "op": "<",
    "left": {"type": "variable", "name": "i"},
    "right": {"type": "literal", "value": 10}
  },
  "update": {
    "type": "assign",
    "target": "i",
    "value": {
      "type": "binary",
      "op": "+",
      "left": {"type": "variable", "name": "i"},
      "right": {"type": "literal", "value": 1}
    }
  },
  "body": [
    // Loop body
  ]
}
```

## Best Practices for LLM Generation

### 1. Start with Structure

Always begin with the module structure:
```json
{
  "type": "module",
  "name": "module_name",
  "functions": []
}
```

### 2. Generate Functions Incrementally

Build functions one at a time:
- First, generate the function signature
- Then, add the body statements
- Finally, verify all variable references

### 3. Use Consistent Naming

- Variables: camelCase (e.g., `userName`, `totalCount`)
- Functions: camelCase (e.g., `calculateSum`, `isPrime`)
- Modules: snake_case (e.g., `math_utils`, `string_helpers`)

### 4. Type Safety

Always specify types explicitly:
- Function parameters must have types
- Function returns must have types
- No implicit conversions

### 5. Handle Edge Cases

Generate proper error handling:
```json
{
  "type": "if",
  "cond": {
    "type": "binary",
    "op": "<",
    "left": {"type": "variable", "name": "index"},
    "right": {"type": "literal", "value": 0}
  },
  "then": [
    {"type": "return", "value": {"type": "literal", "value": -1}}
  ]
}
```

## Common LLM Generation Errors

### 1. Missing Required Fields

**Wrong:**
```json
{
  "type": "function",
  "name": "test"
  // Missing params, returns, body
}
```

**Correct:**
```json
{
  "type": "function",
  "name": "test",
  "params": [],
  "returns": "void",
  "body": []
}
```

### 2. Incorrect Type References

**Wrong:**
```json
{"type": "variable", "value": "x"}  // Variables don't have 'value'
```

**Correct:**
```json
{"type": "variable", "name": "x"}
```

### 3. Invalid Statement Nesting

**Wrong:**
```json
{
  "type": "assign",
  "target": "x",
  "value": {
    "type": "if",  // If is a statement, not expression
    // ...
  }
}
```

**Correct:**
Use a function call or restructure the logic.

### 4. Parameter Handling in LLVM

**Common Error:** When generating code for LLVM compilation, parameters must be properly stored:

**Wrong:**
```json
// Directly using parameters without proper storage
{"type": "variable", "name": "param"}
```

**Correct Pattern:**
Parameters are automatically handled by the compiler - just reference them normally:
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

### 5. Array and Map Literals

**Important:** Use the correct literal types:

**Wrong:**
```json
{"type": "literal", "value": [1, 2, 3]}  // Arrays aren't literals
```

**Correct:**
```json
{
  "type": "array_literal",
  "elements": [
    {"type": "literal", "value": 1},
    {"type": "literal", "value": 2},
    {"type": "literal", "value": 3}
  ]
}
```

## Advanced Generation Techniques

### 1. Template-Based Generation

Create templates for common patterns:

**Array Processing Template:**
```json
{
  "type": "function",
  "name": "processArray",
  "params": [{"name": "arr", "type": "array"}],
  "returns": "void",
  "body": [
    {
      "type": "assign",
      "target": "i",
      "value": {"type": "literal", "value": 0}
    },
    {
      "type": "while",
      "cond": {
        "type": "binary",
        "op": "<",
        "left": {"type": "variable", "name": "i"},
        "right": {
          "type": "builtin",
          "name": "collections.length",
          "args": [{"type": "variable", "name": "arr"}]
        }
      },
      "body": [
        // Process element at index i
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
  ]
}
```

### 2. Modular Generation

Generate modules separately and combine:

1. Generate utility functions
2. Generate main logic
3. Generate test functions
4. Combine into final module

### 3. Iterative Refinement

1. Generate basic structure
2. Add type information
3. Implement function bodies
4. Add error handling
5. Optimize generated code

## Practical Generation Tips from Experience

### 1. Variable Initialization

Always initialize variables before use. The validator will catch undefined variables:

```json
// Initialize counters
{"type": "assign", "target": "sum", "value": {"type": "literal", "value": 0}},
// Then use them
{"type": "assign", "target": "sum", "value": {
  "type": "binary",
  "op": "+",
  "left": {"type": "variable", "name": "sum"},
  "right": {"type": "variable", "name": "x"}
}}
```

### 2. Array Operations

For array iteration, combine index tracking with length checks:

```json
{
  "type": "assign",
  "target": "arr",
  "value": {
    "type": "array_literal",
    "elements": [
      {"type": "literal", "value": 5},
      {"type": "literal", "value": 2},
      {"type": "literal", "value": 8}
    ]
  }
},
{
  "type": "for",
  "init": {"type": "assign", "target": "i", "value": {"type": "literal", "value": 0}},
  "cond": {
    "type": "binary",
    "op": "<",
    "left": {"type": "variable", "name": "i"},
    "right": {
      "type": "builtin",
      "name": "collections.length",
      "args": [{"type": "variable", "name": "arr"}]
    }
  },
  "update": {
    "type": "assign",
    "target": "i",
    "value": {
      "type": "binary",
      "op": "+",
      "left": {"type": "variable", "name": "i"},
      "right": {"type": "literal", "value": 1}
    }
  },
  "body": [
    {
      "type": "expr",
      "value": {
        "type": "builtin",
        "name": "io.print",
        "args": [{
          "type": "index",
          "object": {"type": "variable", "name": "arr"},
          "index": {"type": "variable", "name": "i"}
        }]
      }
    }
  ]
}
```

### 3. Debugging Generated Code

When debugging:
1. First validate with `alas-validate`
2. Check for undefined variables
3. Verify all function calls have correct argument counts
4. Ensure array indices are within bounds
5. Test with simple inputs before complex ones

## Validation and Testing

### 1. Use the Validator

Always validate generated code:
```bash
./bin/alas-validate -file generated.alas.json
```

### 2. Test Incrementally

1. Generate and test individual functions
2. Combine into larger programs
3. Test edge cases

### 3. Common Validation Errors

- **Undefined variables**: Ensure all variables are assigned before use
- **Type mismatches**: Verify operation compatibility
- **Missing returns**: All code paths must return if function has return type

## Performance Considerations

### 1. Minimize Nesting

Deep nesting makes generation harder:
```json
// Prefer flat structure where possible
{
  "type": "assign",
  "target": "temp1",
  "value": {/* expression 1 */}
},
{
  "type": "assign",
  "target": "result",
  "value": {/* use temp1 */}
}
```

### 2. Use Standard Library

Leverage built-in functions instead of reimplementing:
```json
// Use math.max instead of custom max function
{
  "type": "builtin",
  "name": "math.max",
  "args": [
    {"type": "variable", "name": "a"},
    {"type": "variable", "name": "b"}
  ]
}
```

## Real-World Generation Patterns

### Sorting Algorithm Template

Here's a complete bubble sort implementation pattern:

```json
{
  "type": "function",
  "name": "bubbleSort",
  "params": [{"name": "arr", "type": "array"}],
  "returns": "array",
  "body": [
    {
      "type": "assign",
      "target": "n",
      "value": {
        "type": "builtin",
        "name": "collections.length",
        "args": [{"type": "variable", "name": "arr"}]
      }
    },
    {
      "type": "for",
      "init": {"type": "assign", "target": "i", "value": {"type": "literal", "value": 0}},
      "cond": {
        "type": "binary",
        "op": "<",
        "left": {"type": "variable", "name": "i"},
        "right": {"type": "variable", "name": "n"}
      },
      "update": {
        "type": "assign",
        "target": "i",
        "value": {
          "type": "binary",
          "op": "+",
          "left": {"type": "variable", "name": "i"},
          "right": {"type": "literal", "value": 1}
        }
      },
      "body": [
        // Inner loop for comparisons
      ]
    },
    {"type": "return", "value": {"type": "variable", "name": "arr"}}
  ]
}
```

### Recursive Function Pattern

When generating recursive functions, ensure base cases come first:

```json
{
  "type": "function",
  "name": "fibonacci",
  "params": [{"name": "n", "type": "int"}],
  "returns": "int",
  "body": [
    // Base case first
    {
      "type": "if",
      "cond": {
        "type": "binary",
        "op": "<=",
        "left": {"type": "variable", "name": "n"},
        "right": {"type": "literal", "value": 1}
      },
      "then": [
        {"type": "return", "value": {"type": "variable", "name": "n"}}
      ]
    },
    // Recursive case
    {
      "type": "return",
      "value": {
        "type": "binary",
        "op": "+",
        "left": {
          "type": "call",
          "name": "fibonacci",
          "args": [{
            "type": "binary",
            "op": "-",
            "left": {"type": "variable", "name": "n"},
            "right": {"type": "literal", "value": 1}
          }]
        },
        "right": {
          "type": "call",
          "name": "fibonacci",
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
```

## Integration Examples

### ChatGPT/Claude Prompt

```
You are an ALaS code generator. ALaS uses JSON syntax where:
- Every program is a module with functions
- All types are explicit
- Expressions and statements are separate

Generate an ALaS program that implements quicksort for an integer array.

IMPORTANT:
- Use array_literal for array creation, not literal
- Initialize all variables before use
- Use for loops with proper init/cond/update structure
- Reference parameters directly without special handling
```

### GitHub Copilot Integration

Use comments to guide generation:
```json
{
  "type": "module",
  "name": "quicksort",
  "functions": [
    // TODO: Generate partition function that takes array, low, high
    // TODO: Generate quicksort function that recursively sorts
    // TODO: Generate main function that tests with [3, 1, 4, 1, 5, 9]
  ]
}
```

## Debugging LLM-Generated Code

### 1. Syntax Errors

- Validate JSON structure first
- Check for missing commas, brackets
- Ensure proper quoting

### 2. Semantic Errors

- Trace variable definitions
- Verify type consistency
- Check function signatures

### 3. Logic Errors

- Add print statements for debugging
- Test with simple inputs first
- Verify algorithm implementation

## Future Enhancements

As ALaS evolves, LLM integration will improve with:

1. **Semantic Templates**: Higher-level pattern descriptions
2. **Error Recovery**: Automatic fixing of common mistakes
3. **Optimization Hints**: Performance-aware generation
4. **Domain-Specific Libraries**: Specialized generation for different domains

## Key Takeaways for LLM Generation

1. **Always validate generated code** - Use `alas-validate` to catch syntax errors
2. **Initialize before use** - All variables must be assigned before being referenced
3. **Use correct literal types** - `array_literal` for arrays, `map_literal` for maps
4. **Parameters are automatic** - Just reference parameter names directly
5. **For loops are supported** - Use the init/cond/update/body structure
6. **Test incrementally** - Start with simple functions and build up
7. **LLVM compilation works** - Generated code can be compiled to native binaries

## Common Pitfalls to Avoid

1. **Don't use internal type constructors** - Stick to the JSON schema
2. **Don't assume implicit conversions** - Be explicit about types
3. **Don't forget array bounds** - Always check length before indexing
4. **Don't nest statements in expressions** - Keep them separate
5. **Don't skip validation** - Always validate before running or compiling

## Conclusion

ALaS is designed to be the ideal target language for AI code generation. By following these guidelines and learning from common implementation patterns, LLMs can consistently generate correct, efficient ALaS programs. The explicit, structured nature of the language minimizes ambiguity and maximizes the success rate of automated code generation.

The LLVM backend is now fully functional, supporting all core language features including functions, recursion, loops, arrays, and basic maps. This enables ALaS programs to be compiled to efficient native code, making it suitable for production use cases where performance matters.