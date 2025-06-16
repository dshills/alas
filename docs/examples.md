# ALaS Examples

This document contains various example programs demonstrating ALaS features and common programming patterns.

## Basic Examples

### Hello World

The simplest ALaS program:

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
            "args": [{"type": "literal", "value": "Hello, World!"}]
          }
        }
      ]
    }
  ]
}
```

### Variables and Arithmetic

```json
{
  "type": "module",
  "name": "arithmetic",
  "functions": [
    {
      "type": "function",
      "name": "main",
      "params": [],
      "returns": "int",
      "body": [
        {
          "type": "assign",
          "target": "a",
          "value": {"type": "literal", "value": 10}
        },
        {
          "type": "assign",
          "target": "b",
          "value": {"type": "literal", "value": 20}
        },
        {
          "type": "assign",
          "target": "sum",
          "value": {
            "type": "binary",
            "op": "+",
            "left": {"type": "variable", "name": "a"},
            "right": {"type": "variable", "name": "b"}
          }
        },
        {
          "type": "return",
          "value": {"type": "variable", "name": "sum"}
        }
      ]
    }
  ]
}
```

## Control Flow Examples

### Conditional Logic

```json
{
  "type": "module",
  "name": "max_value",
  "functions": [
    {
      "type": "function",
      "name": "max",
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
            "op": ">",
            "left": {"type": "variable", "name": "a"},
            "right": {"type": "variable", "name": "b"}
          },
          "then": [
            {"type": "return", "value": {"type": "variable", "name": "a"}}
          ],
          "else": [
            {"type": "return", "value": {"type": "variable", "name": "b"}}
          ]
        }
      ]
    }
  ]
}
```

### While Loop - Sum of Numbers

```json
{
  "type": "module",
  "name": "sum_numbers",
  "functions": [
    {
      "type": "function",
      "name": "sumToN",
      "params": [{"name": "n", "type": "int"}],
      "returns": "int",
      "body": [
        {
          "type": "assign",
          "target": "sum",
          "value": {"type": "literal", "value": 0}
        },
        {
          "type": "assign",
          "target": "i",
          "value": {"type": "literal", "value": 1}
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
              "target": "sum",
              "value": {
                "type": "binary",
                "op": "+",
                "left": {"type": "variable", "name": "sum"},
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
          "value": {"type": "variable", "name": "sum"}
        }
      ]
    }
  ]
}
```

### For Loop - Array Sum

```json
{
  "type": "module",
  "name": "array_sum",
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
              {"type": "literal", "value": 1},
              {"type": "literal", "value": 2},
              {"type": "literal", "value": 3},
              {"type": "literal", "value": 4},
              {"type": "literal", "value": 5}
            ]
          }
        },
        {
          "type": "assign",
          "target": "sum",
          "value": {"type": "literal", "value": 0}
        },
        {
          "type": "assign",
          "target": "i",
          "value": {"type": "literal", "value": 0}
        },
        {
          "type": "for",
          "cond": {
            "type": "binary",
            "op": "<",
            "left": {"type": "variable", "name": "i"},
            "right": {"type": "literal", "value": 5}
          },
          "body": [
            {
              "type": "assign",
              "target": "sum",
              "value": {
                "type": "binary",
                "op": "+",
                "left": {"type": "variable", "name": "sum"},
                "right": {
                  "type": "index",
                  "object": {"type": "variable", "name": "numbers"},
                  "index": {"type": "variable", "name": "i"}
                }
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
          "value": {"type": "variable", "name": "sum"}
        }
      ]
    }
  ]
}
```

## Recursive Functions

### Fibonacci Sequence

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
            {"type": "return", "value": {"type": "variable", "name": "n"}}
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
    }
  ]
}
```

### Factorial

```json
{
  "type": "module",
  "name": "factorial",
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
    }
  ]
}
```

## Data Structures

### Working with Arrays

