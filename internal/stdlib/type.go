package stdlib

import (
	"fmt"
	"strconv"

	"github.com/dshills/alas/internal/runtime"
)

// registerTypeFunctions registers all std.type builtin functions.
func (r *Registry) registerTypeFunctions() {
	r.Register("type.typeOf", typeTypeOf)
	r.Register("type.toString", typeToString)
	r.Register("type.parseInt", typeParseInt)
	r.Register("type.parseFloat", typeParseFloat)
	r.Register("type.isInt", typeIsInt)
	r.Register("type.isFloat", typeIsFloat)
	r.Register("type.isString", typeIsString)
	r.Register("type.isBool", typeIsBool)
	r.Register("type.isArray", typeIsArray)
	r.Register("type.isMap", typeIsMap)
}

// typeTypeOf implements type.typeOf builtin function.
func typeTypeOf(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.typeOf expects 1 argument, got %d", len(args))
	}

	val := args[0]
	switch val.Type {
	case runtime.ValueTypeInt:
		return runtime.NewString("int"), nil
	case runtime.ValueTypeFloat:
		return runtime.NewString("float"), nil
	case runtime.ValueTypeString:
		return runtime.NewString("string"), nil
	case runtime.ValueTypeBool:
		return runtime.NewString("bool"), nil
	case runtime.ValueTypeArray:
		return runtime.NewString("array"), nil
	case runtime.ValueTypeMap:
		return runtime.NewString("map"), nil
	case runtime.ValueTypeVoid:
		return runtime.NewString("void"), nil
	default:
		return runtime.NewString("unknown"), nil
	}
}

// typeToString implements type.toString builtin function.
func typeToString(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.toString expects 1 argument, got %d", len(args))
	}

	val := args[0]
	switch val.Type {
	case runtime.ValueTypeInt:
		intVal, _ := val.AsInt()
		return runtime.NewString(strconv.FormatInt(intVal, 10)), nil
	case runtime.ValueTypeFloat:
		floatVal, _ := val.AsFloat()
		return runtime.NewString(strconv.FormatFloat(floatVal, 'f', -1, 64)), nil
	case runtime.ValueTypeString:
		str, _ := val.AsString()
		return runtime.NewString(str), nil
	case runtime.ValueTypeBool:
		boolVal, _ := val.AsBool()
		return runtime.NewString(strconv.FormatBool(boolVal)), nil
	case runtime.ValueTypeArray:
		return runtime.NewString("[Array]"), nil
	case runtime.ValueTypeMap:
		return runtime.NewString("{Map}"), nil
	case runtime.ValueTypeVoid:
		return runtime.NewString("void"), nil
	default:
		return runtime.NewString("unknown"), nil
	}
}

// typeParseInt implements type.parseInt builtin function.
func typeParseInt(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.parseInt expects 1 argument, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("type.parseInt: %v", err)
	}

	intVal, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("type.parseInt: cannot parse '%s' as integer", str)
	}

	return runtime.NewInt(intVal), nil
}

// typeParseFloat implements type.parseFloat builtin function.
func typeParseFloat(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.parseFloat expects 1 argument, got %d", len(args))
	}

	str, err := args[0].AsString()
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("type.parseFloat: %v", err)
	}

	floatVal, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return runtime.NewVoid(), fmt.Errorf("type.parseFloat: cannot parse '%s' as float", str)
	}

	return runtime.NewFloat(floatVal), nil
}

// typeIsInt implements type.isInt builtin function.
func typeIsInt(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.isInt expects 1 argument, got %d", len(args))
	}

	return runtime.NewBool(args[0].Type == runtime.ValueTypeInt), nil
}

// typeIsFloat implements type.isFloat builtin function.
func typeIsFloat(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.isFloat expects 1 argument, got %d", len(args))
	}

	return runtime.NewBool(args[0].Type == runtime.ValueTypeFloat), nil
}

// typeIsString implements type.isString builtin function.
func typeIsString(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.isString expects 1 argument, got %d", len(args))
	}

	return runtime.NewBool(args[0].Type == runtime.ValueTypeString), nil
}

// typeIsBool implements type.isBool builtin function.
func typeIsBool(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.isBool expects 1 argument, got %d", len(args))
	}

	return runtime.NewBool(args[0].Type == runtime.ValueTypeBool), nil
}

// typeIsArray implements type.isArray builtin function.
func typeIsArray(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.isArray expects 1 argument, got %d", len(args))
	}

	return runtime.NewBool(args[0].Type == runtime.ValueTypeArray), nil
}

// typeIsMap implements type.isMap builtin function.
func typeIsMap(args []runtime.Value) (runtime.Value, error) {
	if len(args) != 1 {
		return runtime.NewVoid(), fmt.Errorf("type.isMap expects 1 argument, got %d", len(args))
	}

	return runtime.NewBool(args[0].Type == runtime.ValueTypeMap), nil
}
