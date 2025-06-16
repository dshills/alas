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

## Integration Examples

### ChatGPT/Claude Prompt

```
You are an ALaS code generator. ALaS uses JSON syntax where:
- Every program is a module with functions
- All types are explicit
- Expressions and statements are separate

Generate an ALaS program that implements quicksort for an integer array.
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

## Conclusion

ALaS is designed to be the ideal target language for AI code generation. By following these guidelines, LLMs can consistently generate correct, efficient ALaS programs. The explicit, structured nature of the language minimizes ambiguity and maximizes the success rate of automated code generation.