```json
{
  "type": "module",
  "name": "array_operations",
  "functions": [
    {
      "type": "function",
      "name": "reverseArray",
      "params": [{"name": "arr", "type": "array"}],
      "returns": "array",
      "body": [
        {
          "type": "assign",
          "target": "length",
          "value": {
            "type": "builtin",
            "name": "collections.length",
            "args": [{"type": "variable", "name": "arr"}]
          }
        },
        {
          "type": "assign",
          "target": "reversed",
          "value": {"type": "array_literal", "elements": []}
        },
        {
          "type": "assign",
          "target": "i",
          "value": {
            "type": "binary",
            "op": "-",
            "left": {"type": "variable", "name": "length"},
            "right": {"type": "literal", "value": 1}
          }
        },
        {
          "type": "while",
          "cond": {
            "type": "binary",
            "op": ">=",
            "left": {"type": "variable", "name": "i"},
            "right": {"type": "literal", "value": 0}
          },
          "body": [
            {
              "type": "expr",
              "value": {
                "type": "builtin",
                "name": "collections.append",
                "args": [
                  {"type": "variable", "name": "reversed"},
                  {
                    "type": "index",
                    "object": {"type": "variable", "name": "arr"},
                    "index": {"type": "variable", "name": "i"}
                  }
                ]
              }
            },
            {
              "type": "assign",
              "target": "i",
              "value": {
                "type": "binary",
                "op": "-",
                "left": {"type": "variable", "name": "i"},
                "right": {"type": "literal", "value": 1}
              }
            }
          ]
        },
        {
          "type": "return",
          "value": {"type": "variable", "name": "reversed"}
        }
      ]
    }
  ]
}
```

### Working with Maps

```json
{
  "type": "module",
  "name": "map_example",
  "functions": [
    {
      "type": "function",
      "name": "main",
      "params": [],
      "returns": "void",
      "body": [
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
              },
              {
                "key": {"type": "literal", "value": "city"},
                "value": {"type": "literal", "value": "New York"}
              }
            ]
          }
        },
        {
          "type": "assign",
          "target": "name",
          "value": {
            "type": "index",
            "object": {"type": "variable", "name": "person"},
            "index": {"type": "literal", "value": "name"}
          }
        },
        {
          "type": "expr",
          "value": {
            "type": "builtin",
            "name": "io.print",
            "args": [{"type": "variable", "name": "name"}]
          }
        }
      ]
    }
  ]
}
```

## Standard Library Usage

### Math Operations

```json
{
  "type": "module",
  "name": "math_demo",
  "functions": [
    {
      "type": "function",
      "name": "calculateHypotenuse",
      "params": [
        {"name": "a", "type": "float"},
        {"name": "b", "type": "float"}
      ],
      "returns": "float",
      "body": [
        {
          "type": "assign",
          "target": "a_squared",
          "value": {
            "type": "builtin",
            "name": "math.pow",
            "args": [
              {"type": "variable", "name": "a"},
              {"type": "literal", "value": 2}
            ]
          }
        },
        {
          "type": "assign",
          "target": "b_squared",
          "value": {
            "type": "builtin",
            "name": "math.pow",
            "args": [
              {"type": "variable", "name": "b"},
              {"type": "literal", "value": 2}
            ]
          }
        },
        {
          "type": "assign",
          "target": "sum",
          "value": {
            "type": "binary",
            "op": "+",
            "left": {"type": "variable", "name": "a_squared"},
            "right": {"type": "variable", "name": "b_squared"}
          }
        },
        {
          "type": "return",
          "value": {
            "type": "builtin",
            "name": "math.sqrt",
            "args": [{"type": "variable", "name": "sum"}]
          }
        }
      ]
    }
  ]
}
```

### String Manipulation

