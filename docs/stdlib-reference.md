# ALaS Standard Library Reference

The ALaS standard library provides built-in functions organized into modules. All standard library functions are called using the `builtin` expression type.

## I/O Module (`io`)

### `io.print`

Prints values to standard output.

**Signature:** `void io.print(value)`

**Parameters:**
- `value`: Any type - The value to print

**Example:**
```json
{
  "type": "builtin",
  "name": "io.print",
  "args": [{"type": "literal", "value": "Hello, World!"}]
}
```

### `io.println`

Prints values to standard output with a newline.

**Signature:** `void io.println(value)`

**Parameters:**
- `value`: Any type - The value to print

## Math Module (`math`)

### `math.abs`

Returns the absolute value of a number.

**Signature:** `float math.abs(number)`

**Parameters:**
- `number`: int or float - The input number

**Returns:** The absolute value

**Example:**
```json
{
  "type": "builtin",
  "name": "math.abs",
  "args": [{"type": "literal", "value": -42}]
}
```

### `math.sqrt`

Returns the square root of a number.

**Signature:** `float math.sqrt(number)`

**Parameters:**
- `number`: int or float - The input number (must be non-negative)

**Returns:** The square root

**Example:**
```json
{
  "type": "builtin",
  "name": "math.sqrt",
  "args": [{"type": "literal", "value": 16}]
}
```

### `math.pow`

Raises a number to a power.

**Signature:** `float math.pow(base, exponent)`

**Parameters:**
- `base`: int or float - The base number
- `exponent`: int or float - The exponent

**Returns:** base raised to the power of exponent

### `math.max`

Returns the maximum of two numbers.

**Signature:** `float math.max(a, b)`

**Parameters:**
- `a`: int or float - First number
- `b`: int or float - Second number

**Returns:** The larger of the two numbers

### `math.min`

Returns the minimum of two numbers.

**Signature:** `float math.min(a, b)`

**Parameters:**
- `a`: int or float - First number
- `b`: int or float - Second number

**Returns:** The smaller of the two numbers

### `math.floor`

Returns the largest integer less than or equal to a number.

**Signature:** `int math.floor(number)`

**Parameters:**
- `number`: float - The input number

**Returns:** The floor value

### `math.ceil`

Returns the smallest integer greater than or equal to a number.

**Signature:** `int math.ceil(number)`

**Parameters:**
- `number`: float - The input number

**Returns:** The ceiling value

### `math.round`

Rounds a number to the nearest integer.

**Signature:** `int math.round(number)`

**Parameters:**
- `number`: float - The input number

**Returns:** The rounded value

## String Module (`string`)

### `string.length`

Returns the length of a string.

**Signature:** `int string.length(str)`

**Parameters:**
- `str`: string - The input string

**Returns:** The number of characters

**Example:**
```json
{
  "type": "builtin",
  "name": "string.length",
  "args": [{"type": "literal", "value": "Hello"}]
}
```

### `string.toUpper`

Converts a string to uppercase.

**Signature:** `string string.toUpper(str)`

**Parameters:**
- `str`: string - The input string

**Returns:** The uppercase string

### `string.toLower`

Converts a string to lowercase.

**Signature:** `string string.toLower(str)`

**Parameters:**
- `str`: string - The input string

**Returns:** The lowercase string

### `string.concat`

Concatenates two strings.

**Signature:** `string string.concat(str1, str2)`

**Parameters:**
- `str1`: string - First string
- `str2`: string - Second string

**Returns:** The concatenated string

### `string.substring`

Extracts a substring.

**Signature:** `string string.substring(str, start, length)`

**Parameters:**
- `str`: string - The input string
- `start`: int - Starting position (0-based)
- `length`: int - Number of characters to extract

**Returns:** The substring

## Collections Module (`collections`)

### `collections.length`

Returns the length of an array or the number of entries in a map.

**Signature:** `int collections.length(collection)`

**Parameters:**
- `collection`: array or map - The collection to measure

**Returns:** The number of elements

**Example:**
```json
{
  "type": "builtin",
  "name": "collections.length",
  "args": [{"type": "variable", "name": "myArray"}]
}
```

### `collections.contains`

Checks if a collection contains a value.

**Signature:** `bool collections.contains(collection, value)`

**Parameters:**
- `collection`: array or string - The collection to search
- `value`: any - The value to find

**Returns:** true if the value is found, false otherwise

### `collections.append`

Adds an element to the end of an array.

**Signature:** `void collections.append(array, value)`

**Parameters:**
- `array`: array - The array to modify
- `value`: any - The value to append

### `collections.remove`

Removes an element from an array at a specific index.

**Signature:** `void collections.remove(array, index)`

**Parameters:**
- `array`: array - The array to modify
- `index`: int - The index to remove

## Type Module (`type`)

### `type.typeOf`

Returns a string representation of a value's type.

**Signature:** `string type.typeOf(value)`

**Parameters:**
- `value`: any - The value to inspect

**Returns:** Type name as string ("int", "float", "string", "bool", "array", "map")

**Example:**
```json
{
  "type": "builtin",
  "name": "type.typeOf",
  "args": [{"type": "literal", "value": 42}]
}
```

### `type.isInt`

Checks if a value is an integer.

**Signature:** `bool type.isInt(value)`

**Parameters:**
- `value`: any - The value to check

**Returns:** true if the value is an integer

### `type.isFloat`

Checks if a value is a float.

**Signature:** `bool type.isFloat(value)`

**Parameters:**
- `value`: any - The value to check

**Returns:** true if the value is a float

### `type.isString`

Checks if a value is a string.

**Signature:** `bool type.isString(value)`

**Parameters:**
- `value`: any - The value to check

**Returns:** true if the value is a string

### `type.isBool`

Checks if a value is a boolean.

**Signature:** `bool type.isBool(value)`

**Parameters:**
- `value`: any - The value to check

**Returns:** true if the value is a boolean

### `type.isArray`

Checks if a value is an array.

**Signature:** `bool type.isArray(value)`

**Parameters:**
- `value`: any - The value to check

**Returns:** true if the value is an array

### `type.isMap`

Checks if a value is a map.

**Signature:** `bool type.isMap(value)`

**Parameters:**
- `value`: any - The value to check

**Returns:** true if the value is a map

## Complete Example

Here's a program that demonstrates various standard library functions:

```json
{
  "type": "module",
  "name": "stdlib_demo",
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
            "args": [{"type": "literal", "value": "=== Standard Library Demo ==="}]
          }
        },
        {
          "type": "assign",
          "target": "x",
          "value": {
            "type": "builtin",
            "name": "math.sqrt",
            "args": [{"type": "literal", "value": 16}]
          }
        },
        {
          "type": "assign",
          "target": "message",
          "value": {"type": "literal", "value": "hello world"}
        },
        {
          "type": "assign",
          "target": "upper",
          "value": {
            "type": "builtin",
            "name": "string.toUpper",
            "args": [{"type": "variable", "name": "message"}]
          }
        },
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
        },
        {
          "type": "assign",
          "target": "len",
          "value": {
            "type": "builtin",
            "name": "collections.length",
            "args": [{"type": "variable", "name": "numbers"}]
          }
        },
        {
          "type": "expr",
          "value": {
            "type": "builtin",
            "name": "io.print",
            "args": [{"type": "variable", "name": "upper"}]
          }
        }
      ]
    }
  ]
}
```

## Notes

- All standard library functions are pure (no side effects) except for I/O operations
- Type checking is performed at runtime for dynamic operations
- Some functions may not be available in all compilation targets