package stdlib

import (
	"fmt"

	"github.com/dshills/alas/internal/runtime"
)

// BuiltinFunction represents a native function that can be called from ALaS.
type BuiltinFunction func(args []runtime.Value) (runtime.Value, error)

// Registry manages all built-in standard library functions.
type Registry struct {
	functions map[string]BuiltinFunction
}

// NewRegistry creates a new standard library function registry.
func NewRegistry() *Registry {
	r := &Registry{
		functions: make(map[string]BuiltinFunction),
	}

	// Register all standard library modules
	r.registerIOFunctions()
	r.registerMathFunctions()
	r.registerCollectionsFunctions()
	r.registerStringFunctions()
	r.registerTypeFunctions()
	r.registerResultFunctions()

	return r
}

// Register registers a builtin function.
func (r *Registry) Register(name string, fn BuiltinFunction) {
	r.functions[name] = fn
}

// Call calls a builtin function by name.
func (r *Registry) Call(name string, args []runtime.Value) (runtime.Value, error) {
	fn, exists := r.functions[name]
	if !exists {
		return runtime.NewVoid(), fmt.Errorf("builtin function not found: %s", name)
	}

	return fn(args)
}

// HasFunction checks if a builtin function exists.
func (r *Registry) HasFunction(name string) bool {
	_, exists := r.functions[name]
	return exists
}

// ListFunctions returns all registered function names.
func (r *Registry) ListFunctions() []string {
	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	return names
}