```json
{
  "type": "module",
  "name": "string_demo",
  "functions": [
    {
      "type": "function",
      "name": "formatName",
      "params": [
        {"name": "first", "type": "string"},
        {"name": "last", "type": "string"}
      ],
      "returns": "string",
      "body": [
        {
          "type": "assign",
          "target": "firstUpper",
          "value": {
            "type": "builtin",
            "name": "string.toUpper",
            "args": [{"type": "variable", "name": "first"}]
          }
        },
        {
          "type": "assign",
          "target": "lastUpper",
          "value": {
            "type": "builtin",
            "name": "string.toUpper",
            "args": [{"type": "variable", "name": "last"}]
          }
        },
        {
          "type": "assign",
          "target": "space",
          "value": {"type": "literal", "value": " "}
        },
        {
          "type": "assign",
          "target": "fullName",
          "value": {
            "type": "builtin",
            "name": "string.concat",
            "args": [
              {"type": "variable", "name": "firstUpper"},
              {"type": "variable", "name": "space"}
            ]
          }
        },
        {
          "type": "return",
          "value": {
            "type": "builtin",
            "name": "string.concat",
            "args": [
              {"type": "variable", "name": "fullName"},
              {"type": "variable", "name": "lastUpper"}
            ]
          }
        }
      ]
    }
  ]
}
```

## Module System

### Multi-Module Program

**math_utils.alas.json:**
```json
{
  "type": "module",
  "name": "math_utils",
  "exports": ["square", "cube"],
  "functions": [
    {
      "type": "function",
      "name": "square",
      "params": [{"name": "x", "type": "int"}],
      "returns": "int",
      "body": [
        {
          "type": "return",
          "value": {
            "type": "binary",
            "op": "*",
            "left": {"type": "variable", "name": "x"},
            "right": {"type": "variable", "name": "x"}
          }
        }
      ]
    },
    {
      "type": "function",
      "name": "cube",
      "params": [{"name": "x", "type": "int"}],
      "returns": "int",
      "body": [
        {
          "type": "return",
          "value": {
            "type": "binary",
            "op": "*",
            "left": {
              "type": "call",
              "name": "square",
              "args": [{"type": "variable", "name": "x"}]
            },
            "right": {"type": "variable", "name": "x"}
          }
        }
      ]
    }
  ]
}
```

**main.alas.json:**
```json
{
  "type": "module",
  "name": "main",
  "imports": ["math_utils"],
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
            "type": "module_call",
            "module": "math_utils",
            "name": "cube",
            "args": [{"type": "literal", "value": 3}]
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

## Complete Applications

### Prime Number Checker

```json
{
  "type": "module",
  "name": "prime_checker",
  "functions": [
    {
      "type": "function",
      "name": "isPrime",
      "params": [{"name": "n", "type": "int"}],
      "returns": "bool",
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
            {"type": "return", "value": {"type": "literal", "value": false}}
          ]
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
            "left": {
              "type": "binary",
              "op": "*",
              "left": {"type": "variable", "name": "i"},
              "right": {"type": "variable", "name": "i"}
            },
            "right": {"type": "variable", "name": "n"}
          },
          "body": [
            {
              "type": "if",
              "cond": {
                "type": "binary",
                "op": "==",
                "left": {
                  "type": "binary",
                  "op": "%",
                  "left": {"type": "variable", "name": "n"},
                  "right": {"type": "variable", "name": "i"}
                },
                "right": {"type": "literal", "value": 0}
              },
              "then": [
                {"type": "return", "value": {"type": "literal", "value": false}}
              ]
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
          "value": {"type": "literal", "value": true}
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
          "target": "num",
          "value": {"type": "literal", "value": 17}
        },
        {
          "type": "if",
          "cond": {
            "type": "call",
            "name": "isPrime",
            "args": [{"type": "variable", "name": "num"}]
          },
          "then": [
            {
              "type": "expr",
              "value": {
                "type": "builtin",
                "name": "io.print",
                "args": [{"type": "literal", "value": "Is prime!"}]
              }
            }
          ],
          "else": [
            {
              "type": "expr",
              "value": {
                "type": "builtin",
                "name": "io.print",
                "args": [{"type": "literal", "value": "Not prime"}]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

## Tips for Writing ALaS Programs

1. **Start Simple**: Begin with basic functions and gradually add complexity
2. **Use Descriptive Names**: Make your variable and function names meaningful
3. **Modularize**: Break complex logic into smaller functions
4. **Test Incrementally**: Validate and run your program as you build it
5. **Leverage the Standard Library**: Use built-in functions when possible
6. **Comment Your Intent**: Use the metadata fields to document complex logic