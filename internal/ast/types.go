package ast

// Module represents an ALaS module.
type Module struct {
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Exports   []string               `json:"exports,omitempty"`
	Imports   []string               `json:"imports,omitempty"`
	Functions []Function             `json:"functions"`
	Types     []interface{}          `json:"types,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
}

// Function represents a function definition.
type Function struct {
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Params  []Parameter            `json:"params"`
	Returns string                 `json:"returns"`
	Body    []Statement            `json:"body"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// Parameter represents a function parameter.
type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Statement represents any statement in ALaS.
type Statement struct {
	Type   string      `json:"type"`
	Value  *Expression `json:"value,omitempty"`
	Target string      `json:"target,omitempty"`
	Cond   *Expression `json:"cond,omitempty"`
	Then   []Statement `json:"then,omitempty"`
	Else   []Statement `json:"else,omitempty"`
	Body   []Statement `json:"body,omitempty"`
}

// Expression represents any expression in ALaS.
type Expression struct {
	Type     string       `json:"type"`
	Value    interface{}  `json:"value,omitempty"`
	Name     string       `json:"name,omitempty"`
	Op       string       `json:"op,omitempty"`
	Left     *Expression  `json:"left,omitempty"`
	Right    *Expression  `json:"right,omitempty"`
	Args     []Expression `json:"args,omitempty"`
	Elements []Expression `json:"elements,omitempty"` // For array literals
	Pairs    []MapPair    `json:"pairs,omitempty"`    // For map literals
	Index    *Expression  `json:"index,omitempty"`    // For indexing operations
	Object   *Expression  `json:"object,omitempty"`   // For field/index access
}

// MapPair represents a key-value pair in a map literal.
type MapPair struct {
	Key   Expression `json:"key"`
	Value Expression `json:"value"`
}

// Statement types.
const (
	StmtAssign = "assign"
	StmtIf     = "if"
	StmtWhile  = "while"
	StmtFor    = "for"
	StmtReturn = "return"
	StmtExpr   = "expr"
)

// Expression types.
const (
	ExprLiteral  = "literal"
	ExprVariable = "variable"
	ExprBinary   = "binary"
	ExprUnary    = "unary"
	ExprCall     = "call"
	ExprIndex    = "index"
	ExprField    = "field"
	ExprArrayLit = "array_literal"
	ExprMapLit   = "map_literal"
)

// Binary operators.
const (
	OpAdd = "+"
	OpSub = "-"
	OpMul = "*"
	OpDiv = "/"
	OpMod = "%"
	OpEq  = "=="
	OpNe  = "!="
	OpLt  = "<"
	OpLe  = "<="
	OpGt  = ">"
	OpGe  = ">="
	OpAnd = "&&"
	OpOr  = "||"
)

// Unary operators.
const (
	OpNot = "!"
	OpNeg = "-"
)

// Basic types.
const (
	TypeInt    = "int"
	TypeFloat  = "float"
	TypeString = "string"
	TypeBool   = "bool"
	TypeArray  = "array"
	TypeMap    = "map"
	TypeVoid   = "void"
)
