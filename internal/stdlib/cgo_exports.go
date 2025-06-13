package stdlib

// #include <stdint.h>
// #include <stdlib.h>
// #include <string.h>
//
// // C representation of ALaS Value type (using struct instead of union for simplicity)
// typedef struct {
//     int32_t type;  // ValueType enum
//     int64_t int_val;
//     double float_val;
//     char* string_val;
//     void* array_val;
//     void* map_val;
// } CValue;
//
// // Helper to create C string from Go string
// static char* go_string_to_c(const char* s, size_t len) {
//     char* c_str = (char*)malloc(len + 1);
//     if (c_str) {
//         memcpy(c_str, s, len);
//         c_str[len] = '\0';
//     }
//     return c_str;
// }
import "C"
import (
	"unsafe"

	"github.com/dshills/alas/internal/runtime"
)

// ValueType constants matching runtime.ValueType
const (
	CValueTypeInt    = 0
	CValueTypeFloat  = 1
	CValueTypeString = 2
	CValueTypeBool   = 3
	CValueTypeArray  = 4
	CValueTypeMap    = 5
	CValueTypeVoid   = 6
)

// convertCValueToGo converts a C Value to a Go runtime.Value
func convertCValueToGo(cval *C.CValue) runtime.Value {
	switch cval._type {
	case CValueTypeInt:
		return runtime.NewInt(int64(cval.int_val))
	case CValueTypeFloat:
		return runtime.NewFloat(float64(cval.float_val))
	case CValueTypeString:
		str := C.GoString(cval.string_val)
		return runtime.NewString(str)
	case CValueTypeBool:
		return runtime.NewBool(cval.int_val != 0)
	case CValueTypeVoid:
		return runtime.NewVoid()
	// TODO: Handle arrays and maps
	default:
		return runtime.NewVoid()
	}
}

// convertGoValueToC converts a Go runtime.Value to a C Value
// The caller is responsible for freeing any allocated memory
func convertGoValueToC(val runtime.Value) C.CValue {
	var cval C.CValue
	
	switch val.Type {
	case runtime.ValueTypeInt:
		cval._type = CValueTypeInt
		i, _ := val.AsInt()
		cval.int_val = C.int64_t(i)
	case runtime.ValueTypeFloat:
		cval._type = CValueTypeFloat
		f, _ := val.AsFloat()
		cval.float_val = C.double(f)
	case runtime.ValueTypeString:
		cval._type = CValueTypeString
		s, _ := val.AsString()
		cval.string_val = C.CString(s)
	case runtime.ValueTypeBool:
		cval._type = CValueTypeBool
		b, _ := val.AsBool()
		if b {
			cval.int_val = 1
		} else {
			cval.int_val = 0
		}
	case runtime.ValueTypeVoid:
		cval._type = CValueTypeVoid
	// TODO: Handle arrays and maps
	default:
		cval._type = CValueTypeVoid
	}
	
	return cval
}

//export alas_builtin_io_print
func alas_builtin_io_print(val *C.CValue) {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}
	
	// Get the registry and call the function
	registry := NewRegistry()
	registry.Call("io.print", args)
}

//export alas_builtin_math_sqrt
func alas_builtin_math_sqrt(val *C.CValue) C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}
	
	// Get the registry and call the function
	registry := NewRegistry()
	result, err := registry.Call("math.sqrt", args)
	if err != nil {
		// Return NaN or error value
		return convertGoValueToC(runtime.NewFloat(0))
	}
	
	return convertGoValueToC(result)
}

//export alas_builtin_math_abs
func alas_builtin_math_abs(val *C.CValue) C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}
	
	registry := NewRegistry()
	result, err := registry.Call("math.abs", args)
	if err != nil {
		return convertGoValueToC(runtime.NewFloat(0))
	}
	
	return convertGoValueToC(result)
}

//export alas_builtin_collections_length
func alas_builtin_collections_length(val *C.CValue) C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}
	
	registry := NewRegistry()
	result, err := registry.Call("collections.length", args)
	if err != nil {
		return convertGoValueToC(runtime.NewInt(0))
	}
	
	return convertGoValueToC(result)
}

//export alas_builtin_string_toUpper
func alas_builtin_string_toUpper(val *C.CValue) C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}
	
	registry := NewRegistry()
	result, err := registry.Call("string.toUpper", args)
	if err != nil {
		return convertGoValueToC(runtime.NewString(""))
	}
	
	return convertGoValueToC(result)
}

//export alas_builtin_type_typeOf
func alas_builtin_type_typeOf(val *C.CValue) C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}
	
	registry := NewRegistry()
	result, err := registry.Call("type.typeOf", args)
	if err != nil {
		return convertGoValueToC(runtime.NewString("unknown"))
	}
	
	return convertGoValueToC(result)
}

// FreeCString frees a C string allocated by Go
//export alas_free_cstring
func alas_free_cstring(str *C.char) {
	C.free(unsafe.Pointer(str))
}

// Additional exports would be added for all other builtin functions...
// This is a starting point demonstrating the pattern