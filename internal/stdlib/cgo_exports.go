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

// setCValueFields sets the fields of a C.CValue based on a Go runtime.Value
func setCValueFields(cval *C.CValue, val runtime.Value) {
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
	case runtime.ValueTypeArray:
		// TODO: Handle arrays
		cval._type = CValueTypeVoid
	case runtime.ValueTypeMap:
		// TODO: Handle maps
		cval._type = CValueTypeVoid
	default:
		cval._type = CValueTypeVoid
	}
}

// convertGoValueToC converts a Go runtime.Value to a C Value
// The caller is responsible for freeing any allocated memory
func convertGoValueToC(val runtime.Value) C.CValue {
	var cval C.CValue
	setCValueFields(&cval, val)
	return cval
}

// convertGoValueToCPtr converts a Go runtime.Value to a pointer to C Value
// This allocates memory that the caller must free
func convertGoValueToCPtr(val runtime.Value) *C.CValue {
	cval := (*C.CValue)(C.calloc(1, C.size_t(unsafe.Sizeof(C.CValue{}))))
	setCValueFields(cval, val)
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
func alas_builtin_math_sqrt(val *C.CValue) *C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}

	// Get the registry and call the function
	registry := NewRegistry()
	result, err := registry.Call("math.sqrt", args)
	if err != nil {
		// Return error value
		return convertGoValueToCPtr(runtime.NewFloat(0))
	}

	return convertGoValueToCPtr(result)
}

//export alas_builtin_math_abs
func alas_builtin_math_abs(val *C.CValue) *C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}

	registry := NewRegistry()
	result, err := registry.Call("math.abs", args)
	if err != nil {
		return convertGoValueToCPtr(runtime.NewFloat(0))
	}

	return convertGoValueToCPtr(result)
}

//export alas_builtin_collections_length
func alas_builtin_collections_length(val *C.CValue) *C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}

	registry := NewRegistry()
	result, err := registry.Call("collections.length", args)
	if err != nil {
		return convertGoValueToCPtr(runtime.NewInt(0))
	}

	return convertGoValueToCPtr(result)
}

//export alas_builtin_string_toUpper
func alas_builtin_string_toUpper(val *C.CValue) *C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}

	registry := NewRegistry()
	result, err := registry.Call("string.toUpper", args)
	if err != nil {
		return convertGoValueToCPtr(runtime.NewString(""))
	}

	return convertGoValueToCPtr(result)
}

//export alas_builtin_type_typeOf
func alas_builtin_type_typeOf(val *C.CValue) *C.CValue {
	goVal := convertCValueToGo(val)
	args := []runtime.Value{goVal}

	registry := NewRegistry()
	result, err := registry.Call("type.typeOf", args)
	if err != nil {
		return convertGoValueToCPtr(runtime.NewString("unknown"))
	}

	return convertGoValueToCPtr(result)
}

// FreeCString frees a C string allocated by Go
//
//export alas_free_cstring
func alas_free_cstring(str *C.char) {
	C.free(unsafe.Pointer(str))
}

// FreeCValue frees a CValue allocated by Go
//
//export alas_free_cvalue
func alas_free_cvalue(val *C.CValue) {
	if val != nil {
		// Free any string data if it exists
		if val._type == CValueTypeString && val.string_val != nil {
			C.free(unsafe.Pointer(val.string_val))
		}
		// Free the CValue struct itself
		C.free(unsafe.Pointer(val))
	}
}

// Additional exports would be added for all other builtin functions...
// This is a starting point demonstrating the pattern
