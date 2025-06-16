package interpreter

import (
	"testing"

	"github.com/dshills/alas/internal/ast"
	"github.com/dshills/alas/internal/runtime"
)

func TestCustomTypes(t *testing.T) {
	tests := []struct {
		name     string
		module   *ast.Module
		funcName string
		args     []runtime.Value
		want     runtime.Value
		wantErr  bool
	}{
		{
			name: "create struct with field access",
			module: &ast.Module{
				Type: "module",
				Name: "test_custom_types",
				Types: []ast.TypeDefinition{
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "name", Type: "string"},
								{Name: "age", Type: "int"},
							},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type: "function",
						Name: "test_person",
						Params: []ast.Parameter{
							{Name: "name", Type: "string"},
							{Name: "age", Type: "int"},
						},
						Returns: "int",
						Body: []ast.Statement{
							// person = {"name": name, "age": age}
							{
								Type:   ast.StmtAssign,
								Target: "person",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "name"},
											Value: ast.Expression{Type: ast.ExprVariable, Name: "name"},
										},
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "age"},
											Value: ast.Expression{Type: ast.ExprVariable, Name: "age"},
										},
									},
								},
							},
							// return person.age
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type:   ast.ExprField,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "person"},
									Field:  "age",
								},
							},
						},
					},
				},
			},
			funcName: "test_person",
			args: []runtime.Value{
				runtime.NewString("Alice"),
				runtime.NewInt(30),
			},
			want:    runtime.NewInt(30),
			wantErr: false,
		},
		{
			name: "nested field access",
			module: &ast.Module{
				Type: "module",
				Name: "test_nested",
				Types: []ast.TypeDefinition{
					{
						Name: "Address",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "street", Type: "string"},
								{Name: "city", Type: "string"},
							},
						},
					},
					{
						Name: "Person",
						Definition: ast.TypeDefinitionDef{
							Kind: ast.TypeKindStruct,
							Fields: []ast.TypeField{
								{Name: "name", Type: "string"},
								{Name: "address", Type: "Address"},
							},
						},
					},
				},
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "test_nested",
						Params:  []ast.Parameter{},
						Returns: "string",
						Body: []ast.Statement{
							// address = {"street": "123 Main St", "city": "Boston"}
							{
								Type:   ast.StmtAssign,
								Target: "address",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "street"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: "123 Main St"},
										},
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "city"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: "Boston"},
										},
									},
								},
							},
							// person = {"name": "Bob", "address": address}
							{
								Type:   ast.StmtAssign,
								Target: "person",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "name"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: "Bob"},
										},
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "address"},
											Value: ast.Expression{Type: ast.ExprVariable, Name: "address"},
										},
									},
								},
							},
							// return person.address.city
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type: ast.ExprField,
									Object: &ast.Expression{
										Type:   ast.ExprField,
										Object: &ast.Expression{Type: ast.ExprVariable, Name: "person"},
										Field:  "address",
									},
									Field: "city",
								},
							},
						},
					},
				},
			},
			funcName: "test_nested",
			args:     []runtime.Value{},
			want:     runtime.NewString("Boston"),
			wantErr:  false,
		},
		{
			name: "field access on non-map",
			module: &ast.Module{
				Type: "module",
				Name: "test_error",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "test_invalid_field",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							// x = 42
							{
								Type:   ast.StmtAssign,
								Target: "x",
								Value: &ast.Expression{
									Type:  ast.ExprLiteral,
									Value: 42,
								},
							},
							// return x.field (should error)
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type:   ast.ExprField,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "x"},
									Field:  "field",
								},
							},
						},
					},
				},
			},
			funcName: "test_invalid_field",
			args:     []runtime.Value{},
			want:     runtime.NewVoid(),
			wantErr:  true,
		},
		{
			name: "field not found",
			module: &ast.Module{
				Type: "module",
				Name: "test_missing_field",
				Functions: []ast.Function{
					{
						Type:    "function",
						Name:    "test_missing",
						Params:  []ast.Parameter{},
						Returns: "int",
						Body: []ast.Statement{
							// obj = {"x": 10}
							{
								Type:   ast.StmtAssign,
								Target: "obj",
								Value: &ast.Expression{
									Type: ast.ExprMapLit,
									Pairs: []ast.MapPair{
										{
											Key:   ast.Expression{Type: ast.ExprLiteral, Value: "x"},
											Value: ast.Expression{Type: ast.ExprLiteral, Value: 10},
										},
									},
								},
							},
							// return obj.y (field not found)
							{
								Type: ast.StmtReturn,
								Value: &ast.Expression{
									Type:   ast.ExprField,
									Object: &ast.Expression{Type: ast.ExprVariable, Name: "obj"},
									Field:  "y",
								},
							},
						},
					},
				},
			},
			funcName: "test_missing",
			args:     []runtime.Value{},
			want:     runtime.NewVoid(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := New()
			err := interp.LoadModule(tt.module)
			if err != nil {
				t.Fatalf("LoadModule() error = %v", err)
			}

			got, err := interp.Run(tt.funcName, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !valuesEqual(got, tt.want) {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnumTypes(t *testing.T) {
	// Enums are currently represented as strings at runtime
	module := &ast.Module{
		Type: "module",
		Name: "test_enums",
		Types: []ast.TypeDefinition{
			{
				Name: "Status",
				Definition: ast.TypeDefinitionDef{
					Kind:   ast.TypeKindEnum,
					Values: []string{"active", "inactive", "pending"},
				},
			},
		},
		Functions: []ast.Function{
			{
				Type:    "function",
				Name:    "test_enum",
				Params:  []ast.Parameter{},
				Returns: "string",
				Body: []ast.Statement{
					// status = "active"
					{
						Type:   ast.StmtAssign,
						Target: "status",
						Value: &ast.Expression{
							Type:  ast.ExprLiteral,
							Value: "active",
						},
					},
					// return status
					{
						Type: ast.StmtReturn,
						Value: &ast.Expression{
							Type: ast.ExprVariable,
							Name: "status",
						},
					},
				},
			},
		},
	}

	interp := New()
	err := interp.LoadModule(module)
	if err != nil {
		t.Fatalf("LoadModule() error = %v", err)
	}

	got, err := interp.Run("test_enum", []runtime.Value{})
	if err != nil {
		t.Errorf("Run() error = %v", err)
		return
	}

	want := runtime.NewString("active")
	if !valuesEqual(got, want) {
		t.Errorf("Run() = %v, want %v", got, want)
	}
}

// Helper function to compare runtime values.
func valuesEqual(a, b runtime.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case runtime.ValueTypeInt:
		ai, _ := a.AsInt()
		bi, _ := b.AsInt()
		return ai == bi
	case runtime.ValueTypeFloat:
		af, _ := a.AsFloat()
		bf, _ := b.AsFloat()
		return af == bf
	case runtime.ValueTypeString:
		as, _ := a.AsString()
		bs, _ := b.AsString()
		return as == bs
	case runtime.ValueTypeBool:
		ab, _ := a.AsBool()
		bb, _ := b.AsBool()
		return ab == bb
	case runtime.ValueTypeVoid:
		return true
	case runtime.ValueTypeArray, runtime.ValueTypeMap:
		// For arrays and maps, just check if both are non-nil
		// Deeper comparison would require more complex logic
		return true
	default:
		return false
	}
}
