
# ALaS Language Specification (v0.1)

## Overview

ALaS (Artificial Language for Autonomous Systems) is a general-purpose, Turing-complete programming language designed exclusively for AI models to generate, manipulate, and execute. It uses structured, non-human-readable representations (e.g., JSON, binary IR) to enable low-error, high-speed code generation and transformation by LLMs.

## Goals

* Optimize for machine generation, not human ergonomics
* Eliminate syntactic ambiguity and minimize error space
* Use structured representations instead of token streams
* Be modular, Turing-complete, and semantically introspectable

## Key Concepts

### Program

A program consists of one or more modules. Each module contains functions, types, and metadata.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://alas-lang.org/schema/module.json",
  "type": "object",
  "required": ["type", "name", "functions"],
  "properties": {
    "type": {
      "const": "module"
    },
    "name": {
      "type": "string"
    },
    "exports": {
      "type": "array",
      "items": { "type": "string" }
    },
    "imports": {
      "type": "array",
      "items": { "type": "string" }
    },
    "functions": {
      "type": "array",
      "items": { "$ref": "#/definitions/function" }
    },
    "types": {
      "type": "array",
      "items": { "type": "object" }
    },
    "meta": {
      "type": "object",
      "additionalProperties": true
    }
  },
  "definitions": {
    "function": {
      "type": "object",
      "required": ["type", "name", "params", "returns", "body"],
      "properties": {
        "type": { "const": "function" },
        "name": { "type": "string" },
        "params": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["name", "type"],
            "properties": {
              "name": { "type": "string" },
              "type": { "type": "string" }
            }
          }
        },
        "returns": { "type": "string" },
        "body": {
          "type": "array",
          "items": { "$ref": "#/definitions/statement" }
        },
        "meta": {
          "type": "object",
          "additionalProperties": true
        }
      }
    },
    "statement": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": { "type": "string" },
        "value": { "$ref": "#/definitions/expression" },
        "target": { "type": "string" },
        "cond": { "$ref": "#/definitions/expression" },
        "then": {
          "type": "array",
          "items": { "$ref": "#/definitions/statement" }
        },
        "else": {
          "type": "array",
          "items": { "$ref": "#/definitions/statement" }
        },
        "body": {
          "type": "array",
          "items": { "$ref": "#/definitions/statement" }
        }
      }
    },
    "expression": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": { "type": "string" },
        "value": {},
        "name": { "type": "string" },
        "op": { "type": "string" },
        "left": { "$ref": "#/definitions/expression" },
        "right": { "$ref": "#/definitions/expression" },
        "args": {
          "type": "array",
          "items": { "$ref": "#/definitions/expression" }
        }
      }
    }
  }
}
```

## Execution Model

* Deterministic execution model
* Function-level memory scoping
* No global mutable state outside structured `state` constructs

## Extensibility

* Operations are schema-defined and introspectable
* New ops/types can be registered dynamically (e.g., plugin-based)

## Design Constraints

* No comments, whitespace, or identifiers intended for human use
* Must be serializable to compact binary IR for fast transport
* Must be embeddable and streamable for interactive tooling